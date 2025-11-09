package com.nexa.demo

import android.Manifest
import android.app.Activity
import android.content.Context
import android.content.Intent
import android.content.SharedPreferences
import android.content.pm.PackageManager
import android.graphics.Bitmap
import android.graphics.BitmapFactory
import android.net.Uri
import android.os.Bundle
import android.os.Environment
import android.provider.MediaStore
import android.system.Os
import android.util.Log
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
import android.widget.SimpleAdapter
import android.widget.Spinner
import android.widget.TextView
import android.widget.Toast
import androidx.core.app.ActivityCompat
import androidx.core.content.ContextCompat
import androidx.core.content.FileProvider
import androidx.core.content.edit
import androidx.lifecycle.viewmodel.viewModelFactory
import androidx.recyclerview.widget.LinearLayoutManager
import androidx.recyclerview.widget.RecyclerView
import com.google.gson.Gson
import com.google.gson.reflect.TypeToken
import com.liulishuo.okdownload.DownloadContext
import com.liulishuo.okdownload.DownloadTask
import com.liulishuo.okdownload.OkDownload
import com.liulishuo.okdownload.core.cause.EndCause
import com.liulishuo.okdownload.core.connection.DownloadOkHttp3Connection
import com.liulishuo.okdownload.kotlin.listener.createDownloadContextListener
import com.liulishuo.okdownload.kotlin.listener.createListener1
import com.nexa.demo.bean.DownloadableFile
import com.nexa.demo.bean.ModelData
import com.nexa.demo.bean.downloadableFiles
import com.nexa.demo.bean.mmprojTokenFile
import com.nexa.demo.bean.modelDir
import com.nexa.demo.bean.modelFile
import com.nexa.demo.bean.tokenFile
import com.nexa.demo.utils.ExecShell
import com.nexa.demo.utils.ImgUtil
import com.nexa.demo.utils.SharePreferenceKeys
import com.nexa.demo.utils.WavRecorder
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
import com.nexa.sdk.bean.CVResult
import com.nexa.sdk.bean.ChatMessage
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

class MainActivity : Activity() {

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
    private lateinit var embedderWrapper: EmbedderWrapper
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

