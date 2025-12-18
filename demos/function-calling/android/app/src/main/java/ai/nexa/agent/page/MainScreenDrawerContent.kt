package ai.nexa.agent.page

import ai.nexa.agent.R
import ai.nexa.agent.ui.theme.dividerColor
import ai.nexa.agent.ui.theme.drawerBackground
import ai.nexa.agent.ui.theme.drawerIcon
import ai.nexa.agent.ui.theme.ic_discord
import ai.nexa.agent.ui.theme.ic_mail
import ai.nexa.agent.ui.theme.ic_slack
import ai.nexa.agent.ui.theme.outlinedButtonBorder
import ai.nexa.agent.ui.theme.textPrimary
import ai.nexa.agent.ui.theme.textSecondary
import ai.nexa.agent.util.L
import android.content.Intent
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.core.net.toUri

@Composable
fun MainScreenDrawerContent() {
    Column(modifier = Modifier
        .fillMaxWidth()
        .background(Color.LightGray)) {

    }
}

@Composable
fun DrawerContent(
    sessions: Map<String, List<Any>>,
    currentSessionId: String?,
    longPressSessionId: String?,
    onLongPressSession: (String) -> Unit,
    onClearLongPress: () -> Unit,
    onSessionClick: (String) -> Unit,
    onDeleteSessionClick: (String) -> Unit,
    onEditSessionClick: (String) -> Unit,
    onNewChat: () -> Unit = {},
    onModelsClick: () -> Unit = {},
    onSettingsClick: () -> Unit = {},
    onAboutClick: () -> Unit = {}
) {
    val context = LocalContext.current
    Column(
        modifier = Modifier
            .background(
                color = MaterialTheme.colorScheme.drawerBackground,
                shape = RoundedCornerShape(topEnd = 28.dp, bottomEnd = 28.dp)
            )
            .fillMaxHeight()
            .width(320.dp)
            .padding(top = 50.dp)
    ) {
        OutlinedButton(
            onClick = onNewChat,
            modifier = Modifier
                .fillMaxWidth()
                .height(38.dp)
                .padding(horizontal = 24.dp),
            shape = RoundedCornerShape(28.dp),
            border = BorderStroke(1.dp, MaterialTheme.colorScheme.outlinedButtonBorder)
        ) {
            Text("New Chat", fontSize = 18.sp, color = MaterialTheme.colorScheme.textPrimary)
        }
        Spacer(
            Modifier
                .height(20.dp)
                .padding(start = 16.dp, end = 16.dp)
        )

        DrawerMenuItem(R.drawable.ic_grid, "Models", onModelsClick)
        DrawerMenuItem(R.drawable.ic_settings, "Settings", onSettingsClick)
        DrawerMenuItem(R.drawable.ic_info, "About", onAboutClick)

        Spacer(Modifier.height(12.dp))
        HorizontalDivider(
            modifier = Modifier.padding(vertical = 6.dp, horizontal = 16.dp),
            thickness = 1.dp,
            color = MaterialTheme.colorScheme.dividerColor
        )
        Box(modifier = Modifier.weight(1f)) {
            SessionListGrouped(
                groupedSessions = sessions,
                currentSessionId = currentSessionId,
                longPressSessionId = longPressSessionId,
                onLongPressSession = onLongPressSession,
                onClearLongPress = onClearLongPress,
                onSessionClick = onSessionClick,
                onEditSessionClick = onEditSessionClick,
                onDeleteSessionClick = onDeleteSessionClick
            )
        }
        Row(
            modifier = Modifier
                .height(90.dp)
                .fillMaxWidth()
                .padding(24.dp),
            horizontalArrangement = Arrangement.End
        ) {
            Spacer(Modifier.weight(1f))
            IconButton(onClick = {
                val intent =
                    Intent(Intent.ACTION_VIEW, "https://discord.com/invite/nexa-ai".toUri())
                context.startActivity(intent)
            }, Modifier.size(20.dp)) {
                Icon(
                    painterResource(MaterialTheme.colorScheme.ic_discord),
                    contentDescription = "sidcord",
                    tint = Color.Unspecified
                )
            }
            Spacer(Modifier.width(12.dp))
            IconButton(onClick = {
                val intent = Intent(
                    Intent.ACTION_VIEW,
                    "https://github.com/NexaAI/nexa-sdk".toUri()
                )
                context.startActivity(intent)
            }, Modifier.size(20.dp)) {
                Icon(
                    painterResource(MaterialTheme.colorScheme.ic_mail),
                    contentDescription = "git",
                    tint = Color.Unspecified
                )

            }
            Spacer(Modifier.width(12.dp))
            IconButton(onClick = {
                val intent = Intent(
                    Intent.ACTION_VIEW,
                    "https://nexa-ai-community.slack.com/ssb/redirect".toUri()
                )
                context.startActivity(intent)
            }, Modifier.size(20.dp)) {
                Icon(
                    painterResource(MaterialTheme.colorScheme.ic_slack),
                    contentDescription = "slack",
                    tint = Color.Unspecified
                )

            }
        }
    }
}

@Composable
fun SessionListGrouped(
    groupedSessions: Map<String, List<Any>>,
    currentSessionId: String?,
    longPressSessionId: String?,
    onLongPressSession: (String) -> Unit,
    onClearLongPress: () -> Unit,
    onSessionClick: (String) -> Unit,
    onEditSessionClick: (String) -> Unit,
    onDeleteSessionClick: (String) -> Unit
) {
    LazyColumn(
        modifier = Modifier
            .fillMaxSize()
//                .verticalScroll(rememberScrollState())
    ) {
        L.d("DrawerContent", "update chat history list")
        groupedSessions.forEach { (groupTitle, sessions) ->
            stickyHeader {
                Spacer(
                    Modifier
                        .height(18.dp)
                        .background(MaterialTheme.colorScheme.drawerBackground)
                        .fillMaxWidth()
                )
                Text(
                    text = groupTitle,
                    color = MaterialTheme.colorScheme.textSecondary,
                    fontSize = 16.sp,
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(start = 24.dp)
                        .height(20.dp)
                        .background(MaterialTheme.colorScheme.drawerBackground)
                )
                Spacer(
                    Modifier
                        .height(8.dp)
                        .background(MaterialTheme.colorScheme.drawerBackground)
                        .fillMaxWidth()
                )
            }
        }
    }

}

@Composable
fun DrawerMenuItem(iconRes: Int, text: String, onClick: () -> Unit) {
    Row(
        Modifier
            .fillMaxWidth()
            .height(50.dp)
            .padding(top = 6.dp, start = 24.dp, end = 24.dp)
            .clickable { onClick() },
        verticalAlignment = Alignment.CenterVertically
    ) {
        Icon(
            painter = painterResource(id = iconRes),
            contentDescription = text,
            tint = MaterialTheme.colorScheme.drawerIcon,
            modifier = Modifier.size(26.dp)
        )
        Spacer(Modifier.width(12.dp))
        Text(
            text,
            fontSize = 16.sp,
            color = MaterialTheme.colorScheme.textPrimary
        )
    }
}