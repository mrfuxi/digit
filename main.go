package main

import (
	"image"

	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
)

func fontFileName(fontData draw2d.FontData) string {
	return fontData.Name + ".ttf"
}

func main() {
	draw2d.SetFontFolder("./fonts/")
	draw2d.SetFontNamer(fontFileName)

	dest := image.NewRGBA(image.Rect(0, 0, 105, 18))
	gc := draw2dimg.NewGraphicContext(dest)
	gc.FillStroke()

	gc.SetFontData(draw2d.FontData{Name: "OpenSans-Regular"})
	gc.SetFillColor(image.Black) // Text color
	gc.SetFontSize(14)
	gc.FillStringAt("0123456789", 0, 14)
	gc.Close()

	// Save to file
	draw2dimg.SaveToPngFile("digits.png", dest)
}
