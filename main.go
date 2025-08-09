package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	gitignore "github.com/denormal/go-gitignore"
	"github.com/marzdrel/sawmill/processor"
)

const concurrency = 30

var version = "dev"

func getVersion() string {
	if version != "dev" {
		return version
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}

	return "dev"
}

var defaultPatterns = []string{
	"*.go", "*.js", "*.ts", "*.jsx", "*.tsx",
	"*.py", ".rb", "*.rs",
	"*.toml", "*.yml", "*.yaml", "*.json", "*.xml",
	"*.html", "*.css", "*.scss", "*.md",
	"*.txt", "*.conf", "*.ini", "*.sh",
	"*.tf", "Dockerfile.*",
}

type runStats struct {
	FilesProcessed int
	FilesChanged   int
	startTime      time.Time
	verbose        bool
}

func (s *runStats) Log(template string, args ...any) {
	if s.verbose {
		fmt.Printf(template, args...)
	}
}

func (s *runStats) duration() time.Duration {
	duration := time.Since(s.startTime)

	if duration > time.Second {
		return duration.Round(10 * time.Millisecond)
	} else {
		return duration.Round(10 * time.Microsecond)
	}
}

func (s *runStats) Summary() string {
	return fmt.Sprintf(
		"Processed %d files, changed %d files in %s.",
		s.FilesProcessed,
		s.FilesChanged,
		s.duration(),
	)
}

func main() {
	var extensions []string
	var stats runStats

	patternFlag := flag.String("pattern", "",
		"Comma-separated list of file patterns to process")

	verboseFlag := flag.Bool("verbose", false,
		"Enable verbose output")

	ignoreGitignoreFlag := flag.Bool("u", false,
		"Ignore gitignore entries and process all matching files")

	versionFlag := flag.Bool("version", false,
		"Show version information")

	concurrencyFlag := flag.Int("workers", concurrency,
		"Number of concurrent workers to use")

	flag.Parse()

	if *versionFlag {
		fmt.Printf("sawmill version %s\n", getVersion())
		os.Exit(0)
	}

	stats.verbose = *verboseFlag
	stats.startTime = time.Now()

	pattern := *patternFlag
	ignoreGitignore := *ignoreGitignoreFlag

	extensions = defaultPatterns
	if len(pattern) > 0 {
		extensions = strings.Split(pattern, ",")
	}

	root := "."

	var singleFile string
	if flag.NArg() > 0 {
		arg := flag.Arg(0)
		if info, err := os.Stat(arg); err == nil && !info.IsDir() {
			singleFile = arg
		}
	}

	var gi gitignore.GitIgnore
	if _, err := os.Stat(".gitignore"); err == nil {
		if parsedGi, parseErr := gitignore.NewFromFile(".gitignore"); parseErr == nil {
			gi = parsedGi
		}
	}

	jobs := make(chan string)
	results := make(chan processor.Result)
	var wg sync.WaitGroup

	for range *concurrencyFlag {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for job := range jobs {
				result := processor.ProcessFile(job)
				results <- result
			}
		}()
	}

	go func() {
		defer close(jobs)

		if singleFile != "" {
			stats.FilesProcessed++
			stats.Log("Processing: %s\n", singleFile)
			jobs <- singleFile
			return
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

				if gi != nil && !ignoreGitignore {
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

					jobs <- path
				}

				return nil
			})
		if err != nil {
			fmt.Printf("Error walking directory: %v\n", err)
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		if result.Err() != nil {
			fmt.Printf("Error processing file %s: %v\n", result.Path, result.Err())
			continue
		}

		if result.Changed {
			stats.FilesChanged++
			stats.Log("Changed: %s\n", result.Path)
		}
	}

	fmt.Println(stats.Summary())
}
