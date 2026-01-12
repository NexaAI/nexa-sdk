// Copyright 2024-2026 Nexa AI, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package com.nexa.demo.fragments

import androidx.annotation.LayoutRes
import androidx.fragment.app.Fragment

abstract class UpdatePercentFragment(@LayoutRes contentLayoutId: Int) : Fragment(contentLayoutId) {
    abstract fun updatePercent(allPercent: ArrayList<Int>)
    abstract fun updateImages(allImages: ArrayList<String>)
}
