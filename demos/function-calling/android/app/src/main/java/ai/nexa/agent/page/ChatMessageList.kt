package ai.nexa.agent.page

import ai.nexa.agent.R
import ai.nexa.agent.bean.ChatMessageType
import ai.nexa.agent.bean.ChatUiMessage
import ai.nexa.agent.bean.GoogleCalendarEvent
import ai.nexa.agent.constant.Configs
import ai.nexa.agent.model.ChatViewModel
import ai.nexa.agent.state.LocalBannerState
import ai.nexa.agent.state.OperationState
import ai.nexa.agent.ui.theme.ChatMessageAssistantProfilingButton
import ai.nexa.agent.ui.theme.Primary
import ai.nexa.agent.ui.theme.chatImagePreviewBg
import ai.nexa.agent.ui.theme.chatImagePreviewDivider
import ai.nexa.agent.ui.theme.chatImagePreviewTitle
import ai.nexa.agent.ui.theme.chatInputBackground
import ai.nexa.agent.ui.theme.chatInputBorder
import ai.nexa.agent.ui.theme.chatMessageThinkBg
import ai.nexa.agent.ui.theme.chatMessageThinkBorder
import ai.nexa.agent.ui.theme.chatMessageThinkContent
import ai.nexa.agent.ui.theme.chatMessageThinkExpandIcon
import ai.nexa.agent.ui.theme.chatMessageThinkTitle
import ai.nexa.agent.ui.theme.chatMessageUserBg
import ai.nexa.agent.ui.theme.chatMessageUserText
import ai.nexa.agent.ui.theme.commonSettingsUnSelectedText
import ai.nexa.agent.ui.theme.errorBannerIcon
import ai.nexa.agent.ui.theme.errorBannerText
import ai.nexa.agent.ui.theme.ic_close
import ai.nexa.agent.ui.theme.ic_dot
import ai.nexa.agent.util.ImgUtil
import ai.nexa.agent.util.L
import ai.nexa.agent.util.ShareUtil
import ai.nexa.agent.util.chatMessageStyle
import android.graphics.BitmapFactory
import android.media.MediaPlayer
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.MutableState
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateMapOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.layout.LayoutCoordinates
import androidx.compose.ui.layout.onGloballyPositioned
import androidx.compose.ui.platform.LocalClipboardManager
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.AnnotatedString
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.font.FontWeight.Companion.W500
import androidx.compose.ui.text.style.LineBreak
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.withStyle
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import androidx.compose.ui.window.DialogProperties
import coil3.compose.rememberAsyncImagePainter
import com.nexa.studio.ui.chat.markdown.CustomMarkdown
import com.nexa.studio.ui.chat.markdown.rememberStreamingMarkdownState
import kotlinx.coroutines.delay
import java.text.SimpleDateFormat

private val mediaPlayer = MediaPlayer()

