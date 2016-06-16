package grid

import (
	"fmt"
	"image"
	"math"
	"os"
	"path"

	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/mrfuxi/digit/common"
)

const ImageSize = 50

type EdgeType uint8

const (
	EdgeTypeCornerNW EdgeType = 1 << iota
	EdgeTypeCornerNE
	EdgeTypeCornerSE
	EdgeTypeCornerSW
)

type GridInfo struct {
	Fragment EdgeType
	Train    bool
}

type DrawDirections struct {
	GridInfo
}
type Image struct {
	GridInfo
	Image image.Image
}

type Counter struct {
	Image
	ID int
}

var (
	outDir    = path.Join("out")
	TestFile  = path.Join(outDir, "test.dat")
	TrainFile = path.Join(outDir, "train.dat")
)

func prepareDrawDirections(directions chan<- DrawDirections) {
	corners := []EdgeType{
		EdgeTypeCornerNW,
		EdgeTypeCornerNE,
		EdgeTypeCornerSE,
		EdgeTypeCornerSW,
	}

	for _, corner := range corners {
		directions <- DrawDirections{
			GridInfo: GridInfo{
				Fragment: corner,
				Train:    true,
			},
		}
	}

}

func draw(directions <-chan DrawDirections, images chan<- Image) {
	for {
		direction, ok := <-directions
		if !ok {
			break
		}
		digit, err := drawFragment(direction)
		if err != nil {
			fmt.Printf("Counld not draw %#v. Err: %s", direction, err.Error())
			continue
		}
		images <- Image{
			GridInfo: direction.GridInfo,
			Image:    digit,
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
		fileName := fmt.Sprintf("fragment-%06d-%v.png", counter.ID, counter.Fragment)
		if err := draw2dimg.SaveToPngFile(path.Join(outDir, fileName), counter.Image.Image); err != nil {
			fmt.Println(err)
		}
	}
}

func drawFragment(directions DrawDirections) (img image.Image, err error) {
	center := ImageSize / 2.0

	canvas := image.NewRGBA(image.Rect(0, 0, ImageSize, ImageSize))
	gc := draw2dimg.NewGraphicContext(canvas)

	gc.DrawImage(image.Black)      // Background color
	gc.SetStrokeColor(image.White) // Line color
	gc.SetLineWidth(2)

	var startAngle float64
	var diffAngle float64 = 90
	switch directions.Fragment {
	case EdgeTypeCornerNW:
		startAngle = 0
	case EdgeTypeCornerNE:
		startAngle = 90
	case EdgeTypeCornerSE:
		startAngle = 180
	case EdgeTypeCornerSW:
		startAngle = 270
	}

	gc.Translate(center, center)
	gc.Rotate(startAngle * math.Pi / 180.0)

	gc.MoveTo(0, 0)
	gc.LineTo(ImageSize, 0)
	gc.Close()
	gc.FillStroke()

	gc.Rotate(diffAngle * math.Pi / 180.0)

	gc.MoveTo(0, 0)
	gc.LineTo(ImageSize, 0)
	gc.Close()
	gc.FillStroke()

	return canvas, nil
}

func GenerateSudoku() error {
	os.RemoveAll(outDir)
	if err := os.Mkdir(outDir, 0764); err != nil {
		return err
	}

	directions := make(chan DrawDirections, 1)
	images := make(chan Image, 1)
	counters := make(chan Counter, 1)

	common.RoutineRunner(1, true, func() { prepareDrawDirections(directions) }, func() { close(directions) })
	common.RoutineRunner(4, true, func() { draw(directions, images) }, func() { close(images) })
	common.RoutineRunner(1, true, func() { imgCouter(images, counters) }, func() { close(counters) })
	common.RoutineRunner(1, false, func() { imgSaver(counters) }, nil)

	return nil
}
