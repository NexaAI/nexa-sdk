ARGS?=infer Qwen/Qwen3-0.6B-GGUF
#ARGS?=serve


LLAMA_RUNTIME_LIBS=libggml-base libggml-blas libggml-cpu libggml-metal libggml libllama libmtmd libnexa_bridge

OS=linux
EXT=so
RUNTIME_LIBS=$(LLAMA_RUNTIME_LIBS)


.PHONY: run build download clean

run:
	./build/nexa $(ARGS)

build:
	cd runner/nexa-sdk/stub && g++ -O3 -s -fPIC -shared -o libnexa_bridge.so *.cpp
	cd runner && CGO_ENABLED=0 go build -o ../build/nexa ./cmd/nexa-launcher
	cd runner && go build -tags="sonic avx" -o ../build/nexa-cli ./cmd/nexa-cli

doc:
	swag init -d ./runner/server -g ./server.go -o ./runner/server/docs

test:
	cd runner && LD_LIBRARY_PATH=$(PWD)/build/lib go test -v ./nexa-sdk

download:
	mkdir -p build/lib
	@echo "====> Download runtime"
	@for file in $(RUNTIME_LIBS); do \
		echo "Download $$file.$(EXT)"; \
		curl -L -o build/lib/$$file.$(EXT) https://nexa-model-hub-bucket.s3.us-west-1.amazonaws.com/public/nexasdk/$(OS)/$$file.$(EXT); \
	done

clean:
	rm -rf build/nexa
	rm -rf build/nexa-cli
	rm -rf runner/nexa-sdk/stub/*.so

