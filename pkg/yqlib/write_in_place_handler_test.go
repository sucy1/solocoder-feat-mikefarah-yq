package yqlib

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteInPlaceHandlerImpl_CreateTempFile(t *testing.T) {
	// Create a temporary directory and file for testing
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "input.yaml")

	// Create input file with some content
	content := []byte("test: value\n")
	err := os.WriteFile(inputFile, content, 0600)
	if err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	handler := NewWriteInPlaceHandler(inputFile, false)
	tempFile, err := handler.CreateTempFile()

	if err != nil {
		t.Fatalf("CreateTempFile failed: %v", err)
	}

	if tempFile == nil {
		t.Fatal("CreateTempFile returned nil file")
	}

	// Clean up
	tempFile.Close()
	os.Remove(tempFile.Name())
}

func TestWriteInPlaceHandlerImpl_CreateTempFile_NonExistentInput(t *testing.T) {
	// Test with non-existent input file
	handler := NewWriteInPlaceHandler("/non/existent/file.yaml", false)
	tempFile, err := handler.CreateTempFile()

	if err == nil {
		t.Error("Expected error for non-existent input file, got nil")
	}

	if tempFile != nil {
		t.Error("Expected nil temp file for non-existent input file")
		tempFile.Close()
	}
}

func TestWriteInPlaceHandlerImpl_FinishWriteInPlace_Success(t *testing.T) {
	// Create a temporary directory and file for testing
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "input.yaml")

	// Create input file with some content
	content := []byte("test: value\n")
	err := os.WriteFile(inputFile, content, 0600)
	if err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	handler := NewWriteInPlaceHandler(inputFile, false)
	tempFile, err := handler.CreateTempFile()
	if err != nil {
		t.Fatalf("CreateTempFile failed: %v", err)
	}
	defer tempFile.Close()

	// Write some content to temp file
	tempContent := []byte("updated: content\n")
	_, err = tempFile.Write(tempContent)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// Test successful finish
	err = handler.FinishWriteInPlace(true)
	if err != nil {
		t.Fatalf("FinishWriteInPlace failed: %v", err)
	}

	// Verify the original file was updated
	updatedContent, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	if string(updatedContent) != string(tempContent) {
		t.Errorf("File content not updated correctly. Expected %q, got %q",
			string(tempContent), string(updatedContent))
	}
}

func TestWriteInPlaceHandlerImpl_FinishWriteInPlace_Failure(t *testing.T) {
	// Create a temporary directory and file for testing
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "input.yaml")

	// Create input file with some content
	content := []byte("test: value\n")
	err := os.WriteFile(inputFile, content, 0600)
	if err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	handler := NewWriteInPlaceHandler(inputFile, false)
	tempFile, err := handler.CreateTempFile()
	if err != nil {
		t.Fatalf("CreateTempFile failed: %v", err)
	}
	defer tempFile.Close()

	// Write some content to temp file
	tempContent := []byte("updated: content\n")
	_, err = tempFile.Write(tempContent)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// Test failure finish (should not update the original file)
	err = handler.FinishWriteInPlace(false)
	if err != nil {
		t.Fatalf("FinishWriteInPlace failed: %v", err)
	}

	// Verify the original file was NOT updated
	originalContent, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatalf("Failed to read original file: %v", err)
	}

	if string(originalContent) != string(content) {
		t.Errorf("File content should not have been updated. Expected %q, got %q",
			string(content), string(originalContent))
	}
}

func TestWriteInPlaceHandlerImpl_FinishWriteInPlace_Symlink_Success(t *testing.T) {
	// Create a temporary directory and file for testing
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "input.yaml")
	symlinkFile := filepath.Join(tempDir, "symlink.yaml")

	// Create input file with some content
	content := []byte("test: value\n")
	err := os.WriteFile(inputFile, content, 0600)
	if err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	err = os.Symlink(inputFile, symlinkFile)
	if err != nil {
		t.Fatalf("Failed to symlink to input file: %v", err)
	}

	handler := NewWriteInPlaceHandler(symlinkFile, false)
	tempFile, err := handler.CreateTempFile()
	if err != nil {
		t.Fatalf("CreateTempFile failed: %v", err)
	}
	defer tempFile.Close()

	// Write some content to temp file
	tempContent := []byte("updated: content\n")
	_, err = tempFile.Write(tempContent)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// Test successful finish
	err = handler.FinishWriteInPlace(true)
	if err != nil {
		t.Fatalf("FinishWriteInPlace failed: %v", err)
	}

	// Verify that the symlink is still present
	info, err := os.Lstat(symlinkFile)
	if err != nil {
		t.Fatalf("Failed to lstat input file: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Errorf("Input file symlink is no longer present")
	}

	// Verify the original file was updated
	updatedContent, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	if string(updatedContent) != string(tempContent) {
		t.Errorf("File content not updated correctly. Expected %q, got %q",
			string(tempContent), string(updatedContent))
	}
}

