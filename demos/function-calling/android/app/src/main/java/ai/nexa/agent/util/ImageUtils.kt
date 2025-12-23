package ai.nexa.agent.util

import android.util.Base64
import java.io.File
import java.io.FileInputStream
import java.io.FileOutputStream

class ImageUtils {
    companion object {
        fun saveBase64ToFile(base64String: String?, filePath: String): Boolean {
            return saveBase64ToFile(base64String, File(filePath))
        }

        fun saveBase64ToFile(base64String: String?, file: File): Boolean {
            return try {
                val pureBase64 = base64String?.substringAfter(",")
                val decodedBytes = Base64.decode(pureBase64, Base64.DEFAULT)
                FileOutputStream(file).use { fos ->
                    fos.write(decodedBytes)
                }
                true
            } catch (e: Exception) {
                e.printStackTrace()
                false
            }
        }

        fun fileToBase64(file: File): String {
            val bytes = FileInputStream(file).use { inputStream ->
                ByteArray(file.length().toInt()).also {
                    inputStream.read(it)
                }
            }
            return "data:image/${file.extension};base64," +
                    Base64.encodeToString(bytes, Base64.DEFAULT)
        }
    }
}

