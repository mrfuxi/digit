package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/mrfuxi/neural"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	nnSaveFile = flag.String("save-file", "", "Save neural network to file")
	nnLoadFile = flag.String("load-file", "", "Load neural network to file")
	inputSize  = 28 * 28
)

type Record struct {
	Pic  [28 * 28]uint8
	Char string
	Type uint8
}

func prepareMnistData(r io.Reader) (examples []neural.TrainExample) {
	dec := gob.NewDecoder(r)

	i := 0
	for {
		tmp := Record{}
		err := dec.Decode(&tmp)
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		image := tmp.Pic
		label, err := strconv.Atoi(tmp.Char)
		if err != nil {
			panic(err)
		}

		example := neural.TrainExample{
			Input:  make([]float64, inputSize, inputSize),
			Output: make([]float64, 10, 10),
		}

		for j, pix := range image {
			example.Input[j] = (float64(pix) / 255)
		}

		for j := range example.Output {
			example.Output[j] = 0
		}
		example.Output[label] = 1
		examples = append(examples, example)
		i++
	}
	return
}

func loadTestData() ([]neural.TrainExample, []neural.TrainExample, []neural.TrainExample) {
	trainFile, err := os.Open("train.dat")
	if err != nil {
		panic(err)
	}
	defer trainFile.Close()

	testFile, err := os.Open("test.dat")
	if err != nil {
		panic(err)
	}
	defer testFile.Close()

	tmp := prepareMnistData(trainFile)
	trainData := tmp[:len(tmp)-10000]
	validationData := tmp[len(tmp)-10000:]
	testData := prepareMnistData(testFile)
	return trainData, validationData, testData
}

func epocheCallback(nn neural.Evaluator, cost neural.Cost, validationData, testData []neural.TrainExample) neural.EpocheCallback {
	return func(epoche int, dt time.Duration) {
		_, validationErrors := neural.CalculateCorrectness(nn, cost, validationData)
		_, testErrors := neural.CalculateCorrectness(nn, cost, testData)
		if epoche == 1 {
			fmt.Println("epoche,validation error,test error")
		}
		fmt.Printf("%v,%v,%v\n", epoche, validationErrors, testErrors)
	}
}

func main() {
	trainData, validationData, testData := loadTestData()

	activator := neural.NewSigmoidActivator()
	outActivator := neural.NewSoftmaxActivator()
	nn := neural.NewNeuralNetwork(
		[]int{inputSize, 100, 10},
		neural.NewFullyConnectedLayer(activator),
		neural.NewFullyConnectedLayer(outActivator),
	)

	flag.Parse()

	if *nnLoadFile != "" {
		fn, err := os.Open(*nnLoadFile)
		if err != nil {
			log.Fatalln(err)
		}
		if err := neural.Load(nn, fn); err != nil {
			log.Fatalln(err)
		}
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	cost := neural.NewLogLikelihoodCost()
	options := neural.TrainOptions{
		Epochs:         50,
		MiniBatchSize:  10,
		LearningRate:   0.01,
		Regularization: 2,
		Momentum:       0.9,
		TrainerFactory: neural.NewBackpropagationTrainer,
		EpocheCallback: epocheCallback(nn, cost, validationData, testData),
		Cost:           cost,
	}

	t0 := time.Now()
	neural.Train(nn, trainData, options)
	dt := time.Since(t0)

	fmt.Println("Training complete in", dt)

	if *nnSaveFile != "" {
		fn, err := os.OpenFile(*nnSaveFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			log.Fatalln(err)
		}
		if err := neural.Save(nn, fn); err != nil {
			log.Fatalln(err)
		}
	}
}