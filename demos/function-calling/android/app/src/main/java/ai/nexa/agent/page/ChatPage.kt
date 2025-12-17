package ai.nexa.agent.page

import ai.nexa.agent.R
import ai.nexa.agent.constant.FileConfig
import ai.nexa.agent.model.ChatViewModel
import ai.nexa.agent.navigation.RoutePaths
import ai.nexa.agent.state.ChatIntent
import ai.nexa.agent.state.LocalBannerState
import ai.nexa.agent.ui.theme.chatBg
import ai.nexa.agent.ui.theme.chatBtnBorder
import ai.nexa.agent.ui.theme.chatBtnText
import ai.nexa.agent.ui.theme.chatDivider
import ai.nexa.agent.ui.theme.chatInfoText
import ai.nexa.agent.ui.theme.chatMessageInfoCardBorder
import ai.nexa.agent.ui.theme.progressBarColor
import ai.nexa.agent.ui.theme.progressTrackColor
import ai.nexa.agent.ui.theme.settingBtnBg
import ai.nexa.agent.ui.theme.settingBtnBorder
import ai.nexa.agent.ui.theme.settingBtnText
import ai.nexa.agent.ui.theme.settingTextColor
import ai.nexa.agent.ui.theme.textPrimary
import ai.nexa.agent.util.AppUtils
import ai.nexa.agent.util.L
import ai.nexa.agent.util.fkGroteskNeueTrial
import android.Manifest
import android.app.Activity
import android.content.Context
import android.content.Intent
import android.net.Uri
import android.provider.MediaStore
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.AnimatedVisibilityScope
import androidx.compose.animation.ExperimentalSharedTransitionApi
import androidx.compose.animation.SharedTransitionScope
import androidx.compose.animation.core.Animatable
import androidx.compose.animation.core.CubicBezierEasing
import androidx.compose.animation.core.tween
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyListState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Text
import androidx.compose.material3.TextField
import androidx.compose.material3.rememberModalBottomSheetState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.MutableState
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.platform.LocalSoftwareKeyboardController
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight.Companion.W500
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.core.content.FileProvider
import androidx.navigation.NavController
import com.google.accompanist.permissions.ExperimentalPermissionsApi
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.async
import kotlinx.coroutines.awaitAll
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import org.intellij.lang.annotations.JdkConstants
import java.io.File
import kotlin.random.Random

/**
 * @param viewModel 单例
 */
