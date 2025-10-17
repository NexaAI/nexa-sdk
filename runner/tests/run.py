#!/usr/bin/env python3

import os
import traceback

from scripts import config, log, utils

# plugin, model, testcases
testcases: list[tuple[str, str, list[str]]]


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

    missing_models = [tc[1] for tc in testcases]
    for line in res.stdout.splitlines():
        fields = line.split('│')
        if len(fields) < 5:
            continue
        name = fields[1].strip()
        quant = fields[3].strip()
        if name in missing_models:
            # TODO: remove quant check
            if ',' in quant:
                raise RuntimeError(f'Model {name} has multiple quantizations: {quant}')
            missing_models.remove(name)
    if len(missing_models) > 0:
        raise RuntimeError(f'Missing models: {missing_models}')


def run_benchmark():
    log.print("========== Run Benchmark =========")

    failed_cases: list[tuple[str, str, str]] = []
    for i, (plugin, model, tcs) in enumerate(testcases):
        os.makedirs(log.log_dir / plugin, exist_ok=True)
        mp = f'{i+1}/{len(testcases)}'
        log.print(f'==> [{mp}] Plugin: {plugin}, Model: {model}')
        of = None
        ef = None

        for i, tc in enumerate(tcs):
            tcp = f'{i+1}/{len(tcs)}'
            try:
                tc_log = log.log_dir / plugin / f'{model.replace("/", "-")}_{tc}'
                of = open(f'{tc_log}.log', 'w', encoding='utf-8')
                ef = open(f'{tc_log}.json', 'w', encoding='utf-8')
                res = utils.execute_nexa([
                    'infer',
                    model,
                ] + utils.load_param(tc),
                                         debug_log=True,
                                         stdout=of,
                                         stderr=ef,
                                         timeout=300)
                if res.returncode != 0:
                    raise RuntimeError(f'Non-zero exit code: {res.returncode}')

                log.print(f'  --> [{mp}][{tcp}] TestCase: {tc} success')
            except Exception as _:
                log.print(f'  --> [{mp}][{tcp}] TestCase {tc} failed')
                failed_cases.append((plugin, model, tc))
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
        for plugin, model, tc in failed_cases:
            log_file = log.log_dir / plugin / f'{model.replace("/", "-")}_{tc}.log'
            log.print(f'Failed: Plugin: {plugin}, Model: {model}, TestCase: {tc}, see {log_file}')
    log.print(f'Logs saved to {log.log_dir}')


def main():
    log.init()
    init_benchmark()
    check_models()
    run_benchmark()


if __name__ == "__main__":
    main()
