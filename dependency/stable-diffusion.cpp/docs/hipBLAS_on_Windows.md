# Using hipBLAS on Windows

To get hipBLAS in `stable-diffusion.cpp` working on Windows, go through this guide section by section.

## Build Tools for Visual Studio 2022

Skip this step if you already have Build Tools installed.

To install Build Tools, go to [Visual Studio Downloads](https://visualstudio.microsoft.com/vs/), download `Visual Studio 2022 and other Products` and run the installer.

## CMake

Skip this step if you already have CMake installed: running `cmake --version` should output `cmake version x.y.z`.

Download latest `Windows x64 Installer` from [Download | CMake](https://cmake.org/download/) and run it.

## ROCm

Skip this step if you already have Build Tools installed.

The [validation tools](https://rocm.docs.amd.com/en/latest/reference/validation_tools.html) not support on Windows. So you should confirm the Version of `ROCM` by yourself.

Fortunately, `AMD` provides complete help documentation, you can use the help documentation to install [ROCM](https://rocm.docs.amd.com/en/latest/deploy/windows/quick_start.html)

>**If you encounter an error, if it is [AMD ROCm Windows Installation Error 215](https://github.com/RadeonOpenCompute/ROCm/issues/2363), don't worry about this error. ROCM has been installed correctly, but the vs studio plugin installation failed, we can ignore it.**

Then we must set `ROCM` as environment variables before running cmake.

Usually if you install according to the official tutorial and do not modify the ROCM path, then there is a high probability that it is here `C:\Program Files\AMD\ROCm\5.5\bin`

This is what I use to set the clang:
```Commandline
set CC=C:\Program Files\AMD\ROCm\5.5\bin\clang.exe
set CXX=C:\Program Files\AMD\ROCm\5.5\bin\clang++.exe
```

## Ninja

Skip this step if you already have Ninja installed: running `ninja --version` should output `1.11.1`.

Download latest `ninja-win.zip` from [GitHub Releases Page](https://github.com/ninja-build/ninja/releases/tag/v1.11.1) and unzip. Then set as environment variables. I unzipped it in `C:\Program Files\ninja`, so I set it like this:

```Commandline
set ninja=C:\Program Files\ninja\ninja.exe
```
## Building stable-diffusion.cpp

The thing different from the regular CPU build is `-DSD_HIPBLAS=ON` ,
`-G "Ninja"`, `-DCMAKE_C_COMPILER=clang`, `-DCMAKE_CXX_COMPILER=clang++`, `-DAMDGPU_TARGETS=gfx1100`

>**Notice**: check the `clang` and `clang++` information:
```Commandline
clang --version
clang++ --version
```

If you see like this, we can continue:
```
clang version 17.0.0 (git@github.amd.com:Compute-Mirrors/llvm-project e3201662d21c48894f2156d302276eb1cf47c7be)
Target: x86_64-pc-windows-msvc
Thread model: posix
InstalledDir: C:\Program Files\AMD\ROCm\5.5\bin
```

```
clang version 17.0.0 (git@github.amd.com:Compute-Mirrors/llvm-project e3201662d21c48894f2156d302276eb1cf47c7be)
Target: x86_64-pc-windows-msvc
Thread model: posix
InstalledDir: C:\Program Files\AMD\ROCm\5.5\bin
```

>**Notice** that the `gfx1100` is the GPU architecture of my GPU, you can change it to your GPU architecture. Click here to see your architecture [LLVM Target](https://rocm.docs.amd.com/en/latest/release/windows_support.html#windows-supported-gpus)

My GPU is AMD Radeon™ RX 7900 XTX Graphics, so I set it to `gfx1100`.

option:

```commandline
mkdir build
cd build
cmake .. -G "Ninja" -DCMAKE_C_COMPILER=clang -DCMAKE_CXX_COMPILER=clang++ -DSD_HIPBLAS=ON -DCMAKE_BUILD_TYPE=Release -DAMDGPU_TARGETS=gfx1100
cmake --build . --config Release
```

If everything went OK, `build\bin\sd.exe` file should appear.
