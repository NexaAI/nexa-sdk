package com.nexa.android.demo.ui

import androidx.compose.animation.core.animateDpAsState
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.gestures.AnchoredDraggableDefaults
import androidx.compose.foundation.gestures.AnchoredDraggableState
import androidx.compose.foundation.gestures.DraggableAnchors
import androidx.compose.foundation.gestures.Orientation
import androidx.compose.foundation.gestures.anchoredDraggable
import androidx.compose.foundation.gestures.animateTo
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.unit.*
import kotlinx.coroutines.launch
import kotlin.math.roundToInt

enum class PushDrawerValue { Open, Closed }

@Composable
fun rememberPushDrawerState(
    drawerWidth: Dp = 280.dp
): AnchoredDraggableState<PushDrawerValue> {
    val density = LocalDensity.current
    val drawerWidthPx = with(density) { drawerWidth.toPx() }

    return remember {
        AnchoredDraggableState(
            initialValue = PushDrawerValue.Closed,
            anchors = DraggableAnchors {
                PushDrawerValue.Closed at 0f
                PushDrawerValue.Open at drawerWidthPx
            }
        )
    }
}

@Composable
fun PushDrawer(
    drawerState: AnchoredDraggableState<PushDrawerValue>,
    modifier: Modifier = Modifier,
    drawerWidth: Dp = 280.dp,
    drawerContent: @Composable ColumnScope.() -> Unit,
    content: @Composable () -> Unit
) {
    val scope = rememberCoroutineScope()

    // 动画控制缩放和圆角
    val scale by animateFloatAsState(
        targetValue = if (drawerState.currentValue == PushDrawerValue.Open) 0.9f else 1f,
        label = "drawerScale"
    )
    val cornerRadius by animateDpAsState(
        targetValue = if (drawerState.currentValue == PushDrawerValue.Open) 24.dp else 0.dp,
        label = "drawerCorners"
    )

    Box(
        modifier = modifier
            .fillMaxSize()
            .anchoredDraggable(
                state = drawerState,
                orientation = Orientation.Horizontal,
                flingBehavior = AnchoredDraggableDefaults.flingBehavior(drawerState)
            )
    ) {
        // 遮罩层（盖住内容，但不盖 Drawer）
        if (drawerState.currentValue == PushDrawerValue.Open || drawerState.offset > 0f) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(Color.Black.copy(alpha = 0.3f))
                    .pointerInput(Unit) {
                        detectTapGestures {
                            scope.launch { drawerState.animateTo(PushDrawerValue.Closed) }
                        }
                    }
            )
        }

        // Drawer 菜单（要画在遮罩上方，主内容下方）
        Column(
            modifier = Modifier
                .width(drawerWidth)
                .fillMaxHeight()
                .clickable {

                }
                .background(MaterialTheme.colorScheme.surfaceVariant)
        ) {
            drawerContent()
        }

        // 主内容（被推动）
        Surface(
            tonalElevation = if (drawerState.currentValue == PushDrawerValue.Open) 8.dp else 0.dp,
            shadowElevation = if (drawerState.currentValue == PushDrawerValue.Open) 8.dp else 0.dp,
            shape = RoundedCornerShape(cornerRadius),
            modifier = Modifier
                .fillMaxSize()
                .offset { IntOffset(drawerState.offset.roundToInt(), 0) }
                .graphicsLayer {
                    scaleX = scale
                    scaleY = scale
                }
        ) {
            content()
        }
    }

}

