package model_hub

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const msCacheRootEnv = "NEXA_MS_CACHE_DIR"

func msBaseCacheDir() (string, error) {
	if custom := os.Getenv(msCacheRootEnv); custom != "" {
		return custom, nil
	}
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "nexa.ai", "modelscope_cache"), nil
}

func msLocalCacheDir(modelName string) (string, error) {
	base, err := msBaseCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, filepath.FromSlash(modelName)), nil
}

type ModelScope struct{}

func NewModelScope() *ModelScope { return &ModelScope{} }

func (m *ModelScope) CheckAvailable(ctx context.Context, modelName string) error {
	if _, err := exec.LookPath("python3"); err != nil {
		return fmt.Errorf("python3 not found: %w", err)
	}
	return nil
}

func (m *ModelScope) MaxConcurrency() int { return 4 }

func (m *ModelScope) ensureDownloaded(ctx context.Context, modelName string) (string, error) {
	finalCacheDir, err := msLocalCacheDir(modelName)
	if err != nil {
		return "", err
	}

	// quick existence check: if directory exists and not empty, skip
	if info, err := os.Stat(finalCacheDir); err == nil && info.IsDir() {
		var nonEmpty bool
		_ = filepath.WalkDir(finalCacheDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				nonEmpty = true
				return errors.New("done")
			}
			return nil
		})
		if nonEmpty {
			return finalCacheDir, nil
		}
	}

	baseCacheDir, err := msBaseCacheDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(baseCacheDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create base cache dir %s: %w", baseCacheDir, err)
	}

	py := strings.Join([]string{
		"from modelscope.hub.snapshot_download import snapshot_download",
		fmt.Sprintf("snapshot_download('%s', cache_dir=r'%s')", modelName, baseCacheDir),
	}, ";\n")

	cmd := exec.CommandContext(ctx, "python3", "-c", py)
	cmd.Env = os.Environ()
	stderr, _ := cmd.StderrPipe()
	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		return "", err
	}
	// stream logs (debug)
	go func() {
		s := bufio.NewScanner(stdout)
		for s.Scan() {
			slog.Debug("ms sdk", "out", s.Text())
		}
	}()
	go func() {
		s := bufio.NewScanner(stderr)
		for s.Scan() {
			slog.Warn("ms sdk", "err", s.Text())
		}
	}()
	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("modelscope sdk download failed: %w", err)
	}

	return finalCacheDir, nil
}

func (m *ModelScope) ModelInfo(ctx context.Context, modelName string) ([]ModelFileInfo, error) {
	cacheDir, err := m.ensureDownloaded(ctx, modelName)
	if err != nil {
		return nil, err
	}
	res := make([]ModelFileInfo, 0)
	err = filepath.WalkDir(cacheDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(cacheDir, path)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		res = append(res, ModelFileInfo{Name: filepath.ToSlash(rel), Size: info.Size()})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (m *ModelScope) GetFileContent(ctx context.Context, modelName, fileName string, offset, limit int64, writer io.Writer) error {
	cacheDir, err := m.ensureDownloaded(ctx, modelName)
	if err != nil {
		return err
	}

	filePath := filepath.Join(cacheDir, filepath.FromSlash(fileName))
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	if offset > 0 {
		if _, err := f.Seek(offset, io.SeekStart); err != nil {
			return err
		}
	}
	var r io.Reader = f
	if limit > 0 {
		r = io.LimitReader(f, limit)
	}
	_, err = io.Copy(writer, r)
	return err
}
