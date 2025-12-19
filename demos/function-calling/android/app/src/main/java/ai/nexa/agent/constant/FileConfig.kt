package ai.nexa.agent.constant

import android.content.Context
import java.io.File

class FileConfig {
    companion object {
        val modelsDir = fun(context: Context): File {
            return File(context.filesDir, "models").apply { if (!exists()) mkdirs() }
        }
        val chatParentFile = fun(context: Context): File {
            return File(context.filesDir, "chat").apply { if (!exists()) mkdirs() }
        }
    }
}