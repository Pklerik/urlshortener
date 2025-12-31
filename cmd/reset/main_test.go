package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func Test_doGReset(t *testing.T) {
	curDir, _ := os.Getwd()
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		path    string
		wantErr bool
	}{
		{name: "base reset", path: filepath.Join(curDir, "test_folder", "main.go"), wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := doGReset(context.Background(), tt.path)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("doGReset() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("doGReset() succeeded unexpectedly")
			}
			if _, err := os.Stat(filepath.Join(filepath.Dir(tt.path), "main.gen.go")); err == nil {
				// path/to/whatever exists

			} else if errors.Is(err, os.ErrNotExist) {
				t.Fatal("doGReset() do not create main.gen.go")
			}
		})
	}
}
