package processor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessFileStreaming(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		changed  bool
	}{
		{
			name:     "no changes needed",
			input:    "hello world\n",
			expected: "hello world\n",
			changed:  false,
		},
		{
			name:     "trailing whitespace removal",
			input:    "hello world   \n",
			expected: "hello world\n",
			changed:  true,
		},
		{
			name:     "trailing tabs removal",
			input:    "hello world\t\t\n",
			expected: "hello world\n",
			changed:  true,
		},
		{
			name:     "mixed trailing whitespace",
			input:    "hello world \t \n",
			expected: "hello world\n",
			changed:  true,
		},
		{
			name:     "empty lines between content",
			input:    "line1\n\n\nline2\n",
			expected: "line1\n\n\nline2\n",
			changed:  false,
		},
		{
			name:     "trailing empty lines removed",
			input:    "content\n\n\n",
			expected: "content\n",
			changed:  true,
		},
		{
			name:     "empty file",
			input:    "",
			expected: "",
			changed:  false,
		},
		{
			name:     "only empty lines",
			input:    "\n\n\n",
			expected: "",
			changed:  true,
		},
		{
			name:     "no final newline added",
			input:    "content",
			expected: "content\n",
			changed:  false,
		},
		{
			name:     "multiple lines with trailing whitespace",
			input:    "line1  \nline2\t\nline3\n",
			expected: "line1\nline2\nline3\n",
			changed:  true,
		},
		{
			name:     "preserve empty lines in middle",
			input:    "start\n\n\nmiddle\n\n\nend\n",
			expected: "start\n\n\nmiddle\n\n\nend\n",
			changed:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.NewReader(tt.input)
			var output strings.Builder

			changed, err := processFileStreaming(input, &output)
			if err != nil {
				t.Fatalf("processFileStreaming returned error: %v", err)
			}

			if changed != tt.changed {
				t.Errorf("expected changed=%v, got changed=%v", tt.changed, changed)
			}

			if output.String() != tt.expected {
				t.Errorf("expected output %q, got %q", tt.expected, output.String())
			}
		})
	}
}

func TestProcessFile(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name         string
		content      string
		expected     string
		expectChange bool
	}{
		{
			name:         "file with trailing whitespace",
			content:      "hello world   \ntest line\t\n",
			expected:     "hello world\ntest line\n",
			expectChange: true,
		},
		{
			name:         "file with no changes needed",
			content:      "hello world\ntest line\n",
			expected:     "hello world\ntest line\n",
			expectChange: false,
		},
		{
			name:         "empty file",
			content:      "",
			expected:     "",
			expectChange: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tempDir, "test_"+tt.name+".txt")

			err := os.WriteFile(filePath, []byte(tt.content), 0o644)
			if err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			result := ProcessFile(filePath)
			if result.IsErr() {
				t.Fatalf("ProcessFile returned error: %v", result.Err())
			}

			if result.Changed != tt.expectChange {
				t.Errorf("expected Changed=%v, got Changed=%v", tt.expectChange, result.Changed)
			}

			actualContent, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("failed to read processed file: %v", err)
			}

			if string(actualContent) != tt.expected {
				t.Errorf("expected file content %q, got %q", tt.expected, string(actualContent))
			}
		})
	}
}

func TestProcessFileNonExistent(t *testing.T) {
	result := ProcessFile("/non/existent/file.txt")
	if !result.IsErr() {
		t.Error("expected error for non-existent file")
	}
}

func TestResultMethods(t *testing.T) {
	t.Run("result with error", func(t *testing.T) {
		result := makeResult(false, os.ErrNotExist)
		if !result.IsErr() {
			t.Error("expected IsErr() to return true")
		}
		if result.Err() != os.ErrNotExist {
			t.Errorf("expected Err() to return %v, got %v", os.ErrNotExist, result.Err())
		}
	})

	t.Run("result without error", func(t *testing.T) {
		result := makeResult(true, nil)
		if result.IsErr() {
			t.Error("expected IsErr() to return false")
		}
		if result.Err() != nil {
			t.Errorf("expected Err() to return nil, got %v", result.Err())
		}
		if !result.Changed {
			t.Error("expected Changed to be true")
		}
	})
}

func TestCopyFile(t *testing.T) {
	tempDir := t.TempDir()

	srcPath := filepath.Join(tempDir, "source.txt")
	dstPath := filepath.Join(tempDir, "dest.txt")
	content := "test content for copy"

	err := os.WriteFile(srcPath, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	err = copyFile(srcPath, dstPath)
	if err != nil {
		t.Fatalf("copyFile returned error: %v", err)
	}

	copiedContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("failed to read copied file: %v", err)
	}

	if string(copiedContent) != content {
		t.Errorf("expected copied content %q, got %q", content, string(copiedContent))
	}
}

func TestCopyFileNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	dstPath := filepath.Join(tempDir, "dest.txt")

	err := copyFile("/non/existent/file.txt", dstPath)
	if err == nil {
		t.Error("expected error when copying non-existent file")
	}
}
