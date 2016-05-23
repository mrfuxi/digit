package main

import (
	"fmt"
	"image"
	"os"
	"path"

	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
)

var outDir = "out"

func fontFileName(fontData draw2d.FontData) string {
	return fontData.Name + ".ttf"
}

func drawDigitsWithFont(digit string, fontName string, fontSize float64) image.Image {
	canvas := image.NewRGBA(image.Rect(0, 0, 28, 28))
	gc := draw2dimg.NewGraphicContext(canvas)

	gc.DrawImage(image.White)    // Background color
	gc.SetFillColor(image.Black) // Text color

	gc.SetFontData(draw2d.FontData{Name: fontName})
	gc.SetFontSize(fontSize)

	left, top, right, bottom := gc.GetStringBounds(digit)
	height := bottom - top
	width := right - left

	center := 28.0 / 2
	gc.FillStringAt(digit, center-width/2, center+height/2)

	return canvas
}

func main() {
	draw2d.SetFontFolder("./fonts/")
	draw2d.SetFontNamer(fontFileName)

	os.RemoveAll(outDir)
	if err := os.Mkdir(outDir, 0764); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	text := `123456789 +=\|/[]*-$#@`
	for i, c := range text {
		digit := drawDigitsWithFont(string(c), "OpenSans-Regular", 14)

		fileName := fmt.Sprintf("char-%06d.png", i)
		err := draw2dimg.SaveToPngFile(path.Join(outDir, fileName), digit)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
