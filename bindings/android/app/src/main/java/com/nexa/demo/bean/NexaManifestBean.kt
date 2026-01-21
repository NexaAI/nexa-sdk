package com.nexa.demo.bean

import android.annotation.SuppressLint
import kotlinx.serialization.Serializable

@SuppressLint("UnsafeOptInUsageError")
@Serializable
data class NexaManifestBean(
    val ModelName: String? = null,
    val ModelType: String? = null,
    val PluginId: String? = null
)