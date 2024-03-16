package main

import (
	"os"
	"log"
	"k8s.io/component-base/cli"
	"k8s.io/kubernetes/cmd/kube-scheduler/app"
	"my-scheduler-plugins/pkg/plugins"
)

func main() {
	// Register custom plugins to the scheduler framework.
	log.Printf("custom-scheduler starts!\n")
	command := app.NewSchedulerCommand(
		app.WithPlugin(plugins.Name, plugins.New),
	)

	code := cli.Run(command)
	os.Exit(code)
}
