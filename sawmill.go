package main

import (
	"bufio"
	"fmt"
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

func processFile(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, strings.TrimRight(line, " \t"))
	}

	if err = scanner.Err(); err != nil {
		return err
	}

	// Remove any trailing empty lines
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}


	output, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer output.Close()

	writer := bufio.NewWriter(output)
	for i, line := range lines {
		if _, err := writer.WriteString(line); err != nil {
			return err
		}
		// Add newline after each line except the last one
		if i < len(lines)-1 {
			if _, err := writer.WriteString("\n"); err != nil {
				return err
			}
		}
	}

	// Add exactly one final newline if file has content
	if len(lines) > 0 {
		if _, err := writer.WriteString("\n"); err != nil {
			return err
		}
	}

	return writer.Flush()
}