@Composable
fun ChatMessageList(
//    modifier: Modifier = Modifier,
    viewModel: ChatViewModel,
    messages: List<ChatUiMessage>,
//    streamBuffer: StreamBuffer,
//    listState: LazyListState = rememberLazyListState(),
//    onCopyToInput: (String) -> Unit,
//    menuState: MenuState,
//    setMenuState: (MenuState) -> Unit,
    chatMessageListBottomPadding: MutableState<Int>
) {
    val expandedThinks = remember { mutableStateMapOf<String, Boolean>() }
//    val tts = rememberTts()

    // 优化：只收集需要的状态，减少重组
    val state by viewModel.state.collectAsState()
//    val modelType = state.modelType
    val showRegenerateIcon = remember(Unit) {
        true
    }
    val density = LocalDensity.current.density
//
//    DisposableEffect(Unit) {
//        onDispose {
//            L.d("ChatMessageList", "release mediaPlayer")
//            try {
//                // mediaPlayer.release()
//            } catch (e: Exception) {
//                L.e("ChatMessageList", "release mediaPlayer error: ${e.message}")
//            }
//        }
//    }

    Box(modifier = Modifier.fillMaxSize()) {
        LazyColumn(
//            state = listState,
//            modifier = modifier.fillMaxSize()
//                .onSizeChanged {
//                    val tempHeight = it.height / density - Configs.userChatMsgBoxHeight
//                    chatMessageListBottomPadding.value = max(chatMessageListBottomPadding.value, tempHeight.toInt())
//                }
        ) {
            items(messages, key = { item ->
                when (item.type) {
                    ChatMessageType.BOT -> item.id + item.content.hashCode()
                    else -> item.id
                }
            }) { msg ->
                when (msg.type) {
                    ChatMessageType.Error -> {
                        ErrorMessageContent(
                            msg,
                            onRegenerate = {
//                                viewModel.onIntent(ChatIntent.Regenerate(msg))
                            },
                            showRegenerateIcon
                        )
                    }

                    ChatMessageType.IN_THINK -> {
                        L.d("nfl", "current state:IN_THINK")
                        val expanded = expandedThinks[msg.id] ?: false
                        val title =
                            if (false) {
                                "Thinking Completed"
                            } else {
                                animatedThinkTitle()
                            }
                        ThinkBubble(
                            title = title,
                            content = msg.content,
                            expanded = expanded,
                            onExpandToggle = { expandedThinks[msg.id] = !expanded }
                        )
                    }

                    ChatMessageType.THINK -> {
                        val expanded = expandedThinks[msg.id] ?: false
                        val title = "Thinking Completed"
                        L.d("nfl", "current state:THINK")
                        ThinkBubble(
                            title = title,
                            content = msg.content,
                            expanded = expanded,
                            onExpandToggle = { expandedThinks[msg.id] = !expanded }
                        )
                    }

                    ChatMessageType.USER_Audio -> {

                        msg.audioList.forEach { filePath ->
//                            AudioFilePreviewItem(filePath = filePath)
                        }

                    }

                    ChatMessageType.USER_IMAGE -> {
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(start = 16.dp, end = 16.dp, top = 16.dp),
                            horizontalArrangement = Arrangement.End
                        ) {
                            UserImageGallery(
                                imagePaths = msg.imageList,
                                modifier = Modifier.padding(8.dp)
                            )
//                            UserImageBubble(
//                                imagePath = msg.content,
//                                modifier = Modifier.padding(8.dp)
//                            )
                        }
                    }

                    else -> {
                        if (msg.role == "user") {
                            var coords by remember { mutableStateOf<LayoutCoordinates?>(null) }
                            MessageBubble(
                                msg,
                                modifier = Modifier
                                    .onGloballyPositioned { coords = it }
                                    .pointerInput(msg.hashCode()) {
                                        detectTapGestures(
                                            onLongPress = { localOffset ->
                                                val windowOffset =
                                                    coords?.localToWindow(localOffset)
//                                                setMenuState(
//                                                    MenuState(
//                                                        show = true,
//                                                        type = "user",
//                                                        offset = windowOffset ?: localOffset,
//                                                        message = msg
//                                                    )
//                                                )
                                            }
                                        )
                                    }
                            )
                        } else {
                            var coords by remember { mutableStateOf<LayoutCoordinates?>(null) }
                            val clipboardManager = LocalClipboardManager.current

                            AssistantMessageRichContent(
                                deviceId = "NPU",
                                message = msg,
                                onCopy = { clipboardManager.setText(AnnotatedString(msg.content)) },
                                onSpeak = { text ->
//                                    tts.speak(text, TextToSpeech.QUEUE_FLUSH, null, null)
                                },
//                                onRegenerate = { message ->
//                                    viewModel.onIntent(
//                                        ChatIntent.Regenerate(
//                                            message
//                                        )
//                                    )
//                                },
                                showRegenerateIcon = showRegenerateIcon,
                                modifier = Modifier
                                    .onGloballyPositioned { coords = it }
                                    .pointerInput(msg.hashCode()) {
                                        detectTapGestures(
                                            onLongPress = { localOffset ->
                                                val windowOffset =
                                                    coords?.localToWindow(localOffset)
//                                                setMenuState(
//                                                    MenuState(
//                                                        show = true,
//                                                        type = "assistant",
//                                                        offset = windowOffset ?: localOffset,
//                                                        message = msg
//                                                    )
//                                                )
                                            }
                                        )
                                    },
                                useTextView = false
                            )
                        }
                    }
                }
            }
            item {
                Spacer(Modifier.height(0.5.dp))
            }
            if (state.operationState == OperationState.WAITING_SEND_RESPONSE) {
//                item {
//                    ThreeBallLoading(modifier = Modifier.background(Color.Transparent).padding(horizontal = 12.dp))
//                }

                item {
                    ChatLoadingAnimation(
                        modifier = Modifier
                            .padding(start = 16.dp)
                            .size(22.dp)
                    )
                }
            }
            item {
                Spacer(modifier = Modifier.height(chatMessageListBottomPadding.value.dp))
            }
        }
    }
}

@Composable
fun ChatLoadingAnimation(
    modifier: Modifier = Modifier,
    durationMs: Int = 450
) {
    val frames = listOf(
        R.drawable.chat_loading1,
        R.drawable.chat_loading2,
        R.drawable.chat_loading3,
        R.drawable.chat_loading4,
    )
    val frameCount = frames.size
    var frameIndex by remember { mutableStateOf(0) }

    LaunchedEffect(Unit) {
        while (true) {
            frameIndex = (frameIndex + 1) % frameCount
            delay((durationMs / frameCount).toLong())

        }
    }

    Image(
        painter = painterResource(id = frames[frameIndex]),
        contentDescription = "Loading",
        modifier = modifier
    )
}

