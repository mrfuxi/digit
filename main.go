package main

import (
	"errors"
	"os"

	"github.com/urfave/cli"
)

var ErrInputMissing = errors.New("Input file missing")

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
						nn := buildNN()

						err := load(c.String("input"), nn)
						if err != nil {
							return err
						}

						runTraining(nn)
						save(c.String("output"), nn)
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
						nn := buildNN()
						if c.String("input") == "" {
							return ErrInputMissing
						}

						if err := load(c.String("input"), nn); err != nil {
							return err
						}

						validate(nn)
						return nil
					},
				},
			},
		},
		{
			Name:  "gen",
			Usage: "Generating train and test data",
			Action: func(c *cli.Context) error {
				text := c.Args().First()
				return generatData(text)
			},
		},
	}

	app.Run(os.Args)
}
