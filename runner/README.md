## Build

### Prerequisites
Before building, make sure the following tools are installed:

- **unzip**  
  - On Windows:  
    ```powershell
    winget install -e GnuWin32.UnZip
    ```
- **curl**

---

### Install `nexasdk-bridge`

There are two ways to install the bridge library:

1. **From S3 bucket**  
```bash
make download
```
2. **From local files**

* **Unix/macOS:** Link the bridge library to the `build` folder:

```bash
make link
```

* **Windows:** Copy the bridge library from local:

```bash
make xcopy
```

---

### Build Project

Once the prerequisites and bridge library are installed, build the project:

```bash
make build
```

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