@Composable
fun UserImageGallery(
    imagePaths: List<String>,
    modifier: Modifier = Modifier
) {
    var previewImage by remember { mutableStateOf<String?>(null) }

    if (imagePaths.size == 1) {
        UserImageBubble(
            imagePath = imagePaths[0],
            modifier = modifier
                .clickable { previewImage = imagePaths[0] }
        )
    } else {
        LazyRow(
            modifier = modifier,
            horizontalArrangement = Arrangement.spacedBy(5.dp)
        ) {
            items(imagePaths) { path ->
                Box(
                    modifier = Modifier
                        .size(100.dp)
                        .clip(RoundedCornerShape(12.dp))
                        .clickable { previewImage = path }
                ) {
                    Image(
                        painter = rememberAsyncImagePainter(model = path),
                        contentDescription = null,
                        modifier = Modifier.matchParentSize(),
                        contentScale = ContentScale.Crop
                    )
                }
            }
        }
    }

    if (previewImage != null) {
        FullscreenImagePreview(
            imagePath = previewImage,
            onDismiss = { previewImage = null }
        )
    }
}

@Composable
fun FullscreenImagePreview(
    imagePath: String?,
    onDismiss: () -> Unit,
    title: String? = null
) {
    if (imagePath == null) return
    val context = LocalContext.current
    val banner = LocalBannerState.current
    Dialog(
        onDismissRequest = { onDismiss() },
        properties = DialogProperties(usePlatformDefaultWidth = false)
    ) {
        Column(
            Modifier
                .fillMaxSize()
                .background(MaterialTheme.colorScheme.chatImagePreviewBg)
        ) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(48.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                if (!title.isNullOrEmpty()) {
                    Text(
                        text = title,
                        color = MaterialTheme.colorScheme.chatImagePreviewTitle,
                        style = MaterialTheme.typography.titleMedium,
                        modifier = Modifier
                            .padding(start = 16.dp)
                            .weight(1f)
                    )
                } else {
                    Spacer(modifier = Modifier.weight(1f))
                }
                IconButton(
                    onClick = onDismiss,
                    modifier = Modifier
                        .padding(end = 8.dp)
                        .size(48.dp)
                ) {
                    Icon(
                        modifier = Modifier.size(16.dp),
                        painter = painterResource(id = MaterialTheme.colorScheme.ic_close),
                        tint = Color.Unspecified,
                        contentDescription = "close"
                    )
                }
            }
            HorizontalDivider(
                thickness = 0.5.dp,
                color = MaterialTheme.colorScheme.chatImagePreviewDivider
            )

            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .weight(1f),
                contentAlignment = Alignment.Center
            ) {
                Image(
                    painter = rememberAsyncImagePainter(model = imagePath),
                    contentDescription = null,
                    modifier = Modifier
                        .fillMaxWidth(1f)
                        .fillMaxHeight(1f)
                        .clip(RoundedCornerShape(20.dp)),
                    contentScale = ContentScale.Fit
                )
            }

            val showOperations = false
            if (showOperations) {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    modifier = Modifier
                        .padding(bottom = 34.dp)
                        .align(Alignment.CenterHorizontally)
                ) {
                    IconButton(
                        onClick = {
                            ImgUtil.saveImageToGallery(context, imagePath)
                            banner.show("Save Image to Gallery Success!")
                        },
                        modifier = Modifier
                            .size(48.dp)
                            .clip(CircleShape)
                            .border(1.dp, Color.Black, CircleShape)
                    ) {
                        Icon(
                            painter = painterResource(R.drawable.ic_download),
                            contentDescription = "",
                            modifier = Modifier
                                .clip(CircleShape)
                                .fillMaxSize()
                                .padding(14.dp),
                            tint = Color.Black
                        )
                    }
                    Spacer(modifier = Modifier.width(8.dp))
                    IconButton(
                        onClick = {
                            ShareUtil.shareImg(context, imagePath)
                        },
                        modifier = Modifier
                            .size(48.dp)
                            .clip(CircleShape)
                            .border(1.dp, Color.Black, CircleShape)
                    ) {
                        Icon(
                            painter = painterResource(R.drawable.ic_rec_share),
                            contentDescription = "",
                            modifier = Modifier
                                .clip(CircleShape)
                                .fillMaxSize()
                                .padding(16.dp),
                            tint = Color.Black
                        )
                    }
                }
            }
        }
    }
}

@Composable
fun UserImageBubble(
    imagePath: String,
    modifier: Modifier = Modifier
) {
    var imgSize by remember { mutableStateOf<Pair<Int, Int>?>(null) }

    LaunchedEffect(imagePath) {
        val options = BitmapFactory.Options().apply { inJustDecodeBounds = true }
        BitmapFactory.decodeFile(imagePath, options)
        imgSize = options.outWidth to options.outHeight
    }

    val maxH = 240.dp
    val maxW = 240.dp

    val density = LocalDensity.current

    val boxModifier = imgSize?.let { (w, h) ->
        var scaledW = w
        var scaledH = h
        density.run {
            if (h > w) {
                // 让 h 缩放到 maxH
                val maxPxH = maxH.toPx()
                if (h > 0) {
                    val ratio = maxPxH / h
                    scaledH = maxPxH.toInt()
                    scaledW = (w * ratio).toInt()
                }
            } else {
                val maxPxW = maxW.toPx()
                if (w > 0) {
                    val ratio = maxPxW / w
                    scaledW = maxPxW.toInt()
                    scaledH = (h * ratio).toInt()
                }
            }
        }
        modifier
            .height(with(density) { scaledH.toDp() })
            .width(with(density) { scaledW.toDp() })
    } ?: modifier.size(180.dp)

    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        Image(
            painter = rememberAsyncImagePainter(model = imagePath),
            contentDescription = "User Uploaded Image",
            modifier = boxModifier.clip(RoundedCornerShape(16.dp)),
            contentScale = ContentScale.Fit
        )
    }
}

