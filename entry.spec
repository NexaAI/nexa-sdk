# -*- mode: python ; coding: utf-8 -*-
import sys

# Determine the correct file extension based on the operating system
if sys.platform.startswith('win'):
    lib_extension = '.dll'
elif sys.platform.startswith('darwin'):
    lib_extension = '.dylib'
else:  # Linux and other Unix-like systems
    lib_extension = '.so'

a = Analysis(['./nexa/cli/entry.py'],
             pathex=[],
             binaries=[
                 (f'./nexa/gguf/lib/libggml_llama{lib_extension}', './nexa/gguf/lib'),
                 (f'./nexa/gguf/lib/libllama{lib_extension}', './nexa/gguf/lib'),
                 (f'./nexa/gguf/lib/libllava{lib_extension}', './nexa/gguf/lib'),
                 (f'./nexa/gguf/lib/libstable-diffusion{lib_extension}', './nexa/gguf/lib')
             ],
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