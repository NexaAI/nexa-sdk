MODEL?=./build/Qwen3-0.6B-GGUF/Qwen3-0.6B-Q8_0.gguf

.PHONY: run build clean

run:
	LD_LIBRARY_PATH=./build/lib ./build/nexa infer --model $(MODEL)

build: build/llama.cpp
	mkdir -p build/include build/lib

	cmake -B build/llama.cpp-build build/llama.cpp -DBUILD_SHARED_LIBS=ON -DLLAMA_BUILD_TESTS=OFF -DLLAMA_BUILD_TOOLS=OFF -DLLAMA_BUILD_EXAMPLES=OFF -DLLAMA_BUILD_SERVER=OFF
	cmake --build build/llama.cpp-build -j --config Release
	cp ./build/llama.cpp/include/llama.h ./build/include
	cp ./build/llama.cpp/ggml/include/*.h ./build/include
	cp ./build/llama.cpp-build/bin/*.so ./build/lib

	mkdir -p build/binding-build
	g++ -O2 -fPIC -shared -I./build/include -L./build/lib -lllama -o ./build/binding-build/libbinding.so ./binding/binding.cpp
	cp ./build/binding-build/*.so ./build/lib/
	cp ./binding/binding.h ./build/include

	cd runner && go build -o ../build/nexa ./cmd/nexa-cli


build/llama.cpp:
	git clone https://github.com/ggml-org/llama.cpp.git build/llama.cpp

clean:
	rm -rf build/nexa
	rm -rf build/include build/lib
	rm -rf build/binding-build
	rm -rf build/llama.cpp-build

