package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/dav009/bqt"

	cli "github.com/urfave/cli/v2"
)

func TestCommand() cli.Command {
	return cli.Command{
		Name:    "test",
		Aliases: []string{"t"},
		Usage:   "Run tests either using a local BQ emulator or directly running your queries on the cloud",
		Flags: []cli.Flag{&cli.StringFlag{
			Name:     "tests",
			Value:    "unit_tests/",
			Usage:    "Path to your folder containing json test definitions",
			Required: false,
		},
			&cli.StringFlag{
				Name:     "mode",
				Value:    "local",
				Usage:    "`local` (default) runs your test on a BQ emulator. 'cloud': runs your queries on the cloud",
				Required: false,
			},
		},
		Action: func(cCtx *cli.Context) error {
			mode := cCtx.String("mode")
			testsPath := cCtx.String("tests")
			fmt.Println("Parsing tests in: ", testsPath)
			tests, err := bqt.ParseFolder(testsPath)
			if err != nil {
				return err
			}
			fmt.Println("Parsed Tests: ", len(tests))
			fmt.Println("Running Tests...")
			err = bqt.RunTests(mode, tests)
			if err != nil {
				return err
			}
			return nil
		},
	}
}

func GenerateSQL() cli.Command {
	return cli.Command{
		Name:    "generate",
		Aliases: []string{"g"},
		Usage:   "Generate the SQL queries used for tests",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "tests",
				Value:    "unit_tests/",
				Usage:    "Path to your folder containing json test definitions",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "output",
				Value:    "tests",
				Usage:    "Path where test queries will be gneerated",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "flavor",
				Value:    "test",
				Usage:    "`test`(default): Generate SQL run during tests. simple: generates the sql code of a model",
				Required: false,
			},
		},
		Action: func(cCtx *cli.Context) error {
			testsPath := cCtx.String("tests")
			output := cCtx.String("output")
			flavor := cCtx.String("flavor")
			fmt.Println("Parsing tests in: ", testsPath)
			tests, err := bqt.ParseFolder(testsPath)
			if err != nil {
				return err
			}
			fmt.Println("Parsed Tests: ", len(tests))
			fmt.Println("Generating SQL...")
			if flavor == "simple" {
				for _, t := range tests {
					sqlQueries, err := bqt.GenerateTestSQL(t)
					if err != nil {
						return err
					}
					path := filepath.Join(output, t.Name+".sql")
					err = bqt.SaveSQL(path, sqlQueries.QueryWithMockedData)
				}
			}
			if flavor == "test " {
				for _, t := range tests {
					fmt.Println("Generating Test: ", t.Name)
					sqlQueries, err := bqt.GenerateTestSQL(t)
					if err != nil {
						return err
					}

					path := filepath.Join(output, t.Name+"_"+"ExpectedMinusQuery"+".sql")
					fmt.Println("Saving Test: ", path)
					fmt.Println(sqlQueries.ExpectedMinusQuery)
					err = bqt.SaveSQL(path, sqlQueries.ExpectedMinusQuery)
					if err != nil {
						return err
					}

					path = filepath.Join(output, t.Name+"_"+"QueryMinusExpected"+".sql")
					fmt.Println("Saving Test: ", path)
					err = bqt.SaveSQL(path, sqlQueries.QueryMinusExpected)
					if err != nil {
						return err
					}

				}
			}
			return nil
		},
	}
}

func main() {
	testCmd := TestCommand()
	genCmd := GenerateSQL()
	app := cli.NewApp()
	app.Name = "bqt"
	app.Usage = ""
	app.Commands = []*cli.Command{
		&testCmd,
		&genCmd,
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
