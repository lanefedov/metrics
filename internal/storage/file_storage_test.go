package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestMemStorageSaveAndLoadFile(t *testing.T) {
	t.Parallel()

	store := NewMemStorage()
	ctx := context.Background()
	_ = store.AddCounter(ctx, "requests", 42)
	_ = store.SetGauge(ctx, "alloc", 123.45)

	path := filepath.Join(t.TempDir(), "metrics.json")
	if err := store.SaveToFile(path); err != nil {
		t.Fatalf("save metrics: %v", err)
	}

	restored := NewMemStorage()
	if err := restored.LoadFromFile(path); err != nil {
		t.Fatalf("load metrics: %v", err)
	}

	counter, ok, err := restored.GetCounter(ctx, "requests")
	if err != nil {
		t.Fatalf("get counter: %v", err)
	}
	if !ok || counter != 42 {
		t.Fatalf("counter: got (%d, %t), want (42, true)", counter, ok)
	}

	gauge, ok, err := restored.GetGauge(ctx, "alloc")
	if err != nil {
		t.Fatalf("get gauge: %v", err)
	}
	if !ok || gauge != 123.45 {
		t.Fatalf("gauge: got (%v, %t), want (123.45, true)", gauge, ok)
	}
}

func TestMemStorageLoadMissingFile(t *testing.T) {
	t.Parallel()

	store := NewMemStorage()
	path := filepath.Join(t.TempDir(), "missing.json")
	if err := store.LoadFromFile(path); err != nil {
		t.Fatalf("load missing file: %v", err)
	}
}

func TestMemStorageLoadEmptyFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "metrics.json")
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatalf("write empty file: %v", err)
	}

	store := NewMemStorage()
	if err := store.LoadFromFile(path); err != nil {
		t.Fatalf("load empty file: %v", err)
	}
}
