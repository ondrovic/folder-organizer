package cli

import (
	"fmt"

	"github.com/ondrovic/folder-organizer/internal/types"
	"github.com/ondrovic/folder-organizer/internal/utils"

	"github.com/spf13/cobra"
)

var (
	organizeCmd = &cobra.Command{
		Use:   "organize <config-file-path> <folder-to-organize>",
		Short: "Organize the specified folder based on the JSON configuration",
		Long: `Organize files in a directory by sorting them into subdirectories based on file extensions.
The organization structure is defined in a JSON configuration file.
Example config file:
{
  "categories": {
    "images": [".jpg", ".png", ".gif"],
    "documents": [".pdf", ".docx", ".txt"],
    "videos": [".mp4", ".avi", ".mkv"]
  }
}`,
		Args: cobra.ExactArgs(2),
		RunE: runOrganize,
	}
)

func init() {
	// Add flags to the organize command
	organizeCmd.Flags().IntVarP(&options.NumOfWorkers, "workers", "w", 4, "Number of worker goroutines")
	organizeCmd.Flags().BoolVarP(&options.Recursive, "recursive", "r", true, "Process subdirectories recursively")
	organizeCmd.Flags().BoolVarP(&options.ShowProgress, "progress", "p", true, "Show progress during organization")
	organizeCmd.Flags().BoolVarP(&options.CleanupEmptyDirs, "cleanup", "c", true, "Remove empty directories after organization")
}

func runOrganize(cmd *cobra.Command, args []string) error {
	options.ConfigurationPath = args[0]
	options.Directory = args[1]

	// Configure organization options
	opts := types.OrganizeOptions{
		ConfigPath:   options.ConfigurationPath,
		SourcePath:   options.Directory,
		NumWorkers:   options.NumOfWorkers,
		Recursive:    options.Recursive,
		ShowProgress: options.ShowProgress,
	}

	// Run the organization
	stats, err := utils.OrganizeFiles(opts)
	if err != nil {
		return err
	}

	fmt.Printf("\n\tTotal files: %d\n", stats.TotalFiles)
	fmt.Printf("\tOrganized files: %d\n", stats.OrganizedFiles)
	fmt.Printf("\tSkipped files: %d\n", stats.SkippedFiles)

	if options.CleanupEmptyDirs {
		fmt.Printf("\n\tCleaning up empty directories...\n")
		removedCount, err := utils.CleanupEmptyDirs(options.Directory)
		if err != nil {
			return fmt.Errorf("\terror during cleanup: %w", err)
		}
		fmt.Printf("\tRemoved %d empty directories\n", removedCount)
		fmt.Println("")
	}

	return nil
}
