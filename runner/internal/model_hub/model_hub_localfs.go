package model_hub

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

type LocalFS struct {
	basePath string
}

func NewLocalFS(base string) *LocalFS {
	return &LocalFS{base}
}

func (d *LocalFS) ChinaMainlandOnly() bool {
	return false
}

func (d *LocalFS) MaxConcurrency() int {
	return 4
}

func (d *LocalFS) CheckAvailable(ctx context.Context, name string) error {
	// check is directory exists
	info, err := os.Stat(d.basePath)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return os.ErrNotExist
	}
	return nil
}

func (d *LocalFS) ModelInfo(ctx context.Context, name string) ([]ModelFileInfo, error) {
	res := make([]ModelFileInfo, 0)

	// recursive list files in basePath
	err := filepath.Walk(d.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, err := filepath.Rel(d.basePath, path)
			if err != nil {
				return err
			}
			res = append(res, ModelFileInfo{
				Name: relPath,
				Size: info.Size(),
			})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (d *LocalFS) GetFileContent(ctx context.Context, modelName, fileName string, offset, limit int64, writer io.Writer) error {
	file, err := os.Open(filepath.Join(d.basePath, fileName))
	if err != nil {
		return err
	}
	defer file.Close()

	// seek to offset
	if offset > 0 {
		_, err = file.Seek(offset, io.SeekStart)
		if err != nil {
			return err
		}
	}

	var reader io.Reader = file
	if limit > 0 {
		reader = io.LimitReader(file, limit)
	}

	_, err = io.Copy(writer, reader)
	if err != nil {
		return err
	}

	return nil
}
