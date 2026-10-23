# SDK pytest suite

End-to-end tests for the geniex SDK, driven through the public Python
binding so the suite doubles as a public-surface contract check.

```
tests/
├── assets/                # Real test images (quality_dog.jpg + NOTICE.md)
├── qdc/                   # QDC device dispatcher + per-platform entry scripts
├── test_api.py            # SDK metadata + resolve APIs (no model required)
├── test_llama_cpp.py      # llama_cpp plugin — LLM + VLM + precision + MTP
├── test_qairt.py          # qairt plugin     — LLM + VLM + precision
├── conftest.py            # Top-level fixtures (init, model paths, image)
├── pytest.ini             # Marker registry + suite discovery
├── _models.py             # Model-matrix loader for models.json
├── models.json            # The matrix: role -> [{id, precision, devices}]
└── _quality_data.py       # Keyword-quality prompts shared by both plugins
```

## What each plugin covers

Every plugin file ships the same behavioural cases, plus an explicit
model-manager-pull check that runs first:

| Test                             | What it proves                                                     |
|----------------------------------|--------------------------------------------------------------------|
| `test_model_manager_pull`        | AI Hub / HuggingFace pull for every model this file needs works.   |
| `test_llm_multi_turn`            | Two-turn "Alice" conversation; turn-2 must recall the name.        |
| `test_llm_greedy_is_deterministic` | Same prompt under two seeds decodes to byte-identical text.       |
| `test_vlm_multi_turn`            | Two-turn conversation with an image on turn 1.                     |
| `test_llm_quality_keywords`      | Greedy decode, 3 short Q/A; keyword substring must land.           |
| `test_vlm_quality_keywords`      | Golden-retriever caption must match one of the canonical keywords. |
| `test_mtp_multi_turn`            | (llama_cpp only) same "Alice" convo with `spec_type='draft-mtp'`.  |
| `test_mtp_qwen35_cache_lifecycle` | Qwen3.5 MTP stays active and matches clean greedy decoding after reset or prompt revision. |

Behavioural cases are parametrised over `(model, device_map)` pairs from
`models.json`, so which backends run is a property of the model entry.
Template / contract cases (ChatML sentinels, `enable_thinking`, tools
rendering, mtmd markers) stay pinned to the role's first entry — they assert
tokenizer specifics of one model. Model-manager pull failures are treated as
FAIL, not SKIP, so a broken download surfaces as a red CI leg instead of a
silent green skip.

Every generating cell passes `GREEDY_TEMPERATURE` (a negative argmax sentinel),
not `0.0`: both plugins read `0.0` as "unset" and substitute 0.8, so a pinned
seed alone does not make a cell deterministic.

## Marker registry

The conftest auto-tags items by location and `device_map` value:

| Marker          | Source                                                   |
|-----------------|----------------------------------------------------------|
| `api`           | items in `tests/test_api.py`                             |
| `llama_cpp`     | items in `tests/test_llama_cpp.py`                       |
| `qairt`         | items in `tests/test_qairt.py`                           |
| `device_cpu`    | parametrised with `device_map='cpu'`                     |
| `device_gpu`    | parametrised with `device_map='gpu'`                     |
| `device_npu`    | parametrised with `device_map='npu'`                     |
| `device_hybrid` | parametrised with `device_map='hybrid'`                  |
| `snapdragon`    | any `device_map` in {`gpu`, `npu`, `hybrid`} cell (auto) |
| `llm` / `vlm`   | applied per-test via `@pytest.mark.llm` / `.vlm`         |

`snapdragon`-marked and `qairt` items skip automatically unless
`GENIEX_DEVICE_TEST=1` is set **and** the host is a Snapdragon machine.
QAIRT models are pulled on demand from AI Hub (like the llama_cpp models
from HF), so the device shards need network access but no manual pre-pull.

## Running

```bash
# Anywhere — model-free API checks only
pytest tests -m api

# Anywhere — adds the llama_cpp CPU cells (downloads configured CPU test models on first run)
pytest tests -m "api or (llama_cpp and device_cpu)"

# Snapdragon Windows ARM64 or Qualcomm Linux — full matrix
# (pulls every models.json entry, ~30 GB including gpt-oss-20b and the MTP pair)
GENIEX_DEVICE_TEST=1 pytest tests
```

## Models

`tests/models.json` is the matrix. Each role maps to a list of entries and
every entry names the `devices` it runs on, so adding a model to the whole
behavioural suite is a manifest edit. Loader: `tests/_models.py`.

