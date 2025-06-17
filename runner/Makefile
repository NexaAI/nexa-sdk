MODEL?=./build/Qwen3-0.6B-GGUF/Qwen3-0.6B-Q8_0.gguf

.PHONY: run build clean

run:
	LD_LIBRARY_PATH=./build/lib ./build/nexa infer --model $(MODEL) --stream

build:
	mkdir -p ./build/include
	mkdir -p ./build/lib

	# compile llama.cpp
	cmake -B ./build/llama.cpp nexa-sdk/llama.cpp
	cmake --build ./build/llama.cpp -j --config Release
	cp -a ./build/llama.cpp/bin/*.so ./build/lib
	cp -a ./nexa-sdk/llama.cpp/include/*.h ./build/include/
	cp -a ./nexa-sdk/llama.cpp/ggml/include/*.h ./build/include/
	
	# compile binding
	mkdir -p ./build/binding
	g++ -O2 -fPIC -shared -I./build/include -lllama -L./build/lib -o build/binding/libbinding.so ./nexa-sdk/binding/binding.cpp
	cp -a ./build/binding/*.so ./build/lib/
	cp -a ./nexa-sdk/binding/binding.h build/include/
	
	# compile runner
	go build -o build/nexa ./cmd/nexa-cli

clean:
	rm -rf ./build/{nexa,include,lib,binding,llama.cpp}
