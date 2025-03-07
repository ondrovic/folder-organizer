package utils

import (
	"io/fs"
	"os"
	"path/filepath"
)

// CleanupEmptyDirs removes all empty directories in the specified path
// It works recursively from the bottom up to ensure nested empty directories are properly removed
func CleanupEmptyDirs(rootPath string) (int, error) {
	removedCount := 0

	// Normalize the path to handle spaces and special characters
	rootPath = filepath.Clean(rootPath)

	// This function will be called for each directory after its contents have been processed
	removeIfEmpty := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// fmt.Printf("Warning: Error accessing path: %v\n", err)
			return filepath.SkipDir
		}

		// Skip if not a directory or if it's the root path
		if !d.IsDir() || filepath.Clean(path) == rootPath {
			return nil
		}

		// Check if directory is empty
		entries, err := os.ReadDir(path)
		if err != nil {
			// fmt.Printf("Warning: Error reading directory %s: %v\n", path, err)
			return filepath.SkipDir
		}

		if len(entries) == 0 {
			// Directory is empty, remove it
			if err := os.Remove(path); err != nil {
				// fmt.Printf("Warning: Error removing empty directory %s: %v\n", path, err)
				return nil // Continue with other directories
			}
			removedCount++
		}

		return nil
	}

	// Walk the directory tree from the bottom up (to handle nested empty directories)
	// We need to do multiple passes because we might create new empty directories
	// after removing subdirectories
	for {
		currentRemovedCount := removedCount

		// Do a walkdir to actually remove empty directories
		err := filepath.WalkDir(rootPath, removeIfEmpty)
		if err != nil {
			// fmt.Printf("Warning: Some errors occurred during cleanup, but continuing: %v\n", err)
			// Continue despite errors
		}

		// If we didn't remove any directories in this pass, we're done
		if removedCount == currentRemovedCount {
			break
		}
	}

	return removedCount, nil
}
