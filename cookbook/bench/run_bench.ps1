<# 
Function-Calling Benchmark Script
Runs performance benchmarks for NexaAI models with varying system prompt sizes

Usage:
    .\run_bench.ps1
    
Requirements:
    - nexa CLI tool in ../runner/build/nexa
    - Prompt files: prompt500.txt, prompt1000.txt, prompt2000.txt, prompt4000.txt
    
Output:
    - nexa_infer_NPU_*.log (NPU model results)
    - nexa_infer_GGUF_*.log (GGUF GPU variant results)
    - nexa_infer_GGUF_ngl0_*.log (GGUF CPU-only variant results)
#>

$ErrorActionPreference = "Stop"
$PromptSizes = 500, 1000, 2000, 4000
$TestPrompt = 'send_email to me@123.com say i have arrived'

Write-Host "🚀 Starting Function-Calling Benchmarks" -ForegroundColor Green
Write-Host "Prompt sizes: $($PromptSizes -join ', ') tokens" -ForegroundColor Cyan
Write-Host ""

# Test NPU model
Write-Host "📊 Benchmarking NPU Model..." -ForegroundColor Yellow
foreach ($n in $PromptSizes) {
    $promptFile = "..\cookbook\bench\prompt${n}.txt"
    $outputFile = "nexa_infer_NPU_${n}.log"
    
    if (-not (Test-Path $promptFile)) {
        Write-Host "❌ Prompt file not found: $promptFile" -ForegroundColor Red
        continue
    }
    
    Write-Host "  Processing $n-token prompt..." -NoNewline
    & ".\build\nexa" infer "NexaAI/Qwen3-VL-4B-Instruct-NPU" -s $promptFile -p $TestPrompt | Out-File -FilePath $outputFile
    Write-Host " ✓" -ForegroundColor Green
}

Write-Host ""

# Test GGUF with GPU
Write-Host "📊 Benchmarking GGUF Model (GPU)..." -ForegroundColor Yellow
foreach ($n in $PromptSizes) {
    $promptFile = "..\cookbook\bench\prompt${n}.txt"
    $outputFile = "nexa_infer_GGUF_${n}.log"
    
    Write-Host "  Processing $n-token prompt..." -NoNewline
    & ".\build\nexa" infer "NexaAI/Qwen3-VL-4B-Instruct-GGUF" -s $promptFile -p $TestPrompt | Out-File -FilePath $outputFile
    Write-Host " ✓" -ForegroundColor Green
}

Write-Host ""

# Test GGUF with CPU only
Write-Host "📊 Benchmarking GGUF Model (CPU-only, ngl=0)..." -ForegroundColor Yellow
foreach ($n in $PromptSizes) {
    $promptFile = "..\cookbook\bench\prompt${n}.txt"
    $outputFile = "nexa_infer_GGUF_ngl0_${n}.log"
    
    Write-Host "  Processing $n-token prompt..." -NoNewline
    & ".\build\nexa" infer "NexaAI/Qwen3-VL-4B-Instruct-GGUF" --ngl 0 -s $promptFile -p $TestPrompt | Out-File -FilePath $outputFile
    Write-Host " ✓" -ForegroundColor Green
}

Write-Host ""
Write-Host "✅ Benchmarks complete!" -ForegroundColor Green
Write-Host "📝 Results saved to: nexa_infer_*.log files"
Write-Host "📊 See report.md for analysis and recommendations"