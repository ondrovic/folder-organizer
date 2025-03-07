package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ondrovic/folder-organizer/internal/types"
)

func OrganizeFiles(opts types.OrganizeOptions) (*types.Stats, error) {
	// Load and parse the configuration file
	config, err := loadConfig(opts.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}

	// Build the extension mapping
	mapping, err := buildExtensionMapping(config)
	if err != nil {
		return nil, fmt.Errorf("error building extension mapping: %w", err)
	}

	// Get the extension to folder mapping
	extToFolder := mapping.ExtToPath

	// Create a channel for jobs
	jobs := make(chan types.FileJob, 100)

	// Create a wait group to wait for all workers to finish
	var wg sync.WaitGroup

	// Create stats to track progress
	stats := &types.Stats{}

	// Find all files and count them for progress tracking
	err = filepath.WalkDir(opts.SourcePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			// Get relative path to check if this file is already in a target folder
			relPath, err := filepath.Rel(opts.SourcePath, path)
			if err != nil {
				fmt.Printf("Error getting relative path for %s: %v\n", path, err)
				return nil
			}

			// Skip counting files that are already in a target directory with the right structure
			pathParts := strings.Split(filepath.ToSlash(relPath), "/")
			if len(pathParts) > 2 { // We need at least 3 levels: category/extension/file
				// Get the set of top-level folders from the mapping
				targetFolders := make(map[string]bool)
				for _, folderPath := range extToFolder {
					// Extract just the top-level folder
					topFolder := strings.Split(folderPath, string(filepath.Separator))[0]
					targetFolders[topFolder] = true
				}

				// Check if the first directory component matches any of our target folders
				if targetFolders[pathParts[0]] {
					// Also check if the second component is a valid extension folder
					if len(pathParts) > 2 && pathParts[1] == strings.TrimPrefix(filepath.Ext(d.Name()), ".") {
						return nil // Skip files already in organized folders with correct ext subfolder
					}
				}
			}

			stats.TotalFiles++
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error scanning directory: %w", err)
	}

	// Start the progress display with yacspin if requested
	var stopProgress chan struct{}
	if opts.ShowProgress {
		stopProgress = make(chan struct{})
		err := DisplayProgress(stats, stopProgress)
		if err != nil {
			return nil, fmt.Errorf("error starting progress display: %w", err)
		}
	}

	// Start worker goroutines
	for i := 0; i < opts.NumWorkers; i++ {
		wg.Add(1)
		go worker(jobs, &wg, stats)
	}

	// Walk through the source directory and find files to organize
	err = filepath.WalkDir(opts.SourcePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			// If not recursive, skip subdirectories at the root level
			if !opts.Recursive && path != opts.SourcePath {
				// Get relative path to see if it's an immediate subdirectory
				relPath, err := filepath.Rel(opts.SourcePath, path)
				if err != nil {
					return fmt.Errorf("error getting relative path: %w", err)
				}

				// If it's a subdirectory (not the root), skip it
				if !strings.Contains(relPath, string(filepath.Separator)) && relPath != "." {
					return fs.SkipDir
				}
			}
			return nil
		}

		// Get relative path to check if this file is already in a target folder
		relPath, err := filepath.Rel(opts.SourcePath, path)
		if err != nil {
			fmt.Printf("Error getting relative path for %s: %v\n", path, err)
			stats.IncrementProcessed()
			stats.IncrementSkipped()
			return nil
		}

		// Skip files that are already in a target directory with the right structure
		pathParts := strings.Split(filepath.ToSlash(relPath), "/")
		if len(pathParts) > 2 { // We need at least 3 levels: category/extension/file
			// Get the set of top-level folders from the mapping
			targetFolders := make(map[string]bool)
			for _, folderPath := range extToFolder {
				// Extract just the top-level folder
				topFolder := strings.Split(folderPath, string(filepath.Separator))[0]
				targetFolders[topFolder] = true
			}

			// Check if the first directory component matches any of our target folders
			if targetFolders[pathParts[0]] {
				// Also check if the second component is a valid extension folder
				if len(pathParts) > 2 && strings.HasPrefix(d.Name(), ".") && pathParts[1] == strings.TrimPrefix(filepath.Ext(d.Name()), ".") {
					stats.IncrementProcessed()
					stats.IncrementSkipped()
					return nil // Skip files already in organized folders with correct ext subfolder
				}
			}
		}

		// Get the file extension
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if ext == "" {
			stats.IncrementProcessed()
			stats.IncrementSkipped()
			return nil // Skip files without extension
		}

		// Check if this extension should be organized
		if folder, exists := extToFolder[ext]; exists {
			// Create a job for this file, adding extension folder as additional level
			// Remove the dot from extension for the folder name
			extFolder := ext[1:] // Skip the leading dot
			jobs <- types.FileJob{
				SourcePath: path,
				TargetDir:  filepath.Join(opts.SourcePath, folder, extFolder),
				Filename:   d.Name(),
			}
		} else {
			stats.IncrementProcessed()
			stats.IncrementSkipped()
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	// Close the jobs channel to signal workers to exit
	close(jobs)

	// Wait for all workers to finish
	wg.Wait()

	// Stop the progress display if it was started
	if opts.ShowProgress {
		close(stopProgress)
		time.Sleep(100 * time.Millisecond) // Give time for the last progress update
	}

	return stats, nil
}

// loadConfig loads and parses the JSON configuration file
func loadConfig(path string) (*types.Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var config types.Config
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// parseCategory processes a category entry from the JSON
func parseCategory(data json.RawMessage) (*types.Category, error) {
	category := &types.Category{
		Extensions:    []string{},
		Subcategories: make(map[string]*types.Category),
	}

	// Try to unmarshal as array of strings first (simple extension list)
	var extensions []string
	if err := json.Unmarshal(data, &extensions); err == nil {
		category.Extensions = extensions
		return category, nil
	}

	// Try to unmarshal as array of objects (may contain both subcategories and extensions)
	var objArray []json.RawMessage
	if err := json.Unmarshal(data, &objArray); err == nil {
		for _, item := range objArray {
			// Try as string array first (for extension lists inside arrays)
			var strArray []string
			if err := json.Unmarshal(item, &strArray); err == nil {
				category.Extensions = append(category.Extensions, strArray...)
				continue
			}

			// Try as map (for subcategories)
			var objMap map[string]json.RawMessage
			if err := json.Unmarshal(item, &objMap); err == nil {
				for subName, subData := range objMap {
					subCategory, err := parseCategory(subData)
					if err != nil {
						return nil, fmt.Errorf("error parsing subcategory %s: %w", subName, err)
					}
					category.Subcategories[subName] = subCategory
				}
				continue
			}
		}
		return category, nil
	}

	// Try to unmarshal as object (subcategories)
	var objMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &objMap); err == nil {
		for subName, subData := range objMap {
			subCategory, err := parseCategory(subData)
			if err != nil {
				return nil, fmt.Errorf("error parsing subcategory %s: %w", subName, err)
			}
			category.Subcategories[subName] = subCategory
		}
		return category, nil
	}

	return nil, fmt.Errorf("unable to parse category data")
}

// buildExtensionMapping converts the nested category structure to a flat mapping of extensions to paths
func buildExtensionMapping(config *types.Config) (*types.ExtensionMapping, error) {
	mapping := &types.ExtensionMapping{
		ExtToPath: make(map[string]string),
	}

	for topName, rawData := range config.Categories {
		category, err := parseCategory(rawData)
		if err != nil {
			return nil, fmt.Errorf("error parsing top-level category %s: %w", topName, err)
		}

		// Process the top-level extensions
		for _, ext := range category.Extensions {
			// Ensure extension starts with a dot
			if !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}
			mapping.ExtToPath[ext] = topName
		}

		// Process subcategories recursively
		if err := processSubcategories(mapping, topName, category.Subcategories); err != nil {
			return nil, err
		}
	}

	return mapping, nil
}

