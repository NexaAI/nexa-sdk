package service

import (
	"sync"
	"time"

	"github.com/NexaAI/nexa-sdk/internal/config"
	nexa_sdk "github.com/NexaAI/nexa-sdk/nexa-sdk"
)

func KeepAliveGet(name string, createFunc func() nexa_sdk.LLM) nexa_sdk.LLM {
	return keepAlive.get(name, createFunc)
}

type keepable interface {
	Destroy()
	Reset()
}

type modelKeepInfo struct {
	model    nexa_sdk.LLM
	param    map[string]any // TODO: Reload when param change
	lastTime time.Time
}

// current only support keepalive one model
// TODO: unload model due to free ram/vram
type keepAliveService struct {
	models map[string]*modelKeepInfo
	stopCh chan<- struct{}

	sync.Mutex
}

var keepAlive keepAliveService

func NewKeepAlive() keepAliveService {
	return keepAliveService{
		models: make(map[string]*modelKeepInfo),
	}
}

func (k *keepAliveService) start() {
	stopCh := make(chan struct{})
	k.stopCh = stopCh

	go func() {
		t := time.NewTicker(5 * time.Second)
		for {
			select {

			case <-stopCh:
				k.Lock()
				for name, model := range k.models {
					model.model.Destroy()
					delete(k.models, name)
				}
				k.Unlock()
				return

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

// TODO: use generic type
func (k *keepAliveService) get(name string, create func() nexa_sdk.LLM) nexa_sdk.LLM {
	k.Lock()
	defer k.Unlock()

	model, ok := k.models[name]
	if ok {
		model.model.Reset()
		model.lastTime = time.Now()
		return model.model
	}

	// clear old make sure only one is in memory
	for name, model := range k.models {
		model.model.Destroy()
		delete(k.models, name)
	}

	model = &modelKeepInfo{
		model:    create(),
		lastTime: time.Now(),
	}
	k.models[name] = model

	return model.model
}

func (k *keepAliveService) stop() {
	k.stopCh <- struct{}{}
}
