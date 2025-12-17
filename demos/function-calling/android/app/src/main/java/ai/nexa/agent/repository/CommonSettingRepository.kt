package ai.nexa.agent.repository

import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.booleanPreferencesKey
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.intPreferencesKey
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map

class CommonSettingRepository(
    private val dataStore: DataStore<Preferences>
) {
    companion object {
        private val THEME_KEY = stringPreferencesKey("theme")
        private val IS_THINKING_KEY = booleanPreferencesKey("is_thinking")
        private val AUTO_OFFLOAD_KEY = booleanPreferencesKey("auto_offload")
        private val N_CTX_KEY = intPreferencesKey("n_ctx")
    }

    val themeFlow: Flow<String> =
        dataStore.data.map { it[THEME_KEY] ?: "auto" }

    val ctxFlow: Flow<Int> =
        dataStore.data.map { it[N_CTX_KEY] ?: 1024 }

    suspend fun setTheme(theme: String) {
        dataStore.edit { it[THEME_KEY] = theme }
    }

    suspend fun setIsThinking(value: Boolean) {
        dataStore.edit { it[IS_THINKING_KEY] = value }
    }
    suspend fun getIsThinking(): Boolean {
        return dataStore.data.first()[IS_THINKING_KEY] != false
    }

    suspend fun setAutoOffLoad(value: Boolean) {
        dataStore.edit { it[AUTO_OFFLOAD_KEY] = value }
    }
    suspend fun getAutoOffLoad(): Boolean {
        return dataStore.data.first()[AUTO_OFFLOAD_KEY] != false
    }

    suspend fun setNCtx(value: Int) {
        dataStore.edit { it[N_CTX_KEY] = value }
    }
    suspend fun getNCtx(): Int {
        return dataStore.data.first()[N_CTX_KEY] ?: 1024
    }
}