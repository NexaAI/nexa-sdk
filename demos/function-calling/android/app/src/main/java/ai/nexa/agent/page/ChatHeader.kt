package ai.nexa.agent.page

import ai.nexa.agent.R
import ai.nexa.agent.ui.theme.ChatTitleBarIcon
import ai.nexa.agent.ui.theme.textPrimary
import ai.nexa.agent.util.L
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.wrapContentWidth
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

@Composable
fun ChatHeader(
//    currentModel: ModelData?,
//    modelList: List<ModelData>,
    onModelSelected: (String) -> Unit = {},
    onMenuClick: () -> Unit = {},
    onNewSession: () -> Unit = {},
    onMoreClick: () -> Unit
) {
    val iconVector = MaterialTheme.colorScheme.ChatTitleBarIcon
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .height(48.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        IconButton(
            onClick = onMenuClick,
            modifier = Modifier.size(48.dp)
        ) {
            Icon(
                painter = painterResource(id = R.drawable.menu),
                contentDescription = "Menu",
                modifier = Modifier.size(24.dp),
                tint = iconVector,
            )
        }

        // Dropdown
        Box(
            modifier = Modifier
                .weight(1f)
                .fillMaxHeight()
        ) {
            Row(
                verticalAlignment = Alignment.CenterVertically,
                modifier = Modifier
                    .fillMaxWidth()
                    .fillMaxHeight()
                    .wrapContentWidth(Alignment.CenterHorizontally)
            ) {

                ModelDropdownSelector(
//                    models = modelList,
//                    currentModel = currentModel,
                    onSelect = { it ->
                        L.d("nfl", "Model selected: $it")
                        onModelSelected(it)
                    },
                    onMoreClick = { onMoreClick() }
                )
            }

        }

        IconButton(
            onClick = onNewSession,
            modifier = Modifier.size(48.dp)
        ) {
            Icon(
                painter = painterResource(R.drawable.profile_icon),
                contentDescription = "Profile",
                modifier = Modifier.size(24.dp),
                tint = iconVector,
            )
        }
    }
}

@Composable
fun ModelDropdownSelector(
//    models: List<ModelData>,
//    currentModel: ModelData?,
    onSelect: (String) -> Unit,
    onMoreClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    var showMenu by remember { mutableStateOf(false) }
    val iconVector = MaterialTheme.colorScheme.ChatTitleBarIcon
    Box(modifier = modifier) {
        Row(
            modifier = Modifier
//                .clickable { showMenu = true }
                .wrapContentWidth(Alignment.CenterHorizontally)
                .padding(vertical = 8.dp, horizontal = 10.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text(
//                text = currentModel?.displayName ?: "Selected model",
                text = "Omni Neural",
                fontSize = 16.sp,
                color = MaterialTheme.colorScheme.textPrimary,
                fontWeight = FontWeight.Medium,
            )
//            Icon(
//                painter = painterResource(id = R.drawable.ic_chevron_down),
//                contentDescription = null,
//                tint = iconVector,
//                modifier = Modifier.size(22.dp)
//            )
        }

        if (showMenu) {
//            Popup(
//                alignment = Alignment.TopCenter,
//                offset = IntOffset(0, 120),
//                onDismissRequest = { showMenu = false }
//            ) {
//                Card(
//                    modifier = Modifier
//                        .width(270.dp)
//                        .shadow(3.dp, RoundedCornerShape(8.dp)),
//                    shape = RoundedCornerShape(3.dp)
//                ) {
//                    Column(
//                        background(MaterialTheme.colorScheme.popupBackground)
//                            .padding(top = 5.dp, bottom = 5.dp)
//                    ) {
//                        models.forEach { model ->
//                            Row(
//                                modifier = Modifier
//                                    .fillMaxWidth()
//                                    .clickable {
//                                        if (model.id != currentModel?.id) {
//                                            onSelect(model.id)
//                                            showMenu = false
//                                        }
//                                    }
//                                    .padding(horizontal = 10.dp, vertical = 5.dp),
//                                verticalAlignment = Alignment.CenterVertically
//                            ) {
//                                Icon(
//                                    painter = painterResource(
//                                        id = ModelCategory.getCategoryIcon(
//                                            fromType(model.type)
//                                        )
//                                    ),
//                                    contentDescription = null,
//                                    tint = Color.Unspecified,
//                                    modifier = Modifier.size(24.dp)
//                                )
//                                Spacer(Modifier.width(10.dp))
//                                Column(verticalArrangement = Arrangement.spacedBy(0.dp)) {
//                                    Text(
//                                        model.displayName,
//                                        fontSize = 12.sp,
//                                        color = MaterialTheme.colorScheme.dropdownLabel,
//                                        lineHeight = 12.sp
//                                    )
//                                    Text(
//                                        model.type,
//                                        color = MaterialTheme.colorScheme.dropdownSub,
//                                        fontSize = 12.sp,
//                                        lineHeight = 12.sp
//                                    )
//                                }
//                                Spacer(Modifier.weight(1f))
//                                if (model.id == currentModel?.id) {
//                                    Icon(
//                                        painter = painterResource(id = MaterialTheme.colorScheme.ic_check),
//                                        contentDescription = null,
//                                        tint = MaterialTheme.colorScheme.dropdownCheck,
//                                        modifier = Modifier.size(16.dp)
//                                    )
//                                }
//                            }
//                        }
//                        HorizontalDivider(
//                            modifier = Modifier
//                                .fillMaxWidth()
//                                .padding(horizontal = 10.dp, vertical = 8.dp),
//                            thickness = 0.5.dp,
//                            color = MaterialTheme.colorScheme.dropDivider
//                        )
//                        Row(
//                            Modifier
//                                .fillMaxWidth()
//                                .clickable { onMoreClick(); showMenu = false }
//                                .padding(horizontal = 10.dp, vertical = 5.dp),
//                            verticalAlignment = Alignment.CenterVertically,
//                            horizontalArrangement = Arrangement.Center
//                        ) {
//                            Text(
//                                "More Models",
//                                color = MaterialTheme.colorScheme.dropdownLabel,
//                                fontSize = 12.sp
//                            )
//                            Icon(
//                                painter = painterResource(id = R.drawable.ic_arrow_right),
//                                contentDescription = null,
//                                modifier = Modifier.size(16.dp),
//                                tint = MaterialTheme.colorScheme.dropdownLabel
//                            )
//                        }
//                    }
//                }
//            }
        }
    }
}
