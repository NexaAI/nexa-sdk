# HumanEval: Hand-Written Evaluation Set

This is an evaluation harness for the HumanEval problem solving dataset
described in the paper "[Evaluating Large Language Models Trained on
Code](https://arxiv.org/abs/2107.03374)".

## Installation

Make sure to use python 3.7 or later:

```
$ conda create -n codex python=3.7
$ conda activate codex
```

Check out and install this repository:

```
$ git clone https://github.com/openai/human-eval
$ pip install -e human-eval
```

## Usage

**This program exists to run untrusted model-generated code. Users are strongly
encouraged not to do so outside of a robust security sandbox. The [execution
call](https://github.com/openai/human-eval/blob/master/human_eval/execution.py#L48-L58)
in `execution.py` is deliberately commented out to ensure users read this
disclaimer before running code in a potentially unsafe manner. See the comment in
`execution.py` for more information and instructions.**

After following the above instructions to enable execution, generate samples
and save them in the following JSON Lines (jsonl) format, where each sample is
formatted into a single line like so:

```
{"task_id": "Corresponding HumanEval task ID", "completion": "Completion only without the prompt"}
```

We provide `example_problem.jsonl` and `example_solutions.jsonl` under `data`
to illustrate the format and help with debugging.

Here is nearly functional example code (you just have to provide
`generate_one_completion` to make it work) that saves generated completions to
`samples.jsonl`.

```
from human_eval.data import write_jsonl, read_problems

problems = read_problems()

num_samples_per_task = 200
samples = [
    dict(task_id=task_id, completion=generate_one_completion(problems[task_id]["prompt"]))
    for task_id in problems
    for _ in range(num_samples_per_task)
]
write_jsonl("samples.jsonl", samples)
```

To evaluate the samples, run

```
$ evaluate_functional_correctness samples.jsonl
Reading samples...
32800it [00:01, 23787.50it/s]
Running test suites...
100%|...| 32800/32800 [16:11<00:00, 33.76it/s]
Writing results to samples.jsonl_results.jsonl...
100%|...| 32800/32800 [00:00<00:00, 42876.84it/s]
{'pass@1': ..., 'pass@10': ..., 'pass@100': ...}
```

This script provides more fine-grained information in a new file ending in
`<input_path>_results.jsonl`. Each row now contains whether the completion
`passed` along with the execution `result` which is one of "passed", "timed
out", or "failed".

As a quick sanity-check, the example samples should yield 0.5 pass@1.

```
$ evaluate_functional_correctness data/example_samples.jsonl --problem_file=data/example_problem.jsonl
Reading samples...
6it [00:00, 3397.11it/s]
Running example suites...
100%|...| 6/6 [00:03<00:00,  1.96it/s]
Writing results to data/example_samples.jsonl_results.jsonl...
100%|...| 6/6 [00:00<00:00, 6148.50it/s]
{'pass@1': 0.4999999999999999}
```

Because there is no unbiased way of estimating pass@k when there are fewer
samples than k, the script does not evaluate pass@k for these cases. To
evaluate with other k values, pass `--k=<comma-separated-values-here>`. For
other options, see

```
$ evaluate_functional_correctness --help
```

However, we recommend that you use the default values for the rest.

## Known Issues

While evaluation uses very little memory, you might see the following error
message when the system is running out of RAM. Since this may cause some
correct programs to fail, we recommend that you free some memory and try again.

```
malloc: can't allocate region
```

## Citation

Please cite using the following bibtex entry:

```
@article{chen2021codex,
  title={Evaluating Large Language Models Trained on Code},
  author={Mark Chen and Jerry Tworek and Heewoo Jun and Qiming Yuan and Henrique Ponde de Oliveira Pinto and Jared Kaplan and Harri Edwards and Yuri Burda and Nicholas Joseph and Greg Brockman and Alex Ray and Raul Puri and Gretchen Krueger and Michael Petrov and Heidy Khlaaf and Girish Sastry and Pamela Mishkin and Brooke Chan and Scott Gray and Nick Ryder and Mikhail Pavlov and Alethea Power and Lukasz Kaiser and Mohammad Bavarian and Clemens Winter and Philippe Tillet and Felipe Petroski Such and Dave Cummings and Matthias Plappert and Fotios Chantzis and Elizabeth Barnes and Ariel Herbert-Voss and William Hebgen Guss and Alex Nichol and Alex Paino and Nikolas Tezak and Jie Tang and Igor Babuschkin and Suchir Balaji and Shantanu Jain and William Saunders and Christopher Hesse and Andrew N. Carr and Jan Leike and Josh Achiam and Vedant Misra and Evan Morikawa and Alec Radford and Matthew Knight and Miles Brundage and Mira Murati and Katie Mayer and Peter Welinder and Bob McGrew and Dario Amodei and Sam McCandlish and Ilya Sutskever and Wojciech Zaremba},
  year={2021},
  eprint={2107.03374},
  archivePrefix={arXiv},
  primaryClass={cs.LG}
}
```
