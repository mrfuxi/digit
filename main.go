package main

import (
	"errors"
	"os"

	"github.com/mrfuxi/digit/common"
	"github.com/mrfuxi/digit/digitgen"
	"github.com/mrfuxi/digit/digitnet"
	"github.com/mrfuxi/digit/gridgen"
	"github.com/mrfuxi/digit/gridnet"
	"github.com/urfave/cli"
)

var errInputMissing = errors.New("Input file missing")

func main() {
	netFlags := []cli.Flag{
		cli.StringFlag{
			Name:  "input, i",
			Usage: "Load network from `FILE`",
		},
		cli.StringFlag{
			Name:  "output, o",
			Usage: "Save network to `FILE`",
		},
	}

	app := cli.NewApp()
	app.Commands = []cli.Command{
		{
			Name:  "net",
			Usage: "Training and validating network",
			Subcommands: []cli.Command{
				{
					Name:  "digit",
					Flags: netFlags,
					Usage: "Train digit network",
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
					Name:  "grid",
					Flags: netFlags,
					Usage: "Train grid network",
					Action: func(c *cli.Context) error {
						nn := gridnet.BuildNN()

						err := common.LoadNN(c.String("input"), nn)
						if err != nil {
							return err
						}

						gridnet.RunTraining(nn)
						common.SaveNN(c.String("output"), nn)
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