@OptIn(ExperimentalPermissionsApi::class, ExperimentalSharedTransitionApi::class)
@Composable
fun ChatPage(
    navController: NavController,
    viewModel: ChatViewModel,
    openDrawer: () -> Unit = {},
    sharedTransitionScope: SharedTransitionScope,
    animatedVisibilityScope: AnimatedVisibilityScope
) {
    val context = LocalContext.current
    val chatDir = FileConfig.chatParentFile(context)
    val banner = LocalBannerState.current
    val coroutineScope = rememberCoroutineScope()

    var showServerIpAddressDialog by remember { mutableStateOf(false) }
    val shouldShowEmptyMessageView = remember { false }
    val inputBarAlpha = remember { Animatable(0f) }
    val density = LocalDensity.current
    val px_8 = with(density) { 8.dp.toPx() }
    val inputBarOffsetY = remember { Animatable(px_8) }
    var localEditContent by remember { mutableStateOf("") }
    var outputUri by remember { mutableStateOf<Uri?>(null) }
    var photoFile by remember { mutableStateOf<File?>(null) }
    val cameraChatPermissions = arrayOf(
        Manifest.permission.CAMERA,
        Manifest.permission.RECORD_AUDIO
    )

    val launcher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.StartActivityForResult(),
        onResult = { result ->
            if (result.resultCode == Activity.RESULT_OK) {
                coroutineScope.launch {
                    outputUri?.let { uri ->
                        val savedFiles = saveUrisToInternalImages(context, listOf(uri))
                    }
                }
            }
        }
    )

    fun openSysCameraView() {
        try {
            val file = File.createTempFile("photo_", ".jpg", chatDir)
            val uri = FileProvider.getUriForFile(
                context,
                "${context.packageName}.fileprovider",
                file
            )
            photoFile = file
            outputUri = uri
            val intent = Intent(MediaStore.ACTION_IMAGE_CAPTURE).apply {
                putExtra(MediaStore.EXTRA_OUTPUT, outputUri)
                addFlags(Intent.FLAG_GRANT_WRITE_URI_PERMISSION)
            }
            launcher.launch(intent)
        } catch (e: SecurityException) {
            L.e("nfl", "openSysCameraView error:${e.localizedMessage}")
//            banner.show("Need Camera Permission")
        }
    }

    val cameraPermissionLauncher =
        rememberLauncherForActivityResult(ActivityResultContracts.RequestPermission()) {
            if (it) {
                val imgPath =
                    "${FileConfig.chatParentFile(context).absolutePath}${File.separator}capture_${System.currentTimeMillis()}_temp.jpg"
                navController.navigate("${RoutePaths.CAMERA_TAKE_PICTURE}?${RoutePaths.ARGUMENT_TAKE_PICTURE_PATH}=$imgPath")
                // openSysCameraView()
            } else {
                banner.show("Need Camera Permission")
            }
        }
    val cameraChatPermissionLauncher =
        rememberLauncherForActivityResult(ActivityResultContracts.RequestMultiplePermissions()) { permissions ->
            // 检查所有权限是否已授予
            val allPermissionsGranted = permissions.all {
                it.value
            }
            if (allPermissionsGranted) {
                navController.navigate(RoutePaths.CAMERA)
            }
        }
    val chatMessageListBottomPadding = remember { mutableIntStateOf(0) }

    LaunchedEffect(Unit) {
        launch {
            inputBarOffsetY.animateTo(
                targetValue = 0f,
                animationSpec = tween(600, easing = CubicBezierEasing(0.42f, 0f, 0.58f, 1f))
            )
        }
        launch {
            inputBarAlpha.animateTo(
                targetValue = 1f,
                animationSpec = tween(600, easing = CubicBezierEasing(0.42f, 0f, 0.58f, 1f))
            )
        }
    }

    val importImagesLauncher = rememberImportImagesLauncher(context, coroutineScope, viewModel)
    val importAudiosLauncher = rememberImportAudioLauncher(context, coroutineScope)
    val inputText by remember { mutableStateOf("") }

    fun onImportImageClick() = importImagesLauncher.launch(arrayOf("image/*"))
    fun onImportAudioClick() = importAudiosLauncher.launch(arrayOf("audio/*"))

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.chatBg)
            .statusBarsPadding()
            .navigationBarsPadding()
    ) {
        // ChatBackground 在独立容器中，不受键盘影响，但位置要正确
        if (shouldShowEmptyMessageView) {
            ChatBackground(
//                state,
                onSelectModelClick = {
                    navController.navigate("download?showBack=true")
                },
                sharedTransitionScope = sharedTransitionScope,
                animatedVisibilityScope = animatedVisibilityScope,
                modifier = Modifier
                    .fillMaxSize()
                    .padding(bottom = 80.dp)
            )
        }

        // 主要内容区域
        Column(
            modifier = Modifier
                .fillMaxSize()
        ) {
            ChatHeader(
//                currentModel = state.currentModel,
                onMenuClick = {
                    openDrawer
                    showServerIpAddressDialog = true
                },
                onModelSelected = { modelId ->
//                    viewModel.onIntent(ChatIntent.SelectedModel(modelId))
                },
//                modelList = state.modelList,
                onMoreClick = { navController.navigate("download?showBack=true") },
                onNewSession = {
                    L.d("nfl", "onNewSession")
                    viewModel.onIntent(ChatIntent.NewSession)
                }
            )
//            FakeProgressBar(isLoading = state.loading)
            HorizontalDivider(
                Modifier.fillMaxWidth(),
                1.dp,
                MaterialTheme.colorScheme.chatDivider
            )
            Box(modifier = Modifier.weight(1f)) {
                if (!shouldShowEmptyMessageView) {
                    ChatMessageListSection(
//                        state,
//                        listState,
                        viewModel,
//                        menuState = menuState,
//                        setMenuState = { menuState = it },
                        chatMessageListBottomPadding
                    )
                }
            }
            ChatInputBarSection(
                navController = navController,
                viewModel = viewModel,
                onImportImageClick = ::onImportImageClick,
                onImportAudioClick = ::onImportAudioClick,
                modifier = Modifier
                    .imePadding()
                    .offset(y = inputBarOffsetY.value.dp)
                    .alpha(inputBarAlpha.value),
                editContent = localEditContent,
                onCloseEdit = {
//                    localEditContent = ""
//                    viewModel.onIntent(ChatIntent.InputChanged(""))
                },
                onRecorderClick = {
                    cameraChatPermissionLauncher.launch(cameraChatPermissions)
                },
                onOpenSystemCameraClick = {
                    cameraPermissionLauncher.launch(Manifest.permission.CAMERA)
                }
            )
        }

        if (showServerIpAddressDialog) {
            ServerIpAddressDialog(viewModel, onDismissRequest = {
                showServerIpAddressDialog = false
            })
        }
    }
