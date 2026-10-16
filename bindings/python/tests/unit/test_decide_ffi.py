# Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
# SPDX-License-Identifier: BSD-3-Clause

"""Check the ctypes ownership contract for SDK decision output."""

import importlib.util
from ctypes import c_void_p
from pathlib import Path


def test_decide_output_keeps_owned_type_pointer():
    types_path = Path(__file__).resolve().parents[2] / 'geniex' / '_ffi' / '_types.py'
    spec = importlib.util.spec_from_file_location('geniex_decide_types', types_path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)

    fields = dict(module.geniex_DecideOutput._fields_)
    assert fields['decision_type'] is c_void_p

    output = module.geniex_DecideOutput()
    output.decision_type = 0x12345
    assert output.decision_type == 0x12345
