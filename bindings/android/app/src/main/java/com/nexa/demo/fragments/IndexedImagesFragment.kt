package com.nexa.demo.fragments

import android.os.Bundle
import android.util.Log
import androidx.fragment.app.Fragment
import android.view.View
import com.nexa.demo.R
import com.nexa.demo.adapter.SelectImagesAdapter
import com.nexa.demo.databinding.FragmentIndexedImagesBinding
import com.nexa.demo.utils.bindView

private const val ARG_PARAM1 = "param1"

/**
 * A simple [Fragment] subclass.
 * Use the [IndexedImagesFragment.newInstance] factory method to
 * create an instance of this fragment.
 */
class IndexedImagesFragment : UpdatePercentFragment(R.layout.fragment_indexed_images) {

    private var allImages: ArrayList<String>? = arrayListOf()
    private var param2: String? = null
    private val adapter: SelectImagesAdapter = SelectImagesAdapter(allImages!!)

    private val binding by bindView<FragmentIndexedImagesBinding>()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        arguments?.let {
            allImages = it.getStringArrayList(ARG_PARAM1)
        }
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        Log.d("nfl", "IndexedImagesFragment onViewCreated")
        binding.rvImages.adapter = adapter
    }

    override fun updatePercent(allPercent: ArrayList<Int>) {
        adapter.updatePercent(allPercent)
    }

    override fun updateImages(allImages: ArrayList<String>) {
        adapter.updateImages(allImages)
    }

    override fun onDestroyView() {
        super.onDestroyView()
        Log.d("nfl", "IndexedImagesFragment onDestroyView")
    }

    companion object {

        @JvmStatic
        fun newInstance(allImages: ArrayList<String>) =
            IndexedImagesFragment().apply {
                arguments = Bundle().apply {
                    putStringArrayList(ARG_PARAM1, allImages)
                }
            }
    }
}