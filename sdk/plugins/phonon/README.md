# Experimental Phonon TTS backend

This contribution adds a speech API and an opt-in `phonon_qnn` plugin. It runs
Gradium's metadata-driven Phonon generator through QNN HTP, including its prefill
and step graphs. It does not route Phonon through a causal language model or the
Genie dialog engine. A GGUF file alone is not an input to this backend.

## Scope and dependencies

See [platform portability](PORTABILITY.md) for the shared API, Android/Linux
requirements and the distinction between GGUF weights and compiled QNN bundles.

- Tested: Phonon 31M English and 40M English compiled for SC8380XP on Windows
  ARM64, QAIRT **2.45.0.260326**, default bundled voice, 24 kHz mono audio.
- `device="npu"` is mandatory. CPU and GPU choices fail; there is no fallback.
  Text normalization, scheduling, PCM copies and playback still use host CPU.
- `verify_npu=True` also requires positive **per-request** accelerator profiling
  times for both graphs. Set `PHONON_QNN_PROFILE=1` before creating the model.
- The public C API is in `sdk/include/geniex_tts.h`, with Python, Go and Android
  FFI surfaces. Android Kotlin/JNI compile checks do not establish Android
  hardware support. Linux and VENTUNO Q execution are not tested here.
- Other languages and future Phonon checkpoints may use the same bundle format,
  but must be compiled and tested individually. They are not advertised as tested.
- Model acquisition, AI Hub export/compilation, a `geniex tts` CLI, OpenAI speech
  endpoints and automatic Hugging Face model resolution are outside this patch.

