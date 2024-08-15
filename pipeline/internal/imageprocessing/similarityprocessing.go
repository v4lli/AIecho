package imageprocessing

import (
	"gocv.io/x/gocv"
	"log"
)

func SimilarityProcessing(
	latestImage *ProcessedImage, queue []ProcessedImage,
) []float64 {

	latestGray := latestImage.ImageGrey
	var result []float64
	for _, img := range queue {
		prevGray := img.ImageGrey
		prevFeatures := img.Features
		status := gocv.NewMat()
		nextPts := gocv.NewMat()
		err := gocv.NewMat()
		gocv.CalcOpticalFlowPyrLK(*prevGray, *latestGray, *prevFeatures, nextPts, &status, &err)
		if gocv.CountNonZero(err) > 0 {
			log.Printf("Error calculating optical flow pyrLK: %v", err)
		}
		score := float64(gocv.CountNonZero(status))
		result = append(result, score)
	}
	return result
}

func MovementDetection(scores []float64) bool {
	movementScore := float64(0)
	for index, value := range scores {
		movementScore += value/float64(index) + 1
	}
	movementScore /= float64(len(scores))
	if movementScore < 0.6 {
		return true
	} else {
		return false
	}
}