| Role                   | Model | Devices |
|------------------------|-------|---------|
| `llama_cpp_llm`        | `unsloth/Qwen3-4B-GGUF` Q4_0 | cpu, gpu, npu |
| `llama_cpp_llm`        | `unsloth/gpt-oss-20b-GGUF` Q4_0 | hybrid |
| `llama_cpp_vlm`        | `unsloth/gemma-4-E2B-it-GGUF` Q4_0 + mmproj-F16 | cpu, gpu, npu |
| `llama_cpp_mtp_target` | `google/gemma-4-26B-A4B-it-qat-q4_0-gguf` | npu |
| `llama_cpp_mtp_qwen35` | `bartowski/Qwen_Qwen3.5-0.8B-GGUF` Q8_0, including its MTP head | cpu, npu |
| `llama_cpp_mtp_draft`  | `RachidAR/gemma-4-...-assistant-q4_0-gguf` | — (paired with the target) |
| `qairt_llm`            | `qualcomm/Qwen3-4B` | npu |
| `qairt_vlm`            | `qualcomm/Qwen2.5-VL-7B-Instruct` | npu |

Entry fields: `id`, `precision`, `hub` (default `auto`), `devices`,
`quality_max_new_tokens` (per-model budget for the keyword cells),
`env_override` (an env var that replaces `id`).

### Qwen3.5 MTP cache regression

The Qwen3.5 cases use the same GGUF as target and drafter. Unlike Gemma's
shared-cache MTP path, this exercises `nextn.*` head loading and the separate
drafter KV cache. Four cases compare MTP output with a clean plain greedy
decode: explicit reset, identical prompt, edited question, and replacement
question. Each case requires nonzero draft-token statistics, so a silent
fallback to plain decoding fails the test. The unsupported QCS9075M HTP
cell skips, as in the existing Gemma MTP test; its CPU cell still runs.

The default matrix downloads the Q8_0 model on first use. To reuse a local
Qwen3.5-0.8B GGUF containing its MTP head:

```bash
GENIEX_TEST_MTP_MODEL=/path/to/Qwen3.5-0.8B-Q8_0.gguf \
  pytest tests/test_llama_cpp.py -k 'mtp_qwen35_cache_lifecycle and cpu'
```

llama_cpp and QAIRT share the same LLM/VLM models so a keyword-quality
divergence traces to backend / quantization rather than model identity.
`gpt-oss-20b` is the `hybrid`-only entry and needs a bigger
`quality_max_new_tokens` — its reasoning channel has no `enable_thinking` off
switch.

The primary LLM is Qwen3-4B **base**, not Instruct-2507: Instruct-2507 emits a
long `<think>` preamble before the answer that, on the 256-token budget the
suite uses, pushes the keyword off the end of the completion and turns
`test_llm_quality_keywords` into a thinking-budget test rather than a
backend-quality test.

Swap models via `env_override` per entry, or `GENIEX_TEST_MODELS` for the whole
matrix:

```bash
GENIEX_QAIRT_MODEL=qualcomm/<other-llm> \
GENIEX_QAIRT_VLM_MODEL=qualcomm/<other-vlm> \
GENIEX_DEVICE_TEST=1 pytest tests -m qairt

GENIEX_TEST_MODELS=/path/to/my-matrix.json GENIEX_DEVICE_TEST=1 pytest tests
```

## CI

The project splits into two test workflows:

- **Unit Test** ([`_unit-test.yml`](../.github/workflows/_unit-test.yml)) — runs on
  every PR via [`pr-check.yml`](../.github/workflows/pr-check.yml). Covers
  `test-go`, `test-python`, and `test-sdk-ci` (`-m api`) on
  linux-arm64 + windows-arm64 GitHub runners; no QDC hardware.
- **QDC Test** ([`_qdc-test.yml`](../.github/workflows/_qdc-test.yml))
  — runs on `workflow_dispatch` and on `v*` tag push (via
  [`test.yml`](../.github/workflows/test.yml)). Not on PR. One leg per
  platform; each leg runs the full plugin set (`-m "llama_cpp or qairt"`):

  | Platform | Device   | Framework   |
  |----------|----------|-------------|
  | Linux    | QCS9075M | BASH        |
  | Windows  | SC8480XP | POWERSHELL  |

Each platform's entry script lives under [qdc/](qdc/): Linux uses the
image's preinstalled python 3.12; Windows bootstraps python.org's arm64
embed zip. Both reuse
[sdk/benchmark/qdc/_qdc.py](../sdk/benchmark/qdc/_qdc.py) for submit /
poll / log-collect.

Android (SM8850) is not wired yet — QDC's SM8850 image has neither
python nor termux preinstalled, so it needs a CLI-driven harness that is
out of scope for this iteration.

## Boundary with `bindings/python/tests/`

This directory is the home of SDK + plugin coverage. Tests under
`bindings/python/tests/` cover the binding layer itself (CLI wrapper,
progress callbacks, model_manager Python surface, local pull paths) and
do not launch real generation. Anything that reasons about device
selection, plugin behaviour, or model output belongs here.
