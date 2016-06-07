package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/mrfuxi/neural"
	"github.com/mrfuxi/neural/mat"
)

var (
	inputSize = 28 * 28
)

func prepareMnistData(r io.Reader) (examples []neural.TrainExample) {
	dec := gob.NewDecoder(r)

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
	}
	return
}

func loadTrainData() ([]neural.TrainExample, []neural.TrainExample) {
	trainFile, err := os.Open(trainFile)
	if err != nil {
		panic(err)
	}
	defer trainFile.Close()
	tmp := prepareMnistData(trainFile)
	trainData := tmp[:len(tmp)-10000]
	validationData := tmp[len(tmp)-10000:]
	return trainData, validationData
}

func loadTestData() []neural.TrainExample {
	testFile, err := os.Open(testFile)
	if err != nil {
		panic(err)
	}
	defer testFile.Close()

	testData := prepareMnistData(testFile)
	return testData
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

func load(fileName string, nn neural.Evaluator) error {
	if fileName == "" {
		return nil
	}

	fn, err := os.Open(fileName)
	if err != nil {
		return err
	}
	return neural.Load(nn, fn)
}

func save(fileName string, nn neural.Evaluator) error {
	if fileName == "" {
		return nil
	}

	fn, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	return neural.Save(nn, fn)
}

func buildNN() neural.Evaluator {
	activator := neural.NewSigmoidActivator()
	outActivator := neural.NewSoftmaxActivator()
	nn := neural.NewNeuralNetwork(
		[]int{inputSize, 100, 10},
		neural.NewFullyConnectedLayer(activator),
		neural.NewFullyConnectedLayer(outActivator),
	)
	return nn
}

func runTraining(nn neural.Evaluator) {
	fmt.Println("Loading train data")
	testData := loadTestData()
	trainData, validationData := loadTrainData()

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

	fmt.Println("Start training")

	t0 := time.Now()
	neural.Train(nn, trainData, options)
	dt := time.Since(t0)

	fmt.Println("Training complete in", dt)
}

func validate(nn neural.Evaluator) {
	testData := loadTestData()
	var different float64

	for _, sample := range testData {
		output := nn.Evaluate(sample.Input)

		if mat.ArgMax(output) != mat.ArgMax(sample.Output) {
			different++
		}
	}

	errorRate := different / float64(len(testData))
	fmt.Printf("Error rate: %.2f%%\n", errorRate*100)
}
