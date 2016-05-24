package main

import (
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

func fontFileName(fontData draw2d.FontData) string {
	return fontData.Name
}

func drawDigitsWithFont(char string, fontName string, fontSize float64) image.Image {
	canvas := image.NewRGBA(image.Rect(0, 0, 28, 28))
	gc := draw2dimg.NewGraphicContext(canvas)

	gc.DrawImage(image.White)    // Background color
	gc.SetFillColor(image.Black) // Text color

	gc.SetFontData(draw2d.FontData{Name: fontName})
	gc.SetFontSize(fontSize)

	left, top, right, bottom := gc.GetStringBounds(char)
	height := bottom - top
	width := right - left

	center := 28.0 / 2
	gc.FillStringAt(char, center-width/2, center+height/2)

	return canvas
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

	cnt := 0
	for _, font := range fontFiles {
		if filepath.Ext(font.Name()) != ".ttf" {
			continue
		}
		for _, c := range text {
			for _, fontSize := range fontSizes {
				cnt++
				digit := drawDigitsWithFont(string(c), font.Name(), fontSize)

				fileName := fmt.Sprintf("char-%06d.png", cnt)
				err := draw2dimg.SaveToPngFile(path.Join(outDir, fileName), digit)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}
		}
	}
}
