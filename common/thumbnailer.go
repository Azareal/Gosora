package common

import (
	"image"
	"image/gif"
	"image/jpeg"
	_ "image/png"
	"os"
)

var Thumbnailer ThumbnailerInt

type ThumbnailerInt interface {
	Resize(format string, inPath string, tmpPath string, outPath string, width int) error
}

type RezThumbnailer struct {
}

func (thumb *RezThumbnailer) Resize(format string, inPath string, tmpPath string, outPath string, width int) error {
	// TODO: Sniff the aspect ratio of the image and calculate the dest height accordingly, bug make sure it isn't excessively high
	return nil
}

func (thumb *RezThumbnailer) resize(format string, inPath string, outPath string, width int, height int) error {
	return nil
}

// ! Note: CaireThumbnailer can't handle gifs, so we'll have to either cap their sizes or have another resizer deal with them
type CaireThumbnailer struct {
}

func NewCaireThumbnailer() *CaireThumbnailer {
	return &CaireThumbnailer{}
}

func precodeImage(format string, inPath string, tmpPath string) error {
	imageFile, err := os.Open(inPath)
	if err != nil {
		return err
	}
	defer imageFile.Close()

	img, _, err := image.Decode(imageFile)
	if err != nil {
		return err
	}

	outFile, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// TODO: Make sure animated gifs work after being encoded
	if format == "gif" {
		return gif.Encode(outFile, img, nil)
	}
	return jpeg.Encode(outFile, img, nil)
}

func (thumb *CaireThumbnailer) Resize(format string, inPath string, tmpPath string, outPath string, width int) error {
	err := precodeImage(format, inPath, tmpPath)
	if err != nil {
		return err
	}
	return nil

	// TODO: Caire doesn't work. Try something else. Or get them to fix the index out of range. We get enough wins from re-encoding as jpeg anyway
	/*imageFile, err := os.Open(tmpPath)
	if err != nil {
		return err
	}
	defer imageFile.Close()

	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	p := &caire.Processor{NewWidth: width, Scale: true}
	return p.Process(imageFile, outFile)*/
}

/*
type LilliputThumbnailer struct {

}
*/
