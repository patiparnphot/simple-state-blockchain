package main

import (
	"os"

	"github.com/patiparnphot/simple-state-blockchain/cli"
)

func main() {
	defer os.Exit(0)
	cmd := cli.CommandLine{}
	cmd.Run()

}
