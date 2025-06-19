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

// TODO: multi gguf file
func (s *Store) Pull(name string) error {
	// make request
	req, e := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/api/models/%s/tree/main", HF_ENDPOINT, name),
		nil)
	if e != nil {
		return e
	}
	//req.Header.Set("Authorization", config.Get().HFToken)

	// do request
	r, e := http.DefaultClient.Do(req)
	if e != nil {
		return e
	}
	defer r.Body.Close()

	// read body
	body, e := io.ReadAll(r.Body)
	if e != nil {
		return e
	}

	// parse body
	files := make([]HFFileInfo, 0)
	e = sonic.Unmarshal(body, &files)
	if e != nil {
		return e
	}

	// select file
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

	// get download url from metadata
	encName := s.encodeName(name)
	e = os.MkdirAll(path.Join(s.modelDir(), encName), 0770)
	if e != nil {
		return e
	}
	modelfile, e := os.Create(path.Join(s.modelDir(), encName, "modelfile"))
	if e != nil {
		return e
	}
	defer modelfile.Close()

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

	_, e = io.Copy(modelfile, resDownload.Body)
	if e != nil {
		return e
	}

	// save manifest
	model := types.Model{
		Name: name,
		Size: file.Size,
	}
	manifestPath := path.Join(s.modelDir(), encName, "manifest")
	manifestData, _ := sonic.Marshal(model) // wont failed, ignoure error
	e = os.WriteFile(manifestPath, manifestData, 0664)
	if e != nil {
		return e
	}

	return nil
}
