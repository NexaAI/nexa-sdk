package com.nexa.demo.bean

import android.annotation.SuppressLint
import android.content.Context
import com.nexa.demo.FileConfig
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonBuilder
import org.json.JSONObject
import java.io.File

@SuppressLint("UnsafeOptInUsageError")
@Serializable
data class ModelData(
    val id: String,
    val displayName: String,
    val modelName: String,
    /**
     * support plugin_id
     * 0: default, will use cpu;
     * 0x1: cpu, 0x10:gpu, 0x100:npu
     * cpu: 1,
     * gpu: 16,
     * npu: 256,
     * cpu_gpu: 17,
     * cpu_npu: 257,
     * gpu_npu: 272,
     * cpu_gpu_npu: 273
     */
    val pluginIds:Int? = 0,
    val mmprojOrTokenName: String,
    val tokenName: String = "",
    val embeddingName: String = "",
    val extConfigName: String = "",
    val modelUrl: String? = null,
    val mmprojOrTokenUrl: String? = null,
    val embeddingUrl: String? = null,
    val extConfigUrl: String? = null,
    val sizeGb: Double,
    val params: String,
    val features: List<String>,
    val type: String,
    /**
     * Used to indicate the version number of ModelData, mainly for handling the storage location of downloaded files.
     * 0: Default value — files are stored directly under the files/models directory.
     * 1: Files are stored under the files/models/modelId/ directory.
     */
    val versionCode: Int? = 0,
    // NPU-Vision-name
    val patchEmbedName: String = "",
    val vitModelName: String = "",
    val vitConfigFileName: String = "",
    val audioEncoderHelper0Name: String = "",
    val audioEncoderHelper1Name: String = "",
    val audioEncoderModelName: String = "",
    val audioEncoderConfigFileName: String = "",
    // NPU-Vision-url
    val tokenUrl: String? = null,
    val patchEmbedPathUrl: String? = null,
    val vitModelPathUrl: String? = null,
    val vitConfigFilePathUrl: String? = null,
    val audioEncoderHelper0PathUrl: String? = null,
    val audioEncoderHelper1PathUrl: String? = null,
    val audioEncoderModelPathUrl: String? = null,
    val audioEncoderConfigFilePathUrl: String? = null,
    val files: ArrayList<DownloadFileConfig>? = null
) {
    var isSupport = true
}

fun ModelData.modelDir(context: Context): File =
    if (versionCode == 1) {
        File(FileConfig.modelsDir(context), id).apply { if (!exists()) mkdirs() }
    } else {
        FileConfig.modelsDir(context)
    }


fun ModelData.modelFile(context: Context): File? =
    modelUrl?.takeIf { it.isNotBlank() }?.let {
        File(modelDir(context), modelName)
    }

fun ModelData.mmprojTokenFile(context: Context): File? =
    mmprojOrTokenUrl?.takeIf { it.isNotBlank() }?.let {
        File(modelDir(context), mmprojOrTokenName)
    }

fun ModelData.tokenFile(context: Context): File? =
    tokenUrl?.takeIf { it.isNotBlank() }?.let {
        File(modelDir(context), tokenName)
    }

fun ModelData.embederFile(context: Context): File? =
    embeddingUrl?.takeIf { it.isNotBlank() }?.let {
        File(modelDir(context), embeddingName)
    }

fun ModelData.extConfigFile(context: Context): File? =
    extConfigUrl?.takeIf { it.isNotBlank() }?.let {
        File(modelDir(context), extConfigName)
    }

fun ModelData.patchEmbedFile(context: Context): File? =
    patchEmbedPathUrl?.takeIf { it.isNotBlank() }?.let {
        File(modelDir(context), patchEmbedName)
    }

fun ModelData.vitModelFile(context: Context): File? =
    vitModelPathUrl?.takeIf { it.isNotBlank() }?.let {
        File(modelDir(context), vitModelName)
    }

fun ModelData.vitConfigFile(context: Context): File? =
    vitConfigFilePathUrl?.takeIf { it.isNotBlank() }?.let {
        File(modelDir(context), vitConfigFileName)
    }

fun ModelData.audioEncoderHelper0File(context: Context): File? =
    audioEncoderHelper0PathUrl?.takeIf { it.isNotBlank() }?.let {
        File(modelDir(context), audioEncoderHelper0Name)
    }

