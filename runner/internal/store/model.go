// Copyright (c) 2025 Nexa AI
//
// LICENSE NOTICE - DUAL LICENSING:
// - NPU models and inference: CC-BY-NC 4.0 (NON-COMMERCIAL USE ONLY)
// - GPU/CPU models and inference: Apache 2.0 (FREE FOR ALL USE)

package store

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/bytedance/sonic"

	"github.com/NexaAI/nexa-sdk/runner/internal/model_hub"
	"github.com/NexaAI/nexa-sdk/runner/internal/types"
)

// List returns all locally stored models by reading their manifest files
func (s *Store) List() ([]types.ModelManifest, error) {
	res := make([]types.ModelManifest, 0)
	models, err := s.scanModelDir()
	if err != nil {
		return nil, err
	}
	for _, model := range models {
		// Parse each model directory's manifest
		model, err := s.GetManifest(model)
		if err != nil {
			slog.Warn("GetManifest Error", "err", err)
			continue
		}

		res = append(res, *model)
	}

	return res, nil
}

// Remove deletes a specific model and all its files
func (s *Store) Remove(name string) error {
	slog.Debug("Remove model", "model", name)

	err := s.LockModel(name)
	if err != nil {
		return err
	}
	defer s.UnlockModel(name)
	return os.RemoveAll(filepath.Join(s.home, "models", name))
}

// Clean removes all stored models and the models directory
func (s *Store) Clean() int {
	slog.Debug("Start clean model")

	models, err := s.scanModelDir()
	if err != nil {
		return 0
	}

	// Get list of all model names to remove
	count := 0
	for _, model := range models {
		if err := s.Remove(model); err != nil {
			slog.Warn("Failed to remove model", "model", model, "err", err)
			continue
		}
		count += 1
	}

	return count
}

func (s *Store) GetManifest(name string) (*types.ModelManifest, error) {
	err := s.LockModel(name)
	if err != nil {
		return nil, err
	}
	defer s.UnlockModel(name)

	dir := filepath.Join(s.home, "models")
	// Read manifest file
	data, e := os.ReadFile(filepath.Join(dir, name, "nexa.manifest"))
	if e != nil {
		return nil, e
	}

	// Parse manifest JSON
	model := types.ModelManifest{}
	e = sonic.Unmarshal(data, &model)
	if e != nil {
		return nil, e
	}
	return &model, nil
}

// Pull downloads a model from HuggingFace and stores it locally
// It fetches the model tree, finds .gguf files, downloads them, and saves metadata
// if model not specify, all is set true, and autodetect true
func (s *Store) Pull(ctx context.Context, mf types.ModelManifest) (infoCh <-chan types.DownloadInfo, errCh <-chan error) {
	infoC := make(chan types.DownloadInfo, 10)
	infoCh = infoC
	errC := make(chan error, 1)
	errCh = errC

	go func() {
		defer close(errC)
		defer close(infoC)

		// clean before
		if err := s.Remove(mf.Name); err != nil {
			errC <- err
			return
		}

		if err := s.LockModel(mf.Name); err != nil {
			errC <- err
			return
		}
		defer s.UnlockModel(mf.Name)

		// filter download file
		var needs []model_hub.ModelFileInfo
		for _, f := range mf.ModelFile {
			if f.Downloaded {
				needs = append(needs, model_hub.ModelFileInfo{Name: f.Name, Size: f.Size})
			}
		}
		if mf.MMProjFile.Name != "" {
			if mf.MMProjFile.Downloaded {
				needs = append(needs, model_hub.ModelFileInfo{Name: mf.MMProjFile.Name, Size: mf.MMProjFile.Size})
			}
		}
		if mf.TokenizerFile.Name != "" {
			if mf.TokenizerFile.Downloaded {
				needs = append(needs, model_hub.ModelFileInfo{Name: mf.TokenizerFile.Name, Size: mf.TokenizerFile.Size})
			}
		}
		for _, f := range mf.ExtraFiles {
			if f.Downloaded {
				needs = append(needs, model_hub.ModelFileInfo{Name: f.Name, Size: f.Size})
			}
		}

		// Create model directory structure
		err := os.MkdirAll(filepath.Join(s.home, "models", mf.Name), 0o770)
		if err != nil {
			errC <- err
			return
		}

		// Create modelfile for storing downloaded content
		resCh, errCh := model_hub.StartDownload(ctx, mf.Name, filepath.Join(s.home, "models", mf.Name), needs)
		for d := range resCh {
			infoC <- d
		}
		for e := range errCh {
			errC <- e
			return
		}

		model := types.ModelManifest{
			Name:          mf.Name,
			ModelName:     mf.ModelName,
			ModelType:     mf.ModelType,
			PluginId:      mf.PluginId,
			DeviceId:      mf.DeviceId,
			MinSDKVersion: mf.MinSDKVersion,
			ModelFile:     mf.ModelFile,
			MMProjFile:    mf.MMProjFile,
			TokenizerFile: mf.TokenizerFile,
			ExtraFiles:    mf.ExtraFiles,
		}
		manifestPath := filepath.Join(s.home, "models", mf.Name, "nexa.manifest")
		manifestData, _ := sonic.Marshal(model) // JSON marshal won't fail, ignore error
		err = os.WriteFile(manifestPath, manifestData, 0o664)
		if err != nil {
			errC <- err
			return
		}
	}()

	return
}

