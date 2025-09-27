package com.nexa.android.demo

import android.Manifest
import android.content.pm.PackageManager
import android.os.Bundle
import android.util.Log
import java.io.File
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.gestures.animateTo
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Menu
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.core.content.ContextCompat
import androidx.lifecycle.viewmodel.compose.viewModel
import com.nexa.android.demo.ui.ChatScreen
import com.nexa.android.demo.ui.ModelLoadingScreen
import com.nexa.android.demo.ui.NavigationDrawerContent
import com.nexa.android.demo.ui.PushDrawer
import com.nexa.android.demo.ui.PushDrawerValue
import com.nexa.android.demo.ui.VLMChatScreen
import com.nexa.android.demo.ui.rememberPushDrawerState
import com.nexa.android.demo.ui.theme.MyApplicationTheme
import com.nexa.android.demo.viewmodel.AIViewModel
import kotlinx.coroutines.launch

class MainActivity : ComponentActivity() {

    private val requestPermissionLauncher = registerForActivityResult(
        ActivityResultContracts.RequestMultiplePermissions()
    ) { permissions ->
        val allGranted = permissions.values.all { it }
        if (allGranted) {
            // Permissions granted, continue with app initialization
            Log.d("Permission", "All permissions granted")
            // Note: Model initialization will be handled by the LaunchedEffect in NexaAIDemo
        } else {
            // Handle permission denial
            Log.w("Permission", "Some permissions denied: $permissions")
        }
    }

    
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        
        // Initialize Nexa SDK environment variables
        initNexaSdk()
        
        // Request permissions
        requestPermissions()
        
        setContent {
            MyApplicationTheme {
                Scaffold(modifier = Modifier.fillMaxSize()) { innerPadding ->
                    NexaAIDemo(
                        modifier = Modifier.padding(innerPadding)
                    )
                }
            }
        }
    }
    
    private fun requestPermissions() {
        val permissions = arrayOf(
            Manifest.permission.READ_EXTERNAL_STORAGE,
            Manifest.permission.WRITE_EXTERNAL_STORAGE,
            Manifest.permission.CAMERA
        )
        
        val permissionsToRequest = permissions.filter {
            ContextCompat.checkSelfPermission(this, it) != PackageManager.PERMISSION_GRANTED
        }
        
        if (permissionsToRequest.isNotEmpty()) {
            requestPermissionLauncher.launch(permissionsToRequest.toTypedArray())
        }
    }
    
    private fun initNexaSdk() {
        try {
            val nativeLibPath = applicationInfo.nativeLibraryDir
            android.system.Os.setenv("ADSP_LIBRARY_PATH", nativeLibPath, true)
            android.system.Os.setenv("LD_LIBRARY_PATH", nativeLibPath, true)
            android.system.Os.setenv("NEXA_PLUGIN_PATH", nativeLibPath, true)
            Log.d("MainActivity", "Os native library path: ${android.system.Os.getenv("ADSP_LIBRARY_PATH")}")
            
            val libDir = File(nativeLibPath)
            if (libDir.exists() && libDir.isDirectory) {
                val files = libDir.listFiles()
                Log.d("MainActivity", "files: ${files?.size}")
                files?.forEach { file ->
                    Log.d("MainActivity", "lib name: ${file.name}, Size: ${file.length()} bytes")
                }
            } else {
                Log.e("MainActivity", "none dir: $nativeLibPath")
            }
        } catch (e: Exception) {
            Log.e("MainActivity", "Failed to initialize Nexa SDK environment", e)
        }
    }
}


@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun NexaAIDemo(
    modifier: Modifier = Modifier,
    viewModel: AIViewModel = viewModel()
) {
    val context = LocalContext.current
    
    // Initialize ViewModel with context for state persistence
    LaunchedEffect(Unit) {
        viewModel.initialize(context)
    }
    
    val drawerWidth = 280.dp
    val drawerState = rememberPushDrawerState(drawerWidth)
    val scope = rememberCoroutineScope()

    var selectedScreen by remember { mutableStateOf("Model Loading") }
    var inputText by remember { mutableStateOf("") }

            // Get states
            val chatMessages = viewModel.chatMessages.collectAsState().value
            val vlmChatMessages = viewModel.vlmChatMessages.collectAsState().value
            val llmCurrentText = viewModel.llmCurrentText.collectAsState().value
            val vlmCurrentText = viewModel.vlmCurrentText.collectAsState().value
            val llmGenerationState = viewModel.llmGenerationState.collectAsState().value
            val vlmGenerationState = viewModel.vlmGenerationState.collectAsState().value
            val llmState = viewModel.llmState.collectAsState().value
            val vlmState = viewModel.vlmState.collectAsState().value


    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Nexa AI Demo") },
                navigationIcon = {
                    IconButton(onClick = {
                        scope.launch {
                            if (drawerState.currentValue == PushDrawerValue.Open) {
                                drawerState.animateTo(PushDrawerValue.Closed)
                            } else {
                                drawerState.animateTo(PushDrawerValue.Open)
                            }
                        }
                    }) {
                        Icon(Icons.Default.Menu, contentDescription = "Menu")
                    }
                }
            )
        }
    ) { innerPadding ->
        PushDrawer(
            drawerState = drawerState,
            modifier = Modifier.padding(innerPadding),
            drawerWidth = drawerWidth,
            drawerContent = {
                NavigationDrawerContent(
                    selectedScreen = selectedScreen,
                    onScreenSelected = { screen ->
                        selectedScreen = screen
                        scope.launch {
                            drawerState.animateTo(PushDrawerValue.Closed)
                        }
                    }
                )
            },
            content = {
                Column(
                    modifier = Modifier
                        .fillMaxSize()
                ) {

                    // Show content based on selected screen
                    when (selectedScreen) {
                        "Model Loading" -> {
                            ModelLoadingScreen(
                                llmState = llmState,
                                vlmState = vlmState,
                                aiViewModel = viewModel
                            )
                        }
                        "LLM Chat" -> {
                            ChatScreen(
                                chatMessages = chatMessages,
                                currentText = llmCurrentText,
                                generationState = llmGenerationState,
                                inputText = inputText,
                                onInputTextChange = { inputText = it },
                                onSendMessage = {
                                    viewModel.sendMessage(inputText)
                                    inputText = ""
                                }
                            )
                        }
                        "VLM Chat" -> {
                            VLMChatScreen(
                                chatMessages = vlmChatMessages,
                                currentText = vlmCurrentText,
                                generationState = vlmGenerationState,
                                inputText = inputText,
                                onInputTextChange = { inputText = it },
                                onSendMessage = { message, imagePath ->
                                    viewModel.sendVLMMessage(message, imagePath)
                                    inputText = ""
                                }
                            )
                        }
                    }
                }
            }
        )
    }
}





