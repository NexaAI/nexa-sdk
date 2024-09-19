import argparse
import json
import logging
import os
import sys
import datasets
import subprocess
import time
import requests
from nexa.eval import evaluator, utils
from nexa.eval.loggers import EvaluationTracker, WandbLogger
from nexa.eval.tasks import TaskManager
from nexa.eval.utils import handle_non_serializable, make_table, simple_parse_args_string

def evaluate_model(args):
    eval_logger = utils.eval_logger
    eval_logger.setLevel(getattr(logging, f"{args.verbosity}"))
    eval_logger.info(f"Verbosity set to {args.verbosity}")
    os.environ["TOKENIZERS_PARALLELISM"] = "false" 

    if args.wandb_args:
        wandb_logger = WandbLogger(**simple_parse_args_string(args.wandb_args))

    if args.output_path:
        args.hf_hub_log_args += f",output_path={args.output_path}"
    if os.environ.get("HF_TOKEN", None):
        args.hf_hub_log_args += f",token={os.environ.get('HF_TOKEN')}"
    
    evaluation_tracker_args = simple_parse_args_string(args.hf_hub_log_args)
    evaluation_tracker = EvaluationTracker(**evaluation_tracker_args)
    
    task_manager = TaskManager(args.verbosity, include_path=args.include_path)
    
    if args.tasks is None:
        eval_logger.error("Need to specify task to evaluate.")
        sys.exit()
    else:
        task_list = args.tasks.split(",")
        task_names = task_manager.match_tasks(task_list)
    
    eval_logger.info(f"Selected Tasks: {task_names}")

    request_caching_args = evaluator.request_caching_arg_to_dict(cache_requests=args.cache_requests)

    results = evaluator.simple_evaluate(
        model=args.model,
        model_args=args.model_args,
        tasks=task_names,
        num_fewshot=args.num_fewshot,
        batch_size=args.batch_size,
        max_batch_size=args.max_batch_size,
        device=args.device,
        use_cache=args.use_cache,
        limit=args.limit,
        check_integrity=args.check_integrity,
        write_out=args.write_out,
        log_samples=args.log_samples,
        evaluation_tracker=evaluation_tracker,
        system_instruction=args.system_instruction,
        apply_chat_template=args.apply_chat_template,
        fewshot_as_multiturn=args.fewshot_as_multiturn,
        gen_kwargs=args.gen_kwargs,
        task_manager=task_manager,
        verbosity=args.verbosity,
        predict_only=args.predict_only,
        random_seed=args.seed[0],
        numpy_random_seed=args.seed[1],
        torch_random_seed=args.seed[2],
        fewshot_random_seed=args.seed[3],
        **request_caching_args,
    )

    # output_file = "test_results.json"
    # with open(output_file, 'w') as f:
    #     json.dump(results, f, indent=4)

    # print(f"Results have been saved to {output_file}")

    if results is not None:
        if args.log_samples:
            samples = results.pop("samples")
        dumped = json.dumps(
            results, indent=2, default=handle_non_serializable, ensure_ascii=False
        )
        if args.show_config:
            print(dumped)

        batch_sizes = ",".join(map(str, results["config"]["batch_sizes"]))
        evaluation_tracker.save_results_aggregated(results=results, samples=samples if args.log_samples else None)

        if args.log_samples:
            for task_name, config in results["configs"].items():
                evaluation_tracker.save_results_samples(task_name=task_name, samples=results["samples"][task_name])
           
        print(
            f"{args.model} ({args.model_args}), gen_kwargs: ({args.gen_kwargs}), limit: {args.limit}, num_fewshot: {args.num_fewshot}, "
            f"batch_size: {args.batch_size}{f' ({batch_sizes})' if batch_sizes else ''}"
        )

        print(make_table(results))
        if "groups" in results:
            print(make_table(results, "groups"))