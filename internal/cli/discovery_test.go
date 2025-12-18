package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetCommonLocations(t *testing.T) {
	locations := getCommonLocations()

	if len(locations) == 0 {
		t.Error("expected non-empty locations list")
	}

	for _, loc := range locations {
		if loc == "" {
			t.Error("found empty location in list")
		}
	}
}

func TestGetCommonLocations_ContainsExpectedPaths(t *testing.T) {
	locations := getCommonLocations()

	if runtime.GOOS == "windows" {
		found := false
		for _, loc := range locations {
			if filepath.Base(loc) == "claude.cmd" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected to find claude.cmd in Windows locations")
		}
	} else {
		foundLocal := false
		foundNpm := false
		for _, loc := range locations {
			if filepath.Base(loc) == "claude" {
				if filepath.Base(filepath.Dir(loc)) == "bin" {
					foundLocal = true
				}
			}
			if strings.Contains(loc, ".npm-global") {
				foundNpm = true
			}
		}
		if !foundLocal {
			t.Error("expected to find claude in bin directory")
		}
		if !foundNpm {
			t.Error("expected to find npm-global path")
		}
	}
}

func TestValidateWorkingDirectory_EmptyPath(t *testing.T) {
	err := ValidateWorkingDirectory("")
	if err != nil {
		t.Errorf("empty path should return nil, got: %v", err)
	}
}

func TestValidateWorkingDirectory_ValidPath(t *testing.T) {
	tmpDir := t.TempDir()
	err := ValidateWorkingDirectory(tmpDir)
	if err != nil {
		t.Errorf("valid directory should return nil, got: %v", err)
	}
}

func TestValidateWorkingDirectory_NonexistentPath(t *testing.T) {
	err := ValidateWorkingDirectory("/nonexistent/path/12345")
	if err == nil {
		t.Error("nonexistent path should return error")
	}
}

func TestValidateWorkingDirectory_FileNotDir(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	err = ValidateWorkingDirectory(tmpFile.Name())
	if err == nil {
		t.Error("file path should return error")
	}
}
