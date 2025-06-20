MODEL?=Qwen/Qwen3-0.6B-GGUF

.PHONY: run build download clean

run:
	LD_LIBRARY_PATH=./build/lib ./build/nexa infer $(MODEL)

build:
	cd runner && go build -o ../build/nexa ./cmd/nexa-cli


download:
	mkdir -p build/include build/lib

clean:
	rm -rf build/nexa

