package grid

import (
	"fmt"
	"image"
	"math"
	"os"
	"path"

	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/mrfuxi/digit/common"
	"gopkg.in/cheggaaa/pb.v1"
)

const ImageSize = 50

type FragmentType uint8

const (
	FragmentTypeCornerNW FragmentType = 1 << iota
	FragmentTypeCornerNE
	FragmentTypeCornerSE
	FragmentTypeCornerSW
	FragmentTypeEdgeN
	FragmentTypeEdgeE
	FragmentTypeEdgeS
	FragmentTypeEdgeW
)

type GridInfo struct {
	Fragment FragmentType
	Train    bool
}

type DrawDirections struct {
	GridInfo
	DStartAngle float64
	DDiffAngle  float64
	Dx          float64
	Dy          float64
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
	outDir    = path.Join("out_grid")
	TestFile  = path.Join(outDir, "test.dat")
	TrainFile = path.Join(outDir, "train.dat")
)

var (
	progress *pb.ProgressBar
)

func IsCorner(fragment FragmentType) bool {
	switch fragment {
	case FragmentTypeCornerNW:
		return true
	case FragmentTypeCornerNE:
		return true
	case FragmentTypeCornerSE:
		return true
	case FragmentTypeCornerSW:
		return true
	}
	return false
}

func IsEdge(fragment FragmentType) bool {
	switch fragment {
	case FragmentTypeEdgeN:
		return true
	case FragmentTypeEdgeE:
		return true
	case FragmentTypeEdgeS:
		return true
	case FragmentTypeEdgeW:
		return true
	}
	return false
}

func prepareDrawDirections(directions chan<- DrawDirections) {
	fragments := []FragmentType{
		FragmentTypeCornerNW,
		FragmentTypeCornerNE,
		FragmentTypeCornerSE,
		FragmentTypeCornerSW,
		FragmentTypeEdgeN,
		FragmentTypeEdgeE,
		FragmentTypeEdgeS,
		FragmentTypeEdgeW,
	}

	xyMovements := []float64{}
	for d := 0.0; d <= ImageSize/4.0; d += 2.0 {
		xyMovements = append(xyMovements, d)
		if d != 0 {
			xyMovements = append(xyMovements, -d)
		}
	}
	dAngle := []float64{}
	for d := 0.0; d <= 15; d += 5 {
		dAngle = append(dAngle, d, -d)
		if d != 0 {
			dAngle = append(dAngle, -d)
		}
	}

	progress = pb.StartNew(len(fragments) * len(dAngle) * len(dAngle) * len(xyMovements) * len(xyMovements))

	for _, fragment := range fragments {
		for _, ds := range dAngle {
			for _, dd := range dAngle {
				for _, dx := range xyMovements {
					for _, dy := range xyMovements {
						directions <- DrawDirections{
							GridInfo: GridInfo{
								Fragment: fragment,
								Train:    true,
							},
							DStartAngle: ds,
							DDiffAngle:  dd,
							Dx:          dx,
							Dy:          dy,
						}
					}
				}
			}
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
		progress.Increment()
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
	case FragmentTypeCornerNW:
		startAngle = 0
	case FragmentTypeCornerNE:
		startAngle = 90
	case FragmentTypeCornerSE:
		startAngle = 180
	case FragmentTypeCornerSW:
		startAngle = 270

	case FragmentTypeEdgeN:
		startAngle = 0
	case FragmentTypeEdgeE:
		startAngle = 90
	case FragmentTypeEdgeS:
		startAngle = 180
	case FragmentTypeEdgeW:
		startAngle = 270
	}

	startAngle += directions.DStartAngle
	diffAngle += directions.DDiffAngle

	gc.Translate(center+directions.Dx, center+directions.Dy)
	gc.Rotate(startAngle * math.Pi / 180.0)

	if IsCorner(directions.Fragment) {
		gc.MoveTo(0, 0)
	} else {
		gc.MoveTo(-ImageSize, 0)
	}

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

func GenerateSudokuGrid() error {
	os.RemoveAll(outDir)
	if err := os.Mkdir(outDir, 0764); err != nil {
		return err
	}

	directions := make(chan DrawDirections, 1000)
	images := make(chan Image, 100)
	counters := make(chan Counter, 100)

	common.RoutineRunner(1, true, func() { prepareDrawDirections(directions) }, func() { close(directions) })
	common.RoutineRunner(4, true, func() { draw(directions, images) }, func() { close(images) })
	common.RoutineRunner(1, true, func() { imgCouter(images, counters) }, func() { close(counters) })
	common.RoutineRunner(4, false, func() { imgSaver(counters) }, nil)

	return nil
}
