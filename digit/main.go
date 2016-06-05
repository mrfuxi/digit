package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/mrfuxi/neural"
	"github.com/mrfuxi/neural/mat"
	"github.com/urfave/cli"
)

var (
	inputSize = 28 * 28
)

type Record struct {
	Pic  [28 * 28]uint8
	Char string
	Type uint8
}

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
	trainFile, err := os.Open("train.dat")
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
	testFile, err := os.Open("test.dat")
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

func main() {
	nn := buildNN()
	loaded := false

	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "input, i",
			Usage: "Load network from `FILE`",
		},
	}
	app.Before = func(c *cli.Context) error {
		fn := c.GlobalString("input")
		err := load(fn, nn)
		if fn != "" && err == nil {
			loaded = true
		}
		return err
	}
	app.Commands = []cli.Command{
		{
			Name:  "train",
			Usage: "Run training on the network",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "output, o",
					Usage: "Save network to `FILE`",
				},
			},
			After: func(c *cli.Context) error {
				return save(c.String("output"), nn)
			},
			Action: func(c *cli.Context) error {
				runTraining(nn)
				return nil
			},
		},
		{
			Name:  "validate",
			Usage: "Run validation on test data",
			Action: func(c *cli.Context) error {
				if !loaded {
					return errors.New("Neural network not loaded. Nothing to validate")
				}

				validate(nn)
				return nil
			},
		},
	}

	app.Run(os.Args)
}