// Pull downloads a model from HuggingFace and stores it locally
// It fetches the model tree, finds .gguf files, downloads them, and saves metadata
// if model not specify, all is set true, and autodetect true
func (s *Store) PullExtraQuant(ctx context.Context, omf, nmf types.ModelManifest) (infoCh <-chan types.DownloadInfo, errCh <-chan error) {
	infoC := make(chan types.DownloadInfo, 10)
	infoCh = infoC
	errC := make(chan error, 1)
	errCh = errC

	go func() {
		defer close(errC)
		defer close(infoC)

		if err := s.LockModel(nmf.Name); err != nil {
			errC <- err
			return
		}
		defer s.UnlockModel(nmf.Name)

		// filter download file
		var needs []model_hub.ModelFileInfo
		for q, f := range nmf.ModelFile {
			if f.Downloaded && !omf.ModelFile[q].Downloaded {
				needs = append(needs, model_hub.ModelFileInfo{Name: f.Name, Size: f.Size})
			}
		}
		if nmf.TokenizerFile.Downloaded && !omf.TokenizerFile.Downloaded {
			needs = append(needs, model_hub.ModelFileInfo{Name: nmf.TokenizerFile.Name, Size: nmf.TokenizerFile.Size})
		}
		for q, f := range nmf.ExtraFiles {
			if f.Downloaded && !omf.ExtraFiles[q].Downloaded {
				needs = append(needs, model_hub.ModelFileInfo{Name: f.Name, Size: f.Size})
			}
		}

		// Create model directory structure
		err := os.MkdirAll(filepath.Join(s.home, "models", nmf.Name), 0o770)
		if err != nil {
			errC <- err
			return
		}

		resCh, errCh := model_hub.StartDownload(ctx, nmf.Name, filepath.Join(s.home, "models", nmf.Name), needs)
		for d := range resCh {
			infoC <- d
		}
		for e := range errCh {
			errC <- e
			return
		}

		model := types.ModelManifest{
			Name:          nmf.Name,
			ModelName:     nmf.ModelName,
			ModelType:     nmf.ModelType,
			PluginId:      nmf.PluginId,
			DeviceId:      nmf.DeviceId,
			ModelFile:     nmf.ModelFile,
			MMProjFile:    nmf.MMProjFile,
			TokenizerFile: nmf.TokenizerFile,
			ExtraFiles:    nmf.ExtraFiles,
		}
		manifestPath := filepath.Join(s.home, "models", nmf.Name, "nexa.manifest")
		manifestData, _ := sonic.Marshal(model) // JSON marshal won't fail, ignore error
		err = os.WriteFile(manifestPath, manifestData, 0o664)
		if err != nil {
			errC <- err
			return
		}
	}()

	return
}

func (s *Store) DataPath() string {
	return s.home
}

func (s *Store) ModelDirPath() string {
	return filepath.Join(s.home, "models")
}

// ModelfilePath returns the full path to a model's data file
func (s *Store) ModelfilePath(name string, file string) string {
	return filepath.Join(s.home, "models", name, file)
}

func (s *Store) scanModelDir() ([]string, error) {
	orgs, e := os.ReadDir(s.ModelDirPath())
	if e != nil {
		slog.Warn("Failed to read model directory", "err", e)
		return nil, e
	}

	// Parse each model directory's manifest
	res := make([]string, 0)
	for _, org := range orgs {
		if !org.IsDir() {
			continue
		}

		ignoreDirs := []string{".cache"}
		if slices.Contains(ignoreDirs, org.Name()) {
			continue
		}

		repos, e := os.ReadDir(filepath.Join(s.ModelDirPath(), org.Name()))
		if e != nil {
			slog.Warn("Failed to read model subdirectory", "org", org.Name(), "err", e)
			continue
		}

		for _, repo := range repos {
			if !repo.IsDir() {
				continue
			}

			res = append(res, org.Name()+"/"+repo.Name())
		}
	}

	return res, nil
}
