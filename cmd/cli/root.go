package cli

import (
	"github.com/ondrovic/folder-organizer/internal/types"

	"github.com/spf13/cobra"
)

var (
	options = types.CliFlags{}
	RootCmd = &cobra.Command{
		Use:   "folder-organizer",
		Short: "A Cli tool to organize files in a folder",
	}
)

func InitializeCommands() {
	RootCmd.AddCommand(organizeCmd)
}

func Execute() error {
	if err := RootCmd.Execute(); err != nil {
		return err
	}

	return nil
}