@Composable
fun MessageBubble(
    message: ChatUiMessage,
    modifier: Modifier
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(start = 16.dp, end = 16.dp, top = 16.dp),
        horizontalArrangement = Arrangement.End
    ) {
        Box(
            modifier = modifier
                .background(
                    color = MaterialTheme.colorScheme.chatMessageUserBg,
                    shape = RoundedCornerShape(
                        topStart = 16.dp,
                        topEnd = 16.dp,
                        bottomEnd = 0.dp,
                        bottomStart = 16.dp
                    )
                )
                .padding(horizontal = 14.dp, vertical = 10.dp)
                .widthIn(max = 320.dp)
        ) {
            MessageRichContent(
                message
            )
        }
    }
}

@Composable
fun ErrorMessageContent(
    message: ChatUiMessage,
    onRegenerate: () -> Unit,
    showRegenerateIcon: Boolean
) {
    Row(
        verticalAlignment = Alignment.CenterVertically
    ) {
        Icon(
            painterResource(MaterialTheme.colorScheme.errorBannerIcon),
            contentDescription = null,
            tint = Color.Unspecified,
            modifier = Modifier
                .padding(start = 16.dp)
                .size(16.dp)
        )
        Spacer(modifier = Modifier.width(6.dp))
        Text(
            message.content,
            color = MaterialTheme.colorScheme.errorBannerText,
            fontSize = 14.sp,
            modifier = Modifier.weight(1f)
        )
        if (showRegenerateIcon) {
            IconButton(
                onClick = onRegenerate,
                modifier = Modifier
                    .padding(start = 8.dp, end = 16.dp)
                    .size(24.dp)
            ) {
                Icon(
                    modifier = Modifier.size(16.dp),
                    painter = painterResource(id = R.drawable.ic_refresh),
                    tint = MaterialTheme.colorScheme.ChatMessageAssistantProfilingButton,
                    contentDescription = "regenerate"
                )
            }
        }
    }
}

@Composable
fun MessageRichContent(
    message: ChatUiMessage
) {
    val customTitleLineBreak = LineBreak(
        strategy = LineBreak.Strategy.HighQuality,
        strictness = LineBreak.Strictness.Strict,
        wordBreak = LineBreak.WordBreak.Phrase
    )
    Text(
        message.content.trim(),
        style = chatMessageStyle.copy(lineBreak = customTitleLineBreak),
        textAlign = TextAlign.Justify,
        fontSize = 16.sp,
        color = MaterialTheme.colorScheme.chatMessageUserText
    )
}

@Composable
private fun GoogleCalendarItemTitle(title: String) {
    Row(
        verticalAlignment = Alignment.CenterVertically,
        modifier = Modifier.fillMaxWidth()
    ) {
        Box(
            contentAlignment = Alignment.Center,
            modifier = Modifier.size(10.dp)
        ) {
            Icon(
                painter = painterResource(R.drawable.light_green_dot_bg),
                contentDescription = null,
                tint = Color.Unspecified,
                modifier = Modifier.size(10.dp)
            )
            Icon(
                painter = painterResource(R.drawable.light_green_dot_front),
                contentDescription = null,
                tint = Color.Unspecified,
                modifier = Modifier.size(7.dp)
            )
        }

        Spacer(modifier = Modifier.width(5.dp))
        Text(
            title,
            fontWeight = W500,
            textAlign = TextAlign.Justify,
            fontSize = 16.sp,
            color = MaterialTheme.colorScheme.chatMessageUserText
        )
    }
}

@Composable
private fun GoogleCalendarItemContentTag(itemLeftMargin: Int, tip: String, tag: String) {
    Row() {
        Text(
            tip,
            textAlign = TextAlign.Justify,
            modifier = Modifier
                .padding(
                    itemLeftMargin.dp,
                    0.dp,
                    0.dp,
                    0.dp
                ),
            fontSize = 14.sp,
            color = MaterialTheme.colorScheme.chatMessageUserText
        )
        Text(
            tag,
            textAlign = TextAlign.Justify,
            modifier = Modifier
                .padding(
                    5.dp,
                    0.dp,
                    0.dp,
                    0.dp
                )
                .clip(RoundedCornerShape(8.dp))
                .background(color = MaterialTheme.colorScheme.chatInputBackground)
                .border(
                    width = 1.dp,
                    color = MaterialTheme.colorScheme.chatInputBorder,
                    shape = RoundedCornerShape(8.dp)
                ).padding(horizontal = 6.dp),
            fontSize = 14.sp,
            color = MaterialTheme.colorScheme.commonSettingsUnSelectedText
        )
    }
}

