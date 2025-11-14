package com.nexa.demo.fragments

import android.os.Bundle
import androidx.fragment.app.Fragment
import android.view.View
import com.nexa.demo.R
import com.nexa.demo.adapter.SelectImagesAdapter
import com.nexa.demo.databinding.FragmentIndexedImagesBinding
import com.nexa.demo.utils.bindView

// TODO: Rename parameter arguments, choose names that match
// the fragment initialization parameters, e.g. ARG_ITEM_NUMBER
private const val ARG_PARAM1 = "param1"
private const val ARG_PARAM2 = "param2"

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
            param2 = it.getString(ARG_PARAM2)
        }
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        binding.rvImages.adapter = adapter
    }

    override fun updatePercent(allPercent: ArrayList<Int>) {
        adapter.updatePercent(allPercent)
    }

    override fun updateImages(allImages: ArrayList<String>) {
        adapter.updateImages(allImages)
    }

    companion object {

        @JvmStatic
        fun newInstance(allImages: ArrayList<String>, param2: String) =
            IndexedImagesFragment().apply {
                arguments = Bundle().apply {
                    putStringArrayList(ARG_PARAM1, allImages)
                    putString(ARG_PARAM2, param2)
                }
            }
    }
}