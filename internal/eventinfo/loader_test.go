package eventinfo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "event.md")
	content := "EVENT NAME: Test Event\nDATE: Oct 24, 2026"
	
	err := os.WriteFile(tempFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	loadedContent, err := Load(tempFile)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if loadedContent != content {
		t.Errorf("Expected %q, got %q", content, loadedContent)
	}
}

func TestLoad_NotFound(t *testing.T) {
	_, err := Load("does_not_exist.md")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}
