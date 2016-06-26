package gridgen

import (
	"encoding/gob"
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"os"
	"path"

	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/mrfuxi/digit/common"
	"gopkg.in/cheggaaa/pb.v1"
)

const ImageSize = 28
const imageCutOff = 5

type FragmentType uint8

const (
	FragmentTypeEmpty FragmentType = iota
	FragmentTypeCornerNW
	FragmentTypeCornerNE
	FragmentTypeCornerSE
	FragmentTypeCornerSW
	FragmentTypeEdgeN
	FragmentTypeEdgeE
	FragmentTypeEdgeS
	FragmentTypeEdgeW
	FragmentTypeCross
	FragmentTypeLineH
	FragmentTypeLineV
)

type FragmentSuperType uint8

const (
	FragmentSuperTypeEmpty FragmentSuperType = iota
	FragmentSuperTypeCorner
	FragmentSuperTypeEdge
	FragmentSuperTypeCross
)

type GridInfo struct {
	Fragment      FragmentType
	FragmentSuper FragmentSuperType
	Train         bool
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

type Record struct {
	Pic           [ImageSize * ImageSize]uint8
	Fragment      FragmentType
	FragmentSuper FragmentSuperType
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

func IsCross(fragment FragmentType) bool {
	return fragment == FragmentTypeCross
}

func IsEmpty(fragment FragmentType) bool {
	if fragment == FragmentTypeEmpty {
		return true
	}
	if IsLine(fragment) {
		return true
	}

	return false
}

func IsLine(fragment FragmentType) bool {
	switch fragment {
	case FragmentTypeLineH:
		return true
	case FragmentTypeLineV:
		return true
	}

	return false
}

func FragmentTypeToSuper(fragment FragmentType) FragmentSuperType {
	switch {
	case IsCorner(fragment):
		return FragmentSuperTypeCorner
	case IsEdge(fragment):
		return FragmentSuperTypeEdge
	case IsCross(fragment):
		return FragmentSuperTypeCross
	case IsEmpty(fragment):
		return FragmentSuperTypeEmpty
	default:
		panic("What is this?")
	}
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
		FragmentTypeCross,
	}

	lineFragments := []FragmentType{
		FragmentTypeEmpty,
		FragmentTypeLineH,
		FragmentTypeLineV,
	}

	xyMovements := []float64{}
	for d := 0.0; d <= ImageSize/2.0; d += 2.0 {
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

	xyMovementsLine := []float64{}
	for d := 0.0; d <= ImageSize/2.0; d += 2.0 {
		xyMovementsLine = append(xyMovementsLine, d)
		if d != 0 {
			xyMovementsLine = append(xyMovementsLine, -d)
		}
	}
	dAngleLine := []float64{}
	for d := 0.0; d <= 30; d += 5 {
		dAngleLine = append(dAngleLine, d, -d)
		if d != 0 {
			dAngleLine = append(dAngleLine, -d)
		}
	}

	size := len(fragments) * len(dAngle) * len(dAngle) * len(xyMovements) * len(xyMovements)
	sizeLine := len(lineFragments) * len(dAngleLine) * len(xyMovementsLine) * len(xyMovementsLine)

	progress = pb.StartNew(size + sizeLine)

	for _, fragment := range fragments {
		for _, ds := range dAngle {
			for _, dd := range dAngle {
				for _, dx := range xyMovements {
					for _, dy := range xyMovements {
						directions <- DrawDirections{
							GridInfo: GridInfo{
								Fragment:      fragment,
								FragmentSuper: FragmentTypeToSuper(fragment),
								Train:         rand.Intn(100) >= 5,
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

	for _, fragment := range lineFragments {
		for _, ds := range dAngleLine {
			for _, dx := range xyMovementsLine {
				for _, dy := range xyMovementsLine {
					directions <- DrawDirections{
						GridInfo: GridInfo{
							Fragment:      fragment,
							FragmentSuper: FragmentTypeToSuper(fragment),
							Train:         rand.Intn(100) >= 5,
						},
						DStartAngle: ds,
						Dx:          dx,
						Dy:          dy,
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
		img, err := drawFragment(direction)
		if err != nil {
			fmt.Printf("Counld not draw %#v. Err: %s", direction, err.Error())
			progress.Increment()
			continue
		}

		if math.Abs(direction.Dx) > imageCutOff || math.Abs(direction.Dy) > imageCutOff {
			direction.GridInfo.Fragment = FragmentTypeEmpty
			direction.GridInfo.FragmentSuper = FragmentSuperTypeEmpty
		}

		images <- Image{
			GridInfo: direction.GridInfo,
			Image:    img,
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

	if directions.Fragment == FragmentTypeEmpty {
		return canvas, nil
	}

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

	case FragmentTypeLineH:
		startAngle = 0
	case FragmentTypeLineV:
		startAngle = 90
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

	if IsLine(directions.Fragment) {
		return canvas, nil
	}

	gc.Rotate(diffAngle * math.Pi / 180.0)

	if IsCross(directions.Fragment) {
		gc.LineTo(-ImageSize, 0)
	} else {
		gc.MoveTo(0, 0)
	}
	gc.LineTo(ImageSize, 0)
	gc.Close()
	gc.FillStroke()

	return canvas, nil
}

func gobSaver(trainFile string, testFile string, counters <-chan Counter) {
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
			Fragment:      counter.GridInfo.Fragment,
			FragmentSuper: counter.GridInfo.FragmentSuper,
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
		if counter.GridInfo.Train {
			train.Encode(record)
		} else {
			test.Encode(record)
		}
		progress.Increment()
	}
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
	// common.RoutineRunner(4, false, func() { imgSaver(counters) }, nil)
	common.RoutineRunner(1, false, func() { gobSaver(TrainFile, TestFile, counters) }, nil)

	return nil
}
