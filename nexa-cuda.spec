# -*- mode: python ; coding: utf-8 -*-
import os
import sys

os.environ['CMAKE_ARGS'] = '-DGGML_CUDA=ON -DSD_CUBLAS=ON'

a = Analysis(['./nexa/cli/entry.py'],
             pathex=[],
             binaries=[
                 ('./nexa/gguf/lib/empty_file.txt', './nexa/gguf/lib'),
                 ('./nexa/gguf/lib/libggml_llama.so', './nexa/gguf/lib'),
                 ('./nexa/gguf/lib/libllama.so', './nexa/gguf/lib'),
                 ('./nexa/gguf/lib/libllava.so', './nexa/gguf/lib'),
                 ('./nexa/gguf/lib/libstable-diffusion.so', './nexa/gguf/lib')
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
    name='nexa-cuda',
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
    name='nexa-cuda',
)
