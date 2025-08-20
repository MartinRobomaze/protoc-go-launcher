package main

import (
	"fmt"
	"go/build"
	"log"
	"os"
	"os/exec"

	"github.com/MartinRobomaze/protoc-go-launcher/protoc"
)

func main() {
	nArgs := len(os.Args)
	switch {
	case nArgs == 1:
		printUsage()
		log.Fatalf("missing protoc_version and commands provided for protoc.")
	case os.Args[1] != "--protoc_version":
		printUsage()
		log.Fatalf("protoc_version flag missing")
	case nArgs == 2:
		printUsage()
		log.Fatalf("missing protoc_version and commands provided for protoc.")
	case nArgs <= 3:
		printUsage()
		log.Fatalf("No commands provided for protoc")
	}

	protocPath, err := protoc.EnsureProtoc(os.Args[2])
	if err != nil {
		log.Fatal(err.Error())
	}

	pluginsPath, err := protoc.EnsureProtocPlugins()
	if err != nil {
		log.Fatal(err.Error())
	}

	var gopath string
	gopath = os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}

	cmd := exec.Command(protocPath, os.Args[3:]...)

	for _, plugin := range pluginsPath {
		cmd.Args = append(cmd.Args, "--plugin="+plugin)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err = cmd.Run(); err != nil {
		os.Exit(cmd.ProcessState.ExitCode())
	}
}

func printUsage() {
	fmt.Println("Usage: protoc-go-launcher --protoc_version <VERSION> PROTOC_COMMANDS")
}
