package com.nexa.demo.fragments

import android.app.Activity
import android.content.Intent
import android.os.Bundle
import android.util.Log
import androidx.fragment.app.Fragment
import android.view.View
import androidx.activity.result.ActivityResultLauncher
import androidx.activity.result.contract.ActivityResultContracts
import com.google.android.material.shape.CornerFamily
import com.nexa.demo.R
import com.nexa.demo.activity.FolderActivity
import com.nexa.demo.activity.FolderActivity.Companion.KEY_SELECT_DIRS
import com.nexa.demo.adapter.IndexViewPagerAdapter
import com.nexa.demo.databinding.FragmentIndexBinding
import com.nexa.demo.utils.bindView
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.coroutineScope
import kotlinx.coroutines.currentCoroutineContext
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

// TODO: Rename parameter arguments, choose names that match
// the fragment initialization parameters, e.g. ARG_ITEM_NUMBER
private const val ARG_PARAM1 = "param1"
private const val ARG_PARAM2 = "param2"

/**
 * A simple [Fragment] subclass.
 * Use the [IndexFragment.newInstance] factory method to
 * create an instance of this fragment.
 */
class IndexFragment : Fragment(R.layout.fragment_index) {
    // TODO: Rename and change types of parameters
    private var param1: String? = null
    private var param2: String? = null
    private val binding by bindView<FragmentIndexBinding>()
    private lateinit var selectFolderResult: ActivityResultLauncher<Intent>
    private var uiState = UIState.NO_INDEX

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        arguments?.let {
            param1 = it.getString(ARG_PARAM1)
            param2 = it.getString(ARG_PARAM2)
        }
        selectFolderResult = registerForActivityResult(
            ActivityResultContracts.StartActivityForResult()
        ) { result -> //
            if (Activity.RESULT_OK == result.resultCode) {
                Log.d(TAG, "select dirs:${result.data?.getStringArrayListExtra(KEY_SELECT_DIRS)}")
                uiState = UIState.INDEXING
                changeUIState()
                CoroutineScope(Dispatchers.IO).launch {
                    var progress = 0
                    while (progress < 100 && uiState == UIState.INDEXING) {
                        progress++
                        delay(100)
                        withContext(Dispatchers.Main) {
                            binding.lpiIndexing.progress = progress
                            if (progress >= 100) {
                                uiState = UIState.INDEXED
                                changeUIState()
                            }
                        }
                    }
                }
            }
        }
    }

    private fun changeUIState() {
        when (uiState) {
            UIState.NO_INDEX -> {
                binding.llIndexing.visibility = View.GONE
                binding.tvIndexTip.visibility = View.VISIBLE
                binding.llIndexed.visibility = View.GONE
                // enable bottom
            }

            UIState.INDEXING -> {
                binding.llIndexing.visibility = View.VISIBLE
                binding.tvIndexTip.visibility = View.GONE
                binding.llIndexed.visibility = View.GONE
                // disable bottom
            }

            UIState.INDEXED -> {
                binding.llIndexing.visibility = View.GONE
                binding.tvIndexTip.visibility = View.GONE
                binding.llIndexed.visibility = View.VISIBLE
                // enable bottom
            }
        }
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)

        binding.mcvIndex.let { cardView ->
            cardView.setShapeAppearanceModel(
                cardView.shapeAppearanceModel
                    .toBuilder()
                    .setTopLeftCorner(CornerFamily.ROUNDED, 20f)
                    .setTopRightCorner(CornerFamily.ROUNDED, 20f)
                    .build()
            );
        }

        binding.btnImport.setOnClickListener {
            selectFolderResult.launch(Intent(context, FolderActivity::class.java))
        }

        binding.btnCancelIndex.setOnClickListener {
            uiState = UIState.NO_INDEX
            changeUIState()
        }
        binding.vpIndexed.adapter =
            IndexViewPagerAdapter(
                activity!!.supportFragmentManager,
                arrayListOf(
                    IndexedImagesFragment.newInstance("", ""),
                    IndexedVideosFragment.newInstance("", "")
                ),
                arrayListOf("Images", "Videos")
            )
        binding.tlIndexed.setupWithViewPager(binding.vpIndexed)

    }

    enum class UIState {
        NO_INDEX, INDEXING, INDEXED
    }

    companion object {
        const val TAG = "IndexFragment"

        /**
         * Use this factory method to create a new instance of
         * this fragment using the provided parameters.
         *
         * @param param1 Parameter 1.
         * @param param2 Parameter 2.
         * @return A new instance of fragment IndexFragment.
         */
        // TODO: Rename and change types and number of parameters
        @JvmStatic
        fun newInstance(param1: String, param2: String) =
            IndexFragment().apply {
                arguments = Bundle().apply {
                    putString(ARG_PARAM1, param1)
                    putString(ARG_PARAM2, param2)
                }
            }
    }
}