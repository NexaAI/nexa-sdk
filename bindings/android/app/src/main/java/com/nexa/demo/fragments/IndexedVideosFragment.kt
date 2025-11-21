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