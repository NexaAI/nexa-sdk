package ai.nexa.agent.model

import ai.nexa.agent.repository.CommonSettingRepository
import ai.nexa.agent.state.CommonSettingEvent
import ai.nexa.agent.state.CommonSettingIntent
import ai.nexa.agent.state.CommonSettingState
import ai.nexa.agent.util.L
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.launch


class CommonSettingViewModel(
    private val repo: CommonSettingRepository
) : BaseViewModel<CommonSettingState, CommonSettingIntent, CommonSettingEvent>(CommonSettingState()) {

    init {
        onIntent(CommonSettingIntent.LoadTheme)
        initView()
    }

    override fun onIntent(intent: CommonSettingIntent) {
        when (intent) {
            is CommonSettingIntent.LoadTheme -> handleLoadTheme()
            is CommonSettingIntent.SetTheme -> handleSetTheme(intent.theme)
            is CommonSettingIntent.SetIsThinking -> handleSetIsThinking(intent.value)
            is CommonSettingIntent.SetAutoOffLoad -> handleSetAutoOffLoad(intent.value)
            is CommonSettingIntent.SetNCtx -> handleSetNCtx(intent.value)
            CommonSettingIntent.RestCommonSetting -> handleResetCommonSetting()
        }
    }

// region -- intent handlers --

    private fun handleLoadTheme() {
        viewModelScope.launch {
            repo.themeFlow.distinctUntilChanged().collect { theme ->
                updateState { copy(theme = theme, loading = false) }
            }
        }
    }

    private fun handleSetTheme(theme: String) {
        viewModelScope.launch {
            try {
                repo.setTheme(theme)
                L.e("peter", "theme:$theme")
            } catch (e: Exception) {
                sendEvent(CommonSettingEvent.Error)
                updateState { copy(loading = false) }
            }
        }
    }

    private fun handleSetIsThinking(value: Boolean) {
        viewModelScope.launch {
            repo.setIsThinking(value)
            updateState { copy(isThinking = value, loading = false) }
        }
    }

    private fun handleSetAutoOffLoad(value: Boolean) {
        viewModelScope.launch {
            repo.setAutoOffLoad(value)
            updateState { copy(autoOffload = value, loading = false) }
        }
    }

    private fun handleSetNCtx(value: Int) {
        viewModelScope.launch {

            if (value != state.value.nCtx) {
                repo.setNCtx(value)
                updateState { copy(nCtx = value, loading = false) }
            }
        }
    }

    private fun handleResetCommonSetting() {
        viewModelScope.launch {
            repo.setTheme("auto")
            repo.setIsThinking(true)
            repo.setAutoOffLoad(true)
            repo.setNCtx(1024)
            updateState {
                copy(
                    theme = "auto",
                    isThinking = true,
                    autoOffload = true,
                    nCtx = 1024,
                    loading = false
                )
            }
        }
    }

    // endregion
    private fun initView() {
        viewModelScope.launch {
            val think = repo.getIsThinking()
            val auto = repo.getAutoOffLoad()
            val ctx = repo.getNCtx()
            updateState {
                copy(
                    isThinking = think,
                    autoOffload = auto,
                    nCtx = ctx
                )
            }
        }
    }

    fun getIsThinking(onResult: (Boolean) -> Unit) {
        viewModelScope.launch { onResult(repo.getIsThinking()) }
    }

    fun getAutoOffLoad(onResult: (Boolean) -> Unit) {
        viewModelScope.launch { onResult(repo.getAutoOffLoad()) }
    }

    fun getNCtx(onResult: (Int) -> Unit) {
        viewModelScope.launch { onResult(repo.getNCtx()) }
    }
}