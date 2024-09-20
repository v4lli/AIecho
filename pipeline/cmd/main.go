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
		if len(args) < 1 || args[0] == "" {
			log.Fatalf("Dev mode requires a directory for images")
		}
		imageDir = args[0]
		log.Printf("Dev mode enabled using directory %s", imageDir)
	} else {
		if uuid == "" {
			log.Fatalf("Empty UUID is not valid")
		}
		log.Printf("Production mode enabled")
	}

	if debugMode {
		log.Printf("Debug mode enabled")
	} else {
		log.Printf("Debug mode disabled")
	}

	if fastMode {
		maxTokens = 50
		log.Printf("Fast mode enabled")
	} else {
		maxTokens = 60
		log.Printf("Fast mode disabled")
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

	log.Print("Setup done, starting go routines")
	go imageRetrievalRoutine(imageRetrievalFunc, imageChannel1, imageChannel2, i2tImageRetChannel)
	go imagePreProcessingRoutine(imageChannel1, imageProcessingChannel)
	go imageSimilarityRoutine(imageProcessingChannel, similarityChannel)
	go movementDetectionRoutine(similarityChannel, movementChannel)
	go i2tRoutine(imageChannel2, i2tImageRetChannel, i2tBufferChannel, i2tLLMChannel, cfg, maxTokens)
	go llmRoutine(i2tLLMChannel, i2tBufferChannel, movementChannel, cfg)

	log.Printf("Started all routines")
	i2tImageRetChannel <- true

	select {}
}

func imageRetrievalRoutine(
	imageRetrievalFunc func() (gocv.Mat, error), imageChannel1, imageChannel2 chan gocv.Mat,
	i2tImageRetChannel chan bool,
) {
	log.Println("Starting image retrieval routine")
	lastImageRetrieval := time.Now()
	for range i2tImageRetChannel {
		if time.Since(lastImageRetrieval) < time.Second {
			time.Sleep(time.Second - time.Since(lastImageRetrieval))
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
}

func imagePreProcessingRoutine(imgChannel chan gocv.Mat, imageProcessingChannel chan imageprocessing.ProcessedImage) {
	log.Println("Starting image processing routine")
	for img := range imgChannel {
		processedImage, err := imageprocessing.ProcessImage(img)
		if err != nil {
			log.Fatalf("Ending execution due to image processing error: %v", err)
		} else {
			imageProcessingChannel <- processedImage
		}
	}
}

func imageSimilarityRoutine(
	imageProcessingChannel chan imageprocessing.ProcessedImage, similarityChannel chan []float64,
) {
	log.Println("Starting similarity processing routine")
	var imageQueue []imageprocessing.ProcessedImage
	for processedImage := range imageProcessingChannel {
		if len(imageQueue) >= imageQueueLength {
			imageQueue[0].Close()
			imageQueue = imageQueue[1:]
		}
		scores := imageprocessing.SimilarityProcessing(processedImage, imageQueue)
		imageQueue = append(imageQueue, processedImage)
		similarityChannel <- scores
	}
}

func movementDetectionRoutine(similarityChannel chan []float64, movementChannel chan bool) {
	log.Println("Starting movement detection routine")
	for scores := range similarityChannel {
		movement := imageprocessing.MovementDetection(scores)
		movementChannel <- movement
	}
}

func i2tRoutine(
	imageChannel chan gocv.Mat, i2tImageRetChannel chan bool, i2tBufferChannel chan string, i2tLLMChannel chan bool,
	cfg *config.Config, maxTokens int,
) {
	log.Printf("Starting I2T routine")
	for img := range imageChannel {
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
		if !fastMode {
			if i2tDescription.ResType != "error" {
				i2tBufferChannel <- i2tDescription.Content
			}
			if len(i2tBufferChannel) == 3 {
				i2tLLMChannel <- true
			}
		}
		i2tImageRetChannel <- true
	}
}

func llmRoutine(i2tLLMChannel chan bool, i2tBufferChannel chan string, movementChannel chan bool, cfg *config.Config) {
	log.Printf("Starting LLM routine")
	for range i2tLLMChannel {
		var i2tResponses []string
		movement := false
		for i := 0; i < 3; i++ {
			i2tResponses = append(i2tResponses, <-i2tBufferChannel)
			movement = movement || <-movementChannel
		}
		llmDescription := http.RunLLM(cfg, i2tResponses, movement)
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
}
