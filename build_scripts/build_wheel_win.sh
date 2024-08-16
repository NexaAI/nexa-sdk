rm pyproject.toml
rm -r build

# cpu
cp build_scripts/pyproject_win_cpu.toml pyproject.toml
python -m build --wheel
rm pyproject.toml
rm -r build

# cuda
cp build_scripts/pyproject_win_cuda.toml pyproject.toml
python -m build --wheel
rm pyproject.toml
rm -r build

# don't quit immediately
read -p "Press [Enter] key to continue..."