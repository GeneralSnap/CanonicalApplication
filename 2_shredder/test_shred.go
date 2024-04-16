package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func Shred(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("Error accessing file: %w", err)
	}

	size := fileInfo.Size()

	for i := 0; i < 3; i++ {
		file, err := os.OpenFile(path, os.O_WRONLY, 0)
		if err != nil {
			return fmt.Errorf("Error opening file on pass %d: %w", i+1, err)
		}

		if _, err = io.CopyN(file, rand.Reader, size); err != nil {
			file.Close()
			return fmt.Errorf("Error writing random data on pass %d: %w", i+1, err)
		}

		file.Sync()
		file.Close()
	}

	return os.Remove(path)
}

func copyFiles(srcDir, destDir string) error {
	files, err := ioutil.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("Failed to read source directory '%s': %w", srcDir, err)
	}

	if len(files) == 0 {
		return fmt.Errorf("No files to copy in '%s'", srcDir)
	}

	for _, file := range files {
		srcPath := filepath.Join(srcDir, file.Name())
		destPath := filepath.Join(destDir, file.Name())

		input, err := os.ReadFile(srcPath)
		if err != nil {
			fmt.Printf("Failed to read '%s': %s\n", srcPath, err)
			continue
		}

		if err = os.WriteFile(destPath, input, 0644); err != nil {
			fmt.Printf("Failed to write '%s': %s\n", destPath, err)
			continue
		}
	}

	return nil
}

func main() {
	// Hard-coded directory paths
	srcDir := "./source_files"
	destDir := "./test_files"

	// Ensure source directory exists
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		fmt.Printf("Source directory '%s' does not exist\n", srcDir)
		os.Exit(1)
	}

	// Create destination directory if it does not exist
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		if err := os.MkdirAll(destDir, 0755); err != nil {
			fmt.Printf("Failed to create destination directory '%s': %s\n", destDir, err)
			os.Exit(1)
		}
	}

	// Copy files from source to destination
	if err := copyFiles(srcDir, destDir); err != nil {
		fmt.Printf("Error copying files: %s\n", err)
		os.Exit(1)
	}

	// Shred files in the destination directory
	files, err := ioutil.ReadDir(destDir)
	if err != nil {
		fmt.Printf("Failed to read destination directory: %s\n", err)
		os.Exit(1)
	}

	for _, file := range files {
		filePath := filepath.Join(destDir, file.Name())
		err := Shred(filePath)
		if err != nil {
			fmt.Printf("Failed to shred '%s': %s\n", filePath, err)
		} else {
			fmt.Printf("Successfully shredded '%s'\n", filePath)
		}
	}
}