fun ModelData.audioEncoderHelper1File(context: Context): File? =
    audioEncoderHelper1PathUrl?.takeIf { it.isNotBlank() }?.let {
        File(modelDir(context), audioEncoderHelper1Name)
    }

fun ModelData.audioEncoderModelFile(context: Context): File? =
    audioEncoderModelPathUrl?.takeIf { it.isNotBlank() }?.let {
        File(modelDir(context), audioEncoderModelName)
    }

fun ModelData.audioEncoderConfigFile(context: Context): File? =
    audioEncoderConfigFilePathUrl?.takeIf { it.isNotBlank() }?.let {
        File(modelDir(context), audioEncoderConfigFileName)
    }

fun ModelData.downloadableFiles(modelDir: File): List<DownloadableFile> = listOfNotNull(
    modelUrl?.takeIf { it.isNotBlank() }?.let {
        DownloadableFile(File(modelDir, modelName), it)
    },
    mmprojOrTokenUrl?.takeIf { it.isNotBlank() }?.let {
        DownloadableFile(File(modelDir, mmprojOrTokenName), it)
    },
    tokenUrl?.takeIf { it.isNotBlank() }?.let {
        DownloadableFile(File(modelDir, tokenName), it)
    },
    embeddingUrl?.takeIf { it.isNotBlank() }?.let {
        DownloadableFile(File(modelDir, embeddingName), it)
    },
    extConfigUrl?.takeIf { it.isNotBlank() }?.let {
        DownloadableFile(File(modelDir, extConfigName), it)
    },
    patchEmbedPathUrl?.takeIf { it.isNotBlank() }?.let {
        DownloadableFile(File(modelDir, patchEmbedName), it)
    },
    vitModelPathUrl?.takeIf { it.isNotBlank() }?.let {
        DownloadableFile(File(modelDir, vitModelName), it)
    },
    vitConfigFilePathUrl?.takeIf { it.isNotBlank() }?.let {
        DownloadableFile(File(modelDir, vitConfigFileName), it)
    },
    audioEncoderHelper0PathUrl?.takeIf { it.isNotBlank() }?.let {
        DownloadableFile(File(modelDir, audioEncoderHelper0Name), it)
    },
    audioEncoderHelper1PathUrl?.takeIf { it.isNotBlank() }?.let {
        DownloadableFile(File(modelDir, audioEncoderHelper1Name), it)
    },
    audioEncoderModelPathUrl?.takeIf { it.isNotBlank() }?.let {
        DownloadableFile(File(modelDir, audioEncoderModelName), it)
    },
    audioEncoderConfigFilePathUrl?.takeIf { it.isNotBlank() }?.let {
        DownloadableFile(File(modelDir, audioEncoderConfigFileName), it)
    }
).let {
    val temp = arrayListOf<DownloadableFile>()
    files?.takeIf { it.isNotEmpty() }?.let { allFiles ->
        allFiles.forEach { fileConfig ->
            temp.add(
                DownloadableFile(
                    File(
                        modelDir,
                        fileConfig.path + File.separator + fileConfig.name
                    ), fileConfig.url
                )
            )
        }
    }
    it + temp
}


fun ModelData.allModelFilesExist(modelDir: File): Boolean {
    val files = this.downloadableFiles(modelDir).map { it.file }
    return files.all { it.exists() && it.length() > 0 }
}

fun ModelData.getNonExistModelFile(modelDir: File):String? {
    this.downloadableFiles(modelDir).forEach {
        if (!(it.file.exists() && it.file.length() > 0)) {
            return it.file.absolutePath.replace("/data/user/0", "/data/data")
        }
    }
    return null
}

fun ModelData.getNexaManifest(context: Context): NexaManifestBean? {
    try {
        val str = File(modelDir(context), "nexa.manifest").bufferedReader().use { it.readText() }
        return Json { 
            ignoreUnknownKeys = true
        }.decodeFromString<NexaManifestBean>(str)
    } catch (e: Exception) {
        e.printStackTrace()
        return null
    }
}

fun ModelData.getSupportPluginIds(): ArrayList<String> {
    val pluginIds = arrayListOf<String>()
    if (this.pluginIds == 0) {
        pluginIds.add("cpu")
    } else {
        if (this.pluginIds!! and 0x100 == 0x100) {
            pluginIds.add("npu")
        }
        if (this.pluginIds and 0x10 == 0x10) {
            pluginIds.add("gpu")
        }
        if (this.pluginIds and 0x1 == 0x1) {
            pluginIds.add("cpu")
        }
    }
    return pluginIds
}