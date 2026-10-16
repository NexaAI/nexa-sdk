// Copyright (c) 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

// Intentionally empty. The only product of this file is the shared object built from it:
// SONAME libcdsprpc.so, DT_NEEDED libcdsprpc.so.1 (see CMakeLists.txt). Loading it satisfies a
// QNN HTP stub's `NEEDED libcdsprpc.so` on systems that only ship libcdsprpc.so.1.
extern "C" void geniex_cdsprpc_alias_marker(void) {}
