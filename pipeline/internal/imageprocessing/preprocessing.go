package imageprocessing

import (
	"errors"
	"gocv.io/x/gocv"
	"log"
)

type ProcessedImage struct {
	ImageGrey gocv.Mat
	Features  gocv.Mat
}

const (
	MaxCorners   = 100
	QualityLevel = 0.03
	MinDistance  = 7
)

func ProcessImage(img gocv.Mat) (ProcessedImage, error) {
	greyCVImage := gocv.NewMat()
	gocv.CvtColor(img, &greyCVImage, gocv.ColorBGRToGray)
	if greyCVImage.Empty() {
		log.Printf("Error processing image to grey")
		return ProcessedImage{}, errors.New("Error processing image to grey")
	}
	goodFeatures := gocv.NewMat()
	gocv.GoodFeaturesToTrack(greyCVImage, &goodFeatures, MaxCorners, QualityLevel, MinDistance)
	processedImage := ProcessedImage{
		ImageGrey: greyCVImage,
		Features:  goodFeatures,
	}
	img.Close()
	return processedImage, nil
}

func (p ProcessedImage) Close() {
	err := p.Features.Close()
	if err != nil {
		log.Printf("Error closing features")
	}
	err = p.ImageGrey.Close()
	if err != nil {
		log.Printf("Error closing grey image")
	}
}
