# Android Scoped Storage Solution

## Problem
The original implementation tried to access model files directly from `/sdcard/Download/model/` using `File.canRead()`, which returns `false` on Android 10+ due to Scoped Storage restrictions.

## Root Cause
- Android 10+ introduced Scoped Storage
- Direct file access to external storage is restricted
- `File.canRead()` returns `false` even for existing files
- `chmod` commands are ineffective on FUSE-mounted `/sdcard` paths

## Solution: Storage Access Framework (SAF)

### Implementation
1. **Directory Picker Integration**: Added `Intent.ACTION_OPEN_DOCUMENT_TREE` to let users select model directory
2. **Directory Scanning**: Automatically scan selected directory for model files
3. **URI-based Access**: Use `ContentResolver.openInputStream()` to access files via URIs
4. **Temporary Copy**: Copy selected files to app's internal cache directory for SDK compatibility
5. **Automatic File Detection**: Automatically find LLM, VLM, and mmproj files in the directory

### Key Changes
- `openModelFilePicker()`: Opens system directory picker
- `scanDirectoryForModels()`: Scans selected directory for model files
- `initializeModelsWithURIs()`: Initializes models using SAF URIs
- `getFilePathFromURI()`: Copies files to accessible location
- Added "Select Model Directory" button in UI

### Usage
1. User clicks "Select Model Directory" button
2. System directory picker opens
3. User selects the directory containing model files
4. App automatically scans directory for model files
5. Files are copied to internal cache
6. Models are initialized with accessible file paths

### Benefits
- ✅ Works with Android 10+ Scoped Storage
- ✅ No permission issues
- ✅ User-friendly file selection
- ✅ Supports multiple file types
- ✅ Maintains SDK compatibility

## Technical Details
- Files are copied to `context.cacheDir/models/`
- Original URIs are preserved for future access
- Automatic file type detection:
  - LLM: Files containing "Qwen" with .gguf extension
  - VLM: Files containing "SmolVLM" with .gguf extension  
  - mmproj: Files containing "mmproj" with .json extension
- Error handling for file operations
