# PowerShell script to setup ARM64 Python and install nexaai
# This script will:
# 1. Download and install ARM64 Python silently
# 2. Create a virtual environment using absolute paths
# 3. Install nexaai in the virtual environment

$ErrorActionPreference = "Stop"

# Configuration
$PYTHON_VERSION = "3.11.9"  # Adjust version as needed
$PYTHON_INSTALLER_URL = "https://www.python.org/ftp/python/$PYTHON_VERSION/python-$PYTHON_VERSION-arm64.exe"
$PYTHON_INSTALL_DIR = "$env:LOCALAPPDATA\Programs\Python\Python311-ARM64"
$VENV_DIR = Join-Path $PSScriptRoot "venv"
$TEMP_INSTALLER = Join-Path $env:TEMP "python-installer-arm64.exe"

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "ARM64 Python Environment Setup for Nexa AI" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

# Step 1: Download Python ARM64 installer
Write-Host "[1/4] Downloading Python $PYTHON_VERSION ARM64 installer..." -ForegroundColor Yellow
try {
    # Use .NET WebClient for better progress and reliability
    $webClient = New-Object System.Net.WebClient
    $webClient.DownloadFile($PYTHON_INSTALLER_URL, $TEMP_INSTALLER)
    Write-Host "      Download completed: $TEMP_INSTALLER" -ForegroundColor Green
} catch {
    Write-Host "      Error downloading Python installer: $_" -ForegroundColor Red
    exit 1
}