    private val savedImageFiles = mutableListOf<File>()
    private val messages = arrayListOf<Message>()


    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)
        requestPermissions(arrayOf(Manifest.permission.RECORD_AUDIO), 1002)
        okdownload()
        initData()
        initView()
        setListeners()
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
        recyclerView = findViewById<RecyclerView>(R.id.rv_chat)
        adapter = ChatAdapter(messages)
        recyclerView.adapter = adapter
        recyclerView.layoutManager = LinearLayoutManager(this)

        llDownloading = findViewById(R.id.ll_downloading)
        tvDownloadProgress = findViewById(R.id.tv_download_progress)
        pbDownloading = findViewById(R.id.pb_downloading)
        spModelList = findViewById(R.id.sp_model_list)
        spModelList.adapter = object : SimpleAdapter(this, modelList.map {
            val map = mutableMapOf<String, String>()
            map["id"] = it.id
            map
        }, R.layout.item_model, arrayOf("id"), intArrayOf(R.id.tv_model_id)) {

        }
        spModelList.onItemSelectedListener = object : AdapterView.OnItemSelectedListener {
            override fun onItemSelected(
                parent: AdapterView<*>?, view: View?, position: Int, id: Long
            ) {
                selectModelId = modelList[position].id

                messages.clear()
                adapter.notifyDataSetChanged()
                recyclerView.scrollTo(0, 0)
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

    private fun parseModelList() {
        try {
            val baseJson = assets.open("model_list.json").bufferedReader().use { it.readText() }
            modelList = Json.decodeFromString<List<ModelData>>(baseJson)
        } catch (e: Exception) {
            Log.e("nfl", "parseModelList: $e")
        }
    }

    /**
     * Step 0. Preparing to download the model file.
     */
    private fun initData() {
        spDownloaded = getSharedPreferences(SP_DOWNLOADED, MODE_PRIVATE)
//        spDownloaded.edit().putBoolean("Qwen3-0.6B-Q8_0", false).commit()
//        spDownloaded.edit().putBoolean("Qwen3-0.6B-IQ4_NL", false).commit()
//        spDownloaded.edit().putBoolean("LFM2-1.2B-npu", false).commit()
//        spDownloaded.edit().putBoolean("embeddinggemma-300m-npu", false).commit()
//        spDownloaded.edit().putBoolean("jina-v2-rerank-npu", false).commit()
//        spDownloaded.edit().putBoolean("paddleocr-npu", false).commit()
//        spDownloaded.edit().putBoolean("parakeet-tdt-0.6b-v3-npu", false).commit()
//        spDownloaded.edit().putBoolean("OmniNeural-4B", false).commit()
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
        NexaSdk.getInstance().init(this)

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
                btnStop.visibility = View.INVISIBLE
            } else {
                btnStop.visibility = View.VISIBLE
            }
        }
    }

    private fun onLoadModelFailed(tip: String) {
        runOnUiThread {
            vTip.visibility = View.GONE

            if (!spDownloaded.getBoolean(selectModelId, false)) {
                Toast.makeText(this@MainActivity, "please download model first", Toast.LENGTH_SHORT)
                    .show()
            } else {
                Toast.makeText(this@MainActivity, tip, Toast.LENGTH_SHORT)
                    .show()
            }


            // change UI
            btnAddImage.visibility = View.INVISIBLE
            btnAudioRecord.visibility = View.INVISIBLE
            btnUnloadModel.visibility = View.INVISIBLE
            llLoading.visibility = View.INVISIBLE
        }
    }

    private fun hasLoadedModel(): Boolean {
        return isLoadLlmModel || isLoadVlmModel || isLoadEmbedderModel ||
                isLoadRerankerModel || isLoadCVModel || isLoadAsrModel
    }

    private fun setListeners() {

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
        btnDownload.setOnClickListener {
            if (spDownloaded.getBoolean(selectModelId, false) || hasLoadedModel()) {
                Toast.makeText(this@MainActivity, "model already downloaded", Toast.LENGTH_SHORT)
                    .show()
            } else {
                llDownloading.visibility = View.VISIBLE
                tvDownloadProgress.text = "0%"
                modelScope.launch {
                    val selectModelData = modelList.first { it.id == selectModelId }

                    val filesToDownload =
                        selectModelData.downloadableFiles(selectModelData.modelDir(this@MainActivity))
                    Log.d(TAG, "filesToDownload: $filesToDownload")
                    if (filesToDownload.isEmpty()) throw IllegalArgumentException("No download URL")

                    fun getUrlFileSize(client: OkHttpClient, url: String): Long {
                        // Extract hostname for logging
                        val hostname = try {
                            url.substringAfter("://").substringBefore("/")
                        } catch (e: Exception) {
                            "unknown"
                        }

                        Log.d(TAG, "Requesting: $hostname")

                        val builder = Request.Builder().url(url).head()
                        getHfToken(selectModelData, url)?.let {
                            builder.addHeader("Authorization", "Bearer $it")
                        }
                        val request = builder.build()
                        try {
                            client.newCall(request).execute().use { resp ->
                                val size = resp.header("Content-Length")?.toLongOrNull() ?: 0L
                                Log.d(TAG, "Response: code=${resp.code}, size=$size")
                                return size
                            }
                        } catch (e: java.net.UnknownHostException) {
                            Log.e(TAG, "DNS resolution failed for $hostname - Check DNS/network")
                            return 0L
                        } catch (e: java.net.SocketTimeoutException) {
                            Log.e(
                                TAG,
                                "Connection timeout to $hostname - Possible firewall/proxy issue"
                            )
                            return 0L
                        } catch (e: java.net.ConnectException) {
                            Log.e(TAG, "Connection refused by $hostname - Server unreachable")
                            return 0L
                        } catch (e: javax.net.ssl.SSLException) {
                            Log.e(TAG, "SSL/TLS error to $hostname - ${e.message}")
                            return 0L
                        } catch (e: Exception) {
                            Log.e(TAG, "Network error: ${e.javaClass.simpleName} - ${e.message}")
                            return 0L
                        }
                    }

                    val unsafeClient = getUnsafeOkHttpClient().build()

                    val totalSizes: List<Long> = filesToDownload.map { (file, url) ->
                        async {
                            getUrlFileSize(unsafeClient, url)
                        }
                    }.awaitAll()
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
                            if (100 == percent) {
                                llDownloading.visibility = View.GONE
                                spDownloaded.edit().putBoolean(selectModelId, true).commit()
                            } else {
                                tvDownloadProgress.text = "$percent%"
                            }
                        }
                    }

                    fun reportProgress(force: Boolean = false) {
                        val now = System.currentTimeMillis()
                        if (force || now - lastProgressTime > progressInterval) {
                            val elapsedMs = now - startTime
                            val downloaded = alreadyDownloaded.values.sum()
                            val percent =
                                if (totalBytes > 0) ((downloaded * 100) / totalBytes).toInt() else 0
                            val speedAvg =
                                if (elapsedMs > 0) downloaded / (elapsedMs / 1000.0) else 0.0
                            val etaSec =
                                if (speedAvg > 0) ((totalBytes - downloaded) / speedAvg).toLong() else -1L
                            val speedStr = if (speedAvg > 1024 * 1024) {
                                String.format("%.2f MB/s", speedAvg / (1024 * 1024))
                            } else {
                                String.format("%.1f KB/s", speedAvg / 1024)
                            }
                            onProgress(
                                selectModelId, percent, downloaded, totalBytes, etaSec, speedStr
                            )
                            lastProgressTime = now
                        }
                    }
                    ////////////////////////////////////////////////
                    var downloadContext: DownloadContext? = null
                    // val startTime = SystemClock.uptimeMillis()
                    val queueSet = DownloadContext.QueueSet()
                        .setParentPathFile(filesToDownload[0].file.parentFile)
                        .setMinIntervalMillisCallbackProcess(300)
                    val builder = queueSet.commit()
                    filesToDownload.withIndex().forEach { (i, item) ->
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
                    var currentCount = 0

                    downloadContext = builder.setListener(createDownloadContextListener {}).build()
                    downloadContext?.start(
                        createListener1(taskStart = { task, _ ->
                            Log.d(TAG, "download task ${task.id} Start")
                        }, retry = { task, _ ->
                            Log.d(TAG, "download task ${task.id} retry")
                        }, connected = { task, _, _, _ ->
                            Log.d(TAG, "download task ${task.id} connected")
                        }, progress = { task, currentOffset, totalLength ->
                            Log.d(
                                TAG,
                                "download task ${task.id} progress $currentOffset $totalLength"
                            )
                            alreadyDownloaded[task.url] = currentOffset
                            reportProgress(true)
                        }) { task, cause, exception, listener1Model ->
                            // Listener1Assist.Listener1Model
                            if (cause != EndCause.COMPLETED) {
                                Log.e(TAG, "download task ${task.id} error")
                                return@createListener1
                            }
                            // FIXME: Total download progress should be recalculated here, but skipping due to minor difference before completion
                            Log.d(TAG, "download task ${task.id} end")
                            // Download finished
                            currentCount += 1
                            if (currentCount == totalCount) {
                                // Show download complete message
                                reportProgress(force = true)
                                onProgress(selectModelId, 100, totalBytes, totalBytes, 0, "0 KB/s")
                            }
                        }, true
                    )
                }
            }
        }
        /**
         * Step 4. load model
         */
        btnLoadModel.setOnClickListener {
            val selectModelData = modelList.first { it.id == selectModelId }
            if (selectModelData == null) {
                Toast.makeText(this@MainActivity, "model not selected", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }
            if (hasLoadedModel()){
                Toast.makeText(this@MainActivity, "please unload first", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }
            vTip.visibility = View.VISIBLE
            llLoading.visibility = View.VISIBLE
            modelScope.launch {
                resetLoadState()
                when (selectModelData.type) {
                    "chat" -> {
                        // LFM2-1.2B-npu
                        val isNPU = selectModelData.id == "LFM2-1.2B-npu"
                        val conf = ModelConfig(
                            nCtx = 1024,
                            max_tokens = 2048,
                            nThreads = 8,
                            nThreadsBatch = 4,
                            nBatch = 2048,
                            nUBatch = 512,
                            nSeqMax = 1,
                            enable_thinking = enableThinking,
                            npu_lib_folder_path = applicationInfo.nativeLibraryDir,
                            npu_model_folder_path = selectModelData.modelDir(this@MainActivity).absolutePath
                        )
                        // Build and initialize LlmWrapper for chat model
                        LlmWrapper.builder().llmCreateInput(
                            LlmCreateInput(
                                model_name = if (isNPU) "liquid-v2" else "",
                                model_path = selectModelData.modelFile(this@MainActivity)!!.absolutePath,
                                tokenizer_path = selectModelData.tokenFile(this@MainActivity)?.absolutePath,
                                config = conf,
                                plugin_id = if (isNPU) "npu" else "cpu_gpu"
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
                            model_name = "embed-gemma",  // Model name for NPU plugin
                            model_path = selectModelData.modelFile(this@MainActivity)!!.absolutePath,
                            tokenizer_path = selectModelData.tokenFile(this@MainActivity)?.absolutePath,
                            config = ModelConfig(
                                npu_lib_folder_path = applicationInfo.nativeLibraryDir,
                                npu_model_folder_path = selectModelData.modelDir(this@MainActivity).absolutePath
                            ),
                            plugin_id = "npu",
                            device_id = null
                        )

                        EmbedderWrapper.builder()
                            .embedderCreateInput(embedderCreateInput)
                            .build().onSuccess { wrapper ->
                                isLoadEmbedderModel = true
                                embedderWrapper = wrapper
                                onLoadModelSuccess("embedder model loaded")
                            }.onFailure { error ->
                                onLoadModelFailed(error.message.toString())
                            }

                    }

                    "reranker" -> {
                        // Handle reranker model loading with NPU paths using RerankerCreateInput
                        // jina-v2-rerank-npu
                        val rerankerCreateInput = RerankerCreateInput(
                            model_name = "jina-rerank",  // Model name for NPU plugin
                            model_path = selectModelData.modelFile(this@MainActivity)!!.absolutePath,
                            tokenizer_path = selectModelData.tokenFile(this@MainActivity)?.absolutePath,
                            config = ModelConfig(
                                npu_lib_folder_path = applicationInfo.nativeLibraryDir,
                                npu_model_folder_path = selectModelData.modelDir(this@MainActivity).absolutePath
                            ),
                            plugin_id = "npu",
                            device_id = null
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
                            model_name = "paddleocr",
                            config = CVModelConfig(
                                capabilities = CVCapability.OCR,
                                det_model_path = selectModelData.modelDir(this@MainActivity).absolutePath,
                                rec_model_path = selectModelData.modelFile(this@MainActivity)!!.absolutePath,
                                char_dict_path = selectModelData.modelDir(this@MainActivity).absolutePath,
                                npu_model_folder_path = selectModelData.modelDir(this@MainActivity).absolutePath,
                                npu_lib_folder_path = applicationInfo.nativeLibraryDir
                            ),
                            plugin_id = "npu"
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
                            model_name = "parakeet",
                            model_path = selectModelData.modelFile(this@MainActivity)!!.absolutePath,
                            config = ModelConfig(
                                npu_lib_folder_path = applicationInfo.nativeLibraryDir,
                                npu_model_folder_path = selectModelData.modelDir(this@MainActivity).absolutePath
                            ),
                            plugin_id = "npu"
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

                    "multimodal" -> {
                        // VLM model
                        val isNpuVlm = selectModelData.id == "OmniNeural-4B"
                        val config = if (isNpuVlm) {
                            ModelConfig(
                                nCtx = 2048,
                                max_tokens = 2048,
                                nThreads = 8,
                                nThreadsBatch = 4,
                                nBatch = 2048,
                                nUBatch = 512,
                                nSeqMax = 1,
                                enable_thinking = enableThinking,
                                npu_lib_folder_path = applicationInfo.nativeLibraryDir,
                                npu_model_folder_path = selectModelData.modelDir(this@MainActivity).absolutePath
                            )
                        } else {
                            ModelConfig(
                                nCtx = 1024,
                                max_tokens = 2048,
                                nThreads = 4,
                                nThreadsBatch = 4,
                                nBatch = 1,
                                nUBatch = 1,
                                nSeqMax = 1,
                                enable_thinking = enableThinking
                            )
                        }

                        val vlmCreateInput = VlmCreateInput(
                            model_name = if (isNpuVlm) "omni-neural" else "",
                            model_path = selectModelData.modelFile(this@MainActivity)!!.absolutePath,
                            mmproj_path = selectModelData.mmprojTokenFile(this@MainActivity)?.absolutePath,
                            config = config,
                            plugin_id = if (isNpuVlm) "npu" else "cpu_gpu"
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

        /**
         * Step 5. send message
         */
        btnSend.setOnClickListener() {
            if(!hasLoadedModel()) {
                Toast.makeText(this@MainActivity, "please load model first", Toast.LENGTH_SHORT).show()
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

            modelScope.launch {
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
                        val audioFilePath = audioFile!!.absolutePath
                        asrWrapper.transcribe(
                            AsrTranscribeInput(
                                audioFilePath,  // Use hardcoded path instead of inputString
                                "en",  // Language code
                                null   // Optional timestamps
                            )
                        ).onSuccess { transcription ->
                            runOnUiThread {
                                messages.add(Message(transcription.result.transcript ?: "", MessageType.ASSISTANT))
                                reloadRecycleView()
                            }
                        }.onFailure { error ->
                            runOnUiThread {
                                messages.add(Message("Error: ${error.message}", MessageType.PROFILE))
                                reloadRecycleView()
                            }
                        }
                    }

                } else if (isLoadEmbedderModel) {
                    // ADD: Handle embedder inference
                    // Input format: single text or multiple texts separated by "|"
                    val texts = inputString.split("|").map { it.trim() }.toTypedArray()
                    embedderWrapper.embed(texts, EmbeddingConfig()).onSuccess { embeddings ->
                        runOnUiThread {
                            val result = StringBuilder()
                            val embeddingDim = embeddings.size / texts.size
                            
                            texts.forEachIndexed { idx, text ->
                                val start = idx * embeddingDim
                                val end = start + embeddingDim
                                val embedding = embeddings.slice(start until end)
                                
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
                    val selectModelData = modelList.first { it.id == selectModelId }
                    val isNpuVlm = selectModelData.id == "OmniNeural-4B"
                    Log.d(TAG, "isNpuVlm: $isNpuVlm")

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
                            val baseConfig =
                                GenerationConfigSample().toGenerationConfig(grammarString)
                            val configWithMedia = vlmWrapper.injectMediaPathsToConfig(
                                vlmChatList.toTypedArray(),
                                baseConfig
                            )

                            Log.d(TAG, "Config has ${configWithMedia.imageCount} images")

                            vlmWrapper.generateStreamFlow(
                                result.formattedText,
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
                    
                    chatList.add(ChatMessage(role = "user", inputString))
                    // Apply chat template and generate
                    llmWrapper.applyChatTemplate(
                        chatList.toTypedArray(),
                        tools,
                        enableThinking
                    ).onSuccess { templateOutput ->
                        Log.d(TAG, "chat template:${templateOutput.formattedText}")
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
            }
        }
        /**
         * Step 6. others
         */
        btnUnloadModel.setOnClickListener {
            if (!hasLoadedModel()) {
                Toast.makeText(this@MainActivity, "model not loaded", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }
            // Unload model and cleanup
            val handleUnloadResult = fun(result: Int) {
                resetLoadState()
                runOnUiThread {
                    vTip.visibility = View.GONE
                    btnUnloadModel.visibility = View.INVISIBLE
                    btnStop.visibility = View.INVISIBLE
                    btnAddImage.visibility = View.INVISIBLE
                    btnAudioRecord.visibility = View.INVISIBLE
                    Toast.makeText(
                        this@MainActivity, if (result == 0) {
                            "unload success"
                        } else {
                            "unload failed and error code: $result"
                        }, Toast.LENGTH_SHORT
                    ).show()
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
                    embedderWrapper.destroy()
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
                    val decodeSpeed = String.format(null, "%.2f", streamResult.profile.decodingSpeed)
                    val profileData = "TTFT: $ttft ms; Decode Speed: $decodeSpeed t/s"
                    messages.add(Message(profileData, MessageType.PROFILE))
                    reloadRecycleView()
                }
                Log.d(TAG, "Completed: ${streamResult.profile}")
            }

            is LlmStreamResult.Error -> {
                runOnUiThread {
                    val content =
                        streamResult.throwable.message + streamResult.throwable.cause?.message
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

    private  fun clearHistory() {
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
        recyclerView.scrollToPosition(messages.size - 1)
    }
    companion object {
        private const val SP_DOWNLOADED = "sp_downloaded"
        private const val TAG = "MainActivity"
    }
}
