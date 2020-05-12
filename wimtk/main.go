package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/urfave/cli/v2"
)

var verbose bool
var debug bool
var namespace string

var version = "0.5.1"

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
	var statusWatched string

	return &cli.App{
		Name:     "wimtk",
		Usage:    "Various tools for Kubernetes Pods containers",
		HelpName: "wimtk",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"v"},
				Usage:       "Activate verbose mode",
				Destination: &verbose,
			},
			&cli.BoolFlag{
				Name:        "debug",
				Aliases:     []string{"d"},
				Usage:       "Activate debug mode",
				Destination: &debug,
			},
		},
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
				Name:    "wait-status",
				Aliases: []string{"ws"},
				Usage:   "Wait until a list of pods have reach a specific status",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "status-watched",
						Value:       "Running",
						Aliases:     []string{"s"},
						Usage:       "Pod status to Wait for (Running, Pending, ...)",
						Destination: &statusWatched,
					},
				},
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						fmt.Printf("Need at least one Pod\n")
					}
					waitPods(c.Args().Slice(), statusWatched)
					return nil
				},
			},
			{
				Name:    "sync-map",
				Aliases: []string{"sm"},
				Usage:   "Sync a ConfiMap in another namespace into current namespace",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "namespace",
						Value:       "default",
						Aliases:     []string{"n"},
						Usage:       "namespace to take the ConfigMap from",
						Destination: &namespace,
						Required:    true,
					},
				},
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						fmt.Printf("Need at least one configMap\n")
					}
					syncMap(namespace, c.Args().First())
					return nil
				},
			},
			{
				Name:  "noop",
				Usage: "Do nothing (execute next command)",
				Action: func(c *cli.Context) error {
					return nil
				},
			},
			{
				Name:  "version",
				Usage: "Gives current version of wimtk",
				Action: func(c *cli.Context) error {
					fmt.Printf("WimTK version %v\n", version)
					return nil
				},
			},
		},
	}
}

func startCommandIfNeeded(command []string) {
	if len(command) == 0 {
		return
	}

	binary, err := exec.LookPath(command[0])
	panicErr(err)

	VerboseF("Executing %v with args %v\n", binary, command)
	err = syscall.Exec(binary, command, os.Environ())
	panicErr(err)
}

func splitDash(a []string) ([]string, []string) {
	for i, n := range a {
		if "--" == n && i < len(a) {
			return a[0:i], a[i+1:]
		}
	}
	return a, []string{}
}
