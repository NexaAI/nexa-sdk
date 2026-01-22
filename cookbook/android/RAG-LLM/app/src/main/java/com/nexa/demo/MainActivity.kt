package com.nexa.demo

import android.Manifest
import android.app.Activity
import android.content.Context
import android.content.DialogInterface
import android.content.Intent
import android.content.SharedPreferences
import android.content.pm.PackageManager
import android.graphics.Bitmap
import android.graphics.BitmapFactory
import android.graphics.Color
import android.net.Uri
import android.os.Bundle
import android.os.Environment
import android.provider.MediaStore
import android.system.Os
import android.text.Editable
import android.text.TextUtils
import android.text.TextWatcher
import android.util.Log
import android.view.KeyEvent
import android.view.LayoutInflater
import android.view.View
import android.view.inputmethod.InputMethodManager
import android.widget.AdapterView
import android.widget.Button
import android.widget.EditText
import android.widget.HorizontalScrollView
import android.widget.ImageButton
import android.widget.ImageView
import android.widget.LinearLayout
import android.widget.PopupWindow
import android.widget.ProgressBar
import android.widget.SeekBar
import android.widget.SimpleAdapter
import android.widget.Spinner
import android.widget.TextView
import android.widget.Toast
import android.window.OnBackInvokedDispatcher
import androidx.activity.ComponentActivity
import androidx.activity.OnBackPressedCallback
import androidx.activity.result.ActivityResultLauncher
import androidx.activity.result.contract.ActivityResultContracts
import androidx.appcompat.app.AlertDialog
import androidx.compose.ui.unit.TextUnit
import androidx.core.app.ActivityCompat
import androidx.core.content.ContextCompat
import androidx.core.content.FileProvider
import androidx.fragment.app.Fragment
import androidx.fragment.app.FragmentActivity
import androidx.recyclerview.widget.RecyclerView
import com.google.android.material.bottomsheet.BottomSheetDialog
import com.gyf.immersionbar.ktx.immersionBar
import com.hjq.toast.Toaster
import com.liulishuo.okdownload.DownloadContext
import com.liulishuo.okdownload.DownloadTask
import com.liulishuo.okdownload.OkDownload
import com.liulishuo.okdownload.core.cause.EndCause
import com.liulishuo.okdownload.core.connection.DownloadOkHttp3Connection
import com.liulishuo.okdownload.kotlin.listener.createDownloadContextListener
import com.liulishuo.okdownload.kotlin.listener.createListener1
import com.nexa.demo.activity.FolderActivity
import com.nexa.demo.adapter.ChunkAdapter
import com.nexa.demo.bean.DownloadState
import com.nexa.demo.bean.EmbedResultBean
import com.nexa.demo.bean.ModelData
import com.nexa.demo.bean.DownloadableFileWithFallback
import com.nexa.demo.bean.downloadableFiles
import com.nexa.demo.bean.downloadableFilesWithFallback
import com.nexa.demo.bean.downloadableFilesWithNpuList
import com.nexa.demo.bean.getNexaManifest
import com.nexa.demo.bean.getNonExistModelFile
import com.nexa.demo.bean.getSupportPluginIds
import com.nexa.demo.bean.allModelFilesExist
import com.nexa.demo.bean.isDownloaded
import com.nexa.demo.bean.isNpuModel
import com.nexa.demo.bean.withFallbackUrls
import com.nexa.demo.utils.ModelFileListingUtil
import com.nexa.demo.bean.mmprojTokenFile
import com.nexa.demo.bean.modelDir
import com.nexa.demo.bean.modelFile
import com.nexa.demo.bean.tokenFile
import com.nexa.demo.databinding.ActivityMainBinding
import com.nexa.demo.databinding.DialogConfigBinding
import com.nexa.demo.databinding.DialogSelectPluginIdBinding
import com.nexa.demo.databinding.DialogTopkConfigBinding
import com.nexa.demo.fragments.IndexFragment
import com.nexa.demo.listeners.CustomDialogInterface
import com.nexa.demo.utils.DensityUtil
import com.nexa.demo.utils.ExecShell
import com.nexa.demo.utils.GenerateEmbedStringsUtil
import com.nexa.demo.utils.ImgUtil
import com.nexa.demo.utils.PermissionUtil
import com.nexa.demo.utils.SharePreferenceKeys
import com.nexa.demo.utils.WavRecorder
import com.nexa.demo.utils.inflate
import com.nexa.sdk.AsrWrapper
import com.nexa.sdk.CvWrapper
import com.nexa.sdk.EmbedderWrapper
import com.nexa.sdk.LlmWrapper
import com.nexa.sdk.NexaSdk
import com.nexa.sdk.RerankerWrapper
import com.nexa.sdk.VlmWrapper
import com.nexa.sdk.bean.AsrCreateInput
import com.nexa.sdk.bean.AsrTranscribeInput
import com.nexa.sdk.bean.CVCapability
import com.nexa.sdk.bean.CVCreateInput
import com.nexa.sdk.bean.CVModelConfig
import com.nexa.sdk.bean.ChatMessage
import com.nexa.sdk.bean.DeviceIdValue
import com.nexa.sdk.bean.EmbedderCreateInput
import com.nexa.sdk.bean.EmbeddingConfig
import com.nexa.sdk.bean.LlmCreateInput
import com.nexa.sdk.bean.LlmStreamResult
import com.nexa.sdk.bean.ModelConfig
import com.nexa.sdk.bean.RerankConfig
import com.nexa.sdk.bean.RerankerCreateInput
import com.nexa.sdk.bean.VlmChatMessage
import com.nexa.sdk.bean.VlmContent
import com.nexa.sdk.bean.VlmCreateInput
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.async
import kotlinx.coroutines.awaitAll
import kotlinx.coroutines.launch
import kotlinx.serialization.json.Json
import okhttp3.OkHttpClient
import okhttp3.Request
import java.io.File
import java.io.FileNotFoundException
import java.io.FileOutputStream
import java.security.SecureRandom
import java.security.cert.CertificateException
import java.security.cert.X509Certificate
import javax.net.ssl.SSLContext
import javax.net.ssl.SSLSession
import javax.net.ssl.SSLSocketFactory
import javax.net.ssl.TrustManager
import javax.net.ssl.X509TrustManager
import androidx.core.content.edit
import com.nexa.demo.bean.DownloadableFile
import kotlinx.serialization.StringFormat

class MainActivity : FragmentActivity() {

    private val binding: ActivityMainBinding by inflate()
    private var downloadContext: DownloadContext? = null
    private var downloadState = DownloadState.IDLE
    private var downloadingModelData: ModelData? = null
    private lateinit var spDownloaded: SharedPreferences
    private lateinit var llDownloading: LinearLayout
    private lateinit var tvDownloadProgress: TextView
    private lateinit var pbDownloading: ProgressBar
    private lateinit var spModelList: Spinner
    private lateinit var btnDownload: Button
    private lateinit var btnLoadModel: Button
    private lateinit var btnUnloadModel: Button
    private lateinit var btnStop: Button
    private lateinit var etInput: EditText
    private lateinit var btnSend: Button
    private lateinit var btnClearHistory: Button
    private lateinit var btnAddImage: Button
    private lateinit var btnAudioRecord: Button

    private lateinit var recyclerView: RecyclerView
    private lateinit var adapter: ChatAdapter
    private lateinit var bottomPanel: LinearLayout
    private lateinit var btnAudioDone: Button
    private lateinit var btnAudioCancel: Button

    private lateinit var scrollImages: HorizontalScrollView
    private lateinit var topScrollContainer: LinearLayout
    private lateinit var llLoading: LinearLayout
    private lateinit var vTip: View

    private lateinit var llmWrapper: LlmWrapper
    private lateinit var vlmWrapper: VlmWrapper
    var embedderWrapper: EmbedderWrapper? = null
    private lateinit var rerankerWrapper: RerankerWrapper
    private lateinit var cvWrapper: CvWrapper
    private lateinit var asrWrapper: AsrWrapper
    private val modelScope = CoroutineScope(Dispatchers.IO)

    private val chatList = arrayListOf<ChatMessage>()
    private lateinit var llmSystemPrompt: ChatMessage
    private val vlmChatList = arrayListOf<VlmChatMessage>()
    private lateinit var vlmSystemPrompty: VlmChatMessage
    private lateinit var modelList: List<ModelData>
    private var selectModelId = ""
    private var spinnerText = ""

    // ADD: Track which model type is loaded
    private var isLoadLlmModel = false
    private var isLoadVlmModel = false
    private var isLoadEmbedderModel = false
    private var isLoadRerankerModel = false
    private var isLoadCVModel = false
    private var isLoadAsrModel = false

    private var enableThinking = false

    private var wavRecorder: WavRecorder? = null
    private var audioFile: File? = null

    private var lastFormattedPrompt: String? = null  // Store the last formatted prompt
    private var retrievedChunksList = mutableListOf<EmbedResultBean>()  // Store retrieved chunks

    private val savedImageFiles = mutableListOf<File>()
    private val messages = arrayListOf<Message>()

    private val embedMsgList = arrayListOf<ChatMessage>()
    private val embedResultList = arrayListOf<EmbedResultBean>()
    private lateinit var selectFolderResult: ActivityResultLauncher<Intent>

