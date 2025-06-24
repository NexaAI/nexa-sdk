ARGS?=infer Qwen/Qwen3-0.6B-GGUF

.PHONY: run build download clean

run:
	./build/nexa $(ARGS)

build:
	swag init -d ./runner/server -g ./server.go -o ./runner/server/docs
	cd runner && CGO_ENABLED=0 go build -o ../build/nexa ./cmd/nexa-launcher
	cd runner && go build -tags="sonic avx" -o ../build/nexa-cli ./cmd/nexa-cli

download:
	mkdir -p build/include build/lib

test:
	cd runner && LD_LIBRARY_PATH=$(PWD)/build/lib go test -v $(ARGS)

clean:
	rm -rf build/nexa
	rm -rf build/nexa-cli

