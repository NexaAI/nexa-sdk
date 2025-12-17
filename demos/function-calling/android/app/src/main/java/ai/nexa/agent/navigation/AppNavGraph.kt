package ai.nexa.agent.navigation

import ai.nexa.agent.constant.SharePreferenceKeys
import ai.nexa.agent.model.ChatViewModel
import ai.nexa.agent.page.MainScreen
import ai.nexa.agent.util.L
import android.content.Context
import androidx.compose.animation.EnterTransition
import androidx.compose.animation.ExperimentalAnimationApi
import androidx.compose.animation.ExperimentalSharedTransitionApi
import androidx.compose.animation.SharedTransitionLayout
import androidx.compose.animation.core.FastOutSlowInEasing
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeOut
import androidx.compose.animation.slideInHorizontally
import androidx.compose.animation.slideOutHorizontally
import androidx.compose.runtime.Composable
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.IntOffset
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import org.koin.androidx.compose.koinViewModel

const val SPLASH_ANIMATE_DURATION = 450

@OptIn(ExperimentalSharedTransitionApi::class, ExperimentalAnimationApi::class)
@Composable
fun AppNavGraph() {
    val navController = rememberNavController()
    val viewModel: ChatViewModel = koinViewModel()
    val slideSpec = tween<IntOffset>(
        durationMillis = 600,
        easing = FastOutSlowInEasing
    )

    SharedTransitionLayout {
        NavHost(
            navController,
            startDestination = "main",
            enterTransition = {
                if (initialState.destination.route == "main") {
                    EnterTransition.None
                } else {
                    slideInHorizontally(
                        initialOffsetX = { it },
                        animationSpec = slideSpec
                    )
                }
            },
            exitTransition = {
                if (initialState.destination.route == "main") {
                    fadeOut(tween(SPLASH_ANIMATE_DURATION))
                } else {
                    slideOutHorizontally(
                        targetOffsetX = { -it },
                        animationSpec = slideSpec
                    )
                }
            },
            popEnterTransition = {
                if (targetState.destination.route == "main") {
                    EnterTransition.None
                } else {
                    slideInHorizontally(
                        initialOffsetX = { -it },
                        animationSpec = slideSpec
                    )
                }
            },
            popExitTransition = {
                if (initialState.destination.route == "main") {
                    fadeOut(tween(SPLASH_ANIMATE_DURATION))
                } else {
                    slideOutHorizontally(
                        targetOffsetX = { it },
                        animationSpec = slideSpec
                    )
                }
            }
        )
        {

            composable("splash") {
//                SplashScreen(sharedTransitionScope = this@SharedTransitionLayout,
//                    animatedVisibilityScope = this,
//                    onNavigate = { route ->
//                        navController.navigate(route) { popUpTo("splash") { inclusive = true } }
//                    })
            }
            composable(
                route = "main?model_id={model_id}",
                arguments = listOf(
                    navArgument("model_id") {
                        type = NavType.StringType
                        nullable = true
                        defaultValue = null
                    }
                )
            ) { backStackEntry ->
                var isNavigatorPopReload = false
                val sp = LocalContext.current.getSharedPreferences(
                    SharePreferenceKeys.FileName.CommonConfig.fileName,
                    Context.MODE_PRIVATE
                )
                val prepareLoadModelId =
                    sp.getString(SharePreferenceKeys.KEY_PREPARE_LOAD_MODEL_ID, null)
                val modelId =
                    prepareLoadModelId ?: backStackEntry.arguments?.getString("model_id")
                prepareLoadModelId?.let {
                    isNavigatorPopReload = true
//                    sp.edit(true) {
//                        putString(SharePreferenceKeys.KEY_PREPARE_LOAD_MODEL_ID, null)
//                    }
                }
                L.d("nfl", "AppNavGraph Model ID: $modelId")
                MainScreen(navController = navController,
                    viewModel = viewModel,
                    sharedTransitionScope = this@SharedTransitionLayout,
                    animatedVisibilityScope = this)
            }
        }
    }
}