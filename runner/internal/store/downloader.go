package store

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/NexaAI/nexa-sdk/internal/config"
	"github.com/NexaAI/nexa-sdk/internal/types"
	"github.com/bytedance/sonic"
)

const HF_ENDPOINT = "https://huggingface.co"

// HFFileInfo represents file metadata from HuggingFace API response
type HFFileInfo struct {
	Type string `json:"type"`
	OId  string `json:"oid"`
	Size uint64 `json:"size"`
	Path string `json:"path"`
	LFS  struct {
		OId string `json:"oid"`
	} `json:"lfs"`
	XetHash string `json:"xetHash"`
}

// Pull downloads a model from HuggingFace and stores it locally
// It fetches the model tree, finds .gguf files, downloads them, and saves metadata
// TODO: multi gguf file support
func (s *Store) Pull(ctx context.Context, name string) (infoCh <-chan types.DownloadInfo, errCh <-chan error) {
	infoC := make(chan types.DownloadInfo, 10)
	infoCh = infoC
	errC := make(chan error)
	errCh = errC

	go func() {
		defer close(errC)
		defer close(infoC)

		// Fetch model tree from HuggingFace API
		req, e := http.NewRequest(
			http.MethodGet,
			fmt.Sprintf("%s/api/models/%s/tree/main", HF_ENDPOINT, name),
			nil)
		if e != nil {
			errC <- e
			return
		}
		req.Header.Set("Authorization", config.Get().HFToken) // TODO: auth support

		// Execute API request
		r, e := http.DefaultClient.Do(req)
		if e != nil {
			errC <- e
			return
		}
		defer r.Body.Close()

		// Read response body
		body, e := io.ReadAll(r.Body)
		if e != nil {
			errC <- e
			return
		}

		// Parse JSON response into file list
		files := make([]HFFileInfo, 0)
		e = sonic.Unmarshal(body, &files)
		if e != nil {
			errC <- e
			return
		}

		// Find first .gguf file in the model
		var file HFFileInfo
		for _, f := range files {
			if strings.HasSuffix(f.Path, ".gguf") {
				file = f
				break
			}
		}
		if file.Size == 0 {
			errC <- fmt.Errorf("no valid file")
			return
		}

		// Create model directory structure
		encName := s.encodeName(name)
		e = os.MkdirAll(path.Join(s.home, "models", encName), 0770)
		if e != nil {
			errC <- e
			return
		}

		// Create modelfile for storing downloaded content
		modelfile, e := os.Create(path.Join(s.home, "models", encName, "modelfile"))
		if e != nil {
			errC <- e
			return
		}
		defer modelfile.Close()

		// Download the actual model file
		reqDownload, e := http.NewRequest(
			http.MethodGet,
			fmt.Sprintf("%s/%s/resolve/main/%s?download=true", HF_ENDPOINT, name, file.Path),
			nil)
		if e != nil {
			errC <- e
			return
		}
		resDownload, e := http.DefaultClient.Do(reqDownload)
		if e != nil {
			errC <- e
			return
		}
		defer resDownload.Body.Close()

		// Copy downloaded content to local file
		dInfo := types.DownloadInfo{
			Size: file.Size,
		}
		_, e = io.Copy(io.MultiWriter(
			modelfile,
			types.FuncWriter{
				// TODO: reduce channel message
				F: func(b []byte) (int, error) {
					dInfo.Downloaded += uint64(len(b))
					infoC <- dInfo
					return len(b), nil
				},
			},
		), resDownload.Body)
		if e != nil {
			errC <- e
			return
		}

		// Create and save model manifest with metadata
		model := types.Model{
			Name: name,
			Size: file.Size,
		}
		manifestPath := path.Join(s.home, "models", encName, "manifest")
		manifestData, _ := sonic.Marshal(model) // JSON marshal won't fail, ignore error
		e = os.WriteFile(manifestPath, manifestData, 0664)
		if e != nil {
			errC <- e
			return
		}
	}()

	return
}
