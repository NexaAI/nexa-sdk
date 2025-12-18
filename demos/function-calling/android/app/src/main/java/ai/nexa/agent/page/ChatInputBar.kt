package ai.nexa.agent.page

import ai.nexa.agent.R
import ai.nexa.agent.model.ChatViewModel
import ai.nexa.agent.state.LocalBannerState
import ai.nexa.agent.state.OperationState
import ai.nexa.agent.ui.theme.chatInputBackground
import ai.nexa.agent.ui.theme.chatInputBorder
import ai.nexa.agent.ui.theme.editMessageBgColor
import ai.nexa.agent.ui.theme.editMessageCancelColor
import ai.nexa.agent.ui.theme.editMessagePenColor
import ai.nexa.agent.ui.theme.editMessageTextColor
import ai.nexa.agent.ui.theme.handleColor
import ai.nexa.agent.ui.theme.ic_config
import ai.nexa.agent.ui.theme.ic_plus
import ai.nexa.agent.ui.theme.ic_recorder
import ai.nexa.agent.ui.theme.ic_right_btn
import ai.nexa.agent.ui.theme.ic_send
import ai.nexa.agent.ui.theme.ic_stop
import ai.nexa.agent.ui.theme.placeholder
import ai.nexa.agent.ui.theme.textPrimary
import android.Manifest
import android.media.AudioFormat
import android.media.AudioRecord
import android.media.MediaRecorder
import android.text.TextUtils
import android.view.MotionEvent
import androidx.annotation.RequiresPermission
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.defaultMinSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.ime
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.selection.LocalTextSelectionColors
import androidx.compose.foundation.text.selection.TextSelectionColors
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TextField
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableLongStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.input.pointer.pointerInteropFilter
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.platform.LocalFocusManager
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.google.accompanist.permissions.ExperimentalPermissionsApi
import com.google.accompanist.permissions.isGranted
import com.google.accompanist.permissions.rememberPermissionState
import com.nexa.studio.ui.chat.menu.MenuAction
import com.nexa.studio.ui.chat.menu.MessageMenuPopup
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import java.io.File
import kotlin.math.sqrt
import kotlin.random.Random

