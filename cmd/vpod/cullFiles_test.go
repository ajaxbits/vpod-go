package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func createTestFile(path string, size int64) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	// Write the specified number of bytes
	if size > 0 {
		_, err = f.Write(make([]byte, size))
		if err != nil {
			return err
		}
	}
	return nil
}

type testFileInfo struct {
	path     string
	size     int64
	modTime  time.Time
	isDir    bool
	children []string
}

func populateTestDir(tempDir string, files []testFileInfo) error {
	for _, f := range files {
		fullPath := filepath.Join(tempDir, f.path)
		if f.isDir {
			if err := os.Mkdir(fullPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %v", fullPath, err)
			}
			// Create nested files
			for i, child := range f.children {
				childPath := filepath.Join(fullPath, child)
				childSize := int64(50 * (i + 1))
				if err := createTestFile(childPath, childSize); err != nil {
					return fmt.Errorf("failed to create child file %s: %v", childPath, err)
				}
			}
		} else {
			if err := createTestFile(fullPath, f.size); err != nil {
				return fmt.Errorf("failed to create file %s: %v", fullPath, err)
			}
			if err := os.Chtimes(fullPath, f.modTime, f.modTime); err != nil {
				return fmt.Errorf("failed to set mod time for %s: %v", fullPath, err)
			}
		}
	}
	return nil
}

// Check if files exist in the directory
func checkFilesExist(dirPath string, fileNames []string) (map[string]bool, error) {
	result := make(map[string]bool)
	for _, fileName := range fileNames {
		result[fileName] = false
	}

	files, _, err := getFilesWithSize(dirPath, ".m4a")
	if err != nil {
		return nil, err
	}

	// Check which files exist
	for _, file := range files {
		baseName := filepath.Base(file.path)
		if _, exists := result[baseName]; exists {
			result[baseName] = true
		}
	}

	return result, nil
}

