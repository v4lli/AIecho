package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/v4lli/AIecho/pipeline/internal/config"
	"github.com/v4lli/AIecho/pipeline/internal/http"
	"github.com/v4lli/AIecho/pipeline/internal/imageprocessing"
	"github.com/v4lli/AIecho/pipeline/internal/logging"
	"github.com/v4lli/AIecho/pipeline/internal/pipeline"
	"gocv.io/x/gocv"
	"log"
	"time"
)

const imageQueueLength = 3

var (
	uuid      string
	devMode   bool
	debugMode bool
	imageDir  string
	fastMode  bool
	maxTokens int
)

func main() {
	flag.StringVar(&uuid, "uuid", "", "UUID for upstream images")
	flag.BoolVar(&devMode, "dev", false, "Enable dev mode, requires image directory")
	flag.BoolVar(&debugMode, "debug", false, "Enable debug mode, sends errors to client")
	flag.BoolVar(&fastMode, "fast", false, "Enable fast mode")
	flag.Parse()
	args := flag.Args()
	logging.SetupLogging()
	defer logging.CloseLogging()
	log.Printf("Pipeline started")
	if devMode {
		if len(args) < 1 {
			log.Fatalf("Dev mode requires a directory for images")
		}
		imageDir = args[0]
		if imageDir == "" {
			log.Fatalf("Empty image directory is not valid")
		}
		log.Printf("Dev mode enabled using directory %s", imageDir)
	} else {
		if uuid == "" {
			log.Fatalf("Empty UUID is not valid")
		}
		log.Printf("Production mode enabled")
	}
	if fastMode {
		maxTokens = 50
		log.Printf("Fast mode enabled")
	} else {
		maxTokens = 60
		log.Printf("Fast mode disabled")
	}

	var imageRetrievalFunc func() (gocv.Mat, error)
	if devMode {
		imageRetrievalFunc = func() (gocv.Mat, error) {
			return pipeline.RetrieveDevelopmentImages(imageDir)
		}
	} else {
		imageRetrievalFunc = func() (gocv.Mat, error) {
			return pipeline.RetrieveImage(fmt.Sprintf("http://whipcapture:9091/internal/frame/%s/0", uuid))
		}
	}

	cfg := config.LoadConfig("pipeline.env")
	imageChannel1 := make(chan gocv.Mat)
	imageChannel2 := make(chan gocv.Mat)
	imageProcessingChannel := make(chan imageprocessing.ProcessedImage)
	similarityChannel := make(chan []float64)
	movementChannel := make(chan bool)
	i2tImageRetChannel := make(chan bool)
	i2tBufferChannel := make(chan string, 3)
	i2tLLMChannel := make(chan bool)

	lastImageRetrieval := time.Now()
	var imageQueue []imageprocessing.ProcessedImage
	log.Print("Setup done, starting go routines")

	go func() {
		log.Printf("Starting image retrieval routine")
		for range i2tImageRetChannel {
			timePassed := time.Since(lastImageRetrieval)
			if timePassed < time.Second {
				time.Sleep(time.Second - timePassed)
			}
			img, err := imageRetrievalFunc()
			if err != nil {
				log.Fatalf("Ending execution due to image retrieval error: %v", err)
			} else {
				if !fastMode {
					imageChannel1 <- img.Clone()
				}
				imageChannel2 <- img
				lastImageRetrieval = time.Now()
			}
		}
	}()

	go func() {
		log.Printf("Starting image processing routine")
		for img := range imageChannel1 {
			processedImage, err := imageprocessing.ProcessImage(img)
			if err != nil {
				log.Fatalf("Ending execution due to image processing error: %v", err)
			} else {
				imageProcessingChannel <- processedImage
			}
			if err != nil {
				log.Printf("Error closing image: %v", err)
			}
		}
	}()

	go func() {
		log.Printf("Starting similarity processing routine")
		for processedImage := range imageProcessingChannel {
			if len(imageQueue) >= imageQueueLength {
				imageQueue[0].Close()
				imageQueue = imageQueue[1:]
			}
			scores := imageprocessing.SimilarityProcessing(processedImage, imageQueue)
			imageQueue = append(imageQueue, processedImage)
			similarityChannel <- scores
		}
	}()

	go func() {
		log.Printf("Starting movement detection routine")
		for scores := range similarityChannel {
			movement := imageprocessing.MovementDetection(scores)
			movementChannel <- movement
		}
	}()

	go func() {
		log.Printf("Starting I2T routine")
		for img := range imageChannel2 {
			i2tDescription := http.RunImage2Text(cfg, img, maxTokens)
			resultJSON, err := json.Marshal(i2tDescription)
			if err != nil {
				log.Printf("Error marshalling i2t response: %v", err)
			}
			if i2tDescription.ResType == "error" {
				if debugMode {
					fmt.Println(string(resultJSON))
				}
			} else {
				fmt.Println(string(resultJSON))
			}
			i2tImageRetChannel <- true
			if !fastMode {
				i2tBufferChannel <- i2tDescription.Content
				if len(i2tBufferChannel) == 3 {
					i2tLLMChannel <- true
				}
			}
		}
	}()

	go func() {
		log.Printf("Starting LLM routine")
		for range i2tLLMChannel {
			var i2tResponses []string
			for i := 0; i < 3; i++ {
				i2tResponses = append(i2tResponses, <-i2tBufferChannel)
			}
			llmDescription := http.RunLLM(cfg, i2tResponses, true)
			resultJSON, err := json.Marshal(llmDescription)
			if err != nil {
				log.Printf("Error marshalling LLM response: %v", err)
			}

			if llmDescription.ResType == "error" {
				if debugMode {
					fmt.Println(string(resultJSON))
				}
			} else {
				fmt.Println(string(resultJSON))
			}
		}
	}()

	log.Printf("Started all routines")
	i2tImageRetChannel <- true

	select {}
}
