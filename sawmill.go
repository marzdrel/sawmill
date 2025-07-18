package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	gitignore "github.com/denormal/go-gitignore"
)

var defaultPatterns = []string{
	"*.go", "*.js", "*.ts", "*.py", ".rb",
	"*.toml", "*.yml", "*.yaml", "*.json", "*.xml",
	"*.html", "*.css", "*.scss", "*.md",
	"*.txt", "*.conf", "*.ini", "*.sh",
	"*.tf", "Dockerfile.*",
}

type runStats struct {
	FilesProcessed int
	FilesChanged   int
	verbose        bool
}

func (s *runStats) Log(template string, args ...any) {
	if s.verbose {
		fmt.Printf(template, args...)
	}
}

func main() {
	var extensions []string
	var stats runStats

	patternFlag := flag.String("pattern", "",
		"Comma-separated list of file patterns to process")

	verboseFlag := flag.Bool("verbose", false,
		"Enable verbose output")

	flag.Parse()

	stats.verbose = *verboseFlag
	pattern := *patternFlag

	extensions = defaultPatterns
	if len(pattern) > 0 {
		extensions = strings.Split(pattern, ",")
	}

	root := "."

	// Load .gitignore if it exists
	var gi gitignore.GitIgnore
	if _, err := os.Stat(".gitignore"); err == nil {
		if parsedGi, parseErr := gitignore.NewFromFile(".gitignore"); parseErr == nil {
			gi = parsedGi
		}
	}

	err := filepath.Walk(
		root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			// Skip files ignored by gitignore
			if gi != nil {
				// Use relative path for gitignore matching
				relPath := path
				if strings.HasPrefix(path, "./") {
					relPath = path[2:]
				}
				if match := gi.Match(relPath); match != nil && match.Ignore() {
					stats.Log("Skipping ignored file: %s\n", path)
					return nil
				}
			}

			var matched bool

			for _, ext := range extensions {
				matched, err = filepath.Match(ext, info.Name())
				if err != nil {
					return err
				}
				if matched {
					break
				}
			}

			if matched {
				stats.FilesProcessed++
				stats.Log("Processing: %s\n", path)
				result := processFile(path)
				if result.isErr() {
					fmt.Printf("Error processing %s: %v\n", path, err)
				}

				if result.Changed {
					stats.FilesChanged++
					fmt.Printf("File changed: %s\n", path)
				}
			}

			return nil
		})

	fmt.Printf("Processed %d files, changed %d files.\n", stats.FilesProcessed, stats.FilesChanged)
	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}
}

type processResult struct {
	Changed bool
	err     error
}

func (p *processResult) isErr() bool {
	return p.err != nil
}

func makeProcessResult(changed bool, err error) processResult {
	return processResult{
		Changed: changed,
		err:     err,
	}
}

func processFile(filePath string) (result processResult) {
	result = processResult{Changed: false, err: nil}

	var inputFile *os.File

	inputFile, result.err = os.Open(filePath)
	if result.isErr() {
		return
	}
	defer inputFile.Close()

	dir := filepath.Dir(filePath)

	var tempFile *os.File
	tempFile, result.err = os.CreateTemp(dir, "sawmill_*.tmp")
	if result.isErr() {
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	result = makeProcessResult(processFileStreaming(inputFile, tempFile))
	if result.isErr() {
		return
	}

	if !result.Changed {
		return
	}

	result.err = tempFile.Close()
	if result.isErr() {
		return
	}

	result.err = inputFile.Close()

	if result.isErr() {
		return
	}

	result.err = os.Rename(tempFile.Name(), filePath)
	if result.isErr() {
		result.err = copyFile(tempFile.Name(), filePath)
		if result.isErr() {
			return
		}
		result.err = os.Remove(tempFile.Name())
		if result.isErr() {
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
