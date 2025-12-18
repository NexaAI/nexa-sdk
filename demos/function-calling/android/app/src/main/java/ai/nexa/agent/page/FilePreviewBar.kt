package ai.nexa.agent.page

import ai.nexa.agent.R
import ai.nexa.agent.ui.theme.chatMessageAudioBg
import ai.nexa.agent.ui.theme.chatMessageAudioBorder
import ai.nexa.agent.ui.theme.chatMessageAudioFilename
import ai.nexa.agent.ui.theme.chatMessageAudioIcon
import ai.nexa.agent.ui.theme.chatMessageAudioTime
import ai.nexa.agent.ui.theme.ic_delete
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import coil3.compose.rememberAsyncImagePainter
import java.io.File

@Composable
fun FilePreviewBar(
    images: List<File>,
    onRemoveImage: (File) -> Unit,
    audios: List<File>,
    onRemoveAudio: (File) -> Unit
) {


    Row(
        modifier = Modifier
            .padding(bottom = 8.dp)
            .horizontalScroll(rememberScrollState()),
        horizontalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        images.forEach { file ->
            ImagePreviewItem(
                file = file,
                onRemove = onRemoveImage
            )
        }
        audios.forEach { file ->
            AudioFilePreviewItem(
                file = file,
                onRemove = onRemoveAudio
            )
        }
    }
}

@Composable
fun ImagePreviewItem(
    file: File,
    onRemove: (File) -> Unit
) {
    Box(
        modifier = Modifier.size(120.dp)
    ) {
        Image(
            painter = rememberAsyncImagePainter(file),
            contentDescription = null,
            contentScale = ContentScale.Crop,
            modifier = Modifier
                .fillMaxSize()
                .clip(RoundedCornerShape(16.dp))
        )
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(4.dp)
        ) {
            RemoveButton(
                onClick = { onRemove(file) },
                modifier = Modifier.align(Alignment.TopEnd)
            )
        }
    }
}

@Composable
fun AudioFilePreviewItem(
    file: File,
    onRemove: (File) -> Unit
) {
    Box(
        modifier = Modifier
            .height(60.dp)
            .width(200.dp)
            .clip(RoundedCornerShape(16.dp))
            .background(MaterialTheme.colorScheme.chatMessageAudioBg)
            .border(
                1.dp,
                MaterialTheme.colorScheme.chatMessageAudioBorder,
                RoundedCornerShape(16.dp)
            )
    ) {
        Row(
            modifier = Modifier
                .fillMaxSize()
                .padding(start = 12.dp, end = 32.dp, top = 6.dp, bottom = 6.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Icon(
                painter = painterResource(R.drawable.ic_audio),
                contentDescription = null,
                tint = MaterialTheme.colorScheme.chatMessageAudioIcon,
                modifier = Modifier.size(32.dp)
            )
            Spacer(Modifier.width(10.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = file.name,
                    color = MaterialTheme.colorScheme.chatMessageAudioFilename,
                    fontSize = 14.sp,
                    maxLines = 1
                )
                Text(
                    text = "${file.extension.uppercase()}  ${fileSizeString(file.length())}",
                    color = MaterialTheme.colorScheme.chatMessageAudioTime,
                    fontSize = 12.sp
                )
            }
        }
        RemoveButton(
            onClick = { onRemove(file) },
            modifier = Modifier
                .align(Alignment.TopEnd)
                .padding(2.dp)
        )
    }
}

@Composable
fun RemoveButton(
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    IconButton(
        onClick = onClick,
        modifier = modifier.size(20.dp)
    ) {
        Icon(
            painter = painterResource(MaterialTheme.colorScheme.ic_delete),
            contentDescription = "Remove",
            tint = Color.Unspecified,
            modifier = Modifier.size(20.dp)
        )
    }
}

fun fileSizeString(size: Long): String {
    if (size <= 0) return "0KB"
    val kb = size / 1024.0
    return if (kb < 1024) String.format("%.0fKB", kb)
    else String.format("%.2fMB", kb / 1024)
}