@OptIn(ExperimentalPermissionsApi::class)
@Composable
fun ChatInputBar(
    modifier: Modifier = Modifier,
    viewModel: ChatViewModel,
    input: String = "",
    onInputChange: (String) -> Unit = {},
    onSend: (String) -> Unit = {},
    onPlusClick: () -> Unit = {},
    onConfigClick: () -> Unit = {},
    onRecorderClick: () -> Unit = {},
    onMicClick: () -> Unit = {},
    onWaveClick: () -> Unit = {},
    onFileClick: () -> Unit = {},
    onAudioClick: () -> Unit = {},
    onCameraClick: () -> Unit = {},
    onPhotoClick: () -> Unit = {},
    selectedImages: List<File> = emptyList(),
    selectedAudios: List<File> = emptyList(),
    onRemoveImage: (File) -> Unit = {},
    onRemoveAudio: (File) -> Unit = {},
    onStopClick: () -> Unit = {},
    onCloseEdit: () -> Unit = {},
    editContent: String,
) {
    val state by viewModel.state.collectAsState()
    var showUploadMenu by remember { mutableStateOf(false) }
    val imeVisible = WindowInsets.ime.getBottom(LocalDensity.current) > 0
    val bottomPadding = 18.dp
    val focusRequester = remember { FocusRequester() }
    val focusManager = LocalFocusManager.current

    val recordPermissionState = rememberPermissionState(Manifest.permission.RECORD_AUDIO)
    var isRecording by remember { mutableStateOf(false) }
    var audioVolumes by remember { mutableStateOf<List<Float>>(emptyList()) }
    var speechText by remember { mutableStateOf("") }
    var recordingJob by remember { mutableStateOf<Job?>(null) }
    val barCount = 100
    var fakeVolumes by remember { mutableStateOf(List(barCount) { 0f }) }
    var recordingTime by remember { mutableLongStateOf(0L) }
//    var remainingTime by remember { mutableIntStateOf(MAX_DURATION_MINUTE) }
    val context = LocalContext.current
    val banner = LocalBannerState.current

//    val speechRecognizerHelper = remember {
//        SpeechRecognizerHelper(
//            context = context,
//            onResult = { speechText = it },
//            onPartial = {
//                speechText = it
//            },
//            onError = { }
//        )
//    }
    var lastSpeechText by remember { mutableStateOf("") }

    val useWaveRecorder = true

//    LaunchedEffect(isRecording, speechText) {
//        if(!useWaveRecorder) {
//            while (isRecording) {
//                val isNewText = speechText != lastSpeechText
//                lastSpeechText = speechText
//
//                val newValue = if (isNewText && speechText.isNotBlank()) {
//                    0.005f + Random.nextFloat() * (0.1f - 0.005f)
//                } else {
//                    0f
//                }
//                fakeVolumes = fakeVolumes.drop(1) + listOf(newValue.coerceIn(0f, 1f))
//                delay(35)
//            }
//            fakeVolumes = List(barCount) { 0f }
//        }
//    }

//    fun startAudio() {
//        if (recordPermissionState.status.isGranted) {
//            isRecording = true
//            recordingTime = 0L
//            remainingTime = MAX_DURATION_MINUTE
//            if (useWaveRecorder) {
//                wavRecorder = WavRecorder(
//                    onVolumeDbChangeListener = object : WavRecorder.OnVolumeDbChangeListener{
//                        override fun onVolumeDbChange(volumeDb: Double) {
//                            val volumeDbPercent = ((volumeDb - 20) / 20f).toFloat().coerceIn(0f, 1f)
//                            // L.d("nfl", "volumeDb: $volumeDb, volumeDbPercent: $volumeDbPercent")
//                            fakeVolumes = fakeVolumes.drop(1) + listOf(volumeDbPercent)
//                        }
//                    },
//                    onMaxDurationReached = {
//                        // 直接在这里实现停止录制的逻辑
//                        isRecording = false
//                        fakeVolumes = List(barCount) { 0f }
//                        recordingTime = 0L
//                        remainingTime = MAX_DURATION_MINUTE
//                        wavRecorder?.stopRecording()
//                        if (audioFile != null) {
//                            viewModel.onIntent(ChatIntent.AddAudio(listOf(audioFile!!)))
//                        }
//                        wavRecorder = null
//                        audioFile = null
//                        recordingJob?.cancel()
//                    }
//                )
//                audioFile =
//                    File(
//                        FileConfig.chatParentFile(context),
//                        "audio_${System.currentTimeMillis()}.wav"
//                    )
//                L.d("nfl", "audioFile: ${audioFile!!.absolutePath}")
//                wavRecorder!!.startRecording(audioFile!!)
//
//                // 启动计时器
//                recordingJob = CoroutineScope(Dispatchers.Main).launch {
//                    val startTime = System.currentTimeMillis()
//                    while (isRecording) {
//                        recordingTime = System.currentTimeMillis() - startTime
//                        remainingTime = (MAX_DURATION_MINUTE - recordingTime / 1000).toInt().coerceAtLeast(0)
//                        delay(100) // 每100ms更新一次
//                    }
//                }
//            } else {
//                try {
//                    audioVolumes = emptyList()
//                    speechText = ""
//                    speechRecognizerHelper.startListening()
//                } catch (e: SecurityException) {
//                    e.printStackTrace()
//                    isRecording = false
//                    banner.show("Need Audio Record Permission")
//                }
//            }
//        } else {
//            recordPermissionState.launchPermissionRequest()
//        }
//    }
//
//    fun stopAudio(cancel: Boolean? = true) {
//        isRecording = false
//        fakeVolumes = List(barCount) { 0f }
//        recordingTime = 0L
//        remainingTime = MAX_DURATION_MINUTE
//        if (useWaveRecorder) {
//            wavRecorder?.stopRecording()
//            if (cancel != true) {
//                audioFile?.let {
//                    L.d("nfl", "add audio file")
//                    viewModel.onIntent(ChatIntent.AddAudio(listOf(it)))
//                }
//            }
//            wavRecorder = null
//            audioFile = null
//        } else {
//            speechRecognizerHelper.stopListening()
//        }
//        recordingJob?.cancel()
//    }

    Surface(
        color = MaterialTheme.colorScheme.chatInputBackground,
        shape = RoundedCornerShape(20.dp),
        modifier = modifier
//            .pointerInteropFilter {
//                state.currentModel == null
//            }
            .fillMaxWidth()
            .padding(12.dp)
//            .imePadding()
//            .then(
//                if (!imeVisible) Modifier.padding(bottom = bottomPadding) else Modifier
//            )
            .border(
                width = 1.dp,
                color = MaterialTheme.colorScheme.chatInputBorder,
                shape = RoundedCornerShape(20.dp)
            )
            .defaultMinSize(minHeight = 115.dp)
    ) {

        Column(
            modifier = Modifier
                .fillMaxWidth()
                .alpha(
                    if (false) {
//                        focusManager.clearFocus()
                        0.5f
                    } else {
                        1.0f
                    }
                )
                .padding(horizontal = 12.dp, vertical = 12.dp),
            verticalArrangement = Arrangement.Bottom
        ) {
            if (editContent.isNotBlank()) {
//                EditMessageBar(onClose = { onCloseEdit() })
            }
            FilePreviewBar(
                images = selectedImages,
                onRemoveImage = onRemoveImage,
                audios = selectedAudios,
                onRemoveAudio = onRemoveAudio,
            )

            if (isRecording) {
                // FIXME: 先用固定的图片展示
//                Icon(
//                    painterResource(id = MaterialTheme.colorScheme.ic_audio_long_wave),
//                    tint = if (LocalAppDarkTheme.current) {
//                        Color.White
//                    } else {
//                        Color.Black
//                    },
//                    contentDescription = ""
//                )
//                AudioInputBar(
//                    volumes = fakeVolumes,
//                )
            } else {

                AutoGrowChatInput(
                    placeholder = if (false) {
                        "Model not loaded, Please initialize the model"
                    } else {
                        "Type prompt..."
                    },
                    value = input,
                    onValueChange = onInputChange,
                    modifier = Modifier.focusRequester(focusRequester)
                )
            }
            Spacer(modifier = Modifier.height(20.dp))

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Row {
                    if (true) {
                        Box(
                            modifier = Modifier
                                .size(32.dp)
                                .clip(CircleShape) // 可选：保证触摸区域是圆形的
                                .clickable(
                                    indication = null, // 去掉点击涟漪/黑色背景
                                    interactionSource = remember { MutableInteractionSource() }
                                ) {
                                },
                            contentAlignment = Alignment.Center
                        ) {
                            IconButton(
                                onClick = {
                                    onPhotoClick()
//                                    showUploadMenu = !showUploadMenu
                                },
                                modifier = Modifier
                                    .size(32.dp)
                                    .background(
                                        if (showUploadMenu) Color(0xFFE0E0E0) else Color.Transparent
                                    )
                            ) {
                                Icon(
                                    painter = painterResource(id = MaterialTheme.colorScheme.ic_plus),
                                    contentDescription = "plus",
                                    tint = Color.Unspecified,
                                    modifier = Modifier
                                        .size(32.dp)
                                        .background(
                                            if (showUploadMenu) Color(0xFFE0E0E0) else Color.Transparent, // 选中时灰色
                                            shape = CircleShape // 保持圆形背景
                                        )
                                )
                            }
                            if (showUploadMenu) {
                                MessageMenuPopup(
                                    show = true,
                                    onDismiss = { showUploadMenu = false },
                                    offset = Offset(
                                        -34f,
                                        with(LocalDensity.current) { (-115 + 38).dp.toPx() }),
                                    menuActions = listOf(
//                                        MenuAction(
//                                            icon = painterResource(id = R.drawable.ic_audio),
//                                            text = "Audio",
//                                            onClick = { onAudioClick(); showUploadMenu = false }
//                                        ),
                                        MenuAction(
                                            icon = painterResource(id = R.drawable.ic_camera),
                                            text = "Camera",
                                            onClick = { onCameraClick(); showUploadMenu = false }
                                        ),
                                        MenuAction(
                                            icon = painterResource(id = R.drawable.ic_image),
                                            text = "Photos",
                                            onClick = { onPhotoClick(); showUploadMenu = false }
                                        )
                                    )
                                )
                            }
                        }
                        Spacer(modifier = Modifier.width(5.dp))
                    }
//                    IconButton(
//                        onClick = onConfigClick,
//                        modifier = Modifier.size(32.dp)
//                    ) {
//                        Icon(
//                            painter = painterResource(id = MaterialTheme.colorScheme.ic_config),
//                            contentDescription = "Config",
//                            tint = Color.Unspecified,
//                            modifier = Modifier.size(32.dp)
//                        )
//                    }
                }

                Box {
                    Row {
                        if (false) {
                            Box(
                                modifier = Modifier
                                    .size(32.dp)
                                    .clip(CircleShape) // 可选：保证触摸区域是圆形的
                                    .clickable(
                                        indication = null, // 去掉点击涟漪/黑色背景
                                        interactionSource = remember { MutableInteractionSource() }
                                    ) {
                                    },
                                contentAlignment = Alignment.Center
                            ) {
                                IconButton(
                                    onClick = onRecorderClick,
                                    modifier = Modifier
                                        .size(32.dp)
                                ) {
                                    Icon(
                                        painter = painterResource(id = MaterialTheme.colorScheme.ic_recorder),
                                        contentDescription = "plus",
                                        tint = Color.Unspecified,
                                        modifier = Modifier
                                            .size(32.dp)
                                    )
                                }
                            }
                            Spacer(modifier = Modifier.width(5.dp))
                        }

                        if (false) {
                            Spacer(modifier = Modifier.width(5.dp))
                            if (false) {
                                Box(
                                    modifier = Modifier
                                        .size(32.dp)
                                        .clip(CircleShape)
                                        .clickable(enabled = true, onClick = {})
                                        .pointerInteropFilter {
                                            when (it.action) {
                                                MotionEvent.ACTION_DOWN -> {

                                                }

                                                MotionEvent.ACTION_UP -> {

                                                }
                                            }
                                            true
                                        },
                                    contentAlignment = Alignment.Center
                                ) {

                                }
                            }
                            IconButton(
                                onClick = {
//                                    startAudio()
                                },
                                modifier = Modifier
                                    .size(32.dp)
                            ) {
//                                Icon(
//                                    painter = painterResource(id = MaterialTheme.colorScheme.ic_mic),
//                                    contentDescription = "audio",
//                                    tint = Color.Unspecified,
//                                    modifier = Modifier
//                                        .size(32.dp)
//                                )
                            }
                            Spacer(modifier = Modifier.width(5.dp))
                        }

                        if (state.operationState == OperationState.WAITING_SEND_RESPONSE) {
                            IconButton(
                                onClick = { onStopClick() },
                                modifier = Modifier.size(32.dp)
                            ) {
                                Icon(
                                    painter = painterResource(id = MaterialTheme.colorScheme.ic_stop),
                                    contentDescription = "Send",
                                    tint = Color.Unspecified,
                                    modifier = Modifier.size(32.dp)
                                )
                            }
                        } else {
                            if (!TextUtils.isEmpty(input)) {
                                IconButton(
                                    onClick = {
                                        onSend(input)
                                        focusManager.clearFocus()
                                    },
                                    modifier = Modifier.size(32.dp)
                                ) {
                                    Icon(
                                        painter = painterResource(id = MaterialTheme.colorScheme.ic_send),
                                        contentDescription = "Send",
                                        tint = Color.Unspecified,
                                        modifier = Modifier.size(32.dp)
                                    )
                                }
                            }
                        }
                    }

                    if (isRecording) {
                        Row(
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            // 显示倒计时
//                            Text(
//                                text = "${remainingTime}s",
//                                color = if (remainingTime <= 5) Color.Red else MaterialTheme.colorScheme.textPrimary,
//                                style = MaterialTheme.typography.bodyMedium,
//                                modifier = Modifier.padding(horizontal = 8.dp)
//                            )
//
//                            IconButton(
//                                onClick = {
//                                    stopAudio()
//                                },
//                                modifier = Modifier.size(32.dp)
//                            ) {
//                                Icon(
//                                    painter = painterResource(id = MaterialTheme.colorScheme.ic_cancel_btn),
//                                    contentDescription = "",
//                                    tint = Color.Unspecified,
//                                    modifier = Modifier.size(32.dp)
//                                )
//                            }
//                            Spacer(modifier = Modifier.width(10.dp))
//                            IconButton(
//                                onClick = {
//                                    if (MAX_DURATION_MINUTE - remainingTime < 2) {
//                                        banner.show("Hold to record for at least 2 seconds", alignBottom = true)
//                                    } else {
//                                        stopAudio(false)
//                                        onInputChange(speechText)
//                                    }
//                                },
//                                modifier = Modifier.size(32.dp)
//                            ) {
//                                Icon(
//                                    painter = painterResource(id = MaterialTheme.colorScheme.ic_right_btn),
//                                    contentDescription = "",
//                                    tint = Color.Unspecified,
//                                    modifier = Modifier.size(32.dp)
//                                )
//                            }
                        }
                    }
                }
            }
        }
    }
}