# Step 2: Install Python silently
Write-Host "[2/4] Installing Python ARM64 to $PYTHON_INSTALL_DIR..." -ForegroundColor Yellow
try {
    # Install Python with the following options:
    # /quiet - No UI during installation
    # InstallAllUsers=0 - Install for current user only
    # TargetDir - Specify installation directory
    # PrependPath=0 - Don't modify PATH (we'll use absolute paths)
    # Include_test=0 - Don't include test suite
    # Include_pip=1 - Include pip
    # Include_dev=0 - Don't include development headers
    
    $installArgs = @(
        "/quiet",
        "InstallAllUsers=0",
        "TargetDir=$PYTHON_INSTALL_DIR",
        "PrependPath=0",
        "Include_test=0",
        "Include_pip=1",
        "Include_dev=0",
        "Include_launcher=0"
    )
    
    $process = Start-Process -FilePath $TEMP_INSTALLER -ArgumentList $installArgs -Wait -PassThru
    
    if ($process.ExitCode -ne 0) {
        Write-Host "      Python installation failed with exit code: $($process.ExitCode)" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "      Python installed successfully" -ForegroundColor Green
} catch {
    Write-Host "      Error installing Python: $_" -ForegroundColor Red
    exit 1
} finally {
    # Clean up installer
    if (Test-Path $TEMP_INSTALLER) {
        Remove-Item $TEMP_INSTALLER -Force
    }
}

# Step 3: Create virtual environment using absolute path
Write-Host "[3/4] Creating virtual environment at $VENV_DIR..." -ForegroundColor Yellow

$PYTHON_EXE = Join-Path $PYTHON_INSTALL_DIR "python.exe"

# Verify Python installation
if (-not (Test-Path $PYTHON_EXE)) {
    Write-Host "      Error: Python executable not found at $PYTHON_EXE" -ForegroundColor Red
    exit 1
}

# Display Python version
$pythonVersion = & $PYTHON_EXE --version
Write-Host "      Using Python: $pythonVersion" -ForegroundColor Cyan

# Remove existing venv if it exists
if (Test-Path $VENV_DIR) {
    Write-Host "      Removing existing virtual environment..." -ForegroundColor Yellow
    Remove-Item $VENV_DIR -Recurse -Force
}

# Create virtual environment using absolute path
try {
    & $PYTHON_EXE -m venv $VENV_DIR
    Write-Host "      Virtual environment created successfully" -ForegroundColor Green
} catch {
    Write-Host "      Error creating virtual environment: $_" -ForegroundColor Red
    exit 1
}

# Step 4: Install packages in the virtual environment
Write-Host "[4/5] Installing packages in the virtual environment..." -ForegroundColor Yellow

$VENV_PYTHON = Join-Path $VENV_DIR "Scripts\python.exe"
$VENV_PIP = Join-Path $VENV_DIR "Scripts\pip.exe"

if (-not (Test-Path $VENV_PYTHON)) {
    Write-Host "      Error: Virtual environment Python not found at $VENV_PYTHON" -ForegroundColor Red
    exit 1
}

try {
    # Upgrade pip first
    Write-Host "      Upgrading pip..." -ForegroundColor Cyan
    & $VENV_PYTHON -m pip install --upgrade pip --quiet
    
    # Install nexaai
    Write-Host "      Installing nexaai..." -ForegroundColor Cyan
    & $VENV_PIP install nexaai --force-reinstall --no-cache-dir
    
    # Install numpy
    Write-Host "      Installing numpy..." -ForegroundColor Cyan
    & $VENV_PIP install numpy
    
    # Install pre-built pywinpty wheel for ARM64 Windows
    Write-Host "      Installing pywinpty (pre-built wheel for ARM64)..." -ForegroundColor Cyan
    $PYWINPTY_WHEEL_URL = "https://nexa-model-hub-bucket.s3.us-west-1.amazonaws.com/public/nexa_sdk/downloads/pywinpty-2.0.12-cp311-none-win_arm64.whl"
    & $VENV_PIP install $PYWINPTY_WHEEL_URL
    
    # Install Jupyter Notebook and ipykernel
    Write-Host "      Installing Jupyter Notebook and ipykernel..." -ForegroundColor Cyan
    & $VENV_PIP install jupyter notebook ipykernel
    
    Write-Host "      All packages installed successfully" -ForegroundColor Green
} catch {
    Write-Host "      Error installing packages: $_" -ForegroundColor Red
    exit 1
}

# Step 5: Register virtual environment as Jupyter kernel
Write-Host "[5/5] Registering virtual environment as Jupyter kernel..." -ForegroundColor Yellow

$KERNEL_NAME = "nexa-sdk-arm64"
$KERNEL_DISPLAY_NAME = "Python (Nexa SDK ARM64)"

try {
    & $VENV_PYTHON -m ipykernel install --user --name=$KERNEL_NAME --display-name="$KERNEL_DISPLAY_NAME"
    Write-Host "      Kernel '$KERNEL_DISPLAY_NAME' registered successfully" -ForegroundColor Green
} catch {
    Write-Host "      Error registering Jupyter kernel: $_" -ForegroundColor Red
    exit 1
}

# Summary
Write-Host ""
Write-Host "================================================" -ForegroundColor Cyan
Write-Host "Setup Complete!" -ForegroundColor Green
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Python Installation: $PYTHON_INSTALL_DIR" -ForegroundColor White
Write-Host "Virtual Environment: $VENV_DIR" -ForegroundColor White
Write-Host "Jupyter Kernel Name: $KERNEL_NAME" -ForegroundColor White
Write-Host ""
Write-Host "To activate the virtual environment, run:" -ForegroundColor Yellow
Write-Host "  .\venv\Scripts\Activate.ps1" -ForegroundColor Cyan
Write-Host ""
Write-Host "To start Jupyter Notebook:" -ForegroundColor Yellow
Write-Host "  .\venv\Scripts\jupyter-notebook.exe" -ForegroundColor Cyan
Write-Host ""
Write-Host "Or use the Python directly:" -ForegroundColor Yellow
Write-Host "  .\venv\Scripts\python.exe" -ForegroundColor Cyan
Write-Host ""
Write-Host "The kernel '$KERNEL_DISPLAY_NAME' is now available in Jupyter!" -ForegroundColor Green
Write-Host ""

# Activate the virtual environment in the current session
Write-Host "Activating virtual environment in current session..." -ForegroundColor Yellow
$ACTIVATE_SCRIPT = Join-Path $VENV_DIR "Scripts\Activate.ps1"

if (Test-Path $ACTIVATE_SCRIPT) {
    # Use dot-sourcing to run the activation script in the current scope
    . $ACTIVATE_SCRIPT
    Write-Host "Virtual environment activated!" -ForegroundColor Green
    Write-Host "You can now use 'python' and 'pip' commands directly." -ForegroundColor Cyan
} else {
    Write-Host "Warning: Activation script not found at $ACTIVATE_SCRIPT" -ForegroundColor Red
    Write-Host "You can manually activate it later using: .\venv\Scripts\Activate.ps1" -ForegroundColor Yellow
}