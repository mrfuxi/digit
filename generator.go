package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/petar/GoMNIST"
)

const ImageSize = 28

type FType uint8

const (
	FTypeMachine FType = 1 << iota
	FTypeHand
	FTypeTrueHand
)

type fontMap struct {
	name  string
	ftype FType
}

var (
	outDir      = "out"
	testFile    = path.Join(outDir, "test.dat")
	trainFile   = path.Join(outDir, "train.dat")
	mnistDir    = "mnist"
	fontDir     = "fonts"
	fontSubDirs = []fontMap{
		{"hand", FTypeHand},
		{"machine", FTypeMachine},
	}
)

// ErrFont is returned when font could not be loaded, therfore it could not be used
var ErrFont = errors.New("Font issue")
var ErrSize = errors.New("Char to big")
var ErrNoText = errors.New("Text missing")

type CharInfo struct {
	Char  string
	Type  FType
	Train bool
}

type DrawDirections struct {
	CharInfo
	FontName string
	FontSize float64
	Dx       float64
	Dy       float64
}

type Image struct {
	CharInfo
	Image image.Image
}

type Counter struct {
	Image
	ID int
}

type Record struct {
	Pic  [ImageSize * ImageSize]uint8
	Char string
	Type FType
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

func RoutineRunner(count int, async bool, worker func(), finalizer func()) {
	wg := sync.WaitGroup{}
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			worker()
			wg.Done()
		}()
	}

	fun := func() {
		wg.Wait()
		if finalizer != nil {
			finalizer()
		}
	}

	if async {
		go fun()
	} else {
		fun()
	}
}

func drawDigit(directions DrawDirections) (img image.Image, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("pkg: %v", r)
			}
		}
	}()

	canvas := image.NewRGBA(image.Rect(0, 0, ImageSize, ImageSize))
	gc := draw2dimg.NewGraphicContext(canvas)

	gc.DrawImage(image.Black)    // Background color
	gc.SetFillColor(image.White) // Text color

	gc.SetFontData(draw2d.FontData{Name: directions.FontName})
	gc.SetFontSize(directions.FontSize)

	left, top, right, bottom := gc.GetStringBounds(directions.Char)
	height := bottom - top
	width := right - left

	center := ImageSize / 2.0
	gc.FillStringAt(directions.Char, center-width/2+directions.Dx, center+height/2+directions.Dy)

	for i := 0; i < ImageSize; i++ {

		if grayAt(canvas, i, 0) > 0 || grayAt(canvas, i, ImageSize-1) > 0 || grayAt(canvas, 0, i) > 0 || grayAt(canvas, ImageSize-1, i) > 0 {
			return nil, ErrSize
		}
	}

	return canvas, nil
}

func grayAt(src image.Image, x, y int) uint8 {
	srcColor := src.At(x, y)
	dstColor := color.GrayModel.Convert(srcColor).(color.Gray)
	return dstColor.Y
}

func draw(directions <-chan DrawDirections, images chan<- Image) {
	for {
		direction, ok := <-directions
		if !ok {
			break
		}
		digit, err := drawDigit(direction)
		if err != nil {
			fmt.Println(direction.FontName, direction.Char, direction.FontSize, err)
			continue
		}
		images <- Image{
			CharInfo: direction.CharInfo,
			Image:    digit,
		}
	}
}

