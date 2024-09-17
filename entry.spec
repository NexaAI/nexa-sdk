# -*- mode: python ; coding: utf-8 -*-
import sys

# Determine the correct file extension based on the operating system
if sys.platform.startswith('win'):
    lib_extensions = ['.dll', '.lib']
elif sys.platform.startswith('darwin'):
    lib_extensions = ['.dylib']
else:  # Linux and other Unix-like systems
    lib_extensions = ['.so']

binaries = []

if sys.platform.startswith('win'):
    for lib_name in ['ggml_llama', 'llama', 'llava', 'stable-diffusion']:
        for ext in lib_extensions:
            binaries.append((f'./nexa/gguf/lib/{lib_name}{ext}', './nexa/gguf/lib'))
else:
    for lib_name in ['libggml_llama', 'libllama', 'libllava', 'libstable-diffusion']:
        binaries.append((f'./nexa/gguf/lib/{lib_name}{lib_extensions[0]}', './nexa/gguf/lib'))

a = Analysis(['./nexa/cli/entry.py'],
             pathex=[],
             binaries=binaries,
             datas=[],
             hiddenimports=[],
             hookspath=[],
             runtime_hooks=[],
             excludes=[],
             win_no_prefer_redirects=False,
             win_private_assemblies=False,
             cipher=None,
             noarchive=False)

pyz = PYZ(a.pure)

exe = EXE(
    pyz,
    a.scripts,
    [],
    exclude_binaries=True,
    name='nexa',
    debug=False,
    bootloader_ignore_signals=False,
    strip=False,
    upx=True,
    console=True,
    disable_windowed_traceback=False,
    argv_emulation=False,
    target_arch=None,
    codesign_identity=None,
    entitlements_file=None,
)

coll = COLLECT(
    exe,
    a.binaries,
    a.datas,
    strip=False,
    upx=True,
    upx_exclude=[],
    name='nexa',
)