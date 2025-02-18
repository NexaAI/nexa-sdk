@call "C:\Program Files (x86)\Intel\oneAPI\setvars.bat" intel64 --force 

set "CMAKE_GENERATOR=Ninja"
set "CMAKE_ARGS=-DGGML_SYCL=ON -DCMAKE_C_COMPILER=cl -DCMAKE_CXX_COMPILER=icx"

@REM Uncomment the following lines to check if oneAPI is properly installed.
@REM echo Detecting available SYCL devices
@REM echo ----------------------------------------
@REM sycl-ls
@REM echo ----------------------------------------

pip install -e .
@REM pip install nexaai --prefer-binary --index-url https://github.nexa.ai/whl/sycl --extra-index-url https://pypi.org/simple --no-cache-dir
