package pipeline

import (
	"errors"
	"gocv.io/x/gocv"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func RetrieveImage(url string) (gocv.Mat, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error retrieving image: %v", err)
		return gocv.Mat{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		img, _, err := image.Decode(resp.Body)
		if err != nil {
			return gocv.Mat{}, err
		}
		imgMat, err := gocv.ImageToMatRGB(img)
		if err != nil {
			return gocv.Mat{}, err
		}
		return imgMat, nil
	} else if resp.StatusCode == http.StatusNotFound {
		log.Fatal("Image not found, client disconnected")
	}
	return gocv.Mat{}, errors.New("image retrieval error")
}

var (
	imageFiles []string
	index      int
)

func RetrieveDevelopmentImages(imageDir string) (gocv.Mat, error) {
	if imageFiles == nil {
		files, err := readImageDirectory(imageDir)
		if err != nil {
			log.Printf("Error retrieving development images: %v", err)
			return gocv.Mat{}, err
		}
		imageFiles = files
		index = 0
	}
	img := gocv.IMRead(imageFiles[index], gocv.IMReadColor)
	if img.Empty() {
		log.Printf("Error retrieving development images: %v", imageFiles[index])
		return gocv.Mat{}, errors.New("image retrieval error")
	}
	index++
	if index >= len(imageFiles) {
		index = 0
	}
	return img, nil
}

func readImageDirectory(imageDir string) ([]string, error) {
	files, err := os.ReadDir(imageDir)
	if err != nil {
		return nil, err
	}
	var images []string
	for _, file := range files {
		images = append(images, filepath.Join(imageDir, file.Name()))
	}
	return images, nil
}
