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

import android.os.Bundle
import android.view.View
import androidx.fragment.app.Fragment
import com.nexa.demo.R
import com.nexa.demo.adapter.SelectVideosAdapter
import com.nexa.demo.databinding.FragmentIndexedVideosBinding
import com.nexa.demo.utils.bindView

private const val ARG_PARAM1 = "param1"
private const val ARG_PARAM2 = "param2"

/**
 * A simple [Fragment] subclass.
 * Use the [IndexedVideosFragment.newInstance] factory method to
 * create an instance of this fragment.
 */
class IndexedVideosFragment : UpdatePercentFragment(R.layout.fragment_indexed_videos) {

    private var param1: String? = null
    private var param2: String? = null
    private val adapter = SelectVideosAdapter()

    private val binding by bindView<FragmentIndexedVideosBinding>()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        arguments?.let {
            param1 = it.getString(ARG_PARAM1)
            param2 = it.getString(ARG_PARAM2)
        }
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        binding.rvVideos.adapter = adapter
    }

    override fun updatePercent(allPercent: ArrayList<Int>) {

    }

    override fun updateImages(allImages: ArrayList<String>) {

    }

    companion object {

        @JvmStatic
        fun newInstance(param1: String, param2: String) =
            IndexedVideosFragment().apply {
                arguments = Bundle().apply {
                    putString(ARG_PARAM1, param1)
                    putString(ARG_PARAM2, param2)
                }
            }
    }
}
