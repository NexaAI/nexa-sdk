package ai.nexa.agent.util

import android.content.Context
import android.content.Intent
import androidx.core.content.FileProvider.getUriForFile
import java.io.File

class ShareUtil {
    companion object {

        fun shareImg(context: Context, file: File) {
            val shareIntent = Intent(Intent.ACTION_SEND)
            val uri = getUriForFile(context, "${context.packageName}.fileprovider", file)
            shareIntent.addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION)
            shareIntent.putExtra(Intent.EXTRA_STREAM, uri)
            shareIntent.setType("image/*")
            context.startActivity(Intent.createChooser(shareIntent, "发给某人"))
        }

        fun shareImg(context: Context, imagePath: String) {
            shareImg(context, File(imagePath))
        }
    }
}