package digitnet

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/mrfuxi/digit/common"
	"github.com/mrfuxi/digit/digitgen"
	"github.com/mrfuxi/neural"
)

var (
	inputSize = 28 * 28
)

func prepareMnistData(r io.Reader) (examples []neural.TrainExample) {
	dec := gob.NewDecoder(r)

	for {
		tmp := digitgen.Record{}
		err := dec.Decode(&tmp)
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		image := tmp.Pic
		label, err := strconv.Atoi(tmp.Char)
		if err != nil {
			label = 0
		}

		example := neural.TrainExample{
			Input:  make([]float64, inputSize, inputSize),
			Output: make([]float64, 10, 10),
		}

		for j, pix := range image {
			example.Input[j] = (float64(pix)/255)*0.9 + 0.1
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
	trainFile, err := os.Open(digitgen.TrainFile)
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
	testFile, err := os.Open(digitgen.TestFile)
	if err != nil {
		panic(err)
	}
	defer testFile.Close()

	testData := prepareMnistData(testFile)
	return testData
}

func BuildNN() neural.Evaluator {
	activator := neural.NewSigmoidActivator()
	outActivator := neural.NewSoftmaxActivator()
	nn := neural.NewNeuralNetwork(
		[]int{inputSize, 100, 10},
		neural.NewFullyConnectedLayer(activator),
		neural.NewFullyConnectedLayer(outActivator),
	)
	return nn
}

func RunTraining(nn neural.Evaluator) {
	fmt.Println("Loading train data")
	testData := loadTestData()
	trainData, validationData := loadTrainData()

	cost := neural.NewLogLikelihoodCost()
	options := neural.TrainOptions{
		Epochs:         20,
		MiniBatchSize:  10,
		LearningRate:   0.01,
		Regularization: 2,
		Momentum:       0.9,
		TrainerFactory: neural.NewBackpropagationTrainer,
		EpocheCallback: common.EpocheCallback(nn, cost, validationData, testData),
		Cost:           cost,
	}

	fmt.Println("Start training")

	t0 := time.Now()
	neural.Train(nn, trainData, options)
	dt := time.Since(t0)

	fmt.Println("Training complete in", dt)
}
