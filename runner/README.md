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

- copy `ml.h` `qwen3-sdk.dll` `paddleocr-sdk.dll` to `runner/build`
- run `make build`
- set env token `$env:NEXA_HFTOKEN="hf_xxxxxxxxxxxxx"`
- run `./build/nexa pull nexaml/qnn-laptop-libs`
- run `./build/nexa infer qwen3` or `./build/nexa infer paddleoc`
