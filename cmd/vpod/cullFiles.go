package main

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"time"
)

const (
	KB int64 = 1024
	MB       = KB * 1024
	GB       = MB * 1024
)

const (
	audio_storage_path     = "./"
	max_audio_storage_size = 1 * GB
)

func cullFiles(ctx context.Context, logger *slog.Logger) error {
	files, totalSize, err := getFilesWithSize(audio_storage_path, ".m4a")
	if err != nil {
		return err
	}

	if totalSize > max_audio_storage_size {
		logger = logger.With(slog.String(
			"desired_size_bytes",
			strconv.Itoa(int(max_audio_storage_size)),
		))

		logger.Info("audio size bigger than desired -- culling", slog.String(
			"current_size_bytes",
			strconv.Itoa(int(totalSize)),
		))
		slices.SortFunc(files, func(a file, b file) int {
			duration := a.modTime.UTC().Sub(b.modTime.UTC())
			return int(duration)
		})

		remainingSize := totalSize
		for _, file := range files {
			if remainingSize <= max_audio_storage_size {
				break
			}

			err = os.Remove(file.path)
			if err != nil {
				return err
			}
			remainingSize = remainingSize - file.sizeBytes

			logger.Debug(
				"removed file",
				slog.String("path", file.path),
				slog.String(
					"current_size_bytes",
					strconv.Itoa(int(remainingSize)),
				),
			)
		}
		logger.Info(
			"culled excess audio files",
			slog.String(
				"current_size_bytes",
				strconv.Itoa(int(remainingSize)),
			),
		)
	}
	return nil
}

type file struct {
	modTime   time.Time
	path      string
	sizeBytes int64
}

func getFilesWithSize(path string, ext string) ([]file, int64, error) {
	var files []file
	var mu sync.Mutex
	var totalSizeBytes int64

	var calculateSize func(string) error
	calculateSize = func(path string) error {
		fileInfo, err := os.Lstat(path)
		if err != nil {
			return err
		}

		// Skip symbolic links to avoid counting them multiple times
		if fileInfo.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		if fileInfo.IsDir() {
			entries, err := os.ReadDir(path)
			if err != nil {
				return err
			}
			for _, entry := range entries {
				if err := calculateSize(filepath.Join(path, entry.Name())); err != nil {
					return err
				}
			}
		} else {
			if filepath.Ext(fileInfo.Name()) == ext {
				files = append(files, file{
					path:      path,
					modTime:   fileInfo.ModTime(),
					sizeBytes: fileInfo.Size(),
				})
				mu.Lock()
				totalSizeBytes += fileInfo.Size()
				mu.Unlock()
			}
		}
		return nil
	}

	// Start calculation from the root path
	if err := calculateSize(path); err != nil {
		return []file{}, 0, err
	}

	return files, totalSizeBytes, nil
}
