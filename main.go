package main

import (
	"errors"
	"os"

	"github.com/mrfuxi/digit/common"
	"github.com/mrfuxi/digit/digitgen"
	"github.com/mrfuxi/digit/digitnet"
	"github.com/mrfuxi/digit/gridgen"
	"github.com/urfave/cli"
)

var errInputMissing = errors.New("Input file missing")

func main() {
	app := cli.NewApp()
	app.Commands = []cli.Command{
		{
			Name:  "net",
			Usage: "Training and validating network",
			Subcommands: []cli.Command{
				{
					Name:  "train",
					Usage: "Run training on the network",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "input, i",
							Usage: "Load network from `FILE`",
						},
						cli.StringFlag{
							Name:  "output, o",
							Usage: "Save network to `FILE`",
						},
					},
					Action: func(c *cli.Context) error {
						nn := digitnet.BuildNN()

						err := common.LoadNN(c.String("input"), nn)
						if err != nil {
							return err
						}

						digitnet.RunTraining(nn)
						common.SaveNN(c.String("output"), nn)
						return nil
					},
				},
				{
					Name:  "validate",
					Usage: "Run validation on test data",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "input, i",
							Usage: "Load network from `FILE`",
						},
					},
					Action: func(c *cli.Context) error {
						nn := digitnet.BuildNN()
						if c.String("input") == "" {
							return errInputMissing
						}

						if err := common.LoadNN(c.String("input"), nn); err != nil {
							return err
						}

						digitnet.Validate(nn)
						return nil
					},
				},
			},
		},
		{
			Name:  "gen",
			Usage: "Generating train and test data",
			Subcommands: []cli.Command{
				{
					Name:  "digit",
					Usage: "Digits",
					Action: func(c *cli.Context) error {
						text := c.Args().First()
						return digitgen.GeneratDigits(text)
					},
				},
				{
					Name:  "grid",
					Usage: "Fragments of grid",
					Action: func(c *cli.Context) error {
						return gridgen.GenerateSudokuGrid()
					},
				},
			},
		},
	}

	app.Run(os.Args)
}
