package service

import (
	"sync"
	"time"

	"github.com/NexaAI/nexa-sdk/internal/config"
	nexa_sdk "github.com/NexaAI/nexa-sdk/nexa-sdk"
)

// KeepAliveGet retrieves a model from the keepalive cache or creates it if not found
// This avoids the overhead of repeatedly loading/unloading models from disk
func KeepAliveGet(name string, createFunc func() nexa_sdk.LLM) nexa_sdk.LLM {
	return keepAlive.get(name, createFunc)
}

// keepable interface defines objects that can be managed by the keepalive service
// Objects must support cleanup and reset operations
type keepable interface {
	Destroy()
	Reset()
}

// modelKeepInfo holds metadata for a cached model instance
type modelKeepInfo struct {
	model    nexa_sdk.LLM
	param    map[string]any // TODO: Reload when param change
	lastTime time.Time
}

// current only support keepalive one model
// TODO: unload model due to free ram/vram
type keepAliveService struct {
	models map[string]*modelKeepInfo // Map of model name to model info
	stopCh chan<- struct{}           // Channel to stop the cleanup goroutine

	sync.Mutex // Protects concurrent access to models map
}

var keepAlive keepAliveService

func NewKeepAlive() keepAliveService {
	return keepAliveService{
		models: make(map[string]*modelKeepInfo),
	}
}

// start begins the background cleanup process that removes unused models
// Runs a ticker every 5 seconds to check for models that exceed the keepalive timeout
func (k *keepAliveService) start() {
	stopCh := make(chan struct{})
	k.stopCh = stopCh

	go func() {
		t := time.NewTicker(5 * time.Second)
		for {
			select {
			// Stop signal received - cleanup all models and exit
			case <-stopCh:
				k.Lock()
				for name, model := range k.models {
					model.model.Destroy()
					delete(k.models, name)
				}
				k.Unlock()
				return

			// Periodic cleanup - remove models that haven't been used recently
			case <-t.C:
				k.Lock()
				for name, model := range k.models {
					if time.Since(model.lastTime).Milliseconds()/1000 > config.Get().KeepAlive {
						model.model.Destroy()
						delete(k.models, name)
					}
				}
				k.Unlock()
			}
		}
	}()
}

// get retrieves a cached model or creates a new one if not found
// Ensures only one model is kept in memory at a time by clearing others
// TODO: use generic type for better type safety
func (k *keepAliveService) get(name string, create func() nexa_sdk.LLM) nexa_sdk.LLM {
	k.Lock()
	defer k.Unlock()

	// Check if model already exists in cache
	model, ok := k.models[name]
	if ok {
		//model.model.SaveKVCache("./cache")
		model.model.Reset()
		//model.model.LoadKVCache("./cache")
		model.lastTime = time.Now()
		return model.model
	}

	// Clear existing models to ensure only one is in memory
	// This prevents memory overflow but limits to single model usage
	for name, model := range k.models {
		model.model.Destroy()
		delete(k.models, name)
	}

	// Create new model and add to cache
	model = &modelKeepInfo{
		model:    create(),
		lastTime: time.Now(),
	}
	k.models[name] = model

	return model.model
}

// stop signals the cleanup goroutine to terminate
func (k *keepAliveService) stop() {
	k.stopCh <- struct{}{}
}
