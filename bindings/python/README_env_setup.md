# ARM64 Python Environment Setup for Nexa SDK

## Setup

Run the PowerShell script from the project root to create a virtual environment with Python ARM64:

```powershell
.\bindings\python\env_setup.ps1
```

The script will automatically activate the virtual environment when complete.

## Usage

### Activate the Virtual Environment

If the environment was previously installed, activate it with:

```powershell
.\venv\Scripts\Activate.ps1
```

### Option 1: Run Jupyter Notebook

Open the ARM64 notebook directly:

```powershell
jupyter-notebook "bindings\python\notebook\windows(arm64).ipynb"
```

**Important:** When the notebook opens, select the kernel:
- **Kernel** → **Change Kernel** → **Python (Nexa SDK ARM64)**

### Option 2: Run Python Files

Execute any Python binding file:

```powershell
python bindings\python\llm.py
python bindings\python\asr.py
python bindings\python\vlm.py
# ... or any other .py file in bindings\python\
```

---