func prepareDrawDirections(text string, directions chan<- DrawDirections) {
	fontSizes := []float64{14, 16, 18, 20, 22, 24, 26}
	movements := []float64{-4, 0, 4}

	draw2d.SetFontFolder(fontDir)
	draw2d.SetFontNamer(fontFileName)

	var fonts []fontMap
	for _, fontSubDir := range fontSubDirs {
		fontSubDirPath := path.Join(fontDir, fontSubDir.name)
		fontFiles, err := ioutil.ReadDir(fontSubDirPath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, font := range fontFiles {
			fontPath := path.Join(fontSubDir.name, font.Name())
			if filepath.Ext(fontPath) != ".ttf" {
				continue
			} else if err := verifyFont(fontPath); err != nil {
				fmt.Println(err, fontPath)
				continue
			}
			fonts = append(fonts, fontMap{fontPath, fontSubDir.ftype})
		}
	}

	for _, font := range fonts {
		for _, c := range text {
			for _, fontSize := range fontSizes {
				for _, dx := range movements {
					for _, dy := range movements {
						directions <- DrawDirections{
							CharInfo: CharInfo{
								Char:  string(c),
								Type:  font.ftype,
								Train: rand.Intn(100) >= 5,
							},
							FontName: font.name,
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

func imgCouter(images <-chan Image, counters chan<- Counter) {
	cnt := 1
	for img := range images {
		counters <- Counter{
			Image: img,
			ID:    cnt,
		}
		cnt++
	}
}

func imgSaver(counters <-chan Counter) {
	for counter := range counters {
		fileName := fmt.Sprintf("char-%06d-%v.png", counter.ID, counter.Char)
		if err := draw2dimg.SaveToPngFile(path.Join(outDir, fileName), counter.Image.Image); err != nil {
			fmt.Println(err)
		}
	}
}

func gobSaver(counters <-chan Counter) {
	csvFileTrain, err := os.Create(trainFile)
	if err != nil {
		panic(err)
	}
	defer csvFileTrain.Close()

	csvFileTest, err := os.Create(testFile)
	if err != nil {
		panic(err)
	}
	defer csvFileTest.Close()

	train := gob.NewEncoder(csvFileTrain)
	test := gob.NewEncoder(csvFileTest)

	for counter := range counters {
		bounds := counter.Image.Image.Bounds()
		record := Record{
			Char: counter.CharInfo.Char,
			Type: counter.CharInfo.Type,
		}

		pos := 0
		for x := 0; x < bounds.Max.X; x++ {
			for y := 0; y < bounds.Max.Y; y++ {
				clr := counter.Image.Image.At(x, y)
				grayColor := color.GrayModel.Convert(clr).(color.Gray)
				record.Pic[pos] = grayColor.Y
				pos++
			}
		}
		if counter.CharInfo.Train {
			train.Encode(record)
		} else {
			test.Encode(record)
		}
	}
}

func drawMnist(images chan<- Image) {
	train, test, err := GoMNIST.Load(mnistDir)
	if err != nil {
		panic(err)
	}

	for i := 0; i < train.Count(); i++ {
		img, label := train.Get(i)
		images <- Image{
			CharInfo: CharInfo{
				Char:  strconv.Itoa(int(label)),
				Type:  FTypeTrueHand,
				Train: true,
			},
			Image: img,
		}
	}
	for i := 0; i < test.Count(); i++ {
		img, label := train.Get(i)
		images <- Image{
			CharInfo: CharInfo{
				Char:  strconv.Itoa(int(label)),
				Type:  FTypeTrueHand,
				Train: false,
			},
			Image: img,
		}
	}
}

func generatData(text string) error {
	if text == "" {
		return ErrNoText
	}

	os.RemoveAll(outDir)
	if err := os.Mkdir(outDir, 0764); err != nil {
		return err
	}

	directions := make(chan DrawDirections, 100)
	images := make(chan Image, 100)
	counters := make(chan Counter, 100)

	wgProducer := sync.WaitGroup{}
	wgProducer.Add(2)
	go func() {
		wgProducer.Wait()
		close(images)
	}()

	RoutineRunner(1, true, func() { prepareDrawDirections(text, directions) }, func() { close(directions) })
	RoutineRunner(4, true, func() { draw(directions, images) }, func() { wgProducer.Done() })
	RoutineRunner(1, true, func() { drawMnist(images) }, func() { wgProducer.Done() })
	RoutineRunner(1, true, func() { imgCouter(images, counters) }, func() { close(counters) })
	// RoutineRunner(4, false, func() { imgSaver(counters) }, nil)
	RoutineRunner(1, false, func() { gobSaver(counters) }, nil)

	return nil
}
