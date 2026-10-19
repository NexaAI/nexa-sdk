# QAIRT per-operator profiling

Set the output path before creating the model:

```c
geniex_set_qairt_optrace_path("qnn-profiling-data.log");
geniex_llm_create(&input, &llm);
```

The log contains profiling records for all context-binary shards and can be
opened with `qnn-profile-viewer`. Capture requires a QAIRT 2.48+ runtime; use
`geniex_set_qairt_runtime_path` when the bundled runtime is older.

For VLMs, optrace covers the language component.

Pass `NULL` or `""` to disable profiling for subsequently created models.