//    if (menuState.show && menuState.message != null) {
//        MessageMenuPopup(
//            show = true,
//            onDismiss = { menuState = menuState.copy(show = false) },
//            offset = menuState.offset,
//            menuActions = if (menuState.type == "user") {
//                userMenuActions(menuState.message!!, viewModel) { content ->
//                    localEditContent = content
//                    viewModel.onIntent(ChatIntent.InputChanged(content))
//                }
//            } else {
//                botMenuActions(menuState.message!!, onRegenerate = { message ->
//                    viewModel.onIntent(ChatIntent.Regenerate(message))
//                }, showRegenerateIcon = state.modelType.let {
//                    it != ModelType.QNN && it != ModelType.QNN_VISION
//                })
//            }
//        )
//    }
//    // -------- Event -------------
//    HandleChatScreenEvents(viewModel, state, listState, chatMessageListBottomPadding, shouldAutoScroll)
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ServerIpAddressDialog(
    viewModel: ChatViewModel,
    onDismissRequest: () -> Unit,
) {
    val bottomSheetState = rememberModalBottomSheetState()
    val coroutineScope = rememberCoroutineScope()
    var showBottomSheet by remember { mutableStateOf(false) }

    ModalBottomSheet(
        onDismissRequest = onDismissRequest,
        modifier = Modifier.fillMaxWidth(),
        sheetState = bottomSheetState,
        containerColor = Color.White,
        contentColor = Color.White,
    ) {
        var ipAddress by remember { mutableStateOf(viewModel.state.value.serverIpAddress) }
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp)
        ) {
            Text(
                text = "Generation Settings", fontFamily = fkGroteskNeueTrial,
                fontWeight = W500,
                color = MaterialTheme.colorScheme.settingTextColor,
                style = TextStyle(
                    fontSize = 16.sp,
                )
            )
            Spacer(modifier = Modifier.height(50.dp))
            Text(
                text = "Server IP Address", fontFamily = fkGroteskNeueTrial,
                fontWeight = W500,
                color = MaterialTheme.colorScheme.settingTextColor,
                style = TextStyle(
                    fontSize = 14.sp,
                )
            )
            Spacer(modifier = Modifier.height(20.dp))
            TextField(
                value = ipAddress,
                onValueChange = { newText ->
                    ipAddress = newText
                },
                modifier = Modifier
                    .fillMaxWidth()
//                    .heightIn(min = 30.dp, max = 30.dp)
                    .clip(RoundedCornerShape(20.dp))
                    .border(
                        width = 1.dp,
                        color = MaterialTheme.colorScheme.chatMessageInfoCardBorder,
                        shape = RoundedCornerShape(20.dp)
                    ),
//                    .padding(vertical = 4.dp, horizontal = 8.dp),
                singleLine = true,
                maxLines = 1,
                textStyle = TextStyle(
                    fontSize = 14.sp,
                    color = MaterialTheme.colorScheme.textPrimary
                ),
//                cursorBrush = SolidColor(MaterialTheme.colorScheme.handleColor),
            )
            Spacer(modifier = Modifier.height(20.dp))
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.End) {
                OutlinedButton(
                    modifier = Modifier.size(height = 38.dp, width = 78.dp),
                    onClick = {
                        onDismissRequest()
                        viewModel.onIntent(ChatIntent.ChangeServerIPAddress(ipAddress))
                    },
                    border = BorderStroke(1.dp, MaterialTheme.colorScheme.settingBtnBorder),
                    shape = RoundedCornerShape(24.dp),
                    contentPadding = PaddingValues(0.dp)
                ) {
                    Text(
                        "Reset",
                        color = MaterialTheme.colorScheme.settingBtnBorder,
                        fontSize = 16.sp
                    )
                }
                Spacer(modifier = Modifier.width(10.dp))
                Button(
                    modifier = Modifier.size(height = 38.dp, width = 78.dp),
                    onClick = {
                        onDismissRequest()
                    },
                    colors = ButtonDefaults.buttonColors(containerColor = MaterialTheme.colorScheme.settingBtnBg),
                    shape = RoundedCornerShape(24.dp),
                    contentPadding = PaddingValues(0.dp)
                ) {
                    Text(
                        "Close",
                        color = MaterialTheme.colorScheme.settingBtnText,
                        fontSize = 16.sp
                    )
                }
            }
            Spacer(modifier = Modifier.height(50.dp))
        }
    }
}

