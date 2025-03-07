package utils

import (
	"fmt"
	"time"

	"github.com/ondrovic/folder-organizer/internal/types"
	"github.com/theckman/yacspin"
)

// NewProgressSpinner creates and configures a new spinner for displaying progress
func NewProgressSpinner() (*yacspin.Spinner, error) {
	cfg := yacspin.Config{
		Frequency:       100 * time.Millisecond,
		CharSet:         yacspin.CharSets[14],
		Suffix:          " ",
		SuffixAutoColon: false,
		Message:         "Starting",
		StopCharacter:   "✓",
		StopColors:      []string{"fgGreen"},
		StopMessage:     "Organization complete",
	}

	spinner, err := yacspin.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create spinner: %w", err)
	}

	return spinner, nil
}

// UpdateProgress updates the spinner with the current progress information
func UpdateProgress(spinner *yacspin.Spinner, stats *types.Stats) {
	if stats.TotalFiles > 0 {
		percentage := float64(stats.ProcessedFiles) / float64(stats.TotalFiles) * 100
		message := fmt.Sprintf("%d/%d files (%.1f%%) | Organized: %d | Skipped: %d",
			stats.ProcessedFiles, stats.TotalFiles, percentage, stats.OrganizedFiles, stats.SkippedFiles)
		spinner.Message(message)
	} else {
		message := fmt.Sprintf("%d files | Organized: %d | Skipped: %d",
			stats.ProcessedFiles, stats.OrganizedFiles, stats.SkippedFiles)
		spinner.Message(message)
	}
}

// DisplayProgress starts a spinner and continuously updates it with progress information
func DisplayProgress(stats *types.Stats, stop <-chan struct{}) error {
	spinner, err := NewProgressSpinner()
	if err != nil {
		return err
	}

	if err := spinner.Start(); err != nil {
		return fmt.Errorf("failed to start spinner: %w", err)
	}

	go func() {
		defer spinner.Stop()

		for {
			select {
			case <-stop:
				spinner.Message(fmt.Sprintf("Completed! Processed: %d | Organized: %d | Skipped: %d",
					stats.ProcessedFiles, stats.OrganizedFiles, stats.SkippedFiles))
				return
			case <-time.After(500 * time.Millisecond):
				UpdateProgress(spinner, stats)
			}
		}
	}()

	return nil
}
