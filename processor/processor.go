// Package processor provides file processing functionality for removing
// trailing whitespace and ensuring proper newline handling.
package processor

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Result represents the outcome of processing a file.
type Result struct {
	Path    string
	Changed bool
	err     error
}

func (p *Result) IsErr() bool {
	return p.err != nil
}

func (p *Result) Err() error {
	return p.err
}

func makeResult(changed bool, err error) Result {
	return Result{
		Changed: changed,
		err:     err,
	}
}

// ProcessFile processes a file to remove trailing whitespace and ensure
// proper newline handling. Returns a Result indicating if changes were made.
func ProcessFile(filePath string) (result Result) {
	result = Result{Changed: false, err: nil, Path: filePath}

	var inputFile *os.File

	inputFile, result.err = os.Open(filePath)
	if result.IsErr() {
		return
	}
	defer inputFile.Close()

	dir := filepath.Dir(filePath)

	var tempFile *os.File
	tempFile, result.err = os.CreateTemp(dir, "sawmill_*.tmp")
	if result.IsErr() {
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	result = makeResult(processFileStreaming(inputFile, tempFile))
	if result.IsErr() {
		return
	}

	if !result.Changed {
		return
	}

	result.err = tempFile.Close()
	if result.IsErr() {
		return
	}

	result.err = inputFile.Close()

	if result.IsErr() {
		return
	}

	result.err = os.Rename(tempFile.Name(), filePath)
	if result.IsErr() {
		result.err = copyFile(tempFile.Name(), filePath)
		if result.IsErr() {
			return
		}
		result.err = os.Remove(tempFile.Name())
		if result.IsErr() {
			return
		}
	}
	return
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
