package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {

	publishFilesCmd := flag.NewFlagSet("publish-files", flag.ExitOnError)
	var configMapName string
	publishFilesCmd.StringVar(&configMapName, "configMapName", "wimtk", "Name of the ConfigMap Create or Update")

	waitPodsCmd := flag.NewFlagSet("wait-pods", flag.ExitOnError)
	var stateWatched string
	waitPodsCmd.StringVar(&stateWatched, "watchedSate", "Running", "Pod State to Wait for")

	if len(os.Args) < 2 {
		fmt.Println("Expected subcommands")
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "publish-files":
		publishFilesCmd.Parse(os.Args[2:])
		publishFiles(publishFilesCmd.Args(), configMapName)
	case "wait-pods":
		waitPodsCmd.Parse(os.Args[2:])
		waitPods(waitPodsCmd.Args(), stateWatched)
	default:
		usage()
		os.Exit(1)
	}
}

var usage = func() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(flag.CommandLine.Output(), `
Various tools to help with kubernetes accesss from within the Pods

Basic Commands:
  publish-files         Publish a list of files as a ConfigMap
  wait-pods             Wait until a list of pods have reach a specific Phase

`)
	flag.PrintDefaults()
}
