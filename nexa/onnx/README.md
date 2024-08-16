# Notes
For GPU version, you need to install onnxruntime-gpu version package. 
```
pip install optimum[onnxruntime-gpu]
```
If you encounter this error
```
FAIL : Failed to load library libonnxruntime_providers_cuda.so with error: libcublasLt.so.11: cannot open shared object file: No such file or directory
```
Then you need to follow this [doc](https://onnxruntime.ai/docs/install/) and install a compatible version of onnxruntime-gpu.
```
pip install onnxruntime-gpu --extra-index-url https://aiinfra.pkgs.visualstudio.com/PublicPackages/_packaging/onnxruntime-cuda-12/pypi/simple/
```