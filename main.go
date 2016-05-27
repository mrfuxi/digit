package main

import (
	"errors"
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
)

var (
	outDir  = "out"
	fontDir = "fonts"
)

// ErrFont is returned when font could not be loaded, therfore it could not be used
var ErrFont = errors.New("Font issue")

func fontFileName(fontData draw2d.FontData) string {
	return fontData.Name
}

func verifyFont(fontName string) error {
	fontData := draw2d.FontData{Name: fontName}

	canvas := image.NewRGBA(image.Rect(0, 0, 1, 1))
	gc := draw2dimg.NewGraphicContext(canvas)
	gc.SetFontData(fontData)
	if draw2d.GetFont(fontData) == nil {
		return ErrFont
	}
	return nil
}

func drawDigitsWithFont(char, fontName string, fontSize, dx, dy float64) (img image.Image, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("pkg: %v", r)
			}
		}
	}()

	canvas := image.NewRGBA(image.Rect(0, 0, 28, 28))
	gc := draw2dimg.NewGraphicContext(canvas)

	gc.DrawImage(image.White)    // Background color
	gc.SetFillColor(image.Black) // Text color

	fontData := draw2d.FontData{Name: fontName}
	gc.SetFontData(fontData)
	gc.SetFontSize(fontSize)

	left, top, right, bottom := gc.GetStringBounds(char)
	height := bottom - top
	width := right - left

	center := 28.0 / 2
	gc.FillStringAt(char, center-width/2+dx, center+height/2+dy)

	return canvas, nil
}

func main() {
	draw2d.SetFontFolder(fontDir)
	draw2d.SetFontNamer(fontFileName)

	os.RemoveAll(outDir)
	if err := os.Mkdir(outDir, 0764); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fontFiles, err := ioutil.ReadDir(fontDir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	text := `123456789 +=\|/[]*-$#@`
	fontSizes := []float64{10, 14, 16, 18, 20, 22, 24, 26}

	cnt := 1
	for _, font := range fontFiles {
		if filepath.Ext(font.Name()) != ".ttf" {
			continue
		}
		if err := verifyFont(font.Name()); err != nil {
			fmt.Println(err, font.Name())
			continue
		}

		for _, c := range text {
			for _, fontSize := range fontSizes {
				for dx := -4; dx <= 4; dx += 4 {
					for dy := -4; dy <= 4; dy += 4 {

						digit, err := drawDigitsWithFont(string(c), font.Name(), fontSize, float64(dx), float64(dy))
						if err != nil {
							fmt.Println(font.Name(), string(c), fontSize, err)
							continue
						}

						fileName := fmt.Sprintf("char-%06d.png", cnt)
						err = draw2dimg.SaveToPngFile(path.Join(outDir, fileName), digit)
						if err != nil {
							fmt.Println(err)
							os.Exit(1)
						}
						cnt++
					}
				}
			}
		}
	}
}
