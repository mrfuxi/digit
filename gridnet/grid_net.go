package gridnet

import (
	"encoding/gob"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"

	"github.com/mrfuxi/digit/common"
	"github.com/mrfuxi/digit/gridgen"
	"github.com/mrfuxi/neural"
)

const (
	inputSize = gridgen.ImageSize * gridgen.ImageSize
	// outputSize = 4
	outputSize = 10
)

func prepareGridData(r io.Reader) (examples []neural.TrainExample) {
	dec := gob.NewDecoder(r)

	for {
		tmp := gridgen.Record{}
		err := dec.Decode(&tmp)
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		image := tmp.Pic
		label := tmp.Fragment
		if gridgen.IsEmpty(label) {
			label = gridgen.FragmentTypeEmpty
		}
		// label := tmp.FragmentSuper

		example := neural.TrainExample{
			Input:  make([]float64, inputSize, inputSize),
			Output: make([]float64, outputSize, outputSize),
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
	trainFile, err := os.Open(gridgen.TrainFile)
	if err != nil {
		panic(err)
	}
	defer trainFile.Close()
	tmp := prepareGridData(trainFile)
	trainData := tmp[:len(tmp)-5]
	validationData := tmp[len(tmp)-5:]
	return trainData, validationData
}

func loadTestData() []neural.TrainExample {
	testFile, err := os.Open(gridgen.TestFile)
	if err != nil {
		panic(err)
	}
	defer testFile.Close()

	testData := prepareGridData(testFile)
	return testData
}

func BuildNN() neural.Evaluator {
	activator := neural.NewSigmoidActivator()
	outActivator := neural.NewSoftmaxActivator()
	// outActivator := neural.NewSigmoidActivator()
	nn := neural.NewNeuralNetwork(
		[]int{inputSize, 20, outputSize},
		neural.NewFullyConnectedLayer(activator),
		neural.NewFullyConnectedLayer(outActivator),
	)
	return nn
}

func RunTraining(nn neural.Evaluator) {
	fmt.Println("Loading train data")
	testData := loadTestData()
	trainData, validationData := loadTrainData()

	cost := neural.NewCrossEntropyCost()
	// cost := neural.NewLogLikelihoodCost()
	options := neural.TrainOptions{
		Epochs:         5,
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

type randTrainer struct {
	BaseTrainer neural.Trainer
	Sample      neural.TrainExample
}

func NewRandomizedTrainer(network neural.Evaluator, cost neural.CostDerivative) neural.Trainer {
	t := &randTrainer{
		BaseTrainer: neural.NewBackpropagationTrainer(network, cost),
		Sample: neural.TrainExample{
			Input:  make([]float64, inputSize, inputSize),
			Output: make([]float64, outputSize, outputSize),
		},
	}

	return t
}

func (r *randTrainer) Process(sample neural.TrainExample, weightUpdates *neural.WeightUpdates) {
	copy(r.Sample.Input, sample.Input)
	copy(r.Sample.Output, sample.Output)

	// Empty
	if r.Sample.Output[0] == 1 {
		for i := 0; i < inputSize/10; i++ {
			r.Sample.Input[rand.Intn(inputSize)] = 1
		}
	}

	r.BaseTrainer.Process(r.Sample, weightUpdates)
}
