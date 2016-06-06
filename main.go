package main

import (
	"os"

	"github.com/urfave/cli"
)

func main() {
	nn := buildNN()

	app := cli.NewApp()
	app.Commands = []cli.Command{
		{
			Name:  "net",
			Usage: "Training and validating network",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "input, i",
					Usage: "Load network from `FILE`",
				},
			},
			Before: func(c *cli.Context) error {
				err := load(c.String("input"), nn)
				if err != nil {
					return err
				}
				return err
			},
			Subcommands: []cli.Command{
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
				return generatData()
			},
		},
	}

	app.Run(os.Args)
}