func Test_cullFiles_TableDriven(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name              string
		files             []testFileInfo
		maxSize           int64
		expectedFileCount int
		shouldExist       []string
		shouldNotExist    []string
	}{
		{
			name: "Cull older files when over limit",
			files: []testFileInfo{
				{
					path:    "oldest.m4a",
					size:    200,
					modTime: time.Now().Add(-3 * time.Hour),
				},
				{
					path:    "middle.m4a",
					size:    200,
					modTime: time.Now().Add(-2 * time.Hour),
				},
				{
					path:    "newest.m4a",
					size:    200,
					modTime: time.Now().Add(-1 * time.Hour),
				},
			},
			maxSize:           500,
			expectedFileCount: 2,
			shouldExist:       []string{"middle.m4a", "newest.m4a"},
			shouldNotExist:    []string{"oldest.m4a"},
		},
		{
			name: "Keep all files when under limit",
			files: []testFileInfo{
				{
					path:    "test1.m4a",
					size:    200,
					modTime: time.Now().Add(-1 * time.Hour),
				},
				{
					path:    "test2.m4a",
					size:    200,
					modTime: time.Now().Add(-2 * time.Hour),
				},
				{
					path:    "test3.m4a",
					size:    99,
					modTime: time.Now().Add(-3 * time.Hour),
				},
			},
			maxSize:           500,
			expectedFileCount: 3,
			shouldExist:       []string{"test1.m4a", "test2.m4a", "test3.m4a"},
			shouldNotExist:    []string{},
		},
		{
			name: "Cull multiple files when significantly over limit",
			files: []testFileInfo{
				{
					path:    "file1.m4a",
					size:    200,
					modTime: time.Now().Add(-5 * time.Hour),
				},
				{
					path:    "file2.m4a",
					size:    200,
					modTime: time.Now().Add(-4 * time.Hour),
				},
				{
					path:    "file3.m4a",
					size:    200,
					modTime: time.Now().Add(-3 * time.Hour),
				},
				{
					path:    "file4.m4a",
					size:    200,
					modTime: time.Now().Add(-2 * time.Hour),
				},
				{
					path:    "file5.m4a",
					size:    200,
					modTime: time.Now().Add(-1 * time.Hour),
				},
			},
			maxSize:           400,
			expectedFileCount: 2,
			shouldExist:       []string{"file4.m4a", "file5.m4a"},
			shouldNotExist:    []string{"file1.m4a", "file2.m4a", "file3.m4a"},
		},
		{
			name: "Handle mixed file sizes correctly",
			files: []testFileInfo{
				{
					path:    "big.m4a",
					size:    300,
					modTime: time.Now().Add(-3 * time.Hour),
				},
				{
					path:    "medium.m4a",
					size:    150,
					modTime: time.Now().Add(-2 * time.Hour),
				},
				{
					path:    "small.m4a",
					size:    50,
					modTime: time.Now().Add(-1 * time.Hour),
				},
			},
			maxSize:           200,
			expectedFileCount: 2,
			shouldExist:       []string{"medium.m4a", "small.m4a"},
			shouldNotExist:    []string{"big.m4a"},
		},
		// {
		// 	name: "Handle directory with nested files",
		// 	files: []testFileInfo{
		// 		{
		// 			path:    "olddir",
		// 			isDir:   true,
		// 			modTime: time.Now().Add(-3 * time.Hour),
		// 			children: []string{
		// 				"nested1.m4a",
		// 				"nested2.m4a",
		// 			},
		// 		},
		// 		{
		// 			path:    "newest.m4a",
		// 			size:    200,
		// 			modTime: time.Now().Add(-1 * time.Hour),
		// 		},
		// 	},
		// 	maxSize:           300,
		// 	expectedFileCount: 2, // We'll have 3 files (1 top-level + 2 nested), but will cull the oldest
		// 	shouldExist:       []string{"newest.m4a"},
		// 	shouldNotExist:    []string{"nested1.m4a", "nested2.m4a"},
		// },
	}

	// Run test cases
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// setup
			words := strings.Fields(tt.name)
			if len(words) == 0 {
				t.Fatal("Test case has to have a non-empty name")
			}
			tmpDirName := strings.Join(words, "_")

			tempDir, err := os.MkdirTemp("", tmpDirName)
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(tempDir)

			err = populateTestDir(tempDir, tt.files)
			if err != nil {
				t.Fatalf("Failed to populate the test dir: %v", err)
			}
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			// run
			err = cullFiles(context.Background(), logger, tempDir, &tt.maxSize)
			if err != nil {
				t.Fatalf("cullFiles failed: %v", err)
			}

			// Check which files remain
			remainingFiles, totalSize, err := getFilesWithSize(tempDir, ".m4a")
			if err != nil {
				t.Fatalf("getFilesWithSize failed: %v", err)
			}

			// We expect the total size to be no more than maxSize
			if totalSize > tt.maxSize {
				t.Errorf("Expected total size to be at most %d, got %d", tt.maxSize, totalSize)
			}

			// Check that we have the right number of files remaining
			if len(remainingFiles) != tt.expectedFileCount {
				t.Errorf("Expected %d files to remain, got %d", tt.expectedFileCount, len(remainingFiles))
			}

			// Check for files that should exist
			fileExistence, err := checkFilesExist(tempDir, append(tt.shouldExist, tt.shouldNotExist...))
			if err != nil {
				t.Fatalf("Failed to check file existence: %v", err)
			}

			// Verify files that should exist
			for _, fileName := range tt.shouldExist {
				exists, found := fileExistence[fileName]
				if !found {
					t.Errorf("File %s was not checked", fileName)
				} else if !exists {
					t.Errorf("Expected %s to exist, but it was removed", fileName)
				}
			}

			// Verify files that should not exist
			for _, fileName := range tt.shouldNotExist {
				exists, found := fileExistence[fileName]
				if !found {
					t.Errorf("File %s was not checked", fileName)
				} else if exists {
					t.Errorf("Expected %s to be removed, but it still exists", fileName)
				}
			}
		})
	}
}
