package com.nexa.demo

import android.content.SharedPreferences
import android.os.Bundle
import android.system.Os
import android.util.Log
import android.view.View
import android.widget.AdapterView
import android.widget.Button
import android.widget.EditText
import android.widget.LinearLayout
import android.widget.ProgressBar
import android.widget.SimpleAdapter
import android.widget.Spinner
import android.widget.TextView
import android.widget.Toast
import androidx.activity.ComponentActivity
import com.liulishuo.okdownload.DownloadContext
import com.liulishuo.okdownload.DownloadTask
import com.liulishuo.okdownload.OkDownload
import com.liulishuo.okdownload.core.cause.EndCause
import com.liulishuo.okdownload.core.connection.DownloadOkHttp3Connection
import com.liulishuo.okdownload.kotlin.listener.createDownloadContextListener
import com.liulishuo.okdownload.kotlin.listener.createListener1
import com.nexa.demo.bean.ModelData
import com.nexa.demo.bean.downloadableFiles
import com.nexa.demo.bean.mmprojTokenFile
import com.nexa.demo.bean.modelDir
import com.nexa.demo.bean.modelFile
import com.nexa.demo.bean.tokenFile
import com.nexa.sdk.LlmWrapper
import com.nexa.sdk.VlmWrapper
import com.nexa.sdk.bean.ChatMessage
import com.nexa.sdk.bean.LlmCreateInput
import com.nexa.sdk.bean.LlmStreamResult
import com.nexa.sdk.bean.ModelConfig
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
import java.security.SecureRandom
import java.security.cert.CertificateException
import java.security.cert.X509Certificate
import javax.net.ssl.SSLContext
import javax.net.ssl.SSLSession
import javax.net.ssl.SSLSocketFactory
import javax.net.ssl.TrustManager
import javax.net.ssl.X509TrustManager

class MainActivity : ComponentActivity() {

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
    private lateinit var tvResult: TextView
    private lateinit var tvProfileData: TextView

    private lateinit var llmWrapper: LlmWrapper
    private lateinit var vlmWrapper: VlmWrapper
    private val modelScope = CoroutineScope(Dispatchers.IO)