//var wavRecorder: WavRecorder? = null
//var audioFile: File? = null

@Composable
fun EditMessageBar(
    modifier: Modifier = Modifier,
    onClose: () -> Unit
) {
    Box(
        modifier = modifier
            .fillMaxWidth()
            .background(
                MaterialTheme.colorScheme.editMessageBgColor,
                shape = RoundedCornerShape(topStart = 14.dp, topEnd = 14.dp)
            )
            .padding(horizontal = 6.dp, vertical = 3.dp)
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Icon(
                painter = painterResource(id = R.drawable.ic_pen),
                contentDescription = null,
                tint = MaterialTheme.colorScheme.editMessagePenColor,
                modifier = Modifier.size(16.dp),
            )
            Spacer(modifier = Modifier.width(5.dp))
            Text(
                text = "Edit Message",
                color = MaterialTheme.colorScheme.editMessageTextColor,
                modifier = Modifier,
                fontSize = 14.sp
            )
            IconButton(modifier = Modifier.size(30.dp), onClick = onClose) {
                Icon(
                    modifier = Modifier.size(16.dp),
                    painter = painterResource(id = R.drawable.ic_close),
                    contentDescription = "Close",
                    tint = MaterialTheme.colorScheme.editMessageCancelColor,
                )
            }
        }
    }
}

