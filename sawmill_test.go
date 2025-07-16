package main

import (
	"os"
	"testing"
)

func TestProcessFile(t *testing.T) {
	content := "line with spaces   \nsecond line\t\nthird line"

	testFile := "test_temp.txt"
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(testFile)

	err = processFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	result, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}

	expected := "line with spaces\nsecond line\nthird line\n\n"
	if string(result) != expected {
		t.Errorf("Expected %q, got %q", expected, string(result))
	}
}
