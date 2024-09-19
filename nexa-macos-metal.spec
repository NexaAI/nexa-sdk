# -*- mode: python ; coding: utf-8 -*-
import sys

if sys.platform.startswith('darwin'):
    lib_extensions = ['.dylib']
    binaries = []
    for lib_name in ['libggml_llama', 'libllama', 'libllava', 'libstable-diffusion']:
        binaries.append((f'./nexa/gguf/lib/{lib_name}{lib_extensions[0]}', './nexa/gguf/lib'))

    binaries.append(('./nexa/gguf/lib/empty_file.txt', './nexa/gguf/lib'))

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
        name='nexa-metal',
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

    app = BUNDLE(exe,
                 a.binaries,
                 a.datas,
                 name='nexa-metal.app',
                 icon=None,
                 bundle_identifier='com.nexaai.sdk',
                 info_plist={
                     'NSHighResolutionCapable': 'True',
                     'LSBackgroundOnly': 'False',
                     'NSRequiresAquaSystemAppearance': 'False',
                     'CFBundleShortVersionString': '1.0.0',
                 })
