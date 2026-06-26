package tool

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFileReader_Read_Workspace(t *testing.T) {
	dir := t.TempDir()
	content := "line1\nline2\nline3\n"
	if err := os.WriteFile(filepath.Join(dir, "test.go"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	fr := &FileReader{RepoDir: dir, Mode: ModeWorkspace}
	got, err := fr.Read(context.Background(), "test.go")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if got != content {
		t.Errorf("Read() = %q, want %q", got, content)
	}
}

func TestFileReader_Read_WorkspaceNotFound(t *testing.T) {
	dir := t.TempDir()
	fr := &FileReader{RepoDir: dir, Mode: ModeWorkspace}
	_, err := fr.Read(context.Background(), "missing.go")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestFileReader_Read_PathTraversal(t *testing.T) {
	dir := t.TempDir()
	fr := &FileReader{RepoDir: dir, Mode: ModeWorkspace}

	_, err := fr.Read(context.Background(), "../../../etc/passwd")
	if err == nil {
		t.Error("expected error for path traversal")
	}
}

func TestFileReader_Read_SymlinkOutsideRepo(t *testing.T) {
	dir := t.TempDir()
	outside := t.TempDir()
	secretFile := filepath.Join(outside, "secret.txt")
	if err := os.WriteFile(secretFile, []byte("sensitive"), 0644); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(dir, "link.txt")
	if err := os.Symlink(secretFile, link); err != nil {
		t.Skipf("symlinks not supported: %v", err)
	}

	fr := &FileReader{RepoDir: dir, Mode: ModeWorkspace}
	_, err := fr.Read(context.Background(), "link.txt")
	if err == nil {
		t.Error("expected error for symlink pointing outside repo")
	}
}

func TestFileReader_ReadLines_Workspace(t *testing.T) {
	dir := t.TempDir()
	content := "aaa\nbbb\nccc\nddd\n"
	if err := os.WriteFile(filepath.Join(dir, "lines.txt"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	fr := &FileReader{RepoDir: dir, Mode: ModeWorkspace}

	t.Run("all lines", func(t *testing.T) {
		lines, total, err := fr.ReadLines(context.Background(), "lines.txt", 1, 100)
		if err != nil {
			t.Fatal(err)
		}
		if total != 5 {
			t.Errorf("total = %d, want 5", total)
		}
		if len(lines) != 5 {
			t.Errorf("lines count = %d, want 5", len(lines))
		}
	})

	t.Run("start from line 2 with limit", func(t *testing.T) {
		lines, total, err := fr.ReadLines(context.Background(), "lines.txt", 2, 2)
		if err != nil {
			t.Fatal(err)
		}
		if total != 5 {
			t.Errorf("total = %d, want 5", total)
		}
		if len(lines) != 2 {
			t.Fatalf("lines count = %d, want 2", len(lines))
		}
		if lines[0] != "bbb" || lines[1] != "ccc" {
			t.Errorf("lines = %v, want [bbb ccc]", lines)
		}
	})

	t.Run("path traversal rejected", func(t *testing.T) {
		_, _, err := fr.ReadLines(context.Background(), "../../etc/passwd", 1, 10)
		if err == nil {
			t.Error("expected error for path traversal")
		}
	})
}

func TestFileReader_Read_SubdirectoryFile(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "src", "pkg")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	fr := &FileReader{RepoDir: dir, Mode: ModeWorkspace}
	got, err := fr.Read(context.Background(), "src/pkg/main.go")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if got != "package main" {
		t.Errorf("Read() = %q, want %q", got, "package main")
	}
}
