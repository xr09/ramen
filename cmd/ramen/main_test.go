package main

import (
	"testing"
)

func TestParseSize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
		wantErr  bool
	}{
		{"empty", "", 0, true},
		{"invalid", "abc", 0, true},
		{"negative", "-10", 0, true},
		{"zero", "0", 0, true},
		{"just number", "100", 100, false},
		{"MB lowercase", "100m", 100, false},
		{"MB uppercase", "100M", 100, false},
		{"MB explicitly", "100MB", 100, false},
		{"GB lowercase", "1g", 1024, false},
		{"GB uppercase", "2G", 2048, false},
		{"GB explicitly", "1GB", 1024, false},
		{"whitespaces", "  3g  ", 3072, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSize(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("parseSize() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAllocate(t *testing.T) {
	tests := []struct {
		name   string
		sizeMB int
	}{
		{"1MB", 1},
		{"5MB", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mem := allocate(tt.sizeMB)
			expectedSize := tt.sizeMB * 1024 * 1024
			if len(mem) != expectedSize {
				t.Errorf("allocate(%d) = slice of size %d, want %d", tt.sizeMB, len(mem), expectedSize)
			}

			// Verify some pages are touched
			if len(mem) > 0 {
				if mem[0] != 1 {
					t.Errorf("expected first byte to be 1, got %d", mem[0])
				}
			}
		})
	}
}

func TestRunInvalidSize(t *testing.T) {
	err := run("0")
	if err == nil {
		t.Error("expected error for size 0, got nil")
	}

	err = run("-1m")
	if err == nil {
		t.Error("expected error for size -1m, got nil")
	}
}