@OptIn(ExperimentalSharedTransitionApi::class)
@Composable
private fun ChatBackground(
//    state: ChatUiState,
    onSelectModelClick: () -> Unit = {},
    sharedTransitionScope: SharedTransitionScope,
    animatedVisibilityScope: AnimatedVisibilityScope,
    modifier: Modifier = Modifier
) {
    //初加载避免 CenteredCircleWithLogo跳上跳下的，临时处理
    var isInitialLoad by remember { mutableStateOf(true) }

    LaunchedEffect(Unit) {
        delay(1000)
        isInitialLoad = false
    }

    val shouldShowButton = true


    Column(
        modifier = modifier
            .fillMaxSize(),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
//        CenteredCircleWithLogo(
//            sharedTransitionScope = sharedTransitionScope,
//            animatedVisibilityScope = animatedVisibilityScope
//        )
        AnimatedVisibility(shouldShowButton) {
            Column(
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                Spacer(modifier = Modifier.height(12.dp))
                OutlinedButton(
                    onClick = onSelectModelClick,
                    shape = RoundedCornerShape(24.dp),
                    border = BorderStroke(1.dp, MaterialTheme.colorScheme.chatBtnBorder),
                    modifier = Modifier
                        .height(36.dp)
                        .width(133.dp)
                ) {
                    Text(
                        "Select Model",
                        fontSize = 14.sp,
                        color = MaterialTheme.colorScheme.chatBtnText
                    )
                }
                Spacer(Modifier.height(12.dp))
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(
                        painter = painterResource(R.drawable.ic_info),
                        contentDescription = "info",
                        tint = MaterialTheme.colorScheme.chatInfoText,
                        modifier = Modifier.size(16.dp)
                    )
                    Spacer(Modifier.width(4.dp))
                    Text(
                        "No Model Selected",
                        fontSize = 12.sp,
                        color = MaterialTheme.colorScheme.chatInfoText
                    )
                }
            }
        }


    }
}

