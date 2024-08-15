package pipeline

import (
	"image"
	"image/png"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func RetrieveImage(url string) (*image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error retrieving image: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		img, _, err := image.Decode(resp.Body)
		if err != nil {
			return nil, err
		}
		return &img, nil
	} else if resp.StatusCode == http.StatusNotFound {
		log.Fatal("Image not found, client disconnected")
	} else {
		log.Fatalf("Error retrieving image: %v", err)
	}
}

var (
	imageFiles []string
	index      int
)

func RetrieveDevelopmentImages(imageDir string) (*image.Image, error) {
	if imageFiles == nil {
		files, err := readImageDirectory(imageDir)
		if err != nil {
			log.Printf("Error retrieving development images: %v", err)
			return nil, err
		}
		imageFiles = files
		index = 0
	}
	file, err := os.Open(imageFiles[index])
	if err != nil {
		log.Printf("Error retrieving development images: %v", err)
		return nil, err
	}
	index++
	img, err := png.Decode(file)
	if err != nil {
		log.Printf("Error decoding development images: %v", err)
		return nil, err
	}
	return &img, nil
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
