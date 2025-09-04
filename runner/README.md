## Build

### Windows Arm

Setup GO Env

```bash
go env -w CGO_ENABLED=1
go env -w CC=C:/tools/msys64/clangarm64/bin/clang.exe
go env -w CXX=C:/tools/msys64/clangarm64/bin/clang++.exe
```

Due to translation layer, make can not detech arch, need manual specify arch

```bash
make download ARCH=arm64
make build
```

### Windows Arm QNN

Setup GO Env

```bash
go env -w CGO_ENABLED=1
go env -w CC=C:/tools/msys64/clangarm64/bin/clang.exe
go env -w CXX=C:/tools/msys64/clangarm64/bin/clang++.exe
```

Build and Run

- copy follow files from `nexasdk-bridge` to `runner/build`
  - `ml.h -> ml.h`
- copy follow files from `nexaml/nexaml-models/sdk-bridge` to `runner/build`
  - `fftw3.dll`
  - `nexa-mm-process.dll`
  - `nexa-sampling.dll`
  - `nexaproc.dll`
  - `omni-neural-sdk.dll`
  - `qwen3-sdk.dll` -> `qwen3/qwen3-sdk.dll`
  - `qwen3-4B-sdk.dll` -> `qwen3-4B/qwen3-sdk.dll` **MUST RENAME**
  - `yolov12-sdk.dll -> yolov12/yolov12-sdk.dll`
  - `paddle-ocr-proc-lib.dll -> paddleocr/paddle-ocr-proc-lib.dll`
  - `paddleocr-sdk.dll -> paddleocr/yolov12-sdk.dll` **MUST RENAME**
- run `make build`
- set env token `$env:NEXA_HFTOKEN="hf_xxxxxxxxxxxxx"`
- run `./build/nexa infer [huggingface model name]`