@Composable
private fun ChatMessageListSection(
//    state: ChatUiState,
//    listState: LazyListState,
    viewModel: ChatViewModel,
//    menuState: MenuState,
//    setMenuState: (MenuState) -> Unit,
    chatMessageListBottomPadding: MutableState<Int>
) {
    // 优化：使用 remember 缓存消息列表，减少重建
//    val messages = remember(state.currentSession?.messages, state.streamBuffer) {
//        val baseMessages = state.currentSession?.messages ?: emptyList()
//        val streamingThinking = state.streamBuffer.think.toString()
//        val streamingReply = state.streamBuffer.reply.toString()
//
//        buildList {
//            addAll(baseMessages)
//            if (streamingThinking.isNotBlank()) {
//                add(
//                    ChatUiMessage(
//                        content = streamingThinking,
//                        id = "streaming-think",
//                        role = "bot",
//                        type = ChatMessageType.IN_THINK
//                    )
//                )
//            }
//            if (streamingReply.isNotBlank()) {
//                add(
//                    ChatUiMessage(
//                        content = streamingReply,
//                        id = "streaming-reply",
//                        role = "bot",
//                        type = ChatMessageType.BOT
//                    )
//                )
//            }
//        }
//    }

    ChatMessageList(
        viewModel = viewModel,
        messages = viewModel.state.value.chatMsgList,
//        listState = listState,
//        streamBuffer = state.streamBuffer,
//        onCopyToInput = { content -> viewModel.onIntent(ChatIntent.InputChanged(content)) },
//        modifier = Modifier.fillMaxSize(),
//        menuState = menuState,
//        setMenuState = setMenuState,
        chatMessageListBottomPadding = chatMessageListBottomPadding
    )
}

@OptIn(ExperimentalPermissionsApi::class)
@Composable
private fun ChatInputBarSection(
    navController: NavController,
    viewModel: ChatViewModel,
    onImportImageClick: () -> Unit,
    onImportAudioClick: () -> Unit,
    onRecorderClick: () -> Unit,
    modifier: Modifier = Modifier,
    editContent: String,
    onCloseEdit: () -> Unit,
    onOpenSystemCameraClick: () -> Unit
) {
    val state by viewModel.state.collectAsState()
    val keyboardController = LocalSoftwareKeyboardController.current
    val context = LocalContext.current
    val coroutineScope = rememberCoroutineScope()

    ChatInputBar(
        modifier = modifier,
        viewModel = viewModel,
        input = state.input,
        onInputChange = {
            viewModel.onIntent(ChatIntent.InputChanged(it))
        },
        onSend = { newText ->
            // 发送新消息
            viewModel.onIntent(
                ChatIntent.SendMessage(
                    newText,
//                    imageFiles = state.selectedImageFiles,
//                    audioFile = state.selectedAudioFiles
                )
            )
        },
        onPhotoClick = { onImportImageClick() },
        onAudioClick = { onImportAudioClick() },
        selectedImages = state.selectedImageFiles,
//        selectedAudios = state.selectedAudioFiles,
        onRemoveImage = { file -> viewModel.onIntent(ChatIntent.RemoveImage(file)) },
//        onRemoveAudio = { file -> viewModel.onIntent(ChatIntent.RemoveAudio(file)) },
        onConfigClick = { navController.navigate("settings") },
        onRecorderClick = onRecorderClick,
        onStopClick = { viewModel.onIntent(ChatIntent.StopWaitingResponse) },
        editContent = editContent,
        onCloseEdit = { onCloseEdit() },
        onCameraClick = onOpenSystemCameraClick,
    )
}

@Composable
private fun HandleChatScreenEvents(
//    viewModel: ChatViewModel,
//    state: ChatUiState,
    listState: LazyListState,
    bottomPadding: MutableState<Int>,
    shouldAutoScroll: Boolean
) {
//    val eventFlow = viewModel.event
//    val density = LocalDensity.current.density
//    val context = LocalContext.current
//    val banner = LocalBannerState.current
//    val curCoroutineScope = rememberCoroutineScope()
//    var job: Job? = null
//    LaunchedEffect(Unit) {
//        L.d("nfl", "chat screen HandleChatScreenEvents")
//        job = curCoroutineScope.launch {
//            eventFlow.collect { event ->
//                L.d("nfl", "chat screen event:$event")
//                when (event) {
//                }
//            }
//        }
//    }
//    DisposableEffect(Unit) {
//        onDispose {
//            L.d("nfl", "chat screen HandleChatScreenEvents cancel job:$job")
//            job?.cancel()
//        }
//    }
}


