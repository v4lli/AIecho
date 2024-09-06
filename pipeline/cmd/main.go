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
	"sync"
	"time"
)

const imageQueueLength = 3

var (
	uuid      string
	devMode   bool
	imageDir  string
	fastMode  bool
	maxTokens int
)

type Result struct {
	ResType string `json:"type"`
	Content string `json:"content"`
	Urgent  bool   `json:"urgent"`
}

func main() {
	flag.StringVar(&uuid, "uuid", "", "UUID for upstream images")
	flag.BoolVar(&devMode, "dev", false, "Enable dev mode")
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
		log.Printf("Fast mode enabled")
	} else {
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
	imageProcessingChannel := make(chan *imageprocessing.ProcessedImage, 1)
	similarityChannel := make(chan []float64, 1)
	movementChannel := make(chan bool, 1)
	i2tImageRetChannel := make(chan bool, 1)
	i2tBufferChannel := make(chan string, 3)
	i2tLLMChannel := make(chan bool, 1)

	lastImageRetrieval := time.Now()
	var imageQueue []imageprocessing.ProcessedImage
	var mutex sync.Mutex
	log.Print("Setup done, starting go routines")
	go func() {
		for _ = range i2tImageRetChannel {
			//fmt.Printf("GOCV mat count at the beginning of imageRet: %d\n", gocv.MatProfile.Count())
			timePassed := time.Since(lastImageRetrieval)
			if timePassed < time.Second {
				time.Sleep(time.Second - timePassed)
			}
			img, err := imageRetrievalFunc()
			if err != nil {
				log.Fatalf("Ending execution due to image retrieval error: %v", err)
			} else {
				if !fastMode {
					//				imageChannel1 <- img
				}
				imageChannel2 <- img
				lastImageRetrieval = time.Now()
			}

			//fmt.Printf("GOCV mat count at the end of imageRet: %d\n", gocv.MatProfile.Count())
		}
	}()

	go func() {
		for img := range imageChannel1 {
			//fmt.Printf("GOCV mat count at the beginning of imageProc: %d\n", gocv.MatProfile.Count())
			processedImage, err := imageprocessing.ProcessImage(img)
			if err != nil {
				log.Fatalf("Ending execution due to image processing error: %v", err)
			} else {
				imageProcessingChannel <- processedImage
			}
			if err != nil {
				log.Printf("Error closing image: %v", err)
			}
			//fmt.Printf("GOCV mat count at the end of imageProc: %d\n", gocv.MatProfile.Count())
		}
	}()

	go func() {
		for processedImage := range imageProcessingChannel {
			mutex.Lock()
			//fmt.Printf("GOCV mat count at beginning of simProc: %d\n", gocv.MatProfile.Count())
			if len(imageQueue) > imageQueueLength {
				imageQueue[0].Close()
				imageQueue = imageQueue[1:]
			}
			scores := imageprocessing.SimilarityProcessing(processedImage, imageQueue)
			imageQueue = append(imageQueue, *processedImage)
			mutex.Unlock()
			similarityChannel <- scores
			//fmt.Printf("GOCV mat count at end of simProc: %d\n", gocv.MatProfile.Count())
		}
	}()

	go func() {
		for scores := range similarityChannel {
			movement := imageprocessing.MovementDetection(scores)
			movementChannel <- movement
		}
	}()

	go func() {
		for img := range imageChannel2 {
			//fmt.Printf("GOCV mat count at beginning of i2t: %d\n", gocv.MatProfile.Count())
			i2tDescription := http.RunImage2Text(cfg, img, maxTokens)
			if i2tDescription == "" {
				log.Printf("Got empty I2T response")
			}
			result := Result{
				ResType: "tl",
				Content: i2tDescription,
				Urgent:  false,
			}
			resultJSON, err := json.Marshal(result)
			if err != nil {
				log.Printf("Error marshalling i2t response: %v", err)
			}
			fmt.Println(string(resultJSON))
			i2tImageRetChannel <- true
			if !fastMode {
				i2tBufferChannel <- i2tDescription
				if len(i2tBufferChannel) == 3 {
					i2tLLMChannel <- true
				}
			}
			//fmt.Printf("GOCV mat count at end of i2t: %d\n", gocv.MatProfile.Count())
		}
	}()

	go func() {
		for _ = range i2tLLMChannel {
			//fmt.Printf("GOCV mat count at beginning of llm :%d\n", gocv.MatProfile.Count())
			//llmDescription := http.RunLLM(cfg, i2tBufferChannel, <-movementChannel)
			var i2tResponses []string
			for i := 0; i < 3; i++ {
				i2tResponses = append(i2tResponses, <-i2tBufferChannel)
			}
			llmDescription := http.RunLLM(cfg, i2tResponses, true)
			if llmDescription == "" {
				log.Printf("Got empty LLM response")
			}
			result := Result{
				ResType: "desc",
				Content: llmDescription,
				Urgent:  false,
			}
			resultJSON, err := json.Marshal(result)
			if err != nil {
				log.Printf("Error marshalling LLM response: %v", err)
			}
			fmt.Println(string(resultJSON))
		}
		//fmt.Printf("GOCV mat count at end of llm :%d\n", gocv.MatProfile.Count())
	}()
	log.Printf("Started all routines")
	i2tImageRetChannel <- true

	select {}
}
