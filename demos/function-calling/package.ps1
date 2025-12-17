# Find InnoSetup compiler
$iscc = $null

# Check common installation paths
$appDataPath = [Environment]::GetFolderPath("LocalApplicationData")
$commonProgramFiles = "C:\Program Files"

if (Test-Path "$appDataPath\Programs\Inno Setup 6\ISCC.exe") {
    $iscc = "$appDataPath\Programs\Inno Setup 6\ISCC.exe"
} elseif (Test-Path "$commonProgramFiles\Inno Setup 6\ISCC.exe") {
    $iscc = "$commonProgramFiles\Inno Setup 6\ISCC.exe"
} else {
    # Try to find in PATH
    $cmd = Get-Command iscc -ErrorAction SilentlyContinue
    if ($cmd) {
        $iscc = $cmd.Source
    }
}

# Check if InnoSetup was found
if (-not $iscc) {
    Write-Host "Error: InnoSetup not found. Please install InnoSetup 'winget install --id JRSoftware.InnoSetup -e -s winget'" -ForegroundColor Red
    exit 1
}

Write-Host "Found InnoSetup at: $iscc" -ForegroundColor Green

# Ensure installer directory exists
New-Item -ItemType Directory -Force -Path build\installer | Out-Null

# Run InnoSetup compiler
& $iscc function-calling-demo.iss

if ($LASTEXITCODE -ne 0) {
    Write-Host "InnoSetup compilation failed" -ForegroundColor Red
    exit 1
}

Write-Host "Installer created successfully" -ForegroundColor Green