@Composable
fun rememberImportImagesLauncher(
    context: Context,
    coroutineScope: CoroutineScope,
    viewModel: ChatViewModel
) =
    rememberLauncherForActivityResult(ActivityResultContracts.OpenMultipleDocuments()) { uris: List<Uri> ->
        coroutineScope.launch {
            val savedFiles = saveUrisToInternalImages(context, uris)
            L.e("peter", "all file name: ${savedFiles.map { it.name }}")
            viewModel.onIntent(ChatIntent.AddImage(savedFiles))
        }
    }

suspend fun saveUrisToInternalImages(context: Context, uris: List<Uri>): List<File> =
    withContext(Dispatchers.IO) {
        uris.map { uri ->
            val modelsDir = File(context.filesDir, "images")
            if (!modelsDir.exists()) modelsDir.mkdirs()
            val dest = File(modelsDir, AppUtils.getFileName(context, uri))
            if (!dest.exists()) {
                context.contentResolver.openInputStream(uri)?.use { input ->
                    dest.outputStream().use { output ->
                        val buffer = ByteArray(8 * 1024)
                        var bytesRead: Int
                        while (input.read(buffer).also { bytesRead = it } != -1) {
                            output.write(buffer, 0, bytesRead)
                        }
                    }
                }
            }
            dest
        }
    }

@Composable
fun rememberImportAudioLauncher(
    context: Context,
    coroutineScope: CoroutineScope,
) =
    rememberLauncherForActivityResult(ActivityResultContracts.OpenMultipleDocuments()) { uris: List<Uri> ->
        coroutineScope.launch {
            val savedFiles = uris.map { uri ->
                async(Dispatchers.IO) {
                    val audioDir = File(context.filesDir, "audio")
                    if (!audioDir.exists()) audioDir.mkdirs()
                    L.e("peter", "uri: $uri")
                    val dest = File(audioDir, AppUtils.getFileName(context, uri))
                    if (!dest.exists()) {
                        context.contentResolver.openInputStream(uri)?.use { input ->
                            dest.outputStream().use { output ->
                                val buffer = ByteArray(8 * 1024)
                                var bytesRead: Int
                                while (input.read(buffer).also { bytesRead = it } != -1) {
                                    output.write(buffer, 0, bytesRead)
                                }
                            }
                        }
                    }
                    dest
                }
            }.awaitAll()
            L.e("peter", "all file name: ${savedFiles.map { it.name }}")
        }
    }

@Composable
fun FakeProgressBar(
    isLoading: Boolean,
    modifier: Modifier = Modifier,
    onComplete: () -> Unit = {}
) {
    var progress by remember { mutableStateOf(0f) }

    LaunchedEffect(isLoading) {
        if (isLoading) {
            progress = 0f
            while (progress < 0.92f) {
                delay(25)
                progress += Random.nextFloat() * 0.01f + 0.005f
                if (progress > 0.92f) progress = 0.92f
            }
        } else {
            if (progress > 0f && progress < 1f) {
                progress = 1f
                onComplete()
                delay(350)
            }
            progress = 0f
        }
    }

    if (progress > 0f) {
        LinearProgressIndicator(
            gapSize = 0.dp,
            drawStopIndicator = {},
            progress = { progress },
            modifier = modifier
                .fillMaxWidth()
                .height(4.dp),
            color = MaterialTheme.colorScheme.progressBarColor,
            trackColor = MaterialTheme.colorScheme.progressTrackColor
        )
    }
}