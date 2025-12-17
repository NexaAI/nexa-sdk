package ai.nexa.agent.util

import android.content.Context
import android.database.Cursor
import android.net.Uri
import android.provider.OpenableColumns
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.preferencesDataStore
import kotlin.random.Random

val Context.dataStore: DataStore<Preferences> by preferencesDataStore(name = "app_config")
fun ClosedFloatingPointRange<Float>.random() =
    Random.nextFloat() * (endInclusive - start) + start

class AppUtils {

    companion object {
        private const val TAG = "AppUtils"
        fun getFileName(context: Context, uri: Uri, defaultExt: String = "bin"): String {
            var result: String? = null
            if (uri.scheme == "content") {
                val cursor: Cursor? = context.contentResolver.query(uri, null, null, null, null)
                cursor?.use {
                    if (it.moveToFirst()) {
                        val index = it.getColumnIndex(OpenableColumns.DISPLAY_NAME)
                        if (index != -1) {
                            result = it.getString(index)
                        }
                    }
                }
            }
            if (result == null) {
                result = uri.path
                val cut = result?.lastIndexOf('/') ?: -1
                if (cut != -1) {
                    result = result?.substring(cut + 1)
                }
            }

            val ext = result?.substringAfterLast('.', "")?.takeIf { it.isNotEmpty() } ?: defaultExt
            val base = result?.substringBeforeLast('.', result ?: "file")
            val cleanName = (base ?: "file").replace(Regex("[\\s()]+"), "_")
            return "$cleanName.$ext"
        }

    }

}