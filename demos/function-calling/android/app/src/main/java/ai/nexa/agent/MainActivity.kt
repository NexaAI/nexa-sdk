package ai.nexa.agent

import ai.nexa.agent.activity.BaseComposeActivity
import ai.nexa.agent.navigation.AppNavGraph
import android.os.Bundle
import androidx.activity.enableEdgeToEdge
import androidx.compose.runtime.Composable

class MainActivity : BaseComposeActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
    }

    @Composable
    override fun BaseActivityContent() {
        super.BaseActivityContent()
        AppRoot { AppNavGraph() }
    }
}