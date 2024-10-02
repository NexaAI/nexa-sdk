# Do-Not-Answer: A Dataset for Evaluating Safeguards in LLMs

<p align="center" style="display: flex; flex-direction: row; justify-content: center; align-items: center">
📄 <a href="https://aclanthology.org/2024.findings-eacl.61" target="_blank" style="margin-right: 15px; margin-left: 10px">Paper</a> • 
🤗 <a href="https://huggingface.co/datasets/LibrAI/do-not-answer" target="_blank" style="margin-left: 10px">Dataset</a> •  
🔱 <a href="https://huggingface.co/LibrAI" target="_blank" style="margin-left: 10px">Evaluator</a>
</p>

> \[!NOTE\]
>
> For **Chinese** Do-Not-Answer, please refer to the [Chinese Do-Not-Answer Section](#a-chinese-dataset-for-evaluating-the-safeguards-in-large-language-models).

#### 📣 Our paper is featured in the latest [Stanford AI Index Report 2024](https://aiindex.stanford.edu/report/)

## Overview

_Do not answer_ is an open-source dataset to evaluate LLMs' safety mechanism at a low cost. The dataset is curated and filtered to consist only of prompts to which responsible language models do not answer.
Besides human annotations, _Do not answer_ also implements model-based evaluation, where a 600M fine-tuned BERT-like evaluator achieves comparable results with human and GPT-4.

## Instruction Taxonomy

Instructions are organized by a three-level hierarchical taxonomy, covering 61 specific harms.

<p align="center"> <img src="assets/dna.png" style="width: 85%;" id="title-icon"></p>

_Do not answer_ collects 939 instructions, the distribution across five risk areas and 12 harm types is shown below:

<p align="center"> <img src="assets/instruction_dist.png" style="width: 85%;" id="title-icon"></p>

## Response Assessment

We assess six models' responses (GPT-4, ChatGPT, Claude, LLaMA-2, Vicuna, ChatGLM2) from two levels.

- Harmfulness: whether a response is harmful (a binary label).
- Action category: responses are categorized into six distinct groups, each representing a specific response type.
<p align="center"> <img src="assets/action.png" style="width: 85%;" id="title-icon"></p>

### Human Evaluation

High percentage of safe responses across five risk areas for six models demonstrates that they are basically safe.
LLaMA-2 ranks the first, followed by ChatGPT, Claude, GPT-4, and Vicuna, with ChatGLM2 ranking the last by 85 harmful responses among 939.

- Harmlessness Rank
<p align="center"> <img src="assets/harmful_num.png" style="width: 60%;" id="title-icon"></p>

- Action Category Distribution
<p align="center"> <img src="assets/action_dist.png" style="width: 60%;" id="title-icon"></p>

### Automatic Evaluation

Longformer evaluators achieve comparable results with human and GPT-4.

- GPT-4 evaluation
- Fine-tuned Longformer classifier (600M)

| Model    | Human | GPT-4 | Longformer |
| :------- | ----: | ----: | ---------: |
| LLaMA-2  |  99.7 |  99.4 |       99.3 |
| ChatGPT  |  98.5 |  97.7 |       97.9 |
| Claude   |  98.3 |  98.3 |       97.6 |
| GPT-4    |  97.6 |  96.5 |       97.2 |
| vicuna   |  94.5 |  94.9 |       95.0 |
| ChatGLM2 |  90.9 |  92.9 |       92.9 |

## Usage

**NOTICE:** please fill in you API info into `do_not_answer\utils\info.yaml` before using commercial models such as GPT4.

Refer to [notebooks](./notebooks) for details and paper replication.

Refer to [datasets](./datasets) for the dataset used in our paper.

# A Chinese Dataset for Evaluating the Safeguards in Large Language Models

<p align="center" style="display: flex; flex-direction: row; justify-content: center; align-items: center">
📄 <a href="https://arxiv.org/pdf/2402.12193" target="_blank" style="margin-right: 15px; margin-left: 10px">Paper</a> • 
🤗 <a href="https://github.com/Libr-AI/do-not-answer/tree/main/cdna" target="_blank" style="margin-left: 10px">Dataset</a> •  
🔱 <a href="https://github.com/Libr-AI/do-not-answer/tree/main/do_not_answer/evaluator" target="_blank" style="margin-left: 10px">Evaluator</a>
</p>

## Overview

We construct a Chinese LLM safety evaluation by translating and localizing the [_"Do-not-answer"_](https://github.com/Libr-AI/do-not-answer/) dataset and expand it with region-specific questions and align it with country-specific AI generation regulations,

We then extend the resulting 1,014 questions from two prespectives:

1.  **False Negative(FN) questions**: risky questions posed in an
    evasive way, aimed at evaluating an LLM’s sensitivity to perceiving risks, aimed at evaluating an LLM’s sensitivity to perceiving risks
2.  **False Positive(FP) questions**: harmless questions
    containing seemingly risky words, aimed at assessing whether the model is oversensitive, which can limit its helpfulness

This yields 3,042 Chinese questions for evaluating the risk perception and sensitivity to specific words and phrases of LLMs.

## Risk Taxonomy & Statistics

Questions are organized by a two-level hierarchical taxonomy, covering 6 risk types 17 specific harm types.

<p align="center"> <img src="assets/instruction_dist_chinese.png" style="width: 85%;" id="title-icon"></p>

## Response Assessment

To evaluate the responses of LLMs to the questions, we use GPT-4 for automatic evluation and human evaluation to check the gap between human and GPT-4 based evaluation.

### Automatic Evaluation

We use two strategies for automatic evaluation of the model responses:

1. **Question Set Evaluator:** Given a risk area, we
   ask 1–5 risk-specific questions to identify whether
   a response touches on unsafe areas.

2. **Action Type Evaluator:** Given a response, we ask the LLM to select one of the six fine-grained categories of actions how an LLM responds to a question. This is the similar to the auto evaluation in the original [_"Do-not-answer"_](https://github.com/Libr-AI/do-not-answer/) dataset.

#### Overall Results

<p align="center"> <img src="assets/harmful_num_chinese.png" style="width: 60%;" id="title-icon"></p>
Number of harmful responses for five different Chinese LLMs. We can see that LLaMA2, as an
English-centric model, is the safest LLM when testing
using English direct questions from the Do-not-answer
dataset, but it is also the least safe one when evaluated
using our Chinese-centric questions.

#### Results on different risk types

<p align="center"> <img src="assets/risk_type_num_chinese.png" style="width: 60%;" id="title-icon"></p>

LLMs that were mostly pre-trained on Chinese text are better at understanding Chinese questions and offering better answers than
models mainly pre-trained on English text. English-centric LLaMA-2 series perform worse than Chinese-centric models on Chinese safety datasets, especially for
region-specific questions (Type _VI_).

#### Results on different question types

<p align="center"> <img src="assets/question_type_num_chinese.png" style="width: 60%;" id="title-icon"></p>

The number of harmful responses of the indirect-attack questions (yellow bar) is the largest, followed by questions of the direct attack (blue bar). For LLaMA2 and Qwen, the the number of harmful
responses over direct attack and general questions is very close to each other, which suggests that
some chatbots were likely instruction-tuned to be too sensitive to certain words.

### Human Evaluation

<p align="center"> <img src="assets/human_eval.png" style="width: 80%;" id="title-icon"></p>

GPT-4 can identify the majority of safe responses
correctly, but demonstrates random guess performance on harmful responses. For action classification, responses falling into categories of 3 and 4
tend to be classified as 5 by GPT-4. This to some
extent implies that human makes more fine-grained
distinctions between different responding patterns
than GPT-4.

## Citation

If our work is useful for your own, you can cite us with the following BibTex entry:

```bibtex
@inproceedings{wang-etal-2024-answer,
    title = "Do-Not-Answer: Evaluating Safeguards in {LLM}s",
    author = "Wang, Yuxia  and
      Li, Haonan  and
      Han, Xudong  and
      Nakov, Preslav  and
      Baldwin, Timothy",
    editor = "Graham, Yvette  and
      Purver, Matthew",
    booktitle = "Findings of the Association for Computational Linguistics: EACL 2024",
    month = mar,
    year = "2024",
    address = "St. Julian{'}s, Malta",
    publisher = "Association for Computational Linguistics",
    url = "https://aclanthology.org/2024.findings-eacl.61",
    pages = "896--911",
}
```

For Chinese Do-Not-Answer dataset, you can cite us with the following BibTex entry:

```bibtex
@article{wang2024chinese,
  title={A Chinese Dataset for Evaluating the Safeguards in Large Language Models},
  author={Wang, Yuxia and Zhai, Zenan and Li, Haonan and Han, Xudong and Lin, Lizhi and Zhang, Zhenxuan and Zhao, Jingru and Nakov, Preslav and Baldwin, Timothy},
  journal={to appear in ACL 2024 findings},
  year={2024}
}
```

## License

All datasets in this repository are released under the
[Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International License](http://creativecommons.org/licenses/by-nc-sa/4.0/). All source files in this repository are released under the Apache 2.0 license, the text of which can be found in the LICENSE file.
