package ai.nexa.agent.page

import ai.nexa.agent.model.ChatViewModel
import ai.nexa.agent.ui.composable.SetStatusBarColor
import ai.nexa.agent.ui.theme.LocalAppDarkTheme
import ai.nexa.agent.ui.theme.chatStatusBarBg
import androidx.compose.animation.AnimatedContentScope
import androidx.compose.animation.ExperimentalSharedTransitionApi
import androidx.compose.animation.SharedTransitionScope
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.material3.DrawerValue
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.ModalNavigationDrawer
import androidx.compose.material3.Text
import androidx.compose.material3.rememberDrawerState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Modifier
import androidx.navigation.NavController

@OptIn(ExperimentalSharedTransitionApi::class)
@Composable
fun MainScreen(
    navController: NavController,
    viewModel: ChatViewModel,
    sharedTransitionScope: SharedTransitionScope,
    animatedVisibilityScope: AnimatedContentScope
) {

    val statusBarColor = MaterialTheme.colorScheme.chatStatusBarBg
    val drawerState = rememberDrawerState(DrawerValue.Closed)
    val scope = rememberCoroutineScope()

    SetStatusBarColor(
        color = statusBarColor,
        darkIcons = !LocalAppDarkTheme.current
    )

    ModalNavigationDrawer(
        modifier = Modifier.navigationBarsPadding(),
        drawerState = drawerState,
        drawerContent = {
            MainScreenDrawerContent()
        }
    ) {
        ChatPage(
            navController = navController,
//            modelId = modelId,
            viewModel = viewModel,
            openDrawer = {
//                scope.launch {
//                    viewModel.onIntent(ChatIntent.UpdateSessions)
//                    drawerState.open()
//                }
//                keyboardController?.hide()
            },
//            isNavigatorPopReload = isNavigatorPopReload,
            sharedTransitionScope = sharedTransitionScope,
            animatedVisibilityScope = animatedVisibilityScope
        )
    }
}
