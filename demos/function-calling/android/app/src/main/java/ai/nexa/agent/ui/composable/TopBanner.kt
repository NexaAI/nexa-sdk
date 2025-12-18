package ai.nexa.agent.ui.composable

import ai.nexa.agent.ui.theme.errorBannerBg
import ai.nexa.agent.ui.theme.errorBannerBorder
import ai.nexa.agent.ui.theme.errorBannerIcon
import ai.nexa.agent.ui.theme.errorBannerText
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.defaultMinSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

@Composable
fun TopBanner(
    message: String,
    modifier: Modifier = Modifier,
    onDismiss: (() -> Unit)? = null,
) {
    Row(
        modifier = modifier
            .defaultMinSize(minHeight = 36.dp)
            .padding(horizontal = 16.dp)
            .background(color = MaterialTheme.colorScheme.errorBannerBg, RoundedCornerShape(24.dp))
            .border(
                width = 1.dp,
                color = MaterialTheme.colorScheme.errorBannerBorder,
                shape = RoundedCornerShape(20.dp)
            ),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Icon(
            painterResource(MaterialTheme.colorScheme.errorBannerIcon),
            contentDescription = null,
            tint = Color.Unspecified,
            modifier = Modifier
                .padding(start = 8.dp)
                .size(16.dp)
        )
        Spacer(modifier = Modifier.width(6.dp))
        Text(
            message,
            color = MaterialTheme.colorScheme.errorBannerText,
            fontSize = 14.sp,
            modifier = Modifier.weight(1f)
        )
    }
}