@Composable
fun AssistantMessageRichContent(
    deviceId: String,
    modifier: Modifier,
    message: ChatUiMessage,
    onCopy: () -> Unit,
    onSpeak: (String) -> Unit,
//    onRegenerate: (ChatUiMessage) -> Unit,
    showRegenerateIcon: Boolean,
    useTextView: Boolean = true
) {
    if (message.extMsg is GoogleCalendarEvent) {
        val event = message.extMsg.event
        Column(
            modifier = modifier
                .fillMaxWidth()
                .padding(start = 16.dp, end = 16.dp, top = 16.dp)
        ) {
            Text(
                "The event has been successfully added to you canlendar.",
                textAlign = TextAlign.Justify,
                fontSize = 16.sp,
                color = MaterialTheme.colorScheme.chatMessageUserText
            )
            Spacer(modifier = Modifier.height(10.dp))
            Column(
                modifier = Modifier
                    .border(
                        width = 1.dp,
                        color = MaterialTheme.colorScheme.chatMessageThinkBorder,
                        shape = RoundedCornerShape(16.dp)
                    )
                    .padding(15.dp)
            ) {
                val itemLeftMargin = 15
                Text(
                    text = buildAnnotatedString {
                        append("Event added to ")
                        withStyle(style = SpanStyle(color = Primary)) {
                            append("@Google Calendar")
                        }
                    },
                    textAlign = TextAlign.Justify,
                    fontWeight = W500,
                    fontSize = 18.sp,
                    color = MaterialTheme.colorScheme.chatMessageUserText
                )
                Spacer(modifier = Modifier.height(20.dp))
                GoogleCalendarItemTitle("Event name")
                Text(
                    "${event?.summary}",
                    textAlign = TextAlign.Justify,
                    modifier = Modifier.padding(itemLeftMargin.dp, 0.dp, 0.dp, 0.dp),
                    fontSize = 14.sp,
                    color = MaterialTheme.colorScheme.chatMessageUserText
                )
                Spacer(modifier = Modifier.height(10.dp))
//                val startDateStr = "2025-12-20T09:00:00-11:00"
                var date: String = ""
                var startTime: String = ""
                var endTime: String = ""
                event?.start?.dateTime?.split("T")?.let {
                    if (it.size == 2) {
                        date = it[0]

                        it[1].split("-").let { times ->
                            if (times.size == 2) {
                                startTime = times[0]
                                endTime = times[1]
                                GoogleCalendarItemTitle("Event time")
                                GoogleCalendarItemContentTag(
                                    itemLeftMargin, "Date:", " ${
                                        SimpleDateFormat("yyyy-MM-dd").parse(date).toString()
                                            .split("00:00:00").let { dates ->
                                                dates[0]
                                            }
                                    }")
                                Spacer(modifier = Modifier.height(5.dp))
                                Row() {
                                    GoogleCalendarItemContentTag(
                                        itemLeftMargin, "Start:", startTime.take(5)
                                    )
                                    Spacer(modifier = Modifier.height(6.dp))
                                    GoogleCalendarItemContentTag(
                                        itemLeftMargin, "End:", endTime
                                    )
                                }
                                Spacer(modifier = Modifier.height(10.dp))
                            }
                        }
                    }
                }
                GoogleCalendarItemTitle("Event Location")
                Text(
                    "${event?.location}",
                    textAlign = TextAlign.Justify,
                    fontSize = 14.sp,
                    modifier = Modifier.padding(itemLeftMargin.dp, 0.dp, 0.dp, 0.dp),
                    color = MaterialTheme.colorScheme.chatMessageUserText
                )
                Spacer(modifier = Modifier.height(10.dp))
                GoogleCalendarItemTitle("Event Description")
                Text(
                    "${event?.description}",
                    textAlign = TextAlign.Justify,
                    modifier = Modifier.padding(itemLeftMargin.dp, 0.dp, 0.dp, 0.dp),
                    fontSize = 14.sp,
                    color = MaterialTheme.colorScheme.chatMessageUserText
                )
            }
        }
        return
    }
    var expanded by remember { mutableStateOf(false) }

    // Format message content to remove template tags
    fun formatContent(content: String): String {
        message.originContent = content  // Preserve original data

        return if (!content.startsWith("<|")) {
            content
        } else {
            val tag = Configs.gptOssFinalTag
            val position = content.lastIndexOf(tag)
            if (position < 0) {
                ""
            } else {
                if (tag.length + position < content.length) {
                    content.substring(tag.length + position)
                } else {
                    ""
                }
            }
        }
    }

    // Format and cache content
    val formattedContent = remember("message.content") {
        formatContent(message.content)
    }

    // Use optimized streaming markdown state (sync + smart throttling for long text)
    val markdownState = rememberStreamingMarkdownState(
        content = formattedContent,
        debounceMs = 150L  // Debounce for long content (>1500 chars)
    )

    Column(
        modifier = modifier
            .fillMaxWidth()
            .padding(start = 16.dp, end = 16.dp, top = 16.dp)
    ) {
        // Display markdown content (no flickering with immediate parsing)
        CustomMarkdown(
            markdownState = markdownState,
            modifier = Modifier.fillMaxWidth()
        )

//        message.profilingData?.let { prof ->
//            Row(
//                modifier = Modifier.padding(top = 8.dp),
//                horizontalArrangement = Arrangement.spacedBy(3.dp)
//            ) {
//                IconButton(onClick = onCopy, modifier = Modifier.size(24.dp)) {
//                    Icon(
//                        modifier = Modifier.size(16.dp),
//                        painter = painterResource(id = R.drawable.ic_copy),
//                        tint = MaterialTheme.colorScheme.ChatMessageAssistantProfilingButton,
//                        contentDescription = "copy"
//                    )
//                }
////                IconButton(
////                    onClick = { onSpeak(message.content) },
////                    modifier = Modifier.size(24.dp)
////                ) {
////                    Icon(
////                        modifier = Modifier.size(16.dp),
////                        painter = painterResource(id = R.drawable.ic_volume),
////                        tint = MaterialTheme.colorScheme.ChatMessageAssistantProfilingButton,
////                        contentDescription = "volume"
////                    )
////                }
//                if (showRegenerateIcon) {
//                    IconButton(
//                        onClick = { onRegenerate(message) },
//                        modifier = Modifier.size(24.dp)
//                    ) {
//                        Icon(
//                            modifier = Modifier.size(16.dp),
//                            painter = painterResource(id = R.drawable.ic_refresh),
//                            tint = MaterialTheme.colorScheme.ChatMessageAssistantProfilingButton,
//                            contentDescription = "regenerate"
//                        )
//                    }
//                }
//            }
//            Row(
//                modifier = Modifier
//                    .fillMaxWidth()
//                    .padding(vertical = 2.dp),
//                verticalAlignment = Alignment.Bottom
//            ) {
//                Text(
//                    modifier = Modifier.weight(1f),
//                    text = "TTFT: %.2fs;   Decode Speed: %.2f t/s".format(
//                        prof.ttftMs / 1000,
//                        prof.decodingSpeed
//                    ),
//                    color = MaterialTheme.colorScheme.chatMessageAssistantProfiling,
//                    style = MaterialTheme.typography.bodySmall
//                )
//
////                Row(
////                    modifier = Modifier
////                        .clickable { expanded = !expanded }
////                        .padding(start = 8.dp, end = 0.dp)
////                        .height(IntrinsicSize.Min),
////                    verticalAlignment = Alignment.CenterVertically
////                ) {
////                    Text(
////                        text = if (expanded) "Less" else "More",
////                        color = MaterialTheme.colorScheme.chatMessageAssistantProfiling,
////                        style = MaterialTheme.typography.bodySmall
////                    )
////                    Icon(
////                        painter = painterResource(
////                            id = R.drawable.ic_chevron_down
////                        ),
////                        tint = MaterialTheme.colorScheme.chatMessageAssistantProfiling,
////                        contentDescription = if (expanded) "Less" else "More"
////                    )
////                }
//            }
////            AnimatedVisibility(visible = expanded) {
////                InfoCard(deviceId, prof = prof)
////            }
//        }
    }
}

