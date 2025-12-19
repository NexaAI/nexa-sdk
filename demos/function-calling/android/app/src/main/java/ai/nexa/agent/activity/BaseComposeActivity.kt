package ai.nexa.agent.activity

import ai.nexa.agent.model.CommonSettingViewModel
import ai.nexa.agent.state.BannerState
import ai.nexa.agent.state.LocalBannerState
import ai.nexa.agent.ui.composable.TopBanner
import ai.nexa.agent.ui.theme.NexaAgentTheme
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.compose.ui.zIndex
import org.koin.androidx.viewmodel.ext.android.getViewModel

open class BaseComposeActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            NexaStudioApp()
        }
    }

    @Composable
    fun AppRoot(content: @Composable () -> Unit) {
        val bannerState = remember { BannerState() }

        CompositionLocalProvider(LocalBannerState provides bannerState) {
            Box {
                content()
                if (bannerState.visible && bannerState.message != null) {
                    TopBanner(
                        message = bannerState.message!!,
                        modifier = Modifier
                            .align(
                                if (bannerState.alignBottom) {
                                    Alignment.BottomCenter
                                } else {
                                    Alignment.TopCenter
                                }
                            )
                            .fillMaxWidth()
                            .statusBarsPadding()
                            .navigationBarsPadding()
                            .padding(
                                top = if (bannerState.alignBottom) {
                                    0.dp
                                } else {
                                    53.dp
                                },
                                bottom = if (bannerState.alignBottom) {
                                    100.dp
                                } else {
                                    0.dp
                                }
                            )
                            .zIndex(2f),
                        onDismiss = { bannerState.dismiss() }
                    )
                }
            }
        }
    }

    @Composable
    fun NexaStudioApp() {
        val viewModel: CommonSettingViewModel = getViewModel()
        val state by viewModel.state.collectAsState()
        val isDark = when (state.theme) {
            "dark" -> true
            "light" -> false
            else -> isSystemInDarkTheme()
        }
        NexaAgentTheme(darkTheme = isDark) {
            BaseActivityContent()
        }
    }

    @Composable
    open fun BaseActivityContent() {

    }
}