package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/urfave/cli/v2"
)

func main() {
	app := configureApp()
	wimtkArgs, nextCommand := splitDash(os.Args)
	err := app.Run(wimtkArgs)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	startCommandIfNeeded(nextCommand)
}

func configureApp() *cli.App {
	var configMapName string
	var stateWatched string

	return &cli.App{
		Name:     "wimtk",
		Usage:    "Various tools for Kubernetes Pods containers",
		HelpName: "wimtk",
		Commands: []*cli.Command{
			{
				Name:    "publish-files",
				Aliases: []string{"pf"},
				Usage:   "Publish a list of files as a ConfigMap",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "configmap-name",
						Value:       "wimtk",
						Aliases:     []string{"c"},
						Usage:       "Name of the ConfigMap Create or Update",
						Destination: &configMapName,
					},
				},
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						fmt.Printf("Need at least one file\n")
					}
					publishFiles(c.Args().Slice(), configMapName)
					return nil
				},
			},
			{
				Name:    "wait-pods",
				Aliases: []string{"wp"},
				Usage:   "Wait until a list of pods have reach a specific status",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "state-watched",
						Value:       "Running",
						Aliases:     []string{"s"},
						Usage:       "Pod State to Wait for",
						Destination: &stateWatched,
					},
				},
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						fmt.Printf("Need at least one Pod\n")
					}
					waitPods(c.Args().Slice(), configMapName)
					return nil
				},
			},
		},
	}
}

func startCommandIfNeeded(command []string) {
	if len(command) != 0 {
		err := syscall.Exec(command[0], command[1:], os.Environ())
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("**** Please provide full path, you have no shell here ****")
			panic(err)
		}
	}
}

func splitDash(a []string) ([]string, []string) {
	for i, n := range a {
		if "--" == n && i < len(a) {
			return a[0:i], a[i+1:]
		}
	}
	return a, []string{}
}
