package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/v4lli/AIecho/pipeline/internal/config"
	"github.com/v4lli/AIecho/pipeline/internal/http"
	"github.com/v4lli/AIecho/pipeline/internal/imageprocessing"
	"github.com/v4lli/AIecho/pipeline/internal/logging"
	"github.com/v4lli/AIecho/pipeline/internal/pipeline"
	"image"
	"log"
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
	resType string `json:"type"`
	content string `json:"content"`
	urgent  bool   `json:"urgent"`
}

func main() {
	flag.StringVar(&uuid, "uuid", "", "UUID for upstream images")
	flag.BoolVar(&devMode, "dev", false, "Enable dev mode")
	flag.BoolVar(&fastMode, "fast", false, "Enable fast mode")
	flag.Parse()
	args := flag.Args()
	logging.SetupLogging()
	log.Printf("Pipeline started")
	if devMode {
		if len(args) < 1 {
			log.Fatalf("Dev mode requires a directory for images")
		}
		imageDir := args[0]
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

	var imageRetrievalFunc func() (*image.Image, error)
	if devMode {
		imageRetrievalFunc = func() (*image.Image, error) {
			return pipeline.RetrieveImage(fmt.Sprintf("http://whipcapture:9091/internal/frame/%s/0", uuid))
		}
	} else {
		imageRetrievalFunc = func() (*image.Image, error) {
			return pipeline.RetrieveDevelopmentImages(imageDir)
		}
	}

	cfg := config.LoadConfig("pipeline.env")
	imageChannel1 := make(chan *image.Image)
	imageChannel2 := make(chan *image.Image)
	imageProcessingChannel := make(chan *imageprocessing.ProcessedImage, 1)
	similarityChannel := make(chan []float64, 1)
	movementChannel := make(chan bool, 1)
	i2tImageRetChannel := make(chan bool, 1)
	i2tBufferChannel := make(chan string, 3)
	i2tLLMChannel := make(chan bool, 1)

	var lastImageRetrieval time.Time
	var imageQueue []imageprocessing.ProcessedImage

	go func() {
		for _ = range i2tImageRetChannel {
			timePassed := time.Since(lastImageRetrieval)
			if timePassed < time.Second {
				time.Sleep(time.Second - timePassed)
			}
			img, err := imageRetrievalFunc()
			if err != nil {
				log.Fatalf("Ending execution due to image retrieval error: %v", err)
			} else {
				if !fastMode {
					imageChannel1 <- img
				}
				imageChannel2 <- img
			}
		}
	}()

	go func() {
		for img := range imageChannel1 {
			processedImage, err := imageprocessing.ProcessImage(img)
			if err != nil {
				log.Fatalf("Ending execution due to image processing error: %v", err)
			} else {
				imageProcessingChannel <- processedImage
			}
		}
	}()

	go func() {
		for processedImage := range imageProcessingChannel {
			if len(imageQueue) > imageQueueLength {
				imageQueue = imageQueue[1:]
			}
			scores := imageprocessing.SimilarityProcessing(processedImage, imageQueue)
			imageQueue = append(imageQueue, *processedImage)
			similarityChannel <- scores
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
			i2tDescription := http.RunImage2Text(cfg, img, maxTokens)
			if i2tDescription == "" {
				log.Printf("Got empty I2T response")
			}
			result := Result{
				resType: "tl",
				content: i2tDescription,
				urgent:  false,
			}
			fmt.Println(json.Marshal(result))
			i2tImageRetChannel <- true
			if !fastMode {
				i2tBufferChannel <- i2tDescription
				if len(i2tBufferChannel) == 3 {
					i2tLLMChannel <- true
				}
			}
		}
	}()

	go func() {
		for _ = range i2tLLMChannel {
			llmDescription := http.RunLLM(cfg, i2tBufferChannel, <-movementChannel)
			if llmDescription == "" {
				log.Printf("Got empty LLM response")
			}
			result := Result{
				resType: "desc",
				content: llmDescription,
				urgent:  false,
			}
			fmt.Println(json.Marshal(result))
		}
	}()

}
