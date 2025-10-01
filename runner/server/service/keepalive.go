package service

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/NexaAI/nexa-sdk/runner/internal/config"
	"github.com/NexaAI/nexa-sdk/runner/internal/store"
	"github.com/NexaAI/nexa-sdk/runner/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
)

// KeepAliveGet retrieves a model from the keepalive cache or creates it if not found
// This avoids the overhead of repeatedly loading/unloading models from disk
func KeepAliveGet[T any](name string, param types.ModelParam, reset bool) (*T, error) {
	t, err := keepAliveGet[T](name, param, reset)
	if err != nil {
		return nil, err
	}
	return t.(*T), nil
}

var keepAlive keepAliveService

// current only support keepalive one model
type keepAliveService struct {
	models map[string]*modelKeepInfo // Map of model name to model info
	stopCh chan struct{}             // Channel to stop the cleanup goroutine

	sync.Mutex // Protects concurrent access to models map
}

// modelKeepInfo holds metadata for a cached model instance
type modelKeepInfo struct {
	model    keepable
	param    types.ModelParam
	lastTime time.Time
}

// keepable interface defines objects that can be managed by the keepalive service
// Objects must support cleanup and reset operations
type keepable interface {
	Destroy() error
	Reset() error
}

// start begins the background cleanup process that removes unused models
// Runs a ticker every 5 seconds to check for models that exceed the keepalive timeout
func (keepAlive *keepAliveService) start() {
	keepAlive.models = make(map[string]*modelKeepInfo)
	keepAlive.stopCh = make(chan struct{})

	go func() {
		t := time.NewTicker(5 * time.Second)
		for {
			select {
			// Stop signal received - cleanup all models and exit
			case <-keepAlive.stopCh:
				keepAlive.Lock()
				for name, model := range keepAlive.models {
					model.model.Destroy()
					delete(keepAlive.models, name)
				}
				keepAlive.Unlock()
				return

			// Periodic cleanup - remove models that haven't been used recently
			case <-t.C:
				keepAlive.Lock()
				for name, model := range keepAlive.models {
					if time.Since(model.lastTime).Milliseconds()/1000 > config.Get().KeepAlive {
						model.model.Destroy()
						delete(keepAlive.models, name)
					}
				}
				keepAlive.Unlock()
			}
		}
	}()
}

// keepAliveGet retrieves a cached model or creates a new one if not found
// Ensures only one model is kept in memory at a time by clearing others
func keepAliveGet[T any](name string, param types.ModelParam, reset bool) (any, error) {
	keepAlive.Lock()
	defer keepAlive.Unlock()

	// if actualPath, exists := config.GetModelMapping(name); exists {
	// 	// support shortcuts like qwen3 -> Qwen/Qwen3-4B-GGUF
	// 	name = actualPath
	// } else if !strings.Contains(name, "/") {
	// 	// fallback to NexaAI prefix for unknown shortcuts
	// 	name = "NexaAI/" + name
	// }

	// Check if model already exists in cache
	model, ok := keepAlive.models[name]
	if ok && reflect.DeepEqual(model.param, param) {
		if reset {
			model.model.Reset()
		}
		model.lastTime = time.Now()
		return model.model, nil
	}

	// Clear existing models to ensure only one is in memory
	// This prevents memory overflow but limits to single model usage
	// TODO: unload model due to free ram/vram
	for name, model := range keepAlive.models {
		model.model.Destroy()
		delete(keepAlive.models, name)
	}

	s := store.Get()
	manifest, err := s.GetManifest(name)
	if err != nil {
		return nil, err
	}

	// TODO: select one of quant
	var modelfile string
	for _, v := range manifest.ModelFile {
		if v.Downloaded {
			modelfile = s.ModelfilePath(manifest.Name, v.Name)
			break
		}
	}

	var t keepable
	var e error
	switch reflect.TypeFor[T]() {
	case reflect.TypeFor[nexa_sdk.LLM]():
		t, e = nexa_sdk.NewLLM(nexa_sdk.LlmCreateInput{
			ModelName: manifest.ModelName,
			ModelPath: modelfile,
			Config: nexa_sdk.ModelConfig{
				NCtx:         param.NCtx,
				NGpuLayers:   param.NGpuLayers,
				SystemPrompt: param.SystemPrompt,
			},
			PluginID: manifest.PluginId,
			DeviceID: manifest.DeviceId,
		})
	case reflect.TypeFor[nexa_sdk.VLM]():
		var mmproj string
		if manifest.MMProjFile.Name != "" {
			mmproj = s.ModelfilePath(manifest.Name, manifest.MMProjFile.Name)
		}
		var tokenizer string
		if manifest.TokenizerFile.Name != "" {
			tokenizer = s.ModelfilePath(manifest.Name, manifest.TokenizerFile.Name)
		}
		t, e = nexa_sdk.NewVLM(nexa_sdk.VlmCreateInput{
			ModelName:     manifest.ModelName,
			ModelPath:     modelfile,
			MmprojPath:    mmproj,
			TokenizerPath: tokenizer,
			Config: nexa_sdk.ModelConfig{
				NCtx:         param.NCtx,
				NGpuLayers:   param.NGpuLayers,
				SystemPrompt: param.SystemPrompt,
			},
			PluginID: manifest.PluginId,
			DeviceID: manifest.DeviceId,
		})
	case reflect.TypeFor[nexa_sdk.Embedder]():
		t, e = nexa_sdk.NewEmbedder(nexa_sdk.EmbedderCreateInput{
			ModelName: manifest.ModelName,
			ModelPath: modelfile,
			PluginID:  manifest.PluginId,
			DeviceID:  manifest.DeviceId,
		})
	case reflect.TypeFor[nexa_sdk.ImageGen]():
		// For image generation models, use the model directory path instead of specific file
		modelDir := s.ModelfilePath(manifest.Name, "")
		t, e = nexa_sdk.NewImageGen(nexa_sdk.ImageGenCreateInput{
			ModelName: manifest.ModelName,
			ModelPath: modelDir,
			PluginID:  manifest.PluginId,
			DeviceID:  manifest.DeviceId,
		})
	//case reflect.TypeFor[nexa_sdk.Reranker]():
	//	t, e = nexa_sdk.NewReranker(modelfile, nil, param.Device)
	//case reflect.TypeFor[nexa_sdk.TTS]():
	//	t, e = nexa_sdk.NewTTS(modelfile, nil, param.Device)
	default:
		panic(fmt.Sprintf("not support type: %+#v", t))
	}
	if e != nil {
		return nil, e
	}
	model = &modelKeepInfo{
		model:    t,
		param:    param,
		lastTime: time.Now(),
	}
	keepAlive.models[name] = model

	return model.model, nil
}

// stop signals the cleanup goroutine to terminate
func (keepAlive *keepAliveService) stop() {
	keepAlive.stopCh <- struct{}{}
}