func TestWriteInPlaceHandlerImpl_CreateTempFile_Permissions(t *testing.T) {
	// Create a temporary directory and file for testing
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "input.yaml")

	// Create input file with specific permissions
	content := []byte("test: value\n")
	err := os.WriteFile(inputFile, content, 0600)
	if err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	handler := NewWriteInPlaceHandler(inputFile, false)
	tempFile, err := handler.CreateTempFile()
	if err != nil {
		t.Fatalf("CreateTempFile failed: %v", err)
	}
	defer tempFile.Close()

	// Check that temp file has same permissions as input file
	tempFileInfo, err := os.Stat(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to stat temp file: %v", err)
	}

	inputFileInfo, err := os.Stat(inputFile)
	if err != nil {
		t.Fatalf("Failed to stat input file: %v", err)
	}

	if tempFileInfo.Mode() != inputFileInfo.Mode() {
		t.Errorf("Temp file permissions don't match input file. Expected %v, got %v",
			inputFileInfo.Mode(), tempFileInfo.Mode())
	}
}

func TestWriteInPlaceHandlerImpl_Integration(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "integration_test.yaml")

	originalContent := []byte("original: content\n")
	err := os.WriteFile(inputFile, originalContent, 0600)
	if err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	handler := NewWriteInPlaceHandler(inputFile, false)

	tempFile, err := handler.CreateTempFile()
	if err != nil {
		t.Fatalf("CreateTempFile failed: %v", err)
	}

	newContent := []byte("new: content\n")
	_, err = tempFile.Write(newContent)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	err = handler.FinishWriteInPlace(true)
	if err != nil {
		t.Fatalf("FinishWriteInPlace failed: %v", err)
	}

	finalContent, err := os.ReadFile(inputFile)
	if err != nil {
		t.Fatalf("Failed to read final file: %v", err)
	}

	if string(finalContent) != string(newContent) {
		t.Errorf("File not updated correctly. Expected %q, got %q",
			string(newContent), string(finalContent))
	}
}

func TestWriteInPlaceHandlerImpl_BackupIntegrity(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "backup_test.yaml")

	originalContent := []byte("original: content\n")
	err := os.WriteFile(inputFile, originalContent, 0600)
	if err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	handler := NewWriteInPlaceHandler(inputFile, false)
	handlerImpl := handler.(*writeInPlaceHandlerImpl)

	tempFile, err := handler.CreateTempFile()
	if err != nil {
		t.Fatalf("CreateTempFile failed: %v", err)
	}
	tempFile.Close()

	err = handlerImpl.createBackup()
	if err != nil {
		t.Fatalf("createBackup failed: %v", err)
	}

	if handlerImpl.backupSum == "" {
		t.Fatal("Expected backupSum to be set after createBackup")
	}

	backupPath := inputFile + ".bak"
	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}
	if string(backupData) != string(originalContent) {
		t.Errorf("Backup content mismatch. Expected %q, got %q",
			string(originalContent), string(backupData))
	}

	err = handlerImpl.restoreBackup()
	if err != nil {
		t.Fatalf("restoreBackup failed: %v", err)
	}
}

func TestWriteInPlaceHandlerImpl_BackupIntegrityTampered(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "tampered_test.yaml")

	originalContent := []byte("original: content\n")
	err := os.WriteFile(inputFile, originalContent, 0600)
	if err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	handler := NewWriteInPlaceHandler(inputFile, false)
	handlerImpl := handler.(*writeInPlaceHandlerImpl)

	tempFile, err := handler.CreateTempFile()
	if err != nil {
		t.Fatalf("CreateTempFile failed: %v", err)
	}
	tempFile.Close()

	err = handlerImpl.createBackup()
	if err != nil {
		t.Fatalf("createBackup failed: %v", err)
	}

	backupPath := inputFile + ".bak"
	err = os.WriteFile(backupPath, []byte("tampered content\n"), 0600)
	if err != nil {
		t.Fatalf("Failed to tamper backup file: %v", err)
	}

	err = handlerImpl.restoreBackup()
	if err == nil {
		t.Fatal("Expected error when restoring tampered backup, got nil")
	}
}

func TestWriteInPlaceHandlerImpl_NoBackup(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "nobackup_test.yaml")

	originalContent := []byte("original: content\n")
	err := os.WriteFile(inputFile, originalContent, 0600)
	if err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	handler := NewWriteInPlaceHandler(inputFile, true)

	tempFile, err := handler.CreateTempFile()
	if err != nil {
		t.Fatalf("CreateTempFile failed: %v", err)
	}
	tempFile.Close()

	err = handler.FinishWriteInPlace(true)
	if err != nil {
		t.Fatalf("FinishWriteInPlace failed: %v", err)
	}

	backupPath := inputFile + ".bak"
	if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
		t.Error("Expected no backup file to be created with noBackup=true")
	}
}