    private val chatList = arrayListOf<ChatMessage>()
    private lateinit var llmSystemPrompt: ChatMessage
    private val vlmChatList = arrayListOf<VlmChatMessage>()
    private lateinit var vlmSystemPrompty: VlmChatMessage
    private lateinit var downloadDir: File
    private lateinit var modelList: List<ModelData>
    private var selectModelId = ""
    private val pluginId = "llama_cpp"
    private var isLoadVlmModel = false

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)
        okdownload()
        initData()
        if (!downloadDir.exists()) {
            downloadDir.mkdirs()
        }
        val file = File(downloadDir, "readme.txt")
        file.createNewFile()
        initView()
        setListeners()
    }

    private fun initView() {
        tvResult = findViewById(R.id.tv_result)
        tvProfileData = findViewById(R.id.tv_profile_data)
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
        btnSend = findViewById(R.id.btn_send)
    }

    private fun parseModelList() {
        val baseJson = assets.open("model_list.json").bufferedReader().use { it.readText() }
        modelList = Json.decodeFromString<List<ModelData>>(baseJson)
    }

    /**
     * Step 0. Preparing to download the model file.
     */
    private fun initData() {
        spDownloaded = getSharedPreferences(SP_DOWNLOADED, MODE_PRIVATE)
//        spDownloaded.edit().putBoolean("SmolVLM-256M-Instruct-f16", false).commit()
        parseModelList()
        downloadDir = modelList.first().modelDir(this)
        //
        val nativeLibPath: String = applicationContext.applicationInfo.nativeLibraryDir;
        initNexaSdk(nativeLibPath)
        //
        Log.d(TAG, "Os:" + Os.getenv("ADSP_LIBRARY_PATH"))
        val libDir = File(nativeLibPath)
        if (libDir.exists() && libDir.isDirectory) {
            val files = libDir.listFiles()
            Log.d(TAG, "files:" + files.size)
            files?.forEach { file ->
                Log.d("NativeLibs", "file name: ${file.name}, Size: ${file.length()} bytes")
            }
        } else {
            Log.d("NativeLibs", "none dir：$nativeLibPath")
        }
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
        addSystemPrompt(sysPrompt)
    }

    /**
     * Step 1. initNexaSdk environment
     */
    private fun initNexaSdk(nativeLibPath: String) {
        Os.setenv("ADSP_LIBRARY_PATH", nativeLibPath, true)
        Os.setenv("LD_LIBRARY_PATH", nativeLibPath, true)
        Os.setenv("NEXA_PLUGIN_PATH", nativeLibPath, true)
    }

    /**
     * Step 2. add system prompt, such as : output markdown style, contains emoji etc.(Options)
     */
    private fun addSystemPrompt(sysPrompt: String) {
        llmSystemPrompt = ChatMessage("system", sysPrompt)
        chatList.add(llmSystemPrompt)
        // VlmContent.type  = "text", "image", "audio
        vlmSystemPrompty =
            VlmChatMessage(role = "system", contents = listOf(VlmContent("text", sysPrompt)))
        vlmChatList.add(vlmSystemPrompty)
    }

    private fun setListeners() {
        /**
         * Step 3. download model
         */
        btnDownload.setOnClickListener {
            if (spDownloaded.getBoolean(selectModelId, false)) {
                Toast.makeText(this@MainActivity, "model already downloaded", Toast.LENGTH_SHORT)
                    .show()
            } else {
                llDownloading.visibility = View.VISIBLE
                tvDownloadProgress.text = "0%"
                modelScope.launch {
                    val selectModelData = modelList.first { it.id == selectModelId }

                    val filesToDownload = selectModelData.downloadableFiles(downloadDir)
                    if (filesToDownload.isEmpty()) throw IllegalArgumentException("No download URL")

                    fun getUrlFileSize(client: OkHttpClient, url: String): Long {
                        val builder = Request.Builder().url(url).head()
//                    val hfToken = BuildConfig.HF_TOKEN
//                    if (hfToken.isNotEmpty()) builder.addHeader("Authorization", "Bearer $hfToken")
                        val request = builder.build()
                        try {
                            client.newCall(request).execute().use { resp ->
                                return resp.header("Content-Length")?.toLongOrNull() ?: 0L
                            }
                        } catch (e: Exception) {
                            Log.e(TAG, "getUrlFileSize error: $e")
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
                        val task = DownloadTask.Builder(item.url, item.file).build()
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
                            // FIXME:这里应该再计算一次总下载进度，由于在下载完成前，稍微的偏差不影响，所以这里不计算了
                            Log.d(TAG, "download task ${task.id} end")
                            // 下载结束
                            currentCount += 1
                            if (currentCount == totalCount) {
                                // 提示下载完成
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

            modelScope.launch {
                if (selectModelData.type == "chat") {
                    LlmWrapper.builder().llmCreateInput(
                        LlmCreateInput(
                            model_path = selectModelData.modelFile(this@MainActivity)!!.absolutePath,
                            tokenizer_path = selectModelData.tokenFile(this@MainActivity)?.absolutePath,
                            config = ModelConfig(
                                nCtx = 1024,
                                max_tokens = 2048,
                                nThreads = 4,
                                nThreadsBatch = 4,
                                nBatch = 1,
                                nUBatch = 1,
                                nSeqMax = 1
                            ),
                            plugin_id = pluginId
                        )
                    ).build().onSuccess {
                        isLoadVlmModel = false
                        llmWrapper = it
                        runOnUiThread {
                            Toast.makeText(
                                this@MainActivity, "load success", Toast.LENGTH_SHORT
                            ).show()
                        }
                    }.onFailure {
                        runOnUiThread {
                            Toast.makeText(this@MainActivity, it.message, Toast.LENGTH_SHORT)
                                .show()
                        }
                    }

                } else {
                    VlmWrapper.builder().vlmCreateInput(
                        VlmCreateInput(
                            model_path = selectModelData.modelFile(this@MainActivity)!!.absolutePath,
                            tokenizer_path = null,
                            mmproj_path = selectModelData.mmprojTokenFile(this@MainActivity)?.absolutePath,
                            config = ModelConfig(
                                nCtx = 1024,
                                max_tokens = 2048,
                                nThreads = 4,
                                nThreadsBatch = 4,
                                nBatch = 1,
                                nUBatch = 1,
                                nSeqMax = 1
                            ),
                            plugin_id = pluginId
                        ).apply {
                            Log.d(TAG, "VlmCreateInput: $this")
                        }
                    ).build().onSuccess {
                        isLoadVlmModel = true
                        vlmWrapper = it
                        runOnUiThread {
                            Toast.makeText(
                                this@MainActivity, "load success", Toast.LENGTH_SHORT
                            ).show()
                        }
                    }.onFailure {
                        runOnUiThread {
                            Toast.makeText(this@MainActivity, it.message, Toast.LENGTH_SHORT)
                                .show()
                        }
                    }
                }
            }
        }
        /**
         * Step 5. send message
         */
        btnSend.setOnClickListener {
            val inputString = etInput.text.trim().toString()
            //
            etInput.setText("")
            tvResult.text = ""
            tvProfileData.text = ""
            //
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

            if ((!::llmWrapper.isInitialized && !isLoadVlmModel) || (!::vlmWrapper.isInitialized && isLoadVlmModel)) {
                Toast.makeText(this@MainActivity, "model not loaded", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }

            modelScope.launch {
                val sb = StringBuilder()
                if (isLoadVlmModel) {
                    val sendMsg = VlmChatMessage(
                        role = "user", contents = listOf(VlmContent("text", inputString))
                    )
                    vlmChatList.add(
                        sendMsg
                    )
                    vlmWrapper.applyChatTemplate(vlmChatList.toTypedArray(), tools, false)
                        .onSuccess {
                            Log.d(TAG, "chat template:${it.formattedText}")
                            vlmWrapper.generateStreamFlow(
                                it.formattedText,
                                GenerationConfigSample().toGenerationConfig(grammarString)
                            ).collect {
                                handleResult(sb, it)
                            }
                        }.onFailure {
                            runOnUiThread {
                                Toast.makeText(
                                    this@MainActivity, it.message, Toast.LENGTH_SHORT
                                ).show()
                            }
                        }
                } else {
                    chatList.add(ChatMessage("user", inputString))
                    llmWrapper.applyChatTemplate(chatList.toTypedArray(), tools, false).onSuccess {
                        llmWrapper.generateStreamFlow(
                            it.formattedText,
                            GenerationConfigSample().toGenerationConfig(grammarString)
                        ).collect {
                            handleResult(sb, it)
                        }
                    }.onFailure {
                        runOnUiThread {
                            Toast.makeText(
                                this@MainActivity, it.message, Toast.LENGTH_SHORT
                            ).show()
                        }
                    }
                }
            }
        }
        /**
         * Step 6. others
         */
        btnUnloadModel.setOnClickListener {
            if ((!::llmWrapper.isInitialized && !isLoadVlmModel) || (!::vlmWrapper.isInitialized && isLoadVlmModel)) {
                Toast.makeText(this@MainActivity, "model not loaded", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }
            modelScope.launch {
                if (isLoadVlmModel) {
                    vlmWrapper.stopStream()
                    vlmWrapper.destroy()
                    chatList.clear()
                    chatList.add(llmSystemPrompt)
                } else {
                    llmWrapper.stopStream()
                    llmWrapper.destroy()
                    vlmChatList.clear()
                    vlmChatList.add(vlmSystemPrompty)
                }

            }
        }
        btnStop.setOnClickListener {
            if ((!::llmWrapper.isInitialized && !isLoadVlmModel) || (!::vlmWrapper.isInitialized && isLoadVlmModel)) {
                Toast.makeText(this@MainActivity, "model not loaded", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }
            modelScope.launch {
                if (isLoadVlmModel) {
                    vlmWrapper.stopStream()
                } else {
                    llmWrapper.stopStream()
                }
            }
        }
    }

    fun handleResult(sb:StringBuilder, streamResult: LlmStreamResult) {
        when (streamResult) {
            is LlmStreamResult.Token -> {
                runOnUiThread {
                    sb.append(streamResult.text)
                    tvResult.text = sb.toString()
                }
                Log.d(TAG, "Token: ${streamResult.text}")
            }

            is LlmStreamResult.Completed -> {
                runOnUiThread {
                    tvProfileData.text = streamResult.profile.toString()
                }
                Log.d(TAG, "Completed: ${streamResult.profile}")
            }

            is LlmStreamResult.Error -> {
                runOnUiThread {
                    tvResult.text = streamResult.throwable.message
                    tvProfileData.text = streamResult.throwable.cause?.message
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
                    //注意这里不能返回null，否则会报错
                    val x509Certificates = arrayOfNulls<X509Certificate>(0)
                    return x509Certificates
                }

                @Throws(CertificateException::class)
                override fun checkServerTrusted(
                    chain: Array<X509Certificate?>?, authType: String?
                ) {
// 不抛出异常即信任所有服务器证书
                }

                @Throws(CertificateException::class)
                override fun checkClientTrusted(
                    chain: Array<X509Certificate?>?, authType: String?
                ) {
// 默认信任机制
                }
            }
            // 创建一个信任所有证书的 TrustManager
            val trustAllCerts = arrayOf<TrustManager>(x509m)

            // 初始化 SSLContext
            val sslContext = SSLContext.getInstance("SSL")
            sslContext.init(null, trustAllCerts, SecureRandom())

            // 创建 SSLSocketFactory
            val sslSocketFactory: SSLSocketFactory = sslContext.getSocketFactory()

            // 构建 OkHttpClient
            return OkHttpClient.Builder().sslSocketFactory(
                sslSocketFactory, (trustAllCerts[0] as X509TrustManager?)!!
            ).hostnameVerifier { hostname: String?, session: SSLSession? -> true }
        } catch (e: Exception) {
            throw RuntimeException(e)
        }
    }

    companion object {
        private const val SP_DOWNLOADED = "sp_downloaded"
        private const val TAG = "MainActivity"
    }
}