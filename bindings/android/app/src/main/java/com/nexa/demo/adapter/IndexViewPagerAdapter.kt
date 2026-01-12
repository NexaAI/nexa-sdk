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

package com.nexa.demo.adapter

import androidx.fragment.app.Fragment
import androidx.fragment.app.FragmentManager
import androidx.fragment.app.FragmentPagerAdapter
import com.nexa.demo.fragments.UpdatePercentFragment

class IndexViewPagerAdapter(
    fm: FragmentManager,
    private val fragments: MutableList<UpdatePercentFragment>,
    private val titles: MutableList<String>
) : FragmentPagerAdapter(fm) {

    override fun getItem(position: Int): Fragment {
        return fragments.get(position)
    }

    override fun getCount(): Int {
        return fragments.size
    }

    override fun getPageTitle(position: Int): CharSequence? {
        return titles[position]
    }

    fun updatePercent(position: Int, allPercent: ArrayList<Int>) {
        fragments[position].updatePercent(allPercent)
    }
    fun updateImages(position: Int, allImages: ArrayList<String>) {
        fragments[position].updateImages(allImages)
    }
}