//private var countDownTimer: MyCountDownTimer? = null
private var isPlaying = false

//@Composable
//fun AudioFilePreviewItem(
//    filePath: String,
//    modifier: Modifier = Modifier
//) {
//    var duration by remember { mutableIntStateOf(0) }
//    var remainTime by remember { mutableIntStateOf(0) }
//    val maxTimeLength = 20_000
//
//    // 根据音频时长动态计算语音条宽度
//    val audioBarWidth = AudioUtils.calculateAudioBarWidthDp(
//        durationMs = duration
//    )
//
//    fun formatRemainTime(time: Int): Int {
//        return if (time < 1000) {
//            1
//        } else if (time in (maxTimeLength - 999) until maxTimeLength) {
//            maxTimeLength / 1000
//        } else {
//            time / 1000
//        }
//    }
//
//    fun getAudioInfo() {
//        try {
//            //1，创建MediaMetadataRetriever对象
//            val retriever = MediaMetadataRetriever()
//            L.d("nfl", "audio file path:$filePath")
//            //2.设置音视频资源路径
//            retriever.setDataSource(filePath)
//            //3.获取音视频资源总时长
//            duration =
//                retriever.extractMetadata(MediaMetadataRetriever.METADATA_KEY_DURATION)?.toInt()
//                    ?: 0
//            remainTime = formatRemainTime(duration)
//        } catch (e: java.lang.Exception) {
//            e.printStackTrace()
//        }
//    }
//
//    fun startPlayback() {
//        try {
//            mediaPlayer.reset()
//            mediaPlayer.setDataSource(filePath)
//            mediaPlayer.prepare()
//            mediaPlayer.start()
//            duration = mediaPlayer.duration
//            remainTime = formatRemainTime(duration)
//            L.d("nfl", "duration:$duration, remainTime:$remainTime")
//            countDownTimer = object : MyCountDownTimer(duration.toLong(), 1000L) {
//                override fun onTick(millisUntilFinished: Long) {
//                    remainTime = (millisUntilFinished / 1000).toInt()
//                }
//
//                override fun onCancel() {
//                    isPlaying = false
//                    remainTime = formatRemainTime(duration)
//                }
//
//                override fun onFinish() {
//                    isPlaying = false
//                    remainTime = formatRemainTime(duration)
//                }
//
//            }
//            countDownTimer!!.start()
//            isPlaying = true
//        } catch (e: Exception) {
//            e.printStackTrace()
//        }
//    }
//
//    fun pausePlayback() {
//        try {
//            isPlaying = false
//            mediaPlayer.pause()
//            remainTime = formatRemainTime(duration)
//            countDownTimer?.cancel()
//        } catch (e: Exception) {
//            e.printStackTrace()
//        }
//    }
//
//    LaunchedEffect(Unit) {
//        getAudioInfo()
//    }
//
//    Row(
//        modifier = Modifier
//            .fillMaxWidth()
//            .padding(start = 16.dp, end = 16.dp, top = 16.dp),
//        horizontalArrangement = Arrangement.End
//    ) {
//        Row(
//            horizontalArrangement = Arrangement.End,
//            modifier = modifier
//                .clickable(onClick = {
//                    if (isPlaying) {
//                        pausePlayback()
//                    } else {
//                        startPlayback()
//                    }
//                })
//                .background(
//                    color = MaterialTheme.colorScheme.chatMessageUserBg,
//                    shape = RoundedCornerShape(
//                        topStart = 16.dp,
//                        topEnd = 16.dp,
//                        bottomEnd = 0.dp,
//                        bottomStart = 16.dp
//                    )
//                )
//                .padding(horizontal = 14.dp, vertical = 10.dp)
//                .width(audioBarWidth)
//        ) {
//            Spacer(modifier = Modifier.width(10.dp))
//            Text("$remainTime", color = MaterialTheme.colorScheme.chatMessageUserText)
//            Text("\"", color = MaterialTheme.colorScheme.chatMessageUserText)
//            Spacer(modifier = Modifier.width(5.dp))
//            Icon(
//                painter = painterResource(R.drawable.ic_audio_wave),
//                tint = MaterialTheme.colorScheme.chatMessageUserText,
//                modifier = Modifier
//                    .padding(vertical = 5.dp)
//                    .align(Alignment.CenterVertically),
//                contentDescription = null
//            )
//        }
//    }
//}


