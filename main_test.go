package main

import (
	"os"
	"testing"

	"github.com/marzdrel/sawmill/processor"
)

func TestProcessFile(t *testing.T) {
	content := "  line with spaces   \nsecond line\t\nthird line"

	testFile := "test_temp.txt"
	os.WriteFile(testFile, []byte(content), 0o644)
	defer os.Remove(testFile)

	processor.ProcessFile(testFile)
	result, _ := os.ReadFile(testFile)

	expected := "  line with spaces\nsecond line\nthird line\n"
	if string(result) != expected {
		t.Errorf("Expected %q, got %q", expected, string(result))
	}
}

func TestProcessFileNoChanges(t *testing.T) {
	content := "already clean line\nsecond clean line\nthird clean line\n"

	testFile := "test_no_changes.txt"
	os.WriteFile(testFile, []byte(content), 0o644)
	defer os.Remove(testFile)

	stat1, _ := os.Stat(testFile)
	processor.ProcessFile(testFile)
	stat2, _ := os.Stat(testFile)

	result, _ := os.ReadFile(testFile)
	if string(result) != content {
		t.Errorf("File content changed when it shouldn't have")
	}
	if stat1.ModTime() != stat2.ModTime() {
		t.Errorf("File modification time changed when content was unchanged")
	}
}