    private lateinit var chunkAdapter: ChunkAdapter

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        immersionBar {
            statusBarColorInt(Color.WHITE)
            statusBarDarkFont(true)
        }
        requestPermissions(arrayOf(Manifest.permission.RECORD_AUDIO), 1002)
        okdownload()
        initData()
        initView()
        setListeners()
    }

    private val embedStringPrefix =
        "You are a careful assistant. Use ONLY the provided context to answer."

    private fun createEmbedChatMessage(embedString: String): ChatMessage {
        val embedString =
            "$embedStringPrefix\n\n<context>\\n$embedString\\n</context>"
        return ChatMessage(role = "system", embedString)
    }

    private fun addEmbedChatMessage(embedString: String, msgList: ArrayList<ChatMessage>) {
        createEmbedChatMessage(embedString).let { msg ->
            var hasAdded = false
            embedMsgList.forEach {
                if (it.content == msg.content) {
                    hasAdded = true
                    return@forEach
                }
            }
            if (!hasAdded) {
                embedMsgList.add(msg)
            }
        }
        embedMsgList.forEach {
            if (!msgList.contains(it)) {
                msgList.add(it)
            }
        }
    }

    // Format file list with chunk counts for display
    private fun formatFilesInSearch(): String {
        if (embedResultList.isEmpty()) {
            return "Files: 0 | Chunks: 0"
        }

        // Group by file path and count chunks
        val fileChunkCounts = embedResultList.groupBy { it.path }
            .mapValues { it.value.size }
            .toList()
            .sortedBy { File(it.first).name }

        val fileCount = fileChunkCounts.size
        val totalChunks = embedResultList.size

        return "Files: $fileCount | Chunks: $totalChunks"
    }

    private fun resetLoadState() {
        isLoadLlmModel = false
        isLoadVlmModel = false
        isLoadEmbedderModel = false
        isLoadRerankerModel = false
        isLoadCVModel = false
        isLoadAsrModel = false
    }

    private fun initView() {
        chunkAdapter = ChunkAdapter()
        binding.rvCitation.adapter = chunkAdapter
        adapter = ChatAdapter(messages, object : OnCitationsClick {
            override fun onClick(position: Int, retrievedChunks: List<EmbedResultBean>) {
                chunkAdapter.updateData(retrievedChunks)
                binding.llCitations.visibility = View.VISIBLE
            }
        })
        binding.rvChat.adapter = adapter

        llDownloading = findViewById(R.id.ll_downloading)
        tvDownloadProgress = findViewById(R.id.tv_download_progress)
        pbDownloading = findViewById(R.id.pb_downloading)
        spModelList = findViewById(R.id.sp_model_list)
        spModelList.dropDownVerticalOffset = DensityUtil.dpToPx(this, 40f)
        spModelList.post {
            spModelList.dropDownWidth = spModelList.width - DensityUtil.dpToPx(this, 20f)
        }
        val spinnerData = modelList.filter { it.show }.map {
            val map = mutableMapOf<String, String>()
            map["modelId"] = it.id
            map["displayName"] = it.displayName
            map
        }
        spModelList.adapter = object : SimpleAdapter(
            this,
            spinnerData,
            R.layout.item_model,
            arrayOf("displayName"),
            intArrayOf(R.id.tv_model_id)
        ) {

        }
        spModelList.onItemSelectedListener = object : AdapterView.OnItemSelectedListener {
            override fun onItemSelected(
                parent: AdapterView<*>?, view: View?, position: Int, id: Long
            ) {
                selectModelId = spinnerData[position].get("modelId") ?: ""
                spinnerText = spinnerData[position].get("displayName") ?: ""
                Log.d(TAG, "spinnerText:$spinnerText")
                // Use actual file existence check (not just SharedPreferences)
                val selectedModel = modelList.first { it.id == selectModelId }
                val filesExist = isModelDownloaded(selectedModel) == null
                changeOperationUI(
                    if (filesExist) {
                        OperationState.DOWNLOADED
                    } else {
                        OperationState.DEFAULT
                    }
                )
                messages.clear()
                adapter.notifyDataSetChanged()
                binding.rvChat.scrollTo(0, 0)
            }

            override fun onNothingSelected(parent: AdapterView<*>?) {
                selectModelId = ""
            }
        }
        btnDownload = findViewById(R.id.btn_download)
        btnLoadModel = findViewById(R.id.btn_load_model)
        btnUnloadModel = findViewById(R.id.btn_unload_model)
        btnStop = findViewById(R.id.btn_stop)
        etInput = findViewById(R.id.et_input)
        btnAddImage = findViewById(R.id.btn_add_image)
        btnAudioRecord = findViewById(R.id.btn_voice)

        bottomPanel = findViewById(R.id.bottom_panel)
        btnAudioCancel = findViewById(R.id.btn_audio_cancel)
        btnAudioDone = findViewById(R.id.btn_audio_done)

        btnSend = findViewById(R.id.btn_send)
        btnClearHistory = findViewById(R.id.btn_clear_history)
        scrollImages = findViewById(R.id.scroll_images)
        topScrollContainer = findViewById(R.id.ll_images_container)
        llLoading = findViewById(R.id.ll_loading)
        vTip = findViewById<View>(R.id.v_tip)

        btnAudioCancel.setOnClickListener {
            stopRecord(true)
        }

        btnAudioDone.setOnClickListener {
            stopRecord(false)
        }

        findViewById<Button>(R.id.btn_test).setOnClickListener {
            Thread {
                val exeFile = File(filesDir, "nexa_test_llm")
                val chmodProcess = Runtime.getRuntime().exec("chmod 755 " + exeFile.absolutePath);
                chmodProcess.waitFor()
                Log.d("nfl", "exeFile exe? ${exeFile.canExecute()}")
                Log.d("nfl", "Exe Thread:${Thread.currentThread().name}")
                ExecShell().executeCommand(
                    arrayOf(
                        //                        exeFile.absolutePath,
//                        "--test-suite=\"npu\"", "--success "
                        "cat",
                        "/sys/devices/soc0/sku"
//                        "/data/local/tmp/test_cat.txt"
                    )
                ).forEach {
                    Log.d("nfl", "cmd:$it")
                }
            }.start()
        }

        findViewById<View>(R.id.v_tip).setOnClickListener {
            Toast.makeText(this, "please unload model first", Toast.LENGTH_SHORT).show()
        }
    }

    private fun replaceFragment(fragment: Fragment) {
        runOnUiThread {
            Log.d(TAG, "replaceFragment:$fragment")
            binding.flIndex.visibility = View.VISIBLE
            supportFragmentManager.beginTransaction()
                .replace(R.id.fl_index, fragment)
                .commit()
        }
    }

    private fun parseModelList() {
        try {
            val baseJson = assets.open("model_list.json").bufferedReader().use { it.readText() }
            modelList = Json.decodeFromString<List<ModelData>>(baseJson)
        } catch (e: Exception) {
            Log.e("nfl", "parseModelList: $e")
        }
    }

    private fun embedFiles(files: ArrayList<String>, clearOld: Boolean = false) {
        binding.tvDatabaseInfo.text = "Embedding..."
        binding.btnEmbedClear.isEnabled = false
        val modelData =
            modelList.first { it.id == "embeddinggemma-300m-npu-mobile" }
        CoroutineScope(Dispatchers.IO).launch {
            GenerateEmbedStringsUtil.load(this@MainActivity, modelData)
            if (clearOld) {
                embedResultList.clear()
            }
            Log.d(TAG, "embedFiles size:${files.size}")
            val embedStartTime = System.currentTimeMillis()
            GenerateEmbedStringsUtil.embed(files.toTypedArray()).let {
                embedResultList.addAll(it)
                var prefillSpeed = 0.0
                it.forEach { item ->
                    prefillSpeed += item.embedResult.profileData.prefillSpeed
                }
                prefillSpeed /= it.size
                val totalTime = System.currentTimeMillis() - embedStartTime
                runOnUiThread {
                    binding.tvDatabaseInfo.text = formatFilesInSearch()
                    binding.llEmbedProfilling.visibility = View.VISIBLE
                    binding.btnEmbedClear.isEnabled = true
                    binding.tvEmbedProfilling.text = "Total Time: ${
                        String.format(
                            null,
                            "%.2f",
                            totalTime / 1000f
                        )
                    }s; Prefilling Speed: ${
                        String.format(
                            null,
                            "%.2f",
                            prefillSpeed
                        )
                    } tokens/sec"
                }
            }
        }
    }

    private var embedFiles = arrayListOf<String>()

    /**
     * Step 0. Preparing to download the model file.
     */
    private fun initData() {
        onBackPressedDispatcher.addCallback(object : OnBackPressedCallback(true) {
            override fun handleOnBackPressed() {
                if (binding.llCitations.visibility == View.VISIBLE) {
                    binding.llCitations.visibility = View.GONE
                } else {
                    isEnabled = false
                    onBackPressedDispatcher.onBackPressed()
                }
            }

        })
        selectFolderResult = registerForActivityResult(
            ActivityResultContracts.StartActivityForResult()
        ) { result -> //
            if (Activity.RESULT_OK == result.resultCode) {
                result.data?.getStringArrayListExtra(FolderActivity.KEY_SELECT_IMAGES)
                    ?.let { files ->
                        embedFiles.addAll(files)
                        embedFiles(files)
                    }
            }
        }
        spDownloaded = getSharedPreferences(
            SharePreferenceKeys.FileName.ModelDownloaded.fileName,
            MODE_PRIVATE
        )
//        spDownloaded.edit().putBoolean("Qwen3-0.6B-Q8_0", false).commit()
//        spDownloaded.edit().putBoolean("Qwen3-0.6B-IQ4_NL", false).commit()
//        spDownloaded.edit().putBoolean("LFM2-1.2B-npu", false).commit()
//        spDownloaded.edit().putBoolean("embeddinggemma-300m-npu", false).commit()
//        spDownloaded.edit().putBoolean("jina-v2-rerank-npu", false).commit()
//        spDownloaded.edit().putBoolean("paddleocr-npu", false).commit()
//        spDownloaded.edit().putBoolean("parakeet-tdt-0.6b-v3-npu", false).commit()
//        spDownloaded.edit().putBoolean("OmniNeural-4B", false).commit()
//        spDownloaded.edit().putBoolean("Granite-4.0-h-350M-NPU", false).commit()
//        spDownloaded.edit().putBoolean("Granite-4-Micro-NPU", false).commit()
        parseModelList()
        //
        initNexaSdk()
        //
        val sysPrompt = """\
You are Nays Campaign Manager, an AI assistant responsible for managing customer campaigns and investigating campaign-related issues.

When a customer inquiry comes in, you need to:
1. Analyze the customer's request to understand their campaign needs
2. Check if it's related to campaign limits or issues
3. Use the campaign_investigation function when needed to check campaign status
4. Provide appropriate responses based on the investigation results

Your responsibilities include:
- Investigating campaign performance and limits
- Determining if customers have reached their campaign limits
- Providing helpful messages when limits are reached
- Directing customers to support when limits haven't been reached
- Ensuring smooth campaign operations for all customers

When you receive a query about campaigns, you should:
1. First understand what the customer is asking about
2. If it's campaign-related, use the campaign_investigation tool to check the status
3. Based on the tool's response, provide appropriate guidance

Always be professional, helpful, and focused on resolving campaign-related issues efficiently.

Note: You must use the campaign_investigation function whenever a customer asks about campaign limits, issues, or status.
"""
        // It works better with Chinese prompt words.
        val sysPrompt2 = "Must reply in markdown format"
//        addSystemPrompt(sysPrompt2)
    }

    /**
     * Step 1. initNexaSdk environment
     */
    private fun initNexaSdk() {
        // Initialize NexaSdk with context
        NexaSdk.getInstance().init(this, object : NexaSdk.InitCallback {
            override fun onSuccess() {
            }

            override fun onFailure(reason: String) {
                Log.e(TAG, "NexaSdk init failed: $reason")
            }
        })

        val testLocalPath = false
        if (testLocalPath) {
            // FIXME: Set directory according to terminal format
            val pluginNativeLibPath = filesDir.absolutePath
            val pluginAdspLibPath = File(filesDir, "npu/htp-files").absolutePath
            val pluginLdLibraryPath =
                "$pluginNativeLibPath:$pluginNativeLibPath/npu:$pluginAdspLibPath:\$LD_LIBRARY_PATH"
            // FIXME: Set directory with flattened .so files
            val NEXA_PLUGIN_PATH = pluginNativeLibPath
            val LD_LIBRARY_PATH = pluginLdLibraryPath
            val ADSP_LIBRARY_PATH = pluginAdspLibPath
            Log.d("nfl", "NEXA_PLUGIN_PATH:$NEXA_PLUGIN_PATH")
            Log.d("nfl", "LD_LIBRARY_PATH:$LD_LIBRARY_PATH")
            Log.d("nfl", "ADSP_LIBRARY_PATH:$ADSP_LIBRARY_PATH")

            Os.setenv("NEXA_PLUGIN_PATH", NEXA_PLUGIN_PATH, true)
            Os.setenv("LD_LIBRARY_PATH", LD_LIBRARY_PATH, true)
            Os.setenv("ADSP_LIBRARY_PATH", ADSP_LIBRARY_PATH, true)
        }
    }

    /**
     * Step 2. add system prompt, such as : output markdown style, contains emoji etc.(Options)
     */
    private fun addSystemPrompt(sysPrompt: String) {
        llmSystemPrompt = ChatMessage("system", sysPrompt)
        chatList.add(llmSystemPrompt)
        vlmSystemPrompty =
            VlmChatMessage(
                "system",
                listOf(VlmContent("text", sysPrompt))
            )
        vlmChatList.add(vlmSystemPrompty)
    }

    private fun getHfToken(model: ModelData, url: String): String? {
        // Replace with your own HuggingFace token if needed for private models
        return null
    }

    private fun onLoadModelSuccess(tip: String) {
        runOnUiThread {
            Toast.makeText(
                this@MainActivity, tip, Toast.LENGTH_SHORT
            ).show()
            when (pluginType) {
                "cpu" -> {
                    binding.ivDeviceTag.setImageResource(R.drawable.icon_tag_cpu)
                }

                "gpu" -> {
                    binding.ivDeviceTag.setImageResource(R.drawable.icon_tag_gpu)
                }

                "npu" -> {
                    binding.ivDeviceTag.setImageResource(R.drawable.icon_tag_npu)
                }

                "npu_llama" -> {
                    binding.ivDeviceTag.setImageResource(R.drawable.icon_tag_npu)
                }
            }
            changeOperationUI(OperationState.LOADED)
            // change UI
            btnAddImage.visibility = View.INVISIBLE
            btnAudioRecord.visibility = View.INVISIBLE
            if (isLoadVlmModel) {
                btnAddImage.visibility = View.VISIBLE
                btnAudioRecord.visibility = View.VISIBLE
            }
            if (isLoadCVModel) {
                btnAddImage.visibility = View.VISIBLE
            }
            if (isLoadAsrModel) {
                btnAudioRecord.visibility = View.VISIBLE
            }
            //
            btnUnloadModel.visibility = View.VISIBLE
            llLoading.visibility = View.INVISIBLE
            //
            if (isLoadEmbedderModel || isLoadRerankerModel || isLoadAsrModel || isLoadCVModel) {
                btnStop.visibility = View.GONE
            } else {
                btnStop.visibility = View.VISIBLE
            }
        }
    }

    private fun onLoadModelFailed(tip: String) {
        runOnUiThread {
            vTip.visibility = View.GONE

            // Check if files exist locally first
            val selectModelData = modelList.firstOrNull { it.id == selectModelId }
            val fileName = isModelDownloaded(selectModelData!!)
            val filesExist = fileName == null

            if (!filesExist) {
                Toaster.showLong("The \"$fileName\" file is missing. Please download it first.")
            } else {
                Toast.makeText(this@MainActivity, tip, Toast.LENGTH_SHORT)
                    .show()
            }

            // change UI
            btnAddImage.visibility = View.INVISIBLE
            btnAudioRecord.visibility = View.INVISIBLE
            btnUnloadModel.visibility = View.GONE
            llLoading.visibility = View.INVISIBLE
        }
    }

    private fun hasLoadedModel(): Boolean {
        return isLoadLlmModel || isLoadVlmModel || isLoadEmbedderModel ||
                isLoadRerankerModel || isLoadCVModel || isLoadAsrModel
    }

    /**
     * Helper function to check if model files exist locally
     * @return null if all files exist locally. or file's name which is missing.
     */
    private fun isModelDownloaded(modelData: ModelData): String? {
        val modelDir = modelData.modelDir(this@MainActivity)
        val fileName = modelData.getNonExistModelFile(this, modelDir, modelList)
        val filesExist = fileName == null
        // Sync SharedPreferences with actual file existence (both directions)
        if (filesExist && !spDownloaded.getBoolean(modelData.id, false)) {
            Log.d(TAG, "Model files found locally for ${modelData.id}, updating SharedPreferences to true")
            spDownloaded.edit().putBoolean(modelData.id, true).commit()
        } else if (!filesExist && spDownloaded.getBoolean(modelData.id, false)) {
            Log.d(TAG, "Model files missing for ${modelData.id}, updating SharedPreferences to false")
            spDownloaded.edit().putBoolean(modelData.id, false).commit()
        }
        return fileName
    }

    private var pluginType = "npu"
    private fun loadModel(selectModelData: ModelData, modelDataPluginId: String, nGpuLayers: Int) {
        modelScope.launch {
            resetLoadState()
            Log.d(TAG, "load model selectModelData.id:${selectModelData.id}")
            val nexaManifestBean = selectModelData.getNexaManifest(this@MainActivity)
            var pluginId = nexaManifestBean?.PluginId ?: modelDataPluginId
            pluginType = pluginId
            Log.d(TAG, "pluginType:$pluginType")
            var deviceId: String? = null
            var nGpuLayers = nGpuLayers
            when (pluginId) {
                "cpu" -> {
                    pluginId = "cpu_gpu"
                    nGpuLayers = 0
                }

                "gpu" -> {
                    pluginId = "cpu_gpu"
                    deviceId = DeviceIdValue.GPU.value
                    nGpuLayers = 999
                }

                "npu_llama" -> {
                    pluginId = "cpu_gpu"
                    deviceId = DeviceIdValue.NPU.value
                    nGpuLayers = 999
                }
            }

            when (nexaManifestBean?.ModelType ?: selectModelData.type) {
                "chat", "llm" -> {

                    val conf = ModelConfig(
                        nCtx = 4096,
                        nGpuLayers = nGpuLayers,
                        enable_thinking = enableThinking,
                        npu_lib_folder_path = applicationInfo.nativeLibraryDir,
                        npu_model_folder_path = selectModelData.modelDir(this@MainActivity).absolutePath
                    )
                    // Build and initialize LlmWrapper for chat model
                    LlmWrapper.builder().llmCreateInput(
                        LlmCreateInput(
                            model_name = nexaManifestBean?.ModelName ?: "",
                            model_path = selectModelData.modelFile(this@MainActivity)!!.absolutePath,
                            tokenizer_path = selectModelData.tokenFile(this@MainActivity)?.absolutePath,
                            config = conf,
                            plugin_id = pluginId,
                            device_id = deviceId
                        )
                    ).build().onSuccess { wrapper ->
                        isLoadLlmModel = true
                        llmWrapper = wrapper
                        onLoadModelSuccess("llm model loaded")
                    }.onFailure { error ->
                        onLoadModelFailed(error.message.toString())
                    }

                }

                "embedder" -> {
                    // Handle embedder model loading with NPU paths using EmbedderCreateInput
                    // embed-gemma
                    val embedderCreateInput = EmbedderCreateInput(
                        model_name = nexaManifestBean?.ModelName
                            ?: "",  // Model name for NPU plugin
                        model_path = selectModelData.modelFile(this@MainActivity)!!.absolutePath,
                        tokenizer_path = selectModelData.tokenFile(this@MainActivity)?.absolutePath,
                        config = ModelConfig(
                            npu_lib_folder_path = applicationInfo.nativeLibraryDir,
                            npu_model_folder_path = selectModelData.modelDir(this@MainActivity).absolutePath,
                            nGpuLayers = nGpuLayers
                        ),
                        plugin_id = pluginId,
                        device_id = deviceId
                    )

                    EmbedderWrapper.builder()
                        .embedderCreateInput(embedderCreateInput)
                        .build().onSuccess { wrapper ->
                            isLoadEmbedderModel = true
                            embedderWrapper = wrapper
                            onLoadModelSuccess("embedder model loaded")
                            if (selectModelData.id == "embedneural-npu") {
                                runOnUiThread {
                                    binding.ivTopk.visibility = View.VISIBLE
                                }
                                replaceFragment(IndexFragment.newInstance("", ""))
                            }
                        }.onFailure { error ->
                            onLoadModelFailed(error.message.toString())
                        }

                }

                "reranker" -> {
                    // Handle reranker model loading with NPU paths using RerankerCreateInput
                    // jina-v2-rerank-npu
                    val rerankerCreateInput = RerankerCreateInput(
                        model_name = nexaManifestBean?.ModelName
                            ?: "",  // Model name for NPU plugin
                        model_path = selectModelData.modelFile(this@MainActivity)!!.absolutePath,
                        tokenizer_path = selectModelData.tokenFile(this@MainActivity)?.absolutePath,
                        config = ModelConfig(
                            npu_lib_folder_path = applicationInfo.nativeLibraryDir,
                            npu_model_folder_path = selectModelData.modelDir(this@MainActivity).absolutePath,
                            nGpuLayers = nGpuLayers
                        ),
                        plugin_id = pluginId,
                        device_id = deviceId
                    )

                    RerankerWrapper.builder()
                        .rerankerCreateInput(rerankerCreateInput)
                        .build().onSuccess { wrapper ->
                            isLoadRerankerModel = true
                            rerankerWrapper = wrapper
                            onLoadModelSuccess("reranker model loaded")
                        }.onFailure { error ->
                            onLoadModelFailed(error.message.toString())
                        }

                }

                "paddleocr" -> {
                    // paddleocr-npu
                    val cvCreateInput = CVCreateInput(
                        model_name = nexaManifestBean?.ModelName ?: "",
                        config = CVModelConfig(
                            capabilities = CVCapability.OCR,
                            det_model_path = selectModelData.modelDir(this@MainActivity).absolutePath,
                            rec_model_path = selectModelData.modelFile(this@MainActivity)!!.absolutePath,
                            char_dict_path = selectModelData.modelDir(this@MainActivity).absolutePath,
                            npu_model_folder_path = selectModelData.modelDir(this@MainActivity).absolutePath,
                            npu_lib_folder_path = applicationInfo.nativeLibraryDir
                        ),
                        plugin_id = pluginId,
                        device_id = deviceId
                    )
                    CvWrapper.builder()
                        .createInput(cvCreateInput)
                        .build().onSuccess {
                            isLoadCVModel = true
                            cvWrapper = it
                            onLoadModelSuccess("paddleocr model loaded")
                        }.onFailure { error ->
                            onLoadModelFailed(error.message.toString())
                        }
                }

                "asr" -> {
                    // ADD: Handle ASR model loading
                    // parakeet-tdt-0.6b-v3-npu
                    val asrCreateInput = AsrCreateInput(
                        model_name = nexaManifestBean?.ModelName ?: "",
                        model_path = selectModelData.modelFile(this@MainActivity)!!.absolutePath,
                        config = ModelConfig(
                            npu_lib_folder_path = applicationInfo.nativeLibraryDir,
                            npu_model_folder_path = selectModelData.modelDir(this@MainActivity).absolutePath,
                            nGpuLayers = nGpuLayers
                        ),
                        plugin_id = pluginId,
                        device_id = deviceId
                    )

                    AsrWrapper.builder()
                        .asrCreateInput(asrCreateInput)
                        .build().onSuccess { wrapper ->
                            isLoadAsrModel = true
                            asrWrapper = wrapper
                            onLoadModelSuccess("ASR model loaded")
                        }.onFailure { error ->
                            onLoadModelFailed(error.message.toString())
                        }
                }

                "multimodal", "vlm" -> {
                    // VLM model
                    val isNpuVlm = nexaManifestBean?.PluginId == "npu"
                    val config = if (isNpuVlm) {
                        ModelConfig(
                            nCtx = 2048,
                            nThreads = 8,
                            enable_thinking = enableThinking,
                            npu_lib_folder_path = applicationInfo.nativeLibraryDir,
                            npu_model_folder_path = selectModelData.modelDir(this@MainActivity).absolutePath
                        )
                    } else {
                        ModelConfig(
                            nCtx = 1024,
                            nThreads = 4,
                            nBatch = 1,
                            nUBatch = 1,
                            nGpuLayers = nGpuLayers,
                            enable_thinking = enableThinking
                        )
                    }

                    val vlmCreateInput = VlmCreateInput(
                        model_name = nexaManifestBean?.ModelName ?: "",
                        model_path = selectModelData.modelFile(this@MainActivity)!!.absolutePath,
                        mmproj_path = selectModelData.mmprojTokenFile(this@MainActivity)?.absolutePath,
                        config = config,
                        plugin_id = pluginId,
                        device_id = deviceId
                    )

                    VlmWrapper.builder()
                        .vlmCreateInput(vlmCreateInput)
                        .build().onSuccess {
                            isLoadVlmModel = true
                            vlmWrapper = it
                            onLoadModelSuccess("vlm model loaded")
                        }.onFailure { error ->
                            onLoadModelFailed(error.message.toString())
                        }
                }

                else -> {
                    onLoadModelFailed("model type error")
                }
            }
        }
    }

    private fun addDownloadedTag(modelData: ModelData) {
        spDownloaded.edit(commit = true) {
            putBoolean(modelData.id, true)
        }
        modelList.firstOrNull { it.id == modelData.dependencies?.first()}?.let {
            addDownloadedTag(it)
        }
    }

    private fun downloadModel(selectModelData: ModelData) {
        // Check local files first before SharedPreferences
        val fileName = isModelDownloaded(selectModelData)
        if (fileName == null || hasLoadedModel()) {
            addDownloadedTag(selectModelData)
            llDownloading.visibility = View.GONE
            changeOperationUI(OperationState.DOWNLOADED)
            Toast.makeText(this@MainActivity, "model already downloaded", Toast.LENGTH_SHORT)
                .show()
        } else {
            downloadState = DownloadState.DOWNLOADING
            downloadingModelData = selectModelData
            llDownloading.visibility = View.VISIBLE
            tvDownloadProgress.text = "0%"
            modelScope.launch {
                val unsafeClient = getUnsafeOkHttpClient().build()
                
                // Track URL mapping for fallback: primary URL -> fallback URL
                val fallbackUrlMap = mutableMapOf<String, String>()
                // Track failed downloads for fallback retry
                val failedDownloads = mutableListOf<DownloadableFileWithFallback>()
                
                // For NPU models without explicit files list, fetch file list with fallback support
                val filesToDownloadWithFallback: List<DownloadableFileWithFallback> = if (selectModelData.isNpuModel() && 
                    selectModelData.files.isNullOrEmpty() && 
                    !selectModelData.baseUrl.isNullOrEmpty()) {
                    
                    Log.d(TAG, "NPU model detected, fetching file list: ${selectModelData.baseUrl}")
                    
                    // Recursively collect all NPU models that need file listing (main model + all NPU dependencies in the tree)
                    val modelsToFetch = mutableListOf(selectModelData)
                    val visited = mutableSetOf(selectModelData.id)

                    fun collectNpuDependencies(model: ModelData) {
                        model.dependencies?.forEach { depId ->
                            if (depId in visited) return@forEach
                            visited.add(depId)

                            modelList.firstOrNull { it.id == depId }?.let { depModel ->
                                // Check actual file existence, not just SharedPreferences
                                val depModelDir = depModel.modelDir(this@MainActivity)
                                val depFilesExist = depModel.allModelFilesExist(this@MainActivity, depModelDir, modelList)

                                // If it's an NPU model that needs S3 file listing, add it
                                if (depModel.isNpuModel() && 
                                    depModel.files.isNullOrEmpty() && 
                                    !depModel.baseUrl.isNullOrEmpty() &&
                                    !depFilesExist) {
                                    modelsToFetch.add(depModel)
                                    Log.d(TAG, "Added NPU dependency to fetch list: ${depModel.id}")
                                }

                                // Recursively check this dependency's dependencies
                                collectNpuDependencies(depModel)
                            }
                        }
                    }

                    collectNpuDependencies(selectModelData)

                    // Fetch file lists with fallback support for all NPU models in parallel
                    val fileListResults = modelsToFetch.map { model ->
                        async {
                            model to ModelFileListingUtil.listFilesWithFallback(model.baseUrl!!, unsafeClient)
                        }
                    }.awaitAll().toMap()
                    
                    // Build the download list using the fetched file names
                    val allFiles = mutableListOf<DownloadableFileWithFallback>()
                    modelsToFetch.forEach { model ->
                        val result = fileListResults[model]
                        if (result == null || result.files.isEmpty()) {
                            Log.e(TAG, "Failed to fetch file list for ${model.id}")
                        } else {
                            val useHfUrls = result.source == ModelFileListingUtil.FileListResult.Source.HUGGINGFACE
                            Log.d(TAG, "Found ${result.files.size} files for ${model.id} from ${result.source}: ${result.files}")
                            
                            allFiles.addAll(
                                model.downloadableFilesWithFallback(
                                    this@MainActivity,
                                    model.modelDir(this@MainActivity),
                                    modelList,
                                    result.files,
                                    useHfUrls
                                ).filter { df -> 
                                    // Filter out dependencies that were already added
                                    allFiles.none { existing -> existing.primaryUrl == df.primaryUrl }
                                }
                            )
                        }
                    }
                    allFiles
                } else {
                    // For non-NPU models or models with explicit files, use the original method with fallback
                    selectModelData.downloadableFiles(
                        this@MainActivity,
                        selectModelData.modelDir(this@MainActivity),
                        modelList
                    ).withFallbackUrls()
                }
                
                // Build fallback URL map
                filesToDownloadWithFallback.forEach { 
                    fallbackUrlMap[it.primaryUrl] = it.fallbackUrl
                }
                
                // Convert to simple DownloadableFile for initial download attempt
                val filesToDownload = filesToDownloadWithFallback.map { 
                    DownloadableFile(it.file, it.primaryUrl)
                }
                
                Log.d(TAG, "filesToDownload: $filesToDownload")
                if (filesToDownload.isEmpty()) throw IllegalArgumentException("No download URL")

                fun getUrlFileSize(client: OkHttpClient, url: String): Long {
                    val hostname = try {
                        url.substringAfter("://").substringBefore("/")
                    } catch (e: Exception) {
                        "unknown"
                    }

                    Log.d(TAG, "Requesting file size: $hostname")

                    val builder = Request.Builder().url(url).head()
                    getHfToken(selectModelData, url)?.let {
                        builder.addHeader("Authorization", "Bearer $it")
                    }
                    val request = builder.build()
                    var size = 0L
                    try {
                        client.newCall(request).execute().use { resp ->
                            size = resp.header("Content-Length")?.toLongOrNull() ?: 0L
                        }
                    } catch (e: java.net.UnknownHostException) {
                        Log.e(TAG, "DNS resolution failed for $hostname - Check DNS/network")
                    } catch (e: java.net.SocketTimeoutException) {
                        Log.e(TAG, "Connection timeout to $hostname - Possible firewall/proxy issue")
                    } catch (e: java.net.ConnectException) {
                        Log.e(TAG, "Connection refused by $hostname - Server unreachable")
                    } catch (e: javax.net.ssl.SSLException) {
                        Log.e(TAG, "SSL/TLS error to $hostname - ${e.message}")
                    } catch (e: Exception) {
                        Log.e(TAG, "Network error: ${e.javaClass.simpleName} - ${e.message}")
                    }
                    return size
                }
                
                // Try to get file sizes, with fallback to HF if S3 fails
                val fileSizeMap = mutableMapOf<String, Long>()
                val filesToSkip = mutableListOf<String>()
                // Critical file extensions that must be downloaded
                val criticalExtensions = listOf(".nexa", ".gguf", ".manifest")
                filesToDownloadWithFallback.forEach { fileWithFallback ->
                    var size = getUrlFileSize(unsafeClient, fileWithFallback.primaryUrl)
                    if (size == 0L && fileWithFallback.fallbackUrl != fileWithFallback.primaryUrl) {
                        Log.w(TAG, "Primary URL failed, trying fallback for size: ${fileWithFallback.file.name}")
                        size = getUrlFileSize(unsafeClient, fileWithFallback.fallbackUrl)
                        if (size > 0L) {
                            // Update to use fallback URL as primary for this file
                            fallbackUrlMap[fileWithFallback.primaryUrl] = fileWithFallback.primaryUrl // swap
                        }
                    }
                    fileSizeMap[fileWithFallback.primaryUrl] = size
                }
                
                val totalSizes = filesToDownload.map { fileSizeMap[it.url] ?: 0L }
                if (totalSizes.any { it == 0L }) {
                    runOnUiThread {
                        downloadState = DownloadState.IDLE
                        llDownloading.visibility = View.GONE
                        Toaster.show("Download failed - could not get file sizes.")
                    }
                    return@launch
                }
                
                val alreadyDownloaded = mutableMapOf<String, Long>()
                val totalBytes = totalSizes.sum()
                Log.d(TAG, "all model size: $totalBytes")

                val startTime = System.currentTimeMillis()
                var lastProgressTime = 0L
                val progressInterval = 500L
                
                fun onProgress(
                    modelId: String,
                    percent: Int,
                    downloaded: Long,
                    totalBytes: Long,
                    etaSec: Long,
                    speedStr: String
                ) {
                    runOnUiThread {
                        if (10000 == percent) {
                            changeOperationUI(OperationState.DOWNLOADED)
                            llDownloading.visibility = View.GONE
                            addDownloadedTag(downloadingModelData!!)
                            Toaster.show("${downloadingModelData?.displayName} downloaded")
                        } else {
                            tvDownloadProgress.text = "${String.format("%.2f", percent.toFloat() / 100)}%"
                        }
                    }
                }

                fun reportProgress(force: Boolean = false) {
                    val now = System.currentTimeMillis()
                    if (force || now - lastProgressTime > progressInterval) {
                        val elapsedMs = now - startTime
                        val downloaded = alreadyDownloaded.values.sum()
                        val percent =
                            if (totalBytes > 0) ((downloaded * 100 * 100) / totalBytes).toInt() else 0
                        val speedAvg =
                            if (elapsedMs > 0) downloaded / (elapsedMs / 1000.0) else 0.0
                        val etaSec =
                            if (speedAvg > 0) ((totalBytes - downloaded) / speedAvg).toLong() else -1L
                        val speedStr = if (speedAvg > 1024 * 1024) {
                            String.format("%.2f MB/s", speedAvg / (1024 * 1024))
                        } else {
                            String.format("%.1f KB/s", speedAvg / 1024)
                        }
                        Log.d(TAG, "download percent:$percent")
                        onProgress(selectModelId, percent, downloaded, totalBytes, etaSec, speedStr)
                        lastProgressTime = now
                    }
                }
                
                // Function to start download for a list of files
                fun startDownload(
                    downloadFiles: List<DownloadableFile>,
                    isFallbackAttempt: Boolean = false
                ) {
                    if (downloadFiles.isEmpty()) {
                        if (failedDownloads.isEmpty()) {
                            // All downloads complete
                            downloadState = DownloadState.IDLE
                            onProgress(selectModelId, 100 * 100, totalBytes, totalBytes, 0, "0 KB/s")
                        } else {
                            runOnUiThread {
                                downloadState = DownloadState.IDLE
                                llDownloading.visibility = View.GONE
                                Toaster.show("Download failed for some files.")
                            }
                        }
                        return
                    }
                    
                    val queueSet = DownloadContext.QueueSet()
                        .setParentPathFile(downloadFiles[0].file.parentFile)
                        .setMinIntervalMillisCallbackProcess(300)
                    val builder = queueSet.commit()
                    
                    downloadFiles.forEach { item ->
                        val taskBuilder = DownloadTask.Builder(item.url, item.file)
                        getHfToken(selectModelData, item.url)?.let {
                            taskBuilder.addHeader("Authorization", "Bearer $it")
                        }
                        val task = taskBuilder.build()
                        task.info?.let {
                            alreadyDownloaded[it.url] = it.totalOffset
                        }
                        builder.bindSetTask(task)
                    }
                    
                    val totalCount = filesToDownload.size
                    var currentCount = filesToDownload.size - downloadFiles.size // Already completed
                    val pendingFallbacks = mutableListOf<DownloadableFile>()

                    downloadContext = builder.setListener(createDownloadContextListener {}).build()
                    downloadContext?.start(
                        createListener1(taskStart = { task, _ ->
                            Log.d(TAG, "download task ${task.id} Start${if (isFallbackAttempt) " (fallback)" else ""}")
                        }, retry = { task, _ ->
                            Log.d(TAG, "download task ${task.id} retry")
                        }, connected = { task, _, _, _ ->
                            Log.d(TAG, "download task ${task.id} connected")
                        }, progress = { task, currentOffset, totalLength ->
                            Log.d(TAG, "download task ${task.id} progress $currentOffset $totalLength")
                            alreadyDownloaded[task.url] = currentOffset
                            reportProgress(true)
                        }) { task, cause, exception, _ ->
                            when(cause) {
                                EndCause.CANCELED -> {
                                    // do nothing
                                }

                                EndCause.COMPLETED -> {
                                    Log.d(TAG, "download task ${task.id} end")
                                    currentCount += 1
                                    Log.d(TAG, "download task process currentCount:$currentCount, totalCount:$totalCount")
                                    
                                    // Check if all current batch is done
                                    val batchCompleted = downloadFiles.all { df ->
                                        df.url == task.url || df.file.exists()
                                    }
                                    
                                    if (currentCount >= totalCount) {
                                        downloadState = DownloadState.IDLE
                                        onProgress(selectModelId, 100 * 100, totalBytes, totalBytes, 0, "0 KB/s")
                                    }
                                }

                                else -> {
                                    Log.e(TAG, "download task ${task.id} error: $cause, ${exception?.message}")
                                    
                                    // Try fallback URL if available and not already a fallback attempt
                                    if (!isFallbackAttempt) {
                                        val fallbackUrl = fallbackUrlMap[task.url]
                                        if (fallbackUrl != null && fallbackUrl != task.url && task.file != null) {
                                            Log.w(TAG, "Primary download failed, queuing fallback: ${task.file?.name}")
                                            pendingFallbacks.add(DownloadableFile(task.file!!, fallbackUrl))
                                        } else {
                                            // No fallback available, mark as failed
                                            val failedFile = filesToDownloadWithFallback.find { it.primaryUrl == task.url }
                                            if (failedFile != null) {
                                                failedDownloads.add(failedFile)
                                            }
                                        }
                                    } else {
                                        // Fallback also failed
                                        val failedFile = filesToDownloadWithFallback.find { 
                                            it.primaryUrl == task.url || it.fallbackUrl == task.url 
                                        }
                                        if (failedFile != null) {
                                            failedDownloads.add(failedFile)
                                        }
                                    }
                                    
                                    // Check if batch is done (including failures)
                                    currentCount += 1
                                    if (currentCount >= totalCount && pendingFallbacks.isEmpty()) {
                                        if (failedDownloads.isEmpty()) {
                                            downloadState = DownloadState.IDLE
                                            onProgress(selectModelId, 100 * 100, totalBytes, totalBytes, 0, "0 KB/s")
                                        } else {
                                            runOnUiThread {
                                                downloadState = DownloadState.IDLE
                                                llDownloading.visibility = View.GONE
                                                Toaster.show("Download failed for ${failedDownloads.size} file(s).")
                                            }
                                        }
                                    } else if (pendingFallbacks.isNotEmpty()) {
                                        // Start fallback downloads
                                        Log.d(TAG, "Starting ${pendingFallbacks.size} fallback downloads")
                                        modelScope.launch {
                                            startDownload(pendingFallbacks.toList(), isFallbackAttempt = true)
                                        }
                                        pendingFallbacks.clear()
                                    }
                                }
                            }
                        }, true
                    )
                }
                
                // Start initial download with primary URLs
                startDownload(filesToDownload)
            }
        }
    }

    private fun unloadModel(showUnloadTip: Boolean = true) {
        val handleUnloadResult = fun(result: Int) {
            resetLoadState()
            runOnUiThread {
                binding.btnClearHistory.callOnClick()
                changeOperationUI(OperationState.DOWNLOADED)
                vTip.visibility = View.GONE
                btnAddImage.visibility = View.INVISIBLE
                btnAudioRecord.visibility = View.INVISIBLE
                if (showUnloadTip) {
                    Toast.makeText(
                        this@MainActivity, if (result == 0) {
                            "unload success"
                        } else {
                            "unload failed and error code: $result"
                        }, Toast.LENGTH_SHORT
                    ).show()
                }
            }
        }
        modelScope.launch {
            if (isLoadVlmModel) {
                vlmWrapper.stopStream()
                vlmWrapper.destroy()
                vlmChatList.clear()
                // TODO:
                handleUnloadResult(0)
            } else if (isLoadEmbedderModel) {
                // ADD: Unload embedder
                embedderWrapper!!.destroy()
                runOnUiThread {
                    binding.ivTopk.visibility = View.GONE
                    supportFragmentManager.beginTransaction().apply {
                        supportFragmentManager.fragments.forEach {
                            if (it is IndexFragment) {
                                this.remove(it)
                            }
                        }
                    }.commit()
                }
                // TODO:
                handleUnloadResult(0)
            } else if (isLoadRerankerModel) {
                // ADD: Unload reranker
                handleUnloadResult(rerankerWrapper.destroy())
            } else if (isLoadCVModel) {
                // ADD: Unload CV model
                cvWrapper.destroy()
                // TODO:
                handleUnloadResult(0)
            } else if (isLoadAsrModel) {
                // ADD: Unload ASR model
                asrWrapper.destroy()
                // TODO:
                handleUnloadResult(0)
            } else if (isLoadLlmModel) {
                llmWrapper.stopStream()
                llmWrapper.destroy()
                chatList.clear()
                // TODO:
                handleUnloadResult(0)
            } else {
                handleUnloadResult(0)
            }
        }
    }

    private var isEmbedProfillingResultExpand = true
    private fun setListeners() {
        binding.btnTestEmbed.setOnClickListener {
            if (PermissionUtil.checkManageStoragePermission(this)) {
                selectFolderResult.launch(Intent(this, FolderActivity::class.java).apply {
                    this.putExtra(FolderActivity.KEY_SELECT_TYPE, 1)
                })
            } else {
                PermissionUtil.showRequestManageStoragePermissionDialog(this as ComponentActivity)
            }
        }
        binding.btnEmbedClear.setOnClickListener {
            embedResultList.clear()
            embedFiles.clear()
            binding.llEmbedProfilling.visibility = View.GONE
            binding.tvDatabaseInfo.text = formatFilesInSearch()
            binding.btnEmbedClear.isEnabled = false
        }
        binding.llEmbedResultTitle.setOnClickListener {
            isEmbedProfillingResultExpand = !isEmbedProfillingResultExpand
            binding.ivEmbedExpand.setImageResource(if (isEmbedProfillingResultExpand) R.drawable.icon_arrow_down else R.drawable.icon_arrow_right)
            binding.tvEmbedProfilling.visibility =
                if (isEmbedProfillingResultExpand) View.VISIBLE else View.GONE
        }

        binding.ivClose.setOnClickListener {
            binding.llCitations.visibility = View.GONE
        }

        binding.ivConfig.setOnClickListener {
            val chunkSize = RAGConfig.CHUNK_SIZE
            val nChunk = RAGConfig.N_CHUNKS
            var newChunkSize = chunkSize
            var newNChunk = nChunk

            val configBinding = DialogConfigBinding.inflate(layoutInflater)
            fun checkChangedConfig() {
                if (chunkSize != newChunkSize || nChunk != newNChunk) {
                    configBinding.btnReIndex.isEnabled = true
                    configBinding.tvConfigChangedTip.visibility = View.VISIBLE
                } else {
                    configBinding.btnReIndex.isEnabled = false
                    configBinding.tvConfigChangedTip.visibility = View.GONE
                }
            }
            configBinding.acsChunkSize.max = RAGConfig.DEFAULT_MAX_CHUNK_SIZE
            configBinding.acsChunkSize.min = RAGConfig.DEFAULT_MIN_CHUNK_SIZE
            configBinding.acsChunkSize.progress = RAGConfig.CHUNK_SIZE
            configBinding.etChunkSize.setText("${RAGConfig.CHUNK_SIZE}")
            configBinding.etNChunk.setText("${RAGConfig.N_CHUNKS}")
            configBinding.acsChunkSize.setOnSeekBarChangeListener(object :
                SeekBar.OnSeekBarChangeListener {
                override fun onProgressChanged(
                    seekBar: SeekBar?,
                    progress: Int,
                    fromUser: Boolean
                ) {
                    if (fromUser) {
                        configBinding.etChunkSize.setText("$progress")
                    }
                    newChunkSize = progress
                    checkChangedConfig()
                }

                override fun onStartTrackingTouch(seekBar: SeekBar?) {
                }

                override fun onStopTrackingTouch(seekBar: SeekBar?) {
                }

            })
            configBinding.etChunkSize.addTextChangedListener(object : TextWatcher {
                override fun afterTextChanged(s: Editable?) {
                    s?.let {
                        val temp = s.toString()
                        if (TextUtils.isEmpty(temp)) {
                            newChunkSize = RAGConfig.DEFAULT_MIN_CHUNK_SIZE
                        } else {
                            val inputValue = temp.toIntOrNull() ?: RAGConfig.DEFAULT_MIN_CHUNK_SIZE
                            newChunkSize = inputValue.coerceIn(RAGConfig.DEFAULT_MIN_CHUNK_SIZE, RAGConfig.DEFAULT_MAX_CHUNK_SIZE)
                            if (configBinding.acsChunkSize.progress != newChunkSize) {
                                configBinding.acsChunkSize.progress = newChunkSize
                            }
                        }
                        checkChangedConfig()
                    }
                }

                override fun beforeTextChanged(
                    s: CharSequence?,
                    start: Int,
                    count: Int,
                    after: Int
                ) {
                }

                override fun onTextChanged(
                    s: CharSequence?,
                    start: Int,
                    before: Int,
                    count: Int
                ) {
                }

            })
            configBinding.etChunkSize.setOnFocusChangeListener { _, hasFocus ->
                if (!hasFocus) {
                    // When losing focus, correct the displayed value to the valid range
                    val currentText = configBinding.etChunkSize.text.toString()
                    val inputValue = currentText.toIntOrNull() ?: RAGConfig.DEFAULT_MIN_CHUNK_SIZE
                    val correctedValue = inputValue.coerceIn(RAGConfig.DEFAULT_MIN_CHUNK_SIZE, RAGConfig.DEFAULT_MAX_CHUNK_SIZE)
                    if (inputValue != correctedValue || currentText.isEmpty()) {
                        configBinding.etChunkSize.setText("$correctedValue")
                    }
                }
            }
            configBinding.etNChunk.addTextChangedListener(object : TextWatcher {
                override fun afterTextChanged(s: Editable?) {
                    s?.let {
                        val temp = s.toString()
                        newNChunk = if (TextUtils.isEmpty(temp)) {
                            0
                        } else {
                            temp.toInt()
                        }
                        checkChangedConfig()
                    }
                }

                override fun beforeTextChanged(
                    s: CharSequence?,
                    start: Int,
                    count: Int,
                    after: Int
                ) {
                }

                override fun onTextChanged(
                    s: CharSequence?,
                    start: Int,
                    before: Int,
                    count: Int
                ) {
                }

            })

            BottomSheetDialog(this, R.style.TransparentBottomSheetDialog).apply {
                this.setContentView(configBinding.root)
                this.findViewById<View>(com.google.android.material.R.id.design_bottom_sheet)
                    ?.setBackgroundColor(Color.TRANSPARENT)
                configBinding.ivClose.setOnClickListener {
                    this.dismiss()
                }
                configBinding.btnReset.setOnClickListener {
                    configBinding.acsChunkSize.progress = chunkSize
                    configBinding.etChunkSize.setText("$chunkSize")
                    configBinding.etNChunk.setText("$nChunk")
                }
                configBinding.btnReIndex.setOnClickListener {
                    this.dismiss()
                    RAGConfig.CHUNK_SIZE = newChunkSize
                    RAGConfig.N_CHUNKS = newNChunk
                    if (embedFiles.isNotEmpty()) {
                        embedFiles(embedFiles, true)
                    }
                }
            }.show()

        }

        btnAddImage.setOnClickListener {
            openGallery()
        }

        btnAudioRecord.setOnClickListener {
            startRecord()
        }

        btnClearHistory.setOnClickListener {
            clearHistory()
        }
        /**
         * Step 3. download model
         */
        binding.btnCancelDownload.setOnClickListener {
            downloadContext?.stop()
            downloadState = DownloadState.IDLE
            tvDownloadProgress.text = "0%"
            downloadingModelData?.downloadableFiles(
                this@MainActivity,
                downloadingModelData!!.modelDir(this),
                modelList
            )
                ?.forEach {
                    it.file.delete()
                }
            binding.btnDismissDownload.performClick()
        }
        binding.btnRetryDownload.setOnClickListener {
            downloadContext?.stop()
            downloadState = DownloadState.IDLE
            downloadModel(downloadingModelData!!)
        }
        binding.btnDismissDownload.setOnClickListener {
            binding.llDownloading.visibility = View.GONE
        }
        btnDownload.setOnClickListener {
            Log.d(TAG, "downloadState:$downloadState")
            if (downloadState == DownloadState.DOWNLOADING) {
                if (downloadingModelData?.displayName == spinnerText) {
                    binding.llDownloading.visibility = View.VISIBLE
                } else {
                    Toaster.show("${downloadingModelData?.displayName} is currently downloading.")
                }
                return@setOnClickListener
            }
            val selectModelData = modelList.first { it.displayName == spinnerText }
            Log.d(TAG, "selectModelData:${selectModelData.id}")
            downloadModel(selectModelData)
        }
        /**
         * Step 4. load model
         */
        btnLoadModel.setOnClickListener {
            unloadModel(false)
            var selectModelData = modelList.first { it.displayName == spinnerText }
            if (selectModelData == null) {
                Toast.makeText(this@MainActivity, "model not selected", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }
            Log.d(TAG, "current select model data:$selectModelData")
//            if (hasLoadedModel()) {
//                Toast.makeText(this@MainActivity, "please unload first", Toast.LENGTH_SHORT).show()
//                return@setOnClickListener
//            }

            // Check if model files exist locally before attempting to load
            val fileName = isModelDownloaded(selectModelData)
            if (fileName != null) {
                Toaster.showLong("The \"$fileName\" file is missing. Please download it first.")
                return@setOnClickListener
            }

            vTip.visibility = View.VISIBLE
            llLoading.visibility = View.VISIBLE

            val supportPluginIds = selectModelData.getSupportPluginIds()
            Log.d(TAG, "support plugin_id:$supportPluginIds")
            var modelDataPluginId = "npu"
            var nGpuLayers = 0
            if (supportPluginIds.size > 1) {
                val dialogBinding = DialogSelectPluginIdBinding.inflate(layoutInflater)
                supportPluginIds.forEach {
//                    when (it) {
//                        "cpu" -> {
//                            dialogBinding.rbCpu.visibility = View.VISIBLE
//                            dialogBinding.rbCpu.isChecked = true
//                        }
//
//                        "gpu" -> {
//                            dialogBinding.rbGpu.visibility = View.VISIBLE
//                        }
//
//                        "npu" -> {
//                            dialogBinding.rbNpu.visibility = View.VISIBLE
//                            dialogBinding.rbNpu.isChecked = true
//                        }
//                    }
                }
                dialogBinding.rgSelectPluginId.setOnCheckedChangeListener { group, checkedId ->
//                    dialogBinding.llGpuLayers.visibility =
//                        if (checkedId == R.id.rb_gpu) View.VISIBLE else View.GONE
                    when (checkedId) {
                        R.id.rb_npu -> {
                            modelDataPluginId = "npu"
                        }

                        R.id.rb_gpu -> {
                            modelDataPluginId = "gpu"
                        }

                        R.id.rb_cpu -> {
                            modelDataPluginId = "cpu"
                        }

                        R.id.rb_npu_llama -> {
                            modelDataPluginId = "npu_llama"
                        }
                    }
                }

                val dialogOnClickListener = object : CustomDialogInterface.OnClickListener() {
                    override fun onClick(
                        dialog: DialogInterface?,
                        which: Int
                    ) {
                        nGpuLayers = 0
                        if (dialogBinding.llGpuLayers.visibility == View.VISIBLE) {
                            nGpuLayers = dialogBinding.etGpuLayers.text.toString().toInt()
                            if (nGpuLayers == 0) {
                                Toast.makeText(
                                    this@MainActivity,
                                    "nGpuLayers min value is 1",
                                    Toast.LENGTH_SHORT
                                ).show()
                                return
                            }
                        }
                        when (which) {
                            DialogInterface.BUTTON_POSITIVE -> {
                                dialog?.dismiss()
                                loadModel(selectModelData, modelDataPluginId, nGpuLayers)
                            }

                            DialogInterface.BUTTON_NEGATIVE -> {
                                llLoading.visibility = View.INVISIBLE
                                vTip.visibility = View.GONE
                            }
                        }
                    }

                }
                val alertDialog = AlertDialog.Builder(this).setView(dialogBinding.root)
//                    .setNegativeButton("cancel", dialogOnClickListener)
//                    .setPositiveButton("sure", dialogOnClickListener)
                    .setCancelable(false)
                    .create()
                dialogBinding.ivClose.setOnClickListener {
                    llLoading.visibility = View.INVISIBLE
                    vTip.visibility = View.GONE
                    alertDialog.dismiss()
                }
                dialogBinding.btnLoad.setOnClickListener {
                    alertDialog.dismiss()
                    if (modelDataPluginId != "npu") {
                        selectModelId = selectModelData.dependencies!!.first()
                        selectModelData = modelList.first { it.id == selectModelId }
                    }
                    loadModel(selectModelData, modelDataPluginId, 0)
                }
                alertDialog.show()
                dialogOnClickListener.resetPositiveButton(alertDialog)
            } else {
                if ("npu" == supportPluginIds[0]) {
                    modelDataPluginId = "npu"
                }
                loadModel(selectModelData, modelDataPluginId, nGpuLayers)
            }
        }

        /**
         * Step 5. send message
         */
        btnSend.setOnClickListener {
            if (!hasLoadedModel()) {
                Toast.makeText(this@MainActivity, "please load model first", Toast.LENGTH_SHORT)
                    .show()
                return@setOnClickListener
            }

            if (savedImageFiles.isNotEmpty()) {
                messages.add(Message("", MessageType.IMAGES, savedImageFiles.map { it }))
                reloadRecycleView()
            }

            val inputString = etInput.text.trim().toString()
            etInput.setText("")
            etInput.clearFocus()
            val imm = getSystemService(Context.INPUT_METHOD_SERVICE) as InputMethodManager
            imm.hideSoftInputFromWindow(etInput.windowToken, 0)

            if (inputString.isNotEmpty()) {
                messages.add(Message(inputString, MessageType.USER))
                reloadRecycleView()
            }

            val supportFunctionCall = false
            var tools: String? = null
            var grammarString: String? = null
            if (supportFunctionCall) {
                // if this model support 'function call'
                tools =
                    "[{\"type\":\"function\",\"function\":{\"name\": \"campaign_investigation\",\"description\": \"Check campaign limits and determine appropriate action. If customer has reached limit, return a message (hardcoded or generated by model). If limit not reached, contact support.\",\"parameters\": {\"type\": \"object\", \"properties\":{\"campaign_name\":{\"type\": \"string\",\"description\": \"The name of the campaign to investigate\"}}, \"required\":[\"campaign_name\"]}}}]"
                grammarString = """
root ::= "<tool_call>" space object "</tool_call>" space
object ::= "{" space campaign-name-kv "}" space
campaign-name-kv ::= "\"campaign_name\"" space ":" space string
string ::= "\"" char* "\"" space
char ::= [^"\\\x7F\x00-\x1F] | [\\] (["\\bfnrt] | "u" hex hex hex hex)
hex ::= [0-9a-fA-F]
space ::= | " " | "\n" | "\r" | "\t"
"""
            }

            if (!hasLoadedModel()) {
                Toast.makeText(this@MainActivity, "model not loaded", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }

            retrievedChunksList.clear()
            binding.btnStop.isEnabled = true
            modelScope.launch {
                val selectModelData = modelList.first { it.id == selectModelId }
                val isNpu = selectModelData.getNexaManifest(this@MainActivity)?.PluginId == "npu"
                Log.d(TAG, "isNpu: $isNpu")

                val sb = StringBuilder()
                if (isLoadCVModel) {
                    // FIXME: Temporarily select the last image
                    if (savedImageFiles.isEmpty()) {
                        runOnUiThread {
                            Toast.makeText(
                                this@MainActivity,
                                "Please select one picture.",
                                Toast.LENGTH_SHORT
                            ).show()
                        }
                        return@launch
                    }
                    val imagePath = savedImageFiles.last().absolutePath
                    messages.add(Message("", MessageType.IMAGES, savedImageFiles))
                    reloadRecycleView()
                    clearImages()
                    cvWrapper.infer(imagePath).onSuccess {
                        Log.d("nfl", "infer result:$it")
                        runOnUiThread {
                            val content = it.map { result ->
                                "[${result.confidence}] ${result.text}"
                            }.toList().joinToString(separator = "\n")
                            messages.add(Message(content, MessageType.ASSISTANT))
                            reloadRecycleView()
                        }
                    }.onFailure { error ->
                        runOnUiThread {
                            messages.add(Message(error.toString(), MessageType.PROFILE))
                            reloadRecycleView()
                        }
                        Log.d("nfl", "infer result error:$error")
                    }
                } else if (isLoadAsrModel) {
                    if (audioFile == null) {
                        runOnUiThread {
                            Toast.makeText(this@MainActivity, "no audio file", Toast.LENGTH_SHORT)
                                .show()
                        }
                    } else {
//                        val audioFilePath = audioFile!!.absolutePath
                        val audioFilePath = "/sdcard/Download/assets/OSR_us_000_0010_16k.wav"
                        asrWrapper.transcribe(
                            AsrTranscribeInput(
                                audioFilePath,  // Use hardcoded path instead of inputString
                                "en",  // Language code
                                null   // Optional timestamps
                            )
                        ).onSuccess { transcription ->
                            runOnUiThread {
                                messages.add(
                                    Message(
                                        transcription.result.transcript ?: "",
                                        MessageType.ASSISTANT
                                    )
                                )
                                reloadRecycleView()
                            }
                        }.onFailure { error ->
                            runOnUiThread {
                                messages.add(
                                    Message(
                                        "Error: ${error.message}",
                                        MessageType.PROFILE
                                    )
                                )
                                reloadRecycleView()
                            }
                        }
                    }
                } else if (isLoadEmbedderModel) {
                    // ADD: Handle embedder inference
                    // Input format: single text or multiple texts separated by "|"
                    val texts = inputString.split("|").map { it.trim() }.toTypedArray()
                    embedderWrapper!!.embed(texts, EmbeddingConfig()).onSuccess { embeddings ->
                        runOnUiThread {
                            val result = StringBuilder()
                            val embeddingDim = embeddings.embeddings.size / texts.size

                            texts.forEachIndexed { idx, text ->
                                val start = idx * embeddingDim
                                val end = start + embeddingDim
                                val embedding = embeddings.embeddings.slice(start until end)

                                // Calculate mean and variance
                                val mean = embedding.average()
                                val variance = embedding.map { (it - mean) * (it - mean) }.average()

                                result.append("Text ${idx + 1}: \"$text\"\n")
                                result.append("Embedding dimension: $embeddingDim\n")
                                result.append("Mean: ${"%.4f".format(mean)}\n")
                                result.append("Variance: ${"%.4f".format(variance)}\n")
                                result.append("First 5 values: [")
                                result.append(
                                    embedding.take(5).joinToString(", ") { "%.4f".format(it) })
                                result.append("...]\n\n")
                            }

                            messages.add(Message(result.toString(), MessageType.ASSISTANT))
                            reloadRecycleView()
                        }
                    }.onFailure { error ->
                        runOnUiThread {
                            messages.add(Message("Error: ${error.message}", MessageType.PROFILE))
                            reloadRecycleView()
                        }
                    }
                } else if (isLoadRerankerModel) {
                    // Reranker input format: "query\ndoc1\ndoc2\ndoc3..."
                    // First line is query, remaining lines are documents
                    val query = inputString.split("\n")[0]  // Get first line as query
                    val documents =
                        inputString.split("\n").drop(1).toTypedArray()  // Get rest as docs
                    rerankerWrapper.rerank(query, documents, RerankConfig())
                        .onSuccess { rerankerResult ->
                            runOnUiThread {
                                val result = StringBuilder()
                                result.append("Rerank Results:\n")
                                // Sort by score descending to show best matches first
                                rerankerResult.scores?.withIndex()?.sortedByDescending { it.value }
                                    ?.forEach { (idx, score) ->
                                        result.append("${idx + 1}. Score: ${"%.4f".format(score)}\n")
                                        result.append("   ${documents[idx]}\n\n")
                                    }
                                messages.add(Message(result.toString(), MessageType.ASSISTANT))
                                reloadRecycleView()
                            }
                        }.onFailure { error ->
                            runOnUiThread {
                                "Error: ${error.message}".also {
                                    messages.add(Message(it, MessageType.PROFILE))
                                    reloadRecycleView()
                                }
                            }
                        }
                } else if (isLoadVlmModel) {
                    val contents = savedImageFiles.map {
                        VlmContent("image", it.absolutePath)
                    }.toMutableList()
                    audioFile?.let {
                        contents.add(VlmContent("audio", it.absolutePath))
                    }
                    contents.add(VlmContent("text", inputString))
                    audioFile = null
                    clearImages()
                    val sendMsg = VlmChatMessage(role = "user", contents = contents)
                    // VlmContentTransfer(
                    //     this@MainActivity, VlmContent(
                    //         "image", inputString
                    //     )
                    // ).forUrl()

                    // vlmChatList.clear()
                    vlmChatList.add(sendMsg)

                    Log.d(TAG, "before apply chat template:$vlmChatList")
                    vlmWrapper.applyChatTemplate(vlmChatList.toTypedArray(), tools, enableThinking)
                        .onSuccess { result ->
                            Log.d(TAG, "vlm chat template:${result.formattedText}")
                            val baseConfig =
                                GenerationConfigSample().toGenerationConfig(grammarString)
                            val configWithMedia = vlmWrapper.injectMediaPathsToConfig(
                                vlmChatList.toTypedArray(),
                                baseConfig
                            )

                            Log.d(TAG, "Config has ${configWithMedia.imageCount} images")

                            vlmWrapper.generateStreamFlow(
                                if (isNpu || true) inputString else result.formattedText,
                                configWithMedia  // Use the updated config with media paths
                            ).collect { handleResult(sb, it) }
                        }.onFailure {
                            runOnUiThread {
                                Toast.makeText(
                                    this@MainActivity, it.message, Toast.LENGTH_SHORT
                                ).show()
                            }
                        }
                } else {
                    var embedString = ""

                    if (embedResultList.isNotEmpty()) {
                        val queryEmbedding = GenerateEmbedStringsUtil.embedText(inputString)

                        if (queryEmbedding != null) {
                            // Calculate similarity scores for all chunks
                            embedResultList.forEach { embedResultBean ->
                                embedResultBean.score =
                                    GenerateEmbedStringsUtil.computeCosineSimilarity(
                                        queryEmbedding,
                                        embedResultBean.result
                                    )
                                Log.d(
                                    TAG,
                                    "embed score:${embedResultBean.score} path:${embedResultBean.path} chunk:${embedResultBean.chunkIndex}"
                                )
                            }
                        }

                        // Sort by score and get top N_CHUNKS
                        embedResultList.sortWith(Comparator { bean, bean1 -> if (bean.score > bean1.score) -1 else 1 })
                        val topChunks = embedResultList.take(RAGConfig.N_CHUNKS)

                        // Combine chunk texts (display chunk index starting from 1 for UI)
                        embedString = topChunks.joinToString("\n\n") { chunk ->
                            "Chunk ${chunk.chunkIndex + 1} from ${File(chunk.path).name}:\n${chunk.txt ?: ""}"
                        }

                        // Store all top chunks for multiple citations
                        retrievedChunksList.addAll(topChunks)
                    }

                    // Add system prompt for RAG
                    if (embedResultList.isNotEmpty() && chatList.isEmpty()) {
                        val systemPrompt =
                            """You are NexaStudio, a local AI agent built by Nexa AI.  Your job is to answer user questions in a neutral, professional, concise tone."""
                        chatList.add(ChatMessage(role = "system", systemPrompt))
                    }

                    // Format user message with RAG template
                    val userMessage =
                        if (embedResultList.isNotEmpty() && embedString.isNotEmpty()) {
                            """User query: $inputString
---

Context from files: 
$embedString
---

Provide an answer to the user query based on the context information."""
                        } else {
                            inputString
                        }

                    chatList.add(ChatMessage(role = "user", userMessage))
                    // Apply chat template and generate
                    llmWrapper.applyChatTemplate(
                        chatList.toTypedArray(),
                        tools,
                        enableThinking
                    ).onSuccess { templateOutput ->
                        Log.d(TAG, "chat template:${templateOutput.formattedText}")
                        // Store the formatted prompt for display
                        lastFormattedPrompt = templateOutput.formattedText
                        // Fixed: Always use templateOutput.formattedText which contains RAG context
                        llmWrapper.generateStreamFlow(
                            templateOutput.formattedText,
                            GenerationConfigSample().toGenerationConfig(grammarString)
                        ).collect { streamResult ->
                            handleResult(sb, streamResult)
                        }
                    }.onFailure { error ->
                        runOnUiThread {
                            Toast.makeText(
                                this@MainActivity, error.message, Toast.LENGTH_SHORT
                            ).show()
                        }
                    }
                }

                clearImages()

                runOnUiThread {
                    binding.btnStop.isEnabled = false
                }
            }

        }

        /**
         * Step 6. others
         */
        btnUnloadModel.setOnClickListener {
            binding.flIndex.visibility = View.GONE
            if (!hasLoadedModel()) {
                Toast.makeText(this@MainActivity, "model not loaded", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }
            // Unload model and cleanup
            unloadModel()
        }
        btnStop.setOnClickListener {
            if (!hasLoadedModel()) {
                Toast.makeText(
                    this@MainActivity,
                    "model not loaded",
                    Toast.LENGTH_SHORT
                ).show()
                return@setOnClickListener
            }
            // MODIFY: Stop button only works for LLM/VLM (not embedder/reranker)
            if (isLoadEmbedderModel || isLoadRerankerModel || isLoadAsrModel || isLoadCVModel) {
                Toast.makeText(
                    this@MainActivity,
                    "Stop not applicable for embedder/reranker/asr/cv",
                    Toast.LENGTH_SHORT
                ).show()
                return@setOnClickListener
            }
            // Stop streaming
            modelScope.launch {
                if (isLoadVlmModel) {
                    vlmWrapper.stopStream()
                } else if (isLoadLlmModel) {
                    llmWrapper.stopStream()
                }
            }
        }
        binding.ivTopk.setOnClickListener {
            val topKBinding = DialogTopkConfigBinding.inflate(layoutInflater)
            topKBinding.acsTopk.progress = IndexFragment.embedTopK
            topKBinding.etTopk.setText("${IndexFragment.embedTopK}")
            topKBinding.acsTopk.setOnSeekBarChangeListener(object :
                SeekBar.OnSeekBarChangeListener {
                override fun onProgressChanged(
                    seekBar: SeekBar?,
                    progress: Int,
                    fromUser: Boolean
                ) {
                    IndexFragment.embedTopK = progress
                    topKBinding.etTopk.setText("${IndexFragment.embedTopK}")
                }

                override fun onStartTrackingTouch(seekBar: SeekBar?) {
                }

                override fun onStopTrackingTouch(seekBar: SeekBar?) {
                }

            })
            BottomSheetDialog(this, R.style.TransparentBottomSheetDialog).apply {
                this.setContentView(topKBinding.root)
                this.findViewById<View>(com.google.android.material.R.id.design_bottom_sheet)
                    ?.setBackgroundColor(Color.TRANSPARENT)
                topKBinding.ivClose.setOnClickListener {
                    this.dismiss()
                }
            }.show()
        }
    }

    fun handleResult(sb: StringBuilder, streamResult: LlmStreamResult) {
        when (streamResult) {
            is LlmStreamResult.Token -> {
                runOnUiThread {
                    sb.append(streamResult.text)
                    Message(sb.toString(), MessageType.ASSISTANT).let { lastMsg ->
                        val size = messages.size
                        messages[size - 1].let { msg ->
                            if (msg.type != MessageType.ASSISTANT) {
                                messages.add(lastMsg)
                            } else {
                                messages[size - 1] = lastMsg
                            }
                        }
                    }
                    adapter.notifyDataSetChanged()
                }
                Log.d(TAG, "Token: ${streamResult.text}")
            }

            is LlmStreamResult.Completed -> {
                if (isLoadVlmModel) {
                    vlmChatList.add(
                        VlmChatMessage(
                            "assistant",
                            listOf(VlmContent("text", sb.toString()))
                        )
                    )
                } else {
                    chatList.add(ChatMessage("assistant", sb.toString()))
                }

                runOnUiThread {
                    var content = sb.toString()
                    val size = messages.size
                    messages[size - 1] = Message(content, MessageType.ASSISTANT)

                    val ttft = String.format(null, "%.2f", streamResult.profile.ttftMs)
                    val promptTokens = streamResult.profile.promptTokens
                    val prefillSpeed =
                        String.format(null, "%.2f", streamResult.profile.prefillSpeed)

                    val generatedTokens = streamResult.profile.generatedTokens
                    val decodingSpeed =
                        String.format(null, "%.2f", streamResult.profile.decodingSpeed)

                    val profileData =
                        "TTFT: $ttft ms; Prompt Tokens: $promptTokens; \nPrefilling Speed: $prefillSpeed tok/s\nGenerated Tokens: $generatedTokens; Decoding Speed: $decodingSpeed tok/s"
                    messages.add(
                        Message(
                            profileData,
                            MessageType.PROFILE,
                            retrievedChunks = retrievedChunksList.toList(),  // Pass all retrieved chunks for multiple citations
                            formattedPrompt = lastFormattedPrompt  // Pass the formatted prompt
                        )
                    )
                    reloadRecycleView()
                }
                Log.d(TAG, "Completed: ${streamResult.profile}")
            }

            is LlmStreamResult.Error -> {
                runOnUiThread {
                    val content =
                        "your conversation is out of model’s context length, please start a new conversation or click clear button"
                    messages.add(Message(content, MessageType.PROFILE))
                    reloadRecycleView()
                }
                Log.d(TAG, "Error: $streamResult")
            }
        }
    }

    private fun okdownload() {
        val okDownloadBuilder = OkDownload.Builder(this)
        val factory = DownloadOkHttp3Connection.Factory()
        factory.setBuilder(getUnsafeOkHttpClient())
        okDownloadBuilder.connectionFactory(factory)
        try {
            OkDownload.setSingletonInstance(okDownloadBuilder.build())
        } catch (e: java.lang.Exception) {
            Log.e("download", "download init failed")
        }
    }

    private fun getUnsafeOkHttpClient(): OkHttpClient.Builder {
        try {
            val x509m: X509TrustManager = object : X509TrustManager {
                override fun getAcceptedIssuers(): Array<X509Certificate?>? {
                    //Note: Cannot return null here, otherwise it will throw an error
                    val x509Certificates = arrayOfNulls<X509Certificate>(0)
                    return x509Certificates
                }

                @Throws(CertificateException::class)
                override fun checkServerTrusted(
                    chain: Array<X509Certificate?>?, authType: String?
                ) {
// Do not throw exception to trust all server certificates
                }

                @Throws(CertificateException::class)
                override fun checkClientTrusted(
                    chain: Array<X509Certificate?>?, authType: String?
                ) {
// Default trust mechanism
                }
            }
            // Create a TrustManager that trusts all certificates
            val trustAllCerts = arrayOf<TrustManager>(x509m)

            // Initialize SSLContext
            val sslContext = SSLContext.getInstance("SSL")
            sslContext.init(null, trustAllCerts, SecureRandom())

            // Create SSLSocketFactory
            val sslSocketFactory: SSLSocketFactory = sslContext.getSocketFactory()

            // Build OkHttpClient
            return OkHttpClient.Builder().sslSocketFactory(
                sslSocketFactory, (trustAllCerts[0] as X509TrustManager?)!!
            ).hostnameVerifier { hostname: String?, session: SSLSession? -> true }
        } catch (e: Exception) {
            throw RuntimeException(e)
        }
    }

    private fun openGallery() {
        val intent = Intent(Intent.ACTION_PICK, null)
        intent.setDataAndType(MediaStore.Images.Media.EXTERNAL_CONTENT_URI, "image/*")
        startActivityForResult(intent, 1)
    }

    override fun onRequestPermissionsResult(
        requestCode: Int,
        permissions: Array<out String>,
        grantResults: IntArray
    ) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults)

        if (requestCode == 0) {
            if (grantResults.isNotEmpty() && grantResults[0] == PackageManager.PERMISSION_GRANTED) {
                openGallery()
            } else {
                Toast.makeText(this, "Not allow", Toast.LENGTH_SHORT).show()
            }
        } else if (requestCode == 2001) {
            if (grantResults.isNotEmpty() && grantResults[0] == PackageManager.PERMISSION_GRANTED) {
                openCamera()
            } else {
                Toast.makeText(this, "Camera not allow", Toast.LENGTH_SHORT).show()
            }
        }
    }

    override fun onActivityResult(requestCode: Int, resultCode: Int, data: Intent?) {
        super.onActivityResult(requestCode, resultCode, data)

        var bitmap: Bitmap? = null
        if (requestCode == 1) {
            if (resultCode == Activity.RESULT_OK && data != null) {
                val inputStream = contentResolver.openInputStream(data.data!!)
                bitmap = BitmapFactory.decodeStream(inputStream)
            }
        } else if (requestCode == 1001 && resultCode == Activity.RESULT_OK) {
            photoFile?.let {
                bitmap = BitmapFactory.decodeFile(it.absolutePath)
            }
        }

        bitmap?.let {
            try {
                val file = File(filesDir, "chat_${System.currentTimeMillis()}.jpg")
                val success = saveBitmapToFile(it, file)
                if (success) {
                    Log.d(TAG, "Save success：${file.absolutePath}")
                    savedImageFiles.add(file)
                    refreshTopScrollContainer()
                } else {
                    Toast.makeText(this, "Save Image failed", Toast.LENGTH_SHORT).show()
                }
            } catch (e: FileNotFoundException) {
                e.printStackTrace()
            }
        }
    }

    private fun saveBitmapToFile(bitmap: Bitmap, file: File): Boolean {
        return try {
            val tempDir = File(this.filesDir, "tmp").apply { if (!exists()) mkdirs() }

            val tempFile = File(
                tempDir,
                "tmp_${System.currentTimeMillis()}.jpg"
            )
            FileOutputStream(tempFile).use { out ->
                bitmap.compress(Bitmap.CompressFormat.JPEG, 100, out)
            }

            val outFile = File(
                tempDir,
                "out_${System.currentTimeMillis()}.jpg"
            )
            ImgUtil.squareCrop(
                ImgUtil.downscaleAndSave(
                    imageFile = tempFile,
                    outFile = outFile,
                    maxSize = 448,
                    format = Bitmap.CompressFormat.JPEG,
                    quality = 90
                ), file, 448
            )
            true
        } catch (e: Exception) {
            e.printStackTrace()
            false
        }
    }

    private fun stopRecord(cancel: Boolean) {
        wavRecorder?.stopRecording()
        wavRecorder = null
        bottomPanel.visibility = View.GONE
        if (cancel) {
            audioFile = null
        }
        refreshTopScrollContainer()
    }

    private fun startRecord() {
        bottomPanel.visibility = View.VISIBLE

        val file = File(filesDir, "audio")
        if (!file.exists()) {
            file.mkdirs()
        }
        audioFile =
            File(file, "audio_${System.currentTimeMillis()}.wav")
        Log.d(TAG, "audioFile: ${audioFile!!.absolutePath}")
        wavRecorder = WavRecorder()

        wavRecorder?.startRecording(audioFile!!)
    }

    private fun clearHistory() {
        if (isLoadLlmModel) {
            chatList.clear()
            modelScope.launch {
                llmWrapper.reset()
            }
        }
        if (isLoadVlmModel) {
            vlmChatList.clear()
            modelScope.launch {
                vlmWrapper.reset()
            }
        }
        messages.clear()
        audioFile = null
        clearImages()
        reloadRecycleView()
    }

    private var popupWindow: PopupWindow? = null
    private fun showPopupMenu(anchorView: View) {
        if (popupWindow?.isShowing == true) {
            popupWindow?.dismiss()
            return
        }

        val popupView = LayoutInflater.from(this).inflate(R.layout.menu_layout, null)

        popupWindow = PopupWindow(
            popupView,
            anchorView.width * 2,
            android.view.ViewGroup.LayoutParams.WRAP_CONTENT,
            true
        )

        popupWindow?.isOutsideTouchable = true
        popupWindow?.elevation = 10f

        val btnCamera = popupView.findViewById<Button>(R.id.btn_camera)
        val btnPhoto = popupView.findViewById<Button>(R.id.btn_photo)

        btnCamera.setOnClickListener {
            popupWindow?.dismiss()
            checkAndOpenCamera()
        }
        btnPhoto.setOnClickListener {
            popupWindow?.dismiss()
            openGallery()
        }

        popupView.measure(
            View.MeasureSpec.UNSPECIFIED,
            View.MeasureSpec.UNSPECIFIED
        )
        val popupHeight = popupView.measuredHeight
        popupWindow?.showAsDropDown(anchorView, 0, -anchorView.height - popupHeight)
    }

    private var photoUri: Uri? = null
    private var photoFile: File? = null

    private fun checkAndOpenCamera() {
        if (ContextCompat.checkSelfPermission(this, Manifest.permission.CAMERA)
            != PackageManager.PERMISSION_GRANTED
        ) {
            ActivityCompat.requestPermissions(
                this,
                arrayOf(Manifest.permission.CAMERA),
                2001
            )
        } else {
            openCamera()
        }
    }

    private fun openCamera() {
        val intent = Intent(MediaStore.ACTION_IMAGE_CAPTURE)
        photoFile = File(
            getExternalFilesDir(Environment.DIRECTORY_PICTURES),
            "photo_${System.currentTimeMillis()}.jpg"
        )
        photoUri = FileProvider.getUriForFile(
            this,
            "${applicationContext.packageName}.fileprovider",
            photoFile!!
        )

        intent.putExtra(MediaStore.EXTRA_OUTPUT, photoUri)
        intent.addFlags(Intent.FLAG_GRANT_WRITE_URI_PERMISSION)
        startActivityForResult(intent, 1001)
    }

    private fun clearImages() {
        savedImageFiles.clear()
        refreshTopScrollContainer()
    }

    private fun refreshTopScrollContainer() {
        runOnUiThread {
            topScrollContainer.removeAllViews()
            if (savedImageFiles.isEmpty() && audioFile == null) {
                scrollImages.visibility = View.GONE
                return@runOnUiThread
            }

            scrollImages.visibility = View.VISIBLE

            for (file in savedImageFiles) {
                val itemView = LayoutInflater.from(this)
                    .inflate(R.layout.item_image_scroll, topScrollContainer, false)
                val ivImage = itemView.findViewById<ImageView>(R.id.iv_image)
                val btnRemove = itemView.findViewById<ImageButton>(R.id.btn_remove)

                ivImage.setImageURI(Uri.fromFile(file))

                btnRemove.setOnClickListener {
                    savedImageFiles.remove(file)
                    refreshTopScrollContainer()
                }
                topScrollContainer.addView(itemView)
            }

            if (audioFile != null) {
                val audioView = LayoutInflater.from(this)
                    .inflate(R.layout.item_audio_scroll, topScrollContainer, false)
                val audioName = audioView.findViewById<TextView>(R.id.tv_audio_name)
                val audioType = audioView.findViewById<TextView>(R.id.tv_audio_type)
                val btnRemove = audioView.findViewById<ImageButton>(R.id.btn_audio_remove)
                audioName.text = audioFile!!.name
                // TODO: hard code
                audioType.text = "wav"

                btnRemove.setOnClickListener {
                    audioFile = null
                    refreshTopScrollContainer()
                }
                topScrollContainer.addView(audioView)
            }
        }
    }

    private fun reloadRecycleView() {
        adapter.notifyDataSetChanged()
        binding.rvChat.scrollToPosition(messages.size - 1)
    }

    companion object {
        private const val TAG = "MainActivity"
    }

    private enum class OperationState {
        DEFAULT, DOWNLOADED, LOADED
    }

    private fun changeOperationUI(state: OperationState) {
        when (state) {
            OperationState.DOWNLOADED -> {
                binding.spModelList.isEnabled = true
                binding.ivDeviceTag.visibility = View.INVISIBLE

                binding.btnDownload.visibility = View.GONE
                binding.btnLoadModel.visibility = View.VISIBLE
                binding.btnLoadModel.text = "Load"
                binding.btnUnloadModel.visibility = View.VISIBLE
                binding.btnUnloadModel.isEnabled = false
                binding.btnStop.visibility = View.VISIBLE
                binding.btnStop.isEnabled = false

                binding.btnTestEmbed.isEnabled = false
            }

            OperationState.LOADED -> {
                binding.spModelList.isEnabled = false
                binding.ivDeviceTag.visibility = View.VISIBLE

                binding.btnDownload.visibility = View.GONE
                binding.btnLoadModel.visibility = View.VISIBLE
                binding.btnLoadModel.text = pluginType.uppercase()
                binding.btnUnloadModel.visibility = View.VISIBLE
                binding.btnUnloadModel.isEnabled = true
                binding.btnStop.visibility = View.VISIBLE
                binding.btnStop.isEnabled = false

                binding.btnTestEmbed.isEnabled = true
            }

            else -> {
                binding.spModelList.isEnabled = true
                binding.ivDeviceTag.visibility = View.INVISIBLE

                binding.btnDownload.visibility = View.VISIBLE
                binding.btnLoadModel.visibility = View.GONE
                binding.btnLoadModel.text = "Load"
                binding.btnUnloadModel.visibility = View.GONE
                binding.btnStop.visibility = View.GONE

                binding.btnTestEmbed.isEnabled = false
            }
        }
    }

}