fun formatTime(ms: Int): String {
    val sec = ms / 1000
    val min = sec / 60
    val s = sec % 60
    return "%02d:%02d".format(min, s)
}


//@Composable
//fun InfoCard(
//    deviceId:String,
//    prof: ProfilingUiData,
//    modifier: Modifier = Modifier
//) {
//    L.e("peter", "prof.usedMem:" + prof.usedMem)
//    L.e("peter", "prof.maxMem:" + prof.maxMem)
//    val usedGb = prof.usedMem / (1024.0 * 1024 * 1024)
//    val maxGb = prof.maxMem / (1024.0 * 1024 * 1024)
//    val valueString = String.format("%.1f", usedGb)
//    val unitString = String.format("GB/%.0fGB", maxGb)
//    Surface(
//        modifier = modifier
//            .fillMaxWidth()
//            .padding(top = 8.dp)
//            .height(IntrinsicSize.Min),
//        shape = RoundedCornerShape(16.dp),
//        color = Color.Transparent,
//        tonalElevation = 0.dp,
//        shadowElevation = 0.dp,
//        border = BorderStroke(1.dp, MaterialTheme.colorScheme.chatMessageInfoCardBorder)
//    ) {
//        Row(
//            modifier = Modifier
//                .padding(vertical = 14.dp)
//                .fillMaxWidth(),
//            horizontalArrangement = Arrangement.SpaceBetween
//        ) {
//            InfoColumn(
//                modifier = Modifier.padding(horizontal = 15.dp),
//                title = "Acceleration",
//                value = deviceId
//            )
//            InfoColumn(
//                modifier = Modifier.padding(horizontal = 15.dp),
//                title = "Prefill Speed",
//                value = "%.2f".format(prof.promptTokens / (prof.promptTimeMs / 1000)),
//                unit = "t/s"
//            )
//            // FIXME: 由于 memery 无法获取，暂时去掉
////            InfoColumn(
////                modifier = Modifier.padding(horizontal = 15.dp),
////                title = "Peak Memory",
////                value = valueString,
////                unit = unitString
////            )
//        }
//    }
//}


//@Composable
//fun InfoColumn(
//    modifier: Modifier = Modifier,
//    title: String,
//    value: String,
//    valueColor: Color = MaterialTheme.colorScheme.chatMessageAssistantInfoValue,
//    valueFontSize: TextUnit = 14.sp,
//    unit: String? = null
//) {
//    Column(
//        modifier = modifier
//    ) {
//        Text(
//            text = title,
//            fontSize = 10.sp,
//            color = MaterialTheme.colorScheme.chatMessageAssistantInfoTitle,
//            style = MaterialTheme.typography.labelSmall
//        )
//        Row(
//            verticalAlignment = Alignment.Bottom
//        ) {
//            Text(
//                text = value,
//                color = valueColor,
//                fontSize = valueFontSize,
//                modifier = Modifier.padding(vertical = 2.dp)
//            )
//            if (unit != null) {
//                Text(
//                    text = unit,
//                    color = MaterialTheme.colorScheme.chatMessageAssistantInfoUnit,
//                    fontSize = 10.sp,
//                    fontWeight = FontWeight.Normal,
//                    modifier = Modifier.padding(start = 2.dp, bottom = 2.dp)
//                )
//            }
//        }
//    }
//}