@Composable
fun AutoGrowChatInput(
    value: String,
    onValueChange: (String) -> Unit,
    placeholder: String = "Type prompt...",
    minLines: Int = 1,
    maxLines: Int = 5,
    modifier: Modifier
) {
    val selectionColors = TextSelectionColors(
        handleColor = MaterialTheme.colorScheme.handleColor,            // 圆球/句柄颜色
        backgroundColor = MaterialTheme.colorScheme.handleColor // 选择高亮色(可按需调)
    )

    CompositionLocalProvider(LocalTextSelectionColors provides selectionColors) {
        BasicTextField(
            value = value,
            onValueChange = onValueChange,
            modifier = modifier
                .fillMaxWidth()
                .heightIn(min = (minLines * 30).dp, max = (maxLines * 30).dp)
                .padding(vertical = 4.dp, horizontal = 8.dp),
            singleLine = false,
            maxLines = maxLines,
            textStyle = TextStyle(
                fontSize = 18.sp,
                color = MaterialTheme.colorScheme.textPrimary
            ),
            cursorBrush = SolidColor(MaterialTheme.colorScheme.handleColor),
            decorationBox = { innerTextField ->
                Box(
                    modifier = Modifier.fillMaxWidth()
                ) {
                    if (value.isEmpty()) {
                        Text(placeholder, color = MaterialTheme.colorScheme.placeholder)
                    }
                    innerTextField()
                }
            }
        )
    }
}

