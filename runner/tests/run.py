#!/usr/bin/env python3

import io
import json
import os
import traceback

from scripts import config, log, utils

# plugin, model, testcases
testcases: list[config.case]


def init_benchmark():
    global testcases
    log.print("========= Init Benchmark =========")

    plugins = config.get_plugins()
    log.print(f'Plugins: {plugins}')
    if len(plugins) == 0:
        raise Exception("No supported plugins found")

    testcases = config.get_testcases(plugins)
    # count total Testcases
    log.print(f'Found {len(testcases)} Models with {sum(len(tc[2]) for tc in testcases)} TestCases')
    if len(testcases) == 0:
        raise Exception("No TestCases found")


def check_models():
    global testcases
    log.print("========== Check Models ==========")

    res = utils.execute_nexa(['list'])
    if res.returncode != 0:
        raise RuntimeError(f'Non-zero exit code: {res.returncode}')

    exist_models: set[str] = set()
    for line in res.stdout.splitlines():
        fields = line.split('│')
        if len(fields) < 5:
            continue
        name = fields[1].strip()
        quant = fields[3].strip()
        if quant == '':
            exist_models.add(name)
        else:
            for quant in quant.split(','):
                exist_models.add(f'{name}:{quant.strip()}')

    for i, (_, model, type, _) in enumerate(testcases):
        if model in exist_models:
            log.print(f'  --> [{i+1}/{len(testcases)}] {model} already exists, skip download')
            continue
        log.print(f'  --> [{i+1}/{len(testcases)}] {model} not found, downloading...')
        res = utils.execute_nexa([
            'pull',
            model,
            '--model-type',
            type,
        ], timeout=3600)
        if res.returncode != 0:
            raise RuntimeError(f'Failed to download model {model}, exit code: {res.returncode}')


def run_benchmark():
    log.print("========== Run Benchmark =========")

    failed_cases: list[tuple[str, str, str, str]] = []
    for i, (plugin, model, _, tcs) in enumerate(testcases):
        os.makedirs(log.log_dir / plugin, exist_ok=True)
        mp = f'{i+1}/{len(testcases)}'
        log.print(f'==> [{mp}] Plugin: {plugin}, Model: {model}')
        of = None
        ef = None

        for i, tc in enumerate(tcs):
            tc = tc()
            tcp = f'{i+1}/{len(tcs)}'
            try:
                tc_log = log.log_dir / plugin / f'{model.replace("/", "-").replace(":", "-")}-{tc.name()}'
                of = open(f'{tc_log}.log', 'w', encoding='utf-8')
                ef = io.StringIO()
                res = utils.execute_nexa(['infer', model] + tc.param(),
                                         debug_log=True,
                                         stdout=of,
                                         stderr=ef,
                                         timeout=300)
                if res.returncode != 0:
                    raise RuntimeError(f'Non-zero exit code: {res.returncode}')

                # check output
                of.write('\n====== Json Log =======\n')
                failed = False
                for line in ef.getvalue().splitlines():
                    of.write(f'{line}\n')
                    if tc.check(json.loads(line)):
                        of.write(f'Passed\n')
                        continue
                    else:
                        of.write(f'Failed\n')
                        failed = True
                if failed:
                    log.print(f'  --> [{mp}][{tcp}] TestCase {tc.name()} Failed')
                    failed_cases.append((plugin, model, tc.name(), 'Check Failed'))
                else:
                    log.print(f'  --> [{mp}][{tcp}] TestCase: {tc.name()} Success')

            except Exception as _:
                log.print(f'  --> [{mp}][{tcp}] TestCase {tc.name()} Error')
                failed_cases.append((plugin, model, tc.name(), 'Error'))
                if of is not None:
                    of.write('\n====== Exception Log =======\n')
                    of.write(traceback.format_exc())
            finally:
                if of is not None:
                    of.close()
                if ef is not None:
                    ef.close()

    log.print("======== Benchmark Result ========")
    if len(failed_cases) == 0:
        log.print('All TestCases passed')
    else:
        for plugin, model, tc, reason in failed_cases:
            log_file = log.log_dir / plugin / f'{model.replace("/", "-").replace(":", "-")}-{tc}.log'
            log.print(f'{reason}: Plugin: {plugin}, Model: {model}, TestCase: {tc}, see {log_file}')
    log.print(f'Logs saved to {log.log_dir}')


def main():
    log.init()
    init_benchmark()
    check_models()
    run_benchmark()


if __name__ == "__main__":
    main()
