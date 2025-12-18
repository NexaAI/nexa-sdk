package ai.nexa.agent.model

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.receiveAsFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

abstract class BaseViewModel<S, I, E>(initialState: S) : ViewModel() {
    protected val _state = MutableStateFlow(initialState)
    val state: StateFlow<S> = _state.asStateFlow()

    private val _event = Channel<E>()
    val event = _event.receiveAsFlow()

    abstract fun onIntent(intent: I)
    protected fun updateState(reduce: S.() -> S) {
        _state.update(reduce)
    }

    protected fun sendEvent(e: E) {
        viewModelScope.launch { _event.send(e) }
    }
}