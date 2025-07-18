package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: sawmill <pattern>")
		fmt.Println("Example: sawmill.go '*.go'")
		os.Exit(1)
	}

	pattern := os.Args[1]
	root := "."

	err := filepath.Walk(
		root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			matched, err := filepath.Match(pattern, info.Name())
			if err != nil {
				return err
			}

			if matched {
				fmt.Printf("Processing: %s\n", path)
				if err := processFile(path); err != nil {
					fmt.Printf("Error processing %s: %v\n", path, err)
				}
			}

			return nil
		})
	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}
}

func processFile(filePath string) error {
	inputFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	dir := filepath.Dir(filePath)
	tempFile, err := os.CreateTemp(dir, "sawmill_*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	changed, err := processFileStreaming(inputFile, tempFile)
	if err != nil {
		return err
	}

	if !changed {
		return nil
	}

	if err := tempFile.Close(); err != nil {
		return err
	}

	if err := inputFile.Close(); err != nil {
		return err
	}

	if err := os.Rename(tempFile.Name(), filePath); err != nil {
		if err := copyFile(tempFile.Name(), filePath); err != nil {
			return err
		}
		return os.Remove(tempFile.Name())
	}
	return nil
}

func processFileStreaming(input io.Reader, output io.Writer) (bool, error) {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	writer := bufio.NewWriter(output)
	defer writer.Flush()

	hasChanged := false
	pendingNewlines := 0
	hasContent := false

	for scanner.Scan() {
		originalLine := scanner.Text()
		processedLine := strings.TrimRight(originalLine, " \t")

		if originalLine != processedLine {
			hasChanged = true
		}

		if processedLine == "" {
			pendingNewlines++
		} else {
			if hasContent {
				for i := 0; i < pendingNewlines; i++ {
					if _, err := writer.WriteString("\n"); err != nil {
						return false, err
					}
				}
				if _, err := writer.WriteString("\n"); err != nil {
					return false, err
				}
			}
			pendingNewlines = 0
			hasContent = true

			if _, err := writer.WriteString(processedLine); err != nil {
				return false, err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	if pendingNewlines > 0 {
		hasChanged = true
	}

	if hasContent {
		if _, err := writer.WriteString("\n"); err != nil {
			return false, err
		}
	}

	return hasChanged, nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
