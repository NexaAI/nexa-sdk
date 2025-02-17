@REM This script is intended for use within a GitHub Actions workflow only.  
@REM Please do not run it manually.

@call "C:\Program Files (x86)\Intel\oneAPI\setvars.bat" intel64 --force 

set "CMAKE_GENERATOR=Ninja"
set "CMAKE_ARGS=-DGGML_SYCL=ON -DCMAKE_C_COMPILER=cl -DCMAKE_CXX_COMPILER=icx"
python -m build --wheel
