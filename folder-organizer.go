package main

import (
	sCli "github.com/ondrovic/common/utils/cli"
	"github.com/ondrovic/folder-organizer/cmd/cli"
	"runtime"
)

func main() {
	if err := sCli.ClearTerminalScreen(runtime.GOOS); err != nil {
		return
	}

	cli.InitializeCommands()

	if err := cli.RootCmd.Execute(); err != nil {
		return
	}
}
