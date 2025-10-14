#!/usr/bin/env python3

import io
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
        log.print("No supported plugins found, exit")
        return

    testcases = config.get_testcases(plugins)
    # count total Testcases
    log.print(f'Found {len(testcases)} models with {sum(len(tc[2]) for tc in testcases)} testcases')
    if len(testcases) == 0:
        log.print("Testcases is empty, exit")
        return


def check_models():
    log.print("========== Check Models ==========")

    res = utils.execute_nexa(['list'])
    if res.returncode != 0:
        raise RuntimeError(f'Non-zero exit code: {res.returncode}')

    missing_models = [tc[1] for tc in testcases]
    for line in res.stdout.splitlines():
        fileds = line.split('│')
        if len(fileds) < 5:
            continue
        name = fileds[1].strip()
        quant = fileds[3].strip()
        if name in missing_models:
            # TODO: remove quant check
            if ',' in quant:
                raise RuntimeError(f'Model {name} has multiple quantizations: {quant}')
            missing_models.remove(name)
    if len(missing_models) > 0:
        raise RuntimeError(f'Missing models: {missing_models}')


def run_benchmark():
    log.print("========== Run Benchmark =========")

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
                of = open(f'{tc_log}.log', 'w')
                ef = open(f'{tc_log}.json', 'w')
                res = utils.execute_nexa([
                    'infer',
                    model,
                    '--test-mode',
                    '--skip-migrate',
                    '--skip-update',
                ] + utils.load_param(tc),
                                         stdout=of,
                                         stderr=ef,
                                         timeout=300)
                if res.returncode != 0:
                    raise RuntimeError(f'Non-zero exit code: {res.returncode}')

                log.print(f'  --> [{mp}][{tcp}] Testcase: {tc} success')
            except Exception as e:
                log.print(f'  --> [{mp}][{tcp}] Testcase {tc} failed')
                if of is not None:
                    of.write('\n====== Exception Logs ======\n')
                    of.write(traceback.format_exc())
            finally:
                if of is not None:
                    of.close()
                if ef is not None:
                    ef.close()


def main():
    log.init()
    init_benchmark()
    check_models()
    run_benchmark()


if __name__ == "__main__":
    main()
