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
