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

type DrawDirections struct {
	Char     string
	FontName string
	FontSize float64
	Dx       float64
	Dy       float64
}

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

func draw(directions DrawDirections) (img image.Image, err error) {
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

	gc.SetFontData(draw2d.FontData{Name: directions.FontName})
	gc.SetFontSize(directions.FontSize)

	left, top, right, bottom := gc.GetStringBounds(directions.Char)
	height := bottom - top
	width := right - left

	center := 28.0 / 2
	gc.FillStringAt(directions.Char, center-width/2+directions.Dx, center+height/2+directions.Dy)

	return canvas, nil
}

func prepareDrawDirections(directions chan<- DrawDirections) {
	text := `123456789 +=\|/[]*-$#@`
	fontSizes := []float64{10, 14, 16, 18, 20, 22, 24, 26}
	movements := []float64{-4, 0, 4}

	draw2d.SetFontFolder(fontDir)
	draw2d.SetFontNamer(fontFileName)

	fontFiles, err := ioutil.ReadDir(fontDir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

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
				for _, dx := range movements {
					for _, dy := range movements {
						directions <- DrawDirections{
							Char:     string(c),
							FontName: font.Name(),
							FontSize: fontSize,
							Dx:       dx,
							Dy:       dy,
						}
					}
				}
			}
		}
	}

}

func main() {
	os.RemoveAll(outDir)
	if err := os.Mkdir(outDir, 0764); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	directions := make(chan DrawDirections, 1)
	go prepareDrawDirections(directions)

	cnt := 1
	for direction := range directions {
		digit, err := draw(direction)
		if err != nil {
			fmt.Println(direction.FontName, direction.Char, direction.FontSize, err)
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