Supply the Gradium source tree explicitly as `GENIEX_PHONON_SOURCE`. The tested
revision is [`b2896b91f138edb5a4aa56277bd6b63091fffa73`](https://github.com/shivaylamba/xn-ptts/commit/b2896b91f138edb5a4aa56277bd6b63091fffa73),
the Windows QNN portability work in [Gradium PR #243](https://github.com/gradium-ai/xn-ptts/pull/243).
That dependency needs an accepted Windows support revision before a final
upstream release. No third-party source or weights are copied into GenieX.
The optional linked Gradium code is used under its MIT license; its license and
the nlohmann JSON license are installed under `licenses/phonon_qnn`.
Gradium's native engine currently writes its own QNN diagnostics to stderr;
messages added by this GenieX plugin use `GENIEX_LOG_*`.

## Build on Windows ARM64

Use native ARM64 Rust/Cargo, CMake, MSVC ARM64, and clang on PATH. Follow
[`notes/build.md`](../../../notes/build.md) for the Windows toolchain. The isolated
build below disables the existing LLM plugins, so it does not require their
Hexagon/OpenCL builds or driver-signing setup. Long path support should be enabled.
No system configuration is changed by these scripts.

Build the trusted Rust text frontend first, from the reviewed Gradium checkout:

```powershell
cargo build --manifest-path C:\src\xn-ptts\qnn\text-ffi\Cargo.toml `
  --target aarch64-pc-windows-msvc --release --locked

& .\sdk\plugins\phonon\build-windows.ps1 `
  -PhononSource C:\src\xn-ptts `
  -QairtRoot C:\Qualcomm\AIStack\QAIRT\2.45.0.260326 `
  -BuildDirectory C:\build\geniex-phonon `
  -InstallDirectory C:\build\geniex-phonon-sdk
```

Use bundles whose compile **and link** QAIRT versions match the runtime; a bundle
for another SoC or a newer incompatible QNN context version must fail to load.
Obtain the models under their original licenses and access terms. Phonon 31M
and 40M do not share the same license. Do not put weights or SDK binaries in Git.

## Synthesize and listen

Install the Python binding's dependencies (`python -m pip install tqdm pytest`).
From the GenieX checkout:

```powershell
& .\sdk\plugins\phonon\run-windows.ps1 `
  -SdkDirectory C:\build\geniex-phonon-sdk `
  -QairtRoot C:\Qualcomm\AIStack\QAIRT\2.45.0.260326 `
  -TextLibrary C:\src\xn-ptts\qnn\text-ffi\target\aarch64-pc-windows-msvc\release\ptts_text.dll `
  -Models C:\models\phonon31,C:\models\phonon40 `
  -OutDirectory C:\results\geniex-phonon -Play
```

This creates six WAV files and a `results.json` per model and plays the first
utterance for each. `-Play` uses Windows audio playback after generation; first
PCM callback time is **not** speaker latency. RTF is generation time divided by
audio duration. Model-load time is reported separately. Callbacks are synchronous
and slow callbacks increase generation time. These are local measurements, not
a hardware performance guarantee.

Python applications can use the API directly after setting the SDK/plugin and
trusted runtime loader paths, as shown in `examples/python/phonon_tts.py`:

```python
from geniex import GenieXTTS

with GenieXTTS(bundle, runtime_path=qairt_libs, text_library=text_dll,
              soc_model="SC8380XP", verify_npu=True) as speech:
    result = speech.synthesize("Welcome to the local voice assistant.")
    # result.pcm is float32 mono PCM; stream via on_audio=callback if desired.
    assert result.prefill_npu_ms > 0 and result.step_npu_ms > 0
```

## Contracts, lifecycle and compatibility

`geniex_init` precedes create; all handles must be destroyed before
`geniex_deinit`. Destroying a handle during synthesis is unsupported. The core
leaves the existing `Plugin` vtable untouched and discovers the optional
`geniex_create_tts_v1` factory export. Old plugins without that symbol return
`NOT_SUPPORTED` when asked for TTS; existing LLM/VLM interfaces are unchanged.
Legacy in-process `geniex_register_plugin` has no TTS export table and therefore
cannot register a TTS factory in this version.

Synthesis returns borrowed float PCM in the C callback; bindings copy it before
returning. Callback `false` or `cancel()` stops between graph executions and
returns cancellation status with partial audio. A graph execution itself cannot
be interrupted. Callback exceptions are contained by language bindings. Python
rejects overlapping synthesis calls; the native Phonon backend also rejects a
busy handle. Do not call `close`/`destroy` from a callback. A native graph failure
invalidates that model session; recreate it rather than retry on another device.

The C structs use `struct_size` guards; mismatched layouts fail before dispatch.
NPU graph counters are cumulative inside the engine and are converted to deltas
at this boundary, including cancelled requests.

## Test commands

`build-windows.ps1` builds and runs native C/C++ model-free tests: malformed
inputs, struct guards, old plugin compatibility, PCM integrity, callback and
cross-thread cancellation, reuse and exception containment.

For Python/Go contract tests, point `GENIEX_LIB_PATH` at the new DLL and
`GENIEX_PLUGIN_PATH` at `<build>/tests/tts/scan-Release`, then set
`GENIEX_TTS_CONTRACT_TEST=1`:

```powershell
$env:PYTHONPATH = "$PWD\bindings\python"
python -m pytest .\bindings\python\tests\unit\test_tts.py
# Go needs cgo with an ARM64-capable clang, SDK include/library paths, and the
# new geniex.dll on PATH, as for the existing Go SDK binding.
cd bindings\go
go test -run '^TestTTSContracts$' -v .
```

For hardware checks, restore the real installed plugin path, set profiling and
the QAIRT DSP library path, then run:

```powershell
python .\sdk\plugins\phonon\test_device.py `
  --models C:\models\phonon31 C:\models\phonon40 `
  --runtime C:\Qualcomm\AIStack\QAIRT\2.45.0.260326\lib\aarch64-windows-msvc `
  --text-library C:\trusted\ptts_text.dll --soc SC8380XP `
  --out C:\results\geniex-phonon\device-tests.json
```

This checks CPU/GPU rejection, wrong SoC, missing runtime/frontend, verification
without profiling, cancellation, callback exceptions, reuse, noncumulative
profiles, multichunk generation and invalid voices. No network is used at
inference time.
