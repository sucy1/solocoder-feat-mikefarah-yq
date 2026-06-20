package main

import (
	"errors"
	"os"

	command "github.com/mikefarah/yq/v4/cmd"
)

type exitCodeError interface {
	ExitCode() int
}

func main() {
	cmd := command.New()

	args := os.Args[1:]

	_, _, err := cmd.Find(args)
	if err != nil && args[0] != "__complete" && args[0] != "__completeNoDesc" {
		newArgs := []string{"eval"}
		cmd.SetArgs(append(newArgs, os.Args[1:]...))
	}

	if err := cmd.Execute(); err != nil {
		var ecErr exitCodeError
		if errors.As(err, &ecErr) {
			os.Exit(ecErr.ExitCode())
		}
		os.Exit(1)
	}
}