// processSubcategories recursively processes nested categories and builds the extension mapping
func processSubcategories(mapping *types.ExtensionMapping, parentPath string, subcats map[string]*types.Category) error {
	for subName, subCat := range subcats {
		currentPath := filepath.Join(parentPath, subName)

		// Process extensions in this subcategory
		for _, ext := range subCat.Extensions {
			// Ensure extension starts with a dot
			if !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}
			mapping.ExtToPath[ext] = currentPath
		}

		// Process deeper subcategories
		if err := processSubcategories(mapping, currentPath, subCat.Subcategories); err != nil {
			return err
		}
	}
	return nil
}

// worker processes file organization jobs
func worker(jobs <-chan types.FileJob, wg *sync.WaitGroup, stats *types.Stats) {
	defer wg.Done()

	for job := range jobs {
		// Ensure the target directory exists
		err := os.MkdirAll(job.TargetDir, 0755)
		if err != nil {
			fmt.Printf("Error creating directory %s: %v\n", job.TargetDir, err)
			stats.IncrementProcessed()
			stats.IncrementSkipped()
			continue
		}

		// Create the target file path
		targetPath := filepath.Join(job.TargetDir, job.Filename)

		// Check if the target file already exists
		if _, err := os.Stat(targetPath); err == nil {
			// File already exists, append a number to the filename
			ext := filepath.Ext(job.Filename)
			baseName := strings.TrimSuffix(job.Filename, ext)
			counter := 1
			for {
				newName := fmt.Sprintf("%s_%d%s", baseName, counter, ext)
				targetPath = filepath.Join(job.TargetDir, newName)
				if _, err := os.Stat(targetPath); os.IsNotExist(err) {
					break
				}
				counter++
			}
		}

		// Move the file using os.Rename which is more efficient
		err = os.Rename(job.SourcePath, targetPath)
		if err != nil {
			// If rename fails (likely cross-device), fall back to copy+delete
			if err := moveFileFallback(job.SourcePath, targetPath); err != nil {
				fmt.Printf("Error moving file %s: %v\n", job.SourcePath, err)
				stats.IncrementProcessed()
				stats.IncrementSkipped()
				continue
			}
		}

		stats.IncrementProcessed()
		stats.IncrementOrganized()
	}
}

// moveFileFallback implements a copy+delete fallback when os.Rename fails (cross-device moves)
func moveFileFallback(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create destination file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("copy file: %w", err)
	}

	// Ensure the file is flushed to disk
	err = destFile.Sync()
	if err != nil {
		return fmt.Errorf("sync file: %w", err)
	}

	// Close files before removing the source
	sourceFile.Close()
	destFile.Close()

	// Remove the source file
	if err := os.Remove(src); err != nil {
		return fmt.Errorf("remove source after copy: %w", err)
	}

	return nil
}
