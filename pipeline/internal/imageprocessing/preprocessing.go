package imageprocessing

import (
	"gocv.io/x/gocv"
	"image"
	"log"
)

type ProcessedImage struct {
	ImageGrey *gocv.Mat
	Features  *gocv.Mat
}

const (
	MaxCorners   = 100
	QualityLevel = 0.03
	MinDistance  = 7
)

func ProcessImage(img *image.Image) (*ProcessedImage, error) {
	cvImage, err := gocv.ImageToMatRGB(*img)
	greyCVImage := gocv.NewMat()
	gocv.CvtColor(cvImage, &greyCVImage, gocv.ColorBGRToGray)
	if err != nil {
		log.Printf("Error processing image to MatRGB %v", err)
		return nil, err
	}
	goodFeatures := gocv.NewMat()
	gocv.GoodFeaturesToTrack(cvImage, &goodFeatures, MaxCorners, QualityLevel, MinDistance)
	processedImage := ProcessedImage{
		ImageGrey: &greyCVImage,
		Features:  &goodFeatures,
	}
	return &processedImage, nil
}
