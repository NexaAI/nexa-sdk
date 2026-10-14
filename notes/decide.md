# Structured decisions

`geniex decide` produces one locally validated, non-executing JSON decision:

```text
geniex decide <model> <classify|choose>
```

## Classification and multiple choice

```bash
# Text-only classification accepts an LLM or VLM.
geniex decide <model> classify \
  --label capital_of_france \
  --label capital_of_germany \
  --label unknown \
  --instruction "Choose the label that correctly answers the question." \
  --context "What is the capital of France?"

# Multiple choice returns stable option IDs, not display text.
geniex decide <model> choose \
  --option paris=Paris \
  --option berlin=Berlin \
  --instruction "Select the capital of France." \
  --context "Question: What is the capital of France?"
```

Add repeatable `--image` for a visual decision; images require a VLM. Add
`--confidence` only when advisory model-reported confidence is useful.

- **Classification** returns exactly one stable label ID supplied with `--label`.
- **Multiple choice** returns a bounded, duplicate-free list of stable option IDs
  supplied with `--option`.
- **Typed extraction** strictly decodes and validates the result locally.

Use stable IDs for selectable values. Option text and supplied context are
untrusted content, not model instructions. For llama.cpp, the command uses a
small generated grammar that enumerates only the supplied IDs; other runtimes
use JSON mode plus the same strict local validation.

`--benchmark <repetitions>` performs one discarded warmup, resets the loaded
model between attempts, writes timing JSON to standard error, and preserves the
decision JSON on standard output. `--threads`, `--threads-batch`, `--batch`, and
`--ubatch` are opt-in controls that default to the SDK values.
