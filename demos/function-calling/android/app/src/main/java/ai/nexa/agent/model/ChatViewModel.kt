package ai.nexa.agent.model


import ai.nexa.agent.bean.ChatMessageType
import ai.nexa.agent.bean.ChatRole
import ai.nexa.agent.bean.ChatUiMessage
import ai.nexa.agent.bean.GoogleCalendarEvent
import ai.nexa.agent.bean.GoogleRequestData
import ai.nexa.agent.bean.GoogleResponseBean
import ai.nexa.agent.constant.SharePreferenceKeys
import ai.nexa.agent.retrofit.RetrofitClient
import ai.nexa.agent.state.ChatIntent
import ai.nexa.agent.state.ChatSingleEvent
import ai.nexa.agent.state.ChatUiState
import ai.nexa.agent.state.OperationState
import ai.nexa.agent.util.ImageUtils
import ai.nexa.agent.util.L
import android.content.Context
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.serialization.json.Json

class ChatViewModel(
    private val appContext: Context,
) : BaseViewModel<ChatUiState, ChatIntent, ChatSingleEvent>(ChatUiState()) {
    private val chatViewModelSp = appContext.getSharedPreferences(
        SharePreferenceKeys.FileName.CommonConfig.fileName,
        Context.MODE_PRIVATE
    )
    private val ioScope = CoroutineScope(Dispatchers.IO)

    override fun onIntent(intent: ChatIntent) {
        when (intent) {
            is ChatIntent.InputChanged -> handleInputChanged(intent)
            is ChatIntent.SendMessage -> handleSendStreamMessage(intent)
            is ChatIntent.StopWaitingResponse -> {
                updateState { copy(input = "", operationState = OperationState.DEFAULT) }
            }

            is ChatIntent.AddImage -> {
                updateState { copy(selectedImageFiles = intent.files) }
            }

            is ChatIntent.NewSession -> {
                updateState {
                    copy(
                        resetSession = !state.value.resetSession,
                        input = "",
                        operationState = OperationState.DEFAULT,
                        chatMsgList = state.value.chatMsgList.apply {
                            this.clear()
                        },
                        selectedImageFiles = emptyList(),
                        selectedAudioFiles = emptyList()
                    )
                }
            }

            is ChatIntent.ChangeServerIPAddress -> {
                updateState { copy(serverIpAddress = intent.ipAddress) }
            }

            is ChatIntent.RemoveImage -> {
                updateState { copy(selectedImageFiles = state.value.selectedImageFiles.filter { it.absolutePath != intent.file.absolutePath }) }
            }
        }
    }

    private fun handleInputChanged(intent: ChatIntent.InputChanged) {
        updateState { copy(input = intent.text) }
    }

    private fun handleSendStreamMessage(intent: ChatIntent.SendMessage) {
        updateState { copy(input = "", operationState = OperationState.WAITING_SEND_RESPONSE) }
        var imageBase64: String? = null
        ioScope.launch {
            updateState {
                copy(chatMsgList = state.value.chatMsgList.apply {
                    state.value.selectedImageFiles.let { images ->
                        if (!images.isEmpty()) {
                            imageBase64 = ImageUtils.fileToBase64(images[0])
                            this.add(
                                ChatUiMessage(
                                    role = ChatRole.USER.role,
                                    content = intent.text,
                                    type = ChatMessageType.USER_IMAGE,
                                    imageList = images.map {
                                        it.absolutePath
                                    }.toMutableList()
                                )
                            )
                        }
                    }

                    this.add(
                        ChatUiMessage(
                            role = ChatRole.USER.role,
                            content = intent.text,
                            imageList = state.value.selectedImageFiles.map {
                                it.absolutePath
                            }.toMutableList()
                        )
                    )
                }, selectedImageFiles = emptyList())
            }
            delay(3000)
            RetrofitClient.switchApiServer(state.value.serverIpAddress)
//            val responseBean = RetrofitClient.instance.getGoogleCalendarResult(GoogleRequestData(image = imageBase64))
            val testStr = "{" +
                    "  \"meta\": null," +
                    "  \"content\": [" +
                    "    {" +
                    "      \"type\": \"text\"," +
                    "      \"text\": \"{\\\"event\\\":{\\\"id\\\":\\\"a06q5sjcq1cp7ta5u765nt4krs\\\",\\\"summary\\\":\\\"The Voice of AGI\\\",\\\"description\\\":\\\"The voice interface from sci-fi's of old like the Hitchikeer's Guide to the Galaxy to Ironman, the Voice Interface lays out a futurre form of connection, command and interaction.\\\",\\\"location\\\":\\\"AGI House SF: 170 St. Germain Ave, San Francsoco CA 94114\\\",\\\"start\\\":{\\\"dateTime\\\":\\\"2025-12-20T09:00:00-11:00\\\",\\\"timeZone\\\":\\\"Pacific/Pago_Pago\\\"},\\\"end\\\":{\\\"dateTime\\\":\\\"2025-12-20T22:00:00-11:00\\\",\\\"timeZone\\\":\\\"Pacific/Pago_Pago\\\"},\\\"status\\\":\\\"confirmed\\\",\\\"htmlLink\\\":\\\"https://www.google.com/calendar/event?eid=YTA2cTVzamNxMWNwN3RhNXU3NjVudDRrcnMgeWFuZ3hpYW5kYTAwN0AxNjMuY29t\\\",\\\"created\\\":\\\"2025-12-16T13:50:01.000Z\\\",\\\"updated\\\":\\\"2025-12-16T13:50:01.724Z\\\",\\\"creator\\\":{\\\"email\\\":\\\"yangxianda007@163.com\\\",\\\"self\\\":true},\\\"organizer\\\":{\\\"email\\\":\\\"yangxianda007@163.com\\\",\\\"self\\\":true},\\\"iCalUID\\\":\\\"a06q5sjcq1cp7ta5u765nt4krs@google.com\\\",\\\"sequence\\\":0,\\\"reminders\\\":{\\\"useDefault\\\":true},\\\"eventType\\\":\\\"default\\\",\\\"calendarId\\\":\\\"primary\\\",\\\"accountId\\\":\\\"normal\\\"}}\",       " +
                    "      \"annotations\": null," +
                    "      \"meta\": null" +
                    "    }" +
                    "  ]," +
                    "  \"structuredContent\": null," +
                    "  \"isError\": false" +
                    "}"
            val responseBean = Json {
                ignoreUnknownKeys = true
            }.decodeFromString<GoogleResponseBean>(testStr.trim())
            val temp = Json {
                ignoreUnknownKeys = true
            }.decodeFromString<GoogleCalendarEvent>(responseBean.content!![0].text!!)
            L.d("nfl", "GoogleCalendarEvent bean:${temp.event?.start?.dateTime}")
            updateState {
                copy(
                    operationState = OperationState.DEFAULT,
                    chatMsgList = state.value.chatMsgList.apply {
                        this.add(
                            ChatUiMessage(
                                role = ChatRole.ASSISTANT.role,
                                content = "intent.text",
                                extMsg = temp
                            )
                        )
                    },
                    selectedImageFiles = emptyList()
                )
            }
        }
    }

    companion object {
        const val TAG = "ChatViewModel"
    }
}