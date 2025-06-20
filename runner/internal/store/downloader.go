package store

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	_ "github.com/NexaAI/nexa-sdk/internal/config"
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
func (s *Store) Pull(name string) error {
	// Fetch model tree from HuggingFace API
	req, e := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/api/models/%s/tree/main", HF_ENDPOINT, name),
		nil)
	if e != nil {
		return e
	}
	//req.Header.Set("Authorization", config.Get().HFToken) // TODO: auth support

	// Execute API request
	r, e := http.DefaultClient.Do(req)
	if e != nil {
		return e
	}
	defer r.Body.Close()

	// Read response body
	body, e := io.ReadAll(r.Body)
	if e != nil {
		return e
	}

	// Parse JSON response into file list
	files := make([]HFFileInfo, 0)
	e = sonic.Unmarshal(body, &files)
	if e != nil {
		return e
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
		return fmt.Errorf("no valid file")
	}

	// Create model directory structure
	encName := s.encodeName(name)
	e = os.MkdirAll(path.Join(s.modelDir(), encName), 0770)
	if e != nil {
		return e
	}

	// Create modelfile for storing downloaded content
	modelfile, e := os.Create(path.Join(s.modelDir(), encName, "modelfile"))
	if e != nil {
		return e
	}
	defer modelfile.Close()

	// Download the actual model file
	reqDownload, e := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/%s/resolve/main/%s?download=true", HF_ENDPOINT, name, file.Path),
		nil)
	if e != nil {
		return e
	}
	resDownload, e := http.DefaultClient.Do(reqDownload)
	if e != nil {
		return e
	}
	defer resDownload.Body.Close()

	// Copy downloaded content to local file
	_, e = io.Copy(modelfile, resDownload.Body)
	if e != nil {
		return e
	}

	// Create and save model manifest with metadata
	model := types.Model{
		Name: name,
		Size: file.Size,
	}
	manifestPath := path.Join(s.modelDir(), encName, "manifest")
	manifestData, _ := sonic.Marshal(model) // JSON marshal won't fail, ignore error
	e = os.WriteFile(manifestPath, manifestData, 0664)
	if e != nil {
		return e
	}

	return nil
}
