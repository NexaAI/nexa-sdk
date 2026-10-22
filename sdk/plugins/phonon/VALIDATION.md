# Local validation: 22 October 2026

GenieX base: `9d456b3131f799e651d87e769ef56e046ff96ac3`.
Local branch: `feat/phonon-tts`. No Qualcomm PR has been opened.

Hardware: Snapdragon X Elite Windows ARM64 laptop, SC8380XP, Windows build 26200.
QAIRT host and v73 DSP runtime: `2.45.0.260326`.
Native build: MSVC ARM64 19.50, Release; Rust ARM64 model-manager; Gradium source
`b2896b91f138edb5a4aa56277bd6b63091fffa73` supplied externally.

## Functional results

| Check | Result |
|---|---|
| Rebuilt current GenieX SDK and optional Phonon plugin | Passed |
| Native C/C++ API contracts, including C header compilation | Passed |
| Existing Python unit suite plus new TTS tests | **62 passed** |
| Go cgo build, streaming, cancellation, panic containment and reuse | Passed |
| Go real NPU synthesis, both English models | Passed, two requests per model |
| Android Kotlin class and ARM64 Android JNI object compilation | Passed |
| Kotlin result constructor vs JNI signature | `(IIJDDDDII)V` verified |
| Real-device regression scenarios | **26 passed** across both models |
| Clang-format 18.1.8 and Ruff 0.6.9 checks, gofmt, git whitespace checks | Passed |
| GitHub Actions execution | Not run; model-free Windows workflow added |

Device scenarios cover NPU-only enumeration, CPU/GPU rejection, wrong SoC,
missing runtime, missing native text frontend, verification without profiling,
callback and explicit cancellation, callback exceptions, model reuse,
noncumulative graph counters, multichunk speech, and invalid voices.

Both graphs report positive
`QNN_HTP_PROFILE_EVENTTYPE_GRAPH_EXECUTE_ACCEL_TIME_MICROSEC` on every completed
measured request. These are accelerator events, not wall-clock timing alone.
The plugin fixes the engine backend to `htp`; it does not retry another backend.
The tests do not claim that text normalization or playback runs on the NPU.

## Measured latency

Text: “Hello from Phonon, running locally through GenieX on the Snapdragon NPU.”
Default voice, 24 kHz mono. Six consecutive requests per model; medians below
use the last five requests after the model has been loaded. Output durations
vary slightly with generation. Audio was checked for valid format, sample
counts and nonsilence, not scored for naturalness.

| English model | Initial warm first PCM median | Initial warm RTF median | Repeat warm first PCM median | Repeat warm RTF median |
|---|---:|---:|---:|---:|
| Phonon 31M | 10.66 ms | 0.0595 | 15.12 ms | 0.0631 |
| Phonon 40M | 12.44 ms | 0.0645 | 23.78 ms | 0.0817 |

The independent Python callback timer in the repeat measured **15.30 ms** and
**23.93 ms** respectively. SDK/model initialization was about **1.02 s** and
**1.07 s** in that repeat and is excluded from first-PCM latency.
The repeat's 40M warm range was **16.04–81.46 ms**. Laptop load and scheduling
affect latency; these measurements are not a fixed 14 ms guarantee. A separate
run with overlapping test activity is retained as a load observation, not used
as the single-model benchmark. Speaker/output-device latency is not measured.

## Evidence and reproducibility

The working session retains JSON, WAV and logs outside the source checkout in
`outputs/geniex-phonon-upstream/`: initial `31/` and `40/` results, `latest/`
repeat results, `device-tests.json`, Python test/build/Go logs, and a local
`test-local.ps1` launcher. Paths are local artifacts, not files expected in a
fresh upstream clone. Portable setup and test commands are in [README.md](README.md).

## Before an upstream PR

The proposed C speech API and optional factory export need maintainer review;
this is a working contribution candidate, not an accepted upstream API.
The Windows Gradium engine dependency currently comes from its Windows support
PR. Agree on an upstream dependency revision/package strategy before release.

No CLI speech command, model catalogue entry, mixed LLM/TTS session test,
Linux/VENTUNO Q execution, Android device test, or non-English model validation
is claimed. No model weights, QAIRT libraries, credentials or generated audio
are included in this contribution.