@Composable
fun ThinkBubble(
    title: String,
    content: String = "",
    expanded: Boolean,
    onExpandToggle: () -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(16.dp)
            .background(
                color = MaterialTheme.colorScheme.chatMessageThinkBg,
                shape = RoundedCornerShape(16.dp)
            )
            .border(
                width = 1.dp,
                color = MaterialTheme.colorScheme.chatMessageThinkBorder,
                shape = RoundedCornerShape(16.dp)
            )
            .clickable { onExpandToggle() }
            .padding(16.dp)
    ) {
        Row(verticalAlignment = Alignment.CenterVertically) {
            Icon(
                painter = painterResource(id = MaterialTheme.colorScheme.ic_dot),
                contentDescription = null,
                tint = Color.Unspecified,
                modifier = Modifier.size(18.dp)
            )
            Spacer(Modifier.width(6.dp))
            Text(
                title,
                fontWeight = FontWeight.Bold,
                color = MaterialTheme.colorScheme.chatMessageThinkTitle,
                fontSize = 16.sp
            )
            Spacer(Modifier.weight(1f))
            Icon(
                painter = painterResource(id = if (expanded) R.drawable.ic_chevron_down else R.drawable.ic_chevron_right),
                contentDescription = null,
                tint = MaterialTheme.colorScheme.chatMessageThinkExpandIcon,
                modifier = Modifier.size(18.dp)
            )
        }
        AnimatedVisibility(visible = expanded) {
            Column() {
                Spacer(Modifier.height(5.dp))
                val customTitleLineBreak = LineBreak(
                    strategy = LineBreak.Strategy.HighQuality,
                    strictness = LineBreak.Strictness.Strict,
                    wordBreak = LineBreak.WordBreak.Phrase
                )
                Text(
                    content,
                    modifier = Modifier.fillMaxWidth(),
                    style = TextStyle.Default.copy(lineBreak = customTitleLineBreak),
                    textAlign = TextAlign.Justify,
                    color = MaterialTheme.colorScheme.chatMessageThinkContent,
                    fontSize = 15.sp
                )
            }
        }
    }
}

fun String.halfToFull(): String {
    val result = StringBuilder()
    for (c in this) {
        when (c) {
            in '\u0020'..'\u007E' -> {
                // 半角字符范围：空格到~（包括数字、字母、标点等）
                val fullChar = (c.code + 0xFEE0).toChar()
                result.append(fullChar)
            }

            else -> result.append(c)
        }
    }
    return result.toString()
}


@Composable
fun animatedThinkTitle(): String {
    val thinkStages = listOf("Thinking.", "Thinking..", "Thinking...")
    var stage by remember { mutableStateOf(0) }

    LaunchedEffect(Unit) {
        while (true) {
            stage = (stage + 1) % thinkStages.size
            delay(400L)
        }
    }
    return thinkStages[stage]
}

//@Composable
//fun userMenuActions(
//    message: ChatUiMessage,
//    viewModel: ChatViewModel,
//    onCopyToInput: (String) -> Unit
//): List<MenuAction> {
//    val clipboardManager = LocalClipboardManager.current
//    return listOf(
//        MenuAction(
//            icon = painterResource(id = R.drawable.ic_copy),
//            text = "Copy",
//            onClick = {
//                clipboardManager.setText(AnnotatedString(message.content))
//            }
//        ),
//        MenuAction(
//            icon = painterResource(id = R.drawable.ic_pen),
//            text = "Edit",
//            onClick = {
//                onCopyToInput(message.content)
//                viewModel.setEditingMessage(message) // 这里记录正在编辑的消息
//            }
//        )
//    )
//}

//@Composable
//fun botMenuActions(
//    message: ChatUiMessage,
//    onRegenerate: (ChatUiMessage) -> Unit,
//    showRegenerateIcon: Boolean
//): List<MenuAction> {
//    val clipboardManager = LocalClipboardManager.current
////    val tts = rememberTts()
//    val result = arrayListOf(
//        MenuAction(
//            icon = painterResource(id = R.drawable.ic_copy),
//            text = "Copy",
//            onClick = { clipboardManager.setText(AnnotatedString(message.content)) }
//        ))
////    MenuAction(
////        icon = painterResource(id = R.drawable.ic_volume),
////        text = "Play Text",
////        onClick = { tts.speak(message.content, TextToSpeech.QUEUE_FLUSH, null, null) }
////    ),
//    if (showRegenerateIcon) {
//        result.add(
//            MenuAction(
//                icon = painterResource(id = R.drawable.ic_refresh),
//                text = "Regenerate",
//                onClick = { onRegenerate(message) }
//            ))
//    }
//    return result
//}