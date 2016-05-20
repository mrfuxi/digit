package main

import (
	"image"

	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
)

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

	digit := drawDigitsWithFont("8", "OpenSans-Regular", 14)
	draw2dimg.SaveToPngFile("digits.png", digit)
}
