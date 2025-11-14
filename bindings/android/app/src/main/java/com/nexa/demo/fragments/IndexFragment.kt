package com.nexa.demo.fragments

import android.app.Activity
import android.content.Intent
import android.os.Bundle
import android.os.Handler
import android.util.Log
import androidx.fragment.app.Fragment
import android.view.View
import androidx.activity.ComponentActivity
import androidx.activity.result.ActivityResultLauncher
import androidx.activity.result.contract.ActivityResultContracts
import androidx.appcompat.app.AppCompatActivity
import com.google.android.material.shape.CornerFamily
import com.nexa.demo.MainActivity
import com.nexa.demo.R
import com.nexa.demo.activity.FolderActivity
import com.nexa.demo.activity.FolderActivity.Companion.KEY_SELECT_DIRS
import com.nexa.demo.activity.FolderActivity.Companion.KEY_SELECT_IMAGES
import com.nexa.demo.adapter.IndexViewPagerAdapter
import com.nexa.demo.databinding.FragmentIndexBinding
import com.nexa.demo.utils.KeyboardUtil
import com.nexa.demo.utils.PermissionUtil
import com.nexa.demo.utils.bindView
import com.nexa.sdk.bean.EmbeddingConfig
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.coroutineScope
import kotlinx.coroutines.currentCoroutineContext
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import java.io.File
import java.util.Arrays

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
    private lateinit var adapter:IndexViewPagerAdapter
    private val titles: MutableList<String> = arrayListOf()
    private lateinit var selectFolderResult: ActivityResultLauncher<Intent>
    private var uiState = UIState.NO_INDEX
    private var allFileCount = 0

    private val allImages = arrayListOf<String>()
    private val allImagePercentList = arrayListOf<Int>()
    private val allImageResultList = arrayListOf<FloatArray>()


    private val allVideos = arrayListOf<String>()
    private val allVideoPercentList = arrayListOf<Int>()
    private val allVideoResultList = arrayListOf<FloatArray>()

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
                allFileCount = 0
                Log.d(TAG, "select dirs:${result.data?.getStringArrayListExtra(KEY_SELECT_IMAGES)}")

                allImages.clear()
                result.data?.getStringArrayListExtra(KEY_SELECT_IMAGES)?.let {
                    allImages.addAll(it)
                }

                allFileCount = allImages.size
                binding.tvIndexDatabase.text = "Database: $allFileCount files"
                if (allFileCount == 0) {
                    binding.lpiIndexing.max = 1
                    binding.lpiIndexing.progress = 1
                    uiState = UIState.INDEXED
                } else {
                    binding.lpiIndexing.max = allFileCount
                    uiState = UIState.INDEXING
                }
                changeUIState()
                adapter.updateImages(0, allImages)

                allImageResultList.clear()
                allVideoResultList.clear()

                CoroutineScope(Dispatchers.IO).launch {
                    allImages.forEachIndexed { index, imagePath ->
                        var temp = imagePath
                        if (index == 0) {
                            temp = "a feline sits on the mat"
                        } else {
                            temp = "the international space station orbits Earth";
                        }

                        (activity as MainActivity).embedderWrapper.let { embedderWrapper ->
                            embedderWrapper?.embed(arrayOf(temp), EmbeddingConfig(batchSize = 1)).let {
                                it?.onSuccess {
                                    allImageResultList.add(index, it)
                                    Log.d("nfl", "embed result size:${it.size}")
                                    Log.d("nfl", "embed result:${it.contentToString()}")
                                }
                                    ?.onFailure {
                                        Log.d("nfl", "embed result failed:$it")
                                    }

                                withContext(Dispatchers.Main) {
                                    binding.lpiIndexing.progress = index + 1
                                    if (index + 1 >= allFileCount) {
                                        uiState = UIState.INDEXED
                                        changeUIState()
                                    }
                                }
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
                binding.vHideBottom.visibility = View.GONE
            }

            UIState.INDEXING -> {
                binding.llIndexing.visibility = View.VISIBLE
                binding.tvIndexTip.visibility = View.GONE

                binding.llIndexed.visibility = View.VISIBLE
                titles.clear()
                titles.add("Images(${allImages.size})")
                titles.add("Videos")
                adapter.notifyDataSetChanged()

                binding.vHideBottom.visibility = View.VISIBLE
            }

            UIState.INDEXED -> {
                binding.llIndexing.visibility = View.GONE
                binding.tvIndexTip.visibility = View.GONE

                binding.llIndexed.visibility = View.VISIBLE
                titles.clear()
                titles.add("Images(${allImages.size})")
                titles.add("Videos")
                adapter.notifyDataSetChanged()

                binding.vHideBottom.visibility = View.GONE
            }
        }
    }

    private lateinit var searchJob: Job
    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)

        binding.mcvIndex.let { cardView ->
            cardView.setShapeAppearanceModel(
                cardView.shapeAppearanceModel
                    .toBuilder()
                    .setTopLeftCorner(CornerFamily.ROUNDED, 80f)
                    .setTopRightCorner(CornerFamily.ROUNDED, 80f)
                    .build()
            );
        }

        titles.add("Images(${allImages.size})")
        titles.add("Videos")
        adapter = IndexViewPagerAdapter(
            activity!!.supportFragmentManager,
            arrayListOf(
                IndexedImagesFragment.newInstance(allImages, ""),
                IndexedVideosFragment.newInstance("", "")
            ),
            titles
        )
        binding.vpIndexed.adapter = adapter
        binding.tlIndexed.setupWithViewPager(binding.vpIndexed)

        binding.btnImport.setOnClickListener {
            if (PermissionUtil.checkManageStoragePermission(activity!!)) {
                selectFolderResult.launch(Intent(context, FolderActivity::class.java))
            } else {
                PermissionUtil.showRequestManageStoragePermissionDialog(activity as ComponentActivity)
            }
        }

        binding.btnCancelIndex.setOnClickListener {
            uiState = UIState.NO_INDEX
            changeUIState()
        }

        binding.btnSearch.setOnClickListener {
            if ("Search" == binding.btnSearch.text) {
                binding.btnSearch.text = "Stop"
                KeyboardUtil.hide(binding.etSearch)
                searchJob = CoroutineScope(Dispatchers.IO).launch {
                    (activity as MainActivity).embedderWrapper?.embed(
//                        arrayOf(binding.etSearch.text.toString()),
                        arrayOf("the cat is resting on the rug"),
                        EmbeddingConfig(batchSize = 1)
                    )?.onSuccess { searchResult ->
                        allImagePercentList.clear()
                        allImageResultList.forEach { imageResult ->
                            allImagePercentList.add(
                                (computeCosineSimilarity(
                                    searchResult,
                                    imageResult
                                ).apply {
                                    Log.d("nfl", "computeCosineSimilarity: $this")
                                } * 100).toInt()
                            )
                        }
                        withContext(Dispatchers.Main) {
                            binding.btnSearch.text = "Search"
                            (binding.vpIndexed.adapter as IndexViewPagerAdapter).let {
                                it.updatePercent(0, allImagePercentList)
                                it.updatePercent(1, allVideoPercentList)
                            }
                        }
                    }?.onFailure {
                        activity?.runOnUiThread {
                            binding.btnSearch.text = "Search"
                        }
                    }
                }
            } else {
                binding.btnSearch.text = "Search"
                searchJob.cancel()
            }
        }
    }

    enum class UIState {
        NO_INDEX, INDEXING, INDEXED
    }

    fun computeCosineSimilarity(
        embedding1: FloatArray?,
        embedding2: FloatArray?
    ): Float {
        if (embedding1 == null || embedding2 == null) return 0.0f
        if (embedding1.isEmpty() || embedding2.isEmpty()) return 0.0f
        if (embedding1.size != embedding2.size) return 0.0f

        var dotProduct = 0.0f
        var norm1 = 0.0f
        var norm2 = 0.0f

        for (i in embedding1.indices) {
            dotProduct += embedding1[i] * embedding2[i]
            norm1 += embedding1[i] * embedding1[i]
            norm2 += embedding2[i] * embedding2[i]
        }

        val epsilon = 1e-8f
        norm1 = kotlin.math.sqrt(norm1 + epsilon)
        norm2 = kotlin.math.sqrt(norm2 + epsilon)
        Log.d("nfl", "norm1 > 0 ? ${norm1 > 0}")
        Log.d("nfl", "norm2 > 0 ? ${norm2 > 0}")
        Log.d("nfl", "dotProduct > 0 ? ${dotProduct > 0}")
        return dotProduct / (norm1 * norm2)
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