//@RequiresPermission(Manifest.permission.RECORD_AUDIO)
//fun startRecordingWithVolume(
//    onVolumeUpdate: (Float) -> Unit,
//): Job {
//    val sampleRate = 16000
//    val minBufferSize = AudioRecord.getMinBufferSize(
//        sampleRate,
//        AudioFormat.CHANNEL_IN_MONO,
//        AudioFormat.ENCODING_PCM_16BIT
//    )
//
//    val audioRecord = AudioRecord(
//        MediaRecorder.AudioSource.MIC,
//        sampleRate,
//        AudioFormat.CHANNEL_IN_MONO,
//        AudioFormat.ENCODING_PCM_16BIT,
//        minBufferSize
//    )
//
//    audioRecord.startRecording()
//    val job = CoroutineScope(Dispatchers.IO).launch {
//        val buffer = ShortArray(minBufferSize)
//        while (isActive) {
//            val read = audioRecord.read(buffer, 0, buffer.size)
//            if (read > 0) {
//                val rms = sqrt(buffer.take(read).map { it * it.toFloat() }.average()).toFloat()
//                withContext(Dispatchers.Main) {
//                    onVolumeUpdate(rms / 32768f)
//                }
//            }
//        }
//        audioRecord.stop()
//        audioRecord.release()
//    }
//    return job
//}