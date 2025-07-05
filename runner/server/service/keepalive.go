package service

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/NexaAI/nexa-sdk/internal/config"
	"github.com/NexaAI/nexa-sdk/internal/store"
	"github.com/NexaAI/nexa-sdk/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/nexa-sdk"
)

// KeepAliveGet retrieves a model from the keepalive cache or creates it if not found
// This avoids the overhead of repeatedly loading/unloading models from disk
func KeepAliveGet[T any](name string, param types.ModelParam) (*T, error) {
	t, err := keepAliveGet[T](name, param)
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
	Destroy()
	Reset()
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
func keepAliveGet[T any](name string, param types.ModelParam) (any, error) {
	keepAlive.Lock()
	defer keepAlive.Unlock()

	// make nexaml repo as default
	if !strings.Contains(name, "/") {
		name = "nexaml/" + name
	}

	// Check if model already exists in cache
	model, ok := keepAlive.models[name]
	if ok && reflect.DeepEqual(model.param, param) {
		model.model.Reset()
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
	modelfile := s.ModelfilePath(manifest.Name, manifest.ModelFile)

	var t keepable
	var e error
	switch reflect.TypeFor[T]() {
	case reflect.TypeFor[nexa_sdk.LLM]():
		t, e = nexa_sdk.NewLLM(modelfile, nil, param.CtxLen, param.Device)
	case reflect.TypeFor[nexa_sdk.VLM]():
		if manifest.MMProjFile == "" {
			return nil, fmt.Errorf("missing mmproj file")
		} else {
			mmproj := s.ModelfilePath(manifest.Name, manifest.MMProjFile)
			t, e = nexa_sdk.NewVLM(modelfile, &mmproj, param.CtxLen, param.Device)
		}
	case reflect.TypeFor[nexa_sdk.Embedder]():
		t, e = nexa_sdk.NewEmbedder(modelfile, nil, param.Device)
	case reflect.TypeFor[nexa_sdk.Reranker]():
		t, e = nexa_sdk.NewReranker(modelfile, nil, param.Device)
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
