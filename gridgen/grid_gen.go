package gridgen

import (
	"encoding/gob"
	"fmt"
	"image"
	"image/color"
	"os"
	"path"

	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/mrfuxi/digit/common"
	"gopkg.in/cheggaaa/pb.v1"
)

const ImageSize = 28
const imageCutOff = 5

type GridInfo struct {
	Fragment      FragmentType
	FragmentSuper FragmentSuperType
	Train         bool
	OffCenter     float64
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

func prepareMeta(drawers chan<- Drawer) {
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

	dAngleLine := []float64{}
	for d := 0.0; d <= 30; d += 5 {
		dAngleLine = append(dAngleLine, d, -d)
		if d != 0 {
			dAngleLine = append(dAngleLine, -d)
		}
	}

	dr := []Drawer{
		&cornerDrawer{Movements: xyMovements, Angles: dAngle},
		&edgeDrawer{Movements: xyMovements, Angles: dAngle},
		&crossDrawer{Movements: xyMovements, Angles: dAngle},
		&lineDrawer{Movements: xyMovements, Angles: dAngleLine},
		&emptyDrawer{Samples: 1000, Noise: 0.015},
		&incompleteEdgeDrawer{Movements: xyMovements, Angles: dAngle},
	}

	size := 0
	for _, d := range dr {
		size += d.Count()
	}
	progress = pb.StartNew(size)

	for _, d := range dr {
		drawers <- d
	}
}

func drawWithDrawer(drawers <-chan Drawer, images chan<- Image) {
	for {
		drawer, ok := <-drawers
		if !ok {
			break
		}

		drawer.Draw(images)
	}
}

func imgCouter(images <-chan Image, counters chan<- Counter) {
	cnt := 1
	for img := range images {
		if img.OffCenter > imageCutOff {
			img.GridInfo.Fragment = FragmentTypeEmpty
			img.GridInfo.FragmentSuper = FragmentSuperTypeEmpty
		}

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

	drawers := make(chan Drawer, 100)
	images := make(chan Image, 100)
	counters := make(chan Counter, 100)

	common.RoutineRunner(1, true, func() { prepareMeta(drawers) }, func() { close(drawers) })
	common.RoutineRunner(4, true, func() { drawWithDrawer(drawers, images) }, func() { close(images) })
	common.RoutineRunner(1, true, func() { imgCouter(images, counters) }, func() { close(counters) })
	common.RoutineRunner(4, false, func() { imgSaver(counters) }, nil)
	// common.RoutineRunner(1, false, func() { gobSaver(TrainFile, TestFile, counters) }, nil)

	return nil
}
