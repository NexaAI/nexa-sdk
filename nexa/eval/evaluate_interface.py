import argparse
import json
import logging
import os
import sys
import datasets
from nexa.eval import evaluator, utils
from nexa.eval.loggers import EvaluationTracker, WandbLogger
from nexa.eval.tasks import TaskManager
from nexa.eval.utils import handle_non_serializable, make_table, simple_parse_args_string

args = argparse.Namespace(
    model="my-local-completions",
    tasks="openai_humaneval",
    # model_args="pretrained=google/gemma-2b",
    model_args="base_url=http://0.0.0.0:8000/v1/completions",
    num_fewshot=None, 
    hf_hub_log_args="",
    batch_size=8,
    max_batch_size=None,
    device="cuda", 
    output_path=f"results/gemma/openai_humaneval", 
    limit=None,
    use_cache=None, 
    cache_requests=None,
    check_integrity=False,
    write_out=False,
    log_samples=False,
    system_instruction=None,
    apply_chat_template=False,
    fewshot_as_multiturn=False,
    include_path=None, 
    gen_kwargs=None,
    verbosity="INFO",
    predict_only=False,
    wandb_args=None,
    show_config=False,
    seed=[0, 1234, 1234, 1234], 
    trust_remote_code=False,
)

# Set up logging
eval_logger = utils.eval_logger
eval_logger.setLevel(getattr(logging, f"{args.verbosity}"))
eval_logger.info(f"Verbosity set to {args.verbosity}")
os.environ["TOKENIZERS_PARALLELISM"] = "false"  # Disable tokenizers parallelism

if args.wandb_args:
    wandb_logger = WandbLogger(**simple_parse_args_string(args.wandb_args))

# Set up output path and token for Hugging Face Hub
if args.output_path:
    args.hf_hub_log_args += f",output_path={args.output_path}"
if os.environ.get("HF_TOKEN", None):
    args.hf_hub_log_args += f",token={os.environ.get('HF_TOKEN')}"

# Parse arguments for evaluation tracker
evaluation_tracker_args = simple_parse_args_string(args.hf_hub_log_args)
evaluation_tracker = EvaluationTracker(**evaluation_tracker_args)

# Ensure output path is specified if logging or predicting only
if args.predict_only:
    args.log_samples = True
if (args.log_samples or args.predict_only) and not args.output_path:
    raise ValueError("Specify --output_path if providing --log_samples or --predict_only")

# Initialize the task manager
task_manager = TaskManager(args.verbosity, include_path=args.include_path)

# Warning if a limit is set, since it should only be used for testing
if args.limit:
    eval_logger.warning(" --limit SHOULD ONLY BE USED FOR TESTING. REAL METRICS SHOULD NOT BE COMPUTED USING LIMIT.")

# Set trust remote code flag if required
if args.trust_remote_code:
    datasets.config.HF_DATASETS_TRUST_REMOTE_CODE = True
    args.model_args = args.model_args + ",trust_remote_code=True"

# Process task list or handle special task flags
if args.tasks is None:
    eval_logger.error("Need to specify task to evaluate.")
    sys.exit()
elif args.tasks == "list":
    print(task_manager.list_all_tasks())
    sys.exit()
elif args.tasks == "list_groups":
    print(task_manager.list_all_tasks(list_subtasks=False, list_tags=False))
    sys.exit()
elif args.tasks == "list_tags":
    print(task_manager.list_all_tasks(list_groups=False, list_subtasks=False))
    sys.exit()
elif args.tasks == "list_subtasks":
    print(task_manager.list_all_tasks(list_groups=False, list_tags=False))
    sys.exit()
else:
    # Parse task names from the provided list or directory
    if os.path.isdir(args.tasks):
        import glob
        task_names = []
        yaml_path = os.path.join(args.tasks, "*.yaml")
        for yaml_file in glob.glob(yaml_path):
            config = utils.load_yaml_config(yaml_file)
            task_names.append(config)
    else:
        task_list = args.tasks.split(",")
        task_names = task_manager.match_tasks(task_list)
        for task in [task for task in task_list if task not in task_names]:
            if os.path.isfile(task):
                config = utils.load_yaml_config(task)
                task_names.append(config)
        task_missing = [
            task for task in task_list if task not in task_names and "*" not in task
        ]  # we don't want errors if a wildcard ("*") task name was used

        if task_missing:
            missing = ", ".join(task_missing)
            eval_logger.error(
                f"Tasks were not found: {missing}\n"
                f"{utils.SPACING}Try `lm-eval --tasks list` for list of available tasks",
            )
            raise ValueError(
                f"Tasks not found: {missing}. Try `lm-eval --tasks {{list_groups,list_subtasks,list_tags,list}}` to list out all available names for task groupings; only (sub)tasks; tags; or all of the above, or pass '--verbosity DEBUG' to troubleshoot task registration issues."
            )

# Log selected tasks     
eval_logger.info(f"Selected Tasks: {task_names}")

# Prepare request caching arguments
request_caching_args = evaluator.request_caching_arg_to_dict(cache_requests=args.cache_requests)

# Perform the evaluation
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

# Process and print the results
if results is not None:
        if args.log_samples:
            samples = results.pop("samples")
        dumped = json.dumps(
            results, indent=2, default=handle_non_serializable, ensure_ascii=False
        )
        if args.show_config:
            print(dumped)

        batch_sizes = ",".join(map(str, results["config"]["batch_sizes"]))

        # Add W&B logging
        if args.wandb_args:
            try:
                wandb_logger.post_init(results)
                wandb_logger.log_eval_result()
                if args.log_samples:
                    wandb_logger.log_eval_samples(samples)
            except Exception as e:
                eval_logger.info(f"Logging to Weights and Biases failed due to {e}")

        evaluation_tracker.save_results_aggregated(
            results=results, samples=samples if args.log_samples else None
        )

        if args.log_samples:
            for task_name, config in results["configs"].items():
                evaluation_tracker.save_results_samples(
                    task_name=task_name, samples=samples[task_name]
                )

        if (
            evaluation_tracker.push_results_to_hub
            or evaluation_tracker.push_samples_to_hub
        ):
            evaluation_tracker.recreate_metadata_card()

        print(
            f"{args.model} ({args.model_args}), gen_kwargs: ({args.gen_kwargs}), limit: {args.limit}, num_fewshot: {args.num_fewshot}, "
            f"batch_size: {args.batch_size}{f' ({batch_sizes})' if batch_sizes else ''}"
        )
        print(make_table(results))
        if "groups" in results:
            print(make_table(results, "groups"))

        if args.wandb_args:
            # Tear down wandb run once all the logging is done.
            wandb_logger.run.finish()
