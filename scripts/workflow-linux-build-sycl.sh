# This script is intended for use within a GitHub Actions workflow only.  
# Please do not run it manually.

# Init oneAPI environment
source /opt/intel/oneapi/setvars.sh

export CMAKE_ARGS="-DGGML_SYCL=ON -DCMAKE_C_COMPILER=icx -DCMAKE_CXX_COMPILER=icpx"
python -m build --wheel
