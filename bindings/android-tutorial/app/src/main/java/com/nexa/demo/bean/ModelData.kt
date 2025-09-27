package com.nexa.demo.bean

import android.annotation.SuppressLint
import android.content.Context
import com.nexa.demo.FileConfig
import kotlinx.serialization.Serializable
import java.io.File

@SuppressLint("UnsafeOptInUsageError")
@Serializable
data class ModelData(
    val id: String,
    val displayName: String,
    val modelName: String,
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
     * 用于标记 ModelData 的版本号，主要用于处理下载文件的存放
     * 0：默认值，直接存放在 files/models 目录下
     * 1：存放在 files/models/modelId/ 目录下
     */
    val versionCode:Int? = 0,
    /**
     * 是否支持从 S3 服务器下载资源
     */
    val supportS3: Boolean? = false,
    // Qnn-Vision-name
    val patchEmbedName: String = "",
    val vitModelName: String = "",
    val vitConfigFileName: String = "",
    val audioEncoderHelper0Name: String = "",
    val audioEncoderHelper1Name: String = "",
    val audioEncoderModelName: String = "",
    val audioEncoderConfigFileName: String = "",
    // Qnn-Vision-url
    val tokenUrl: String? = null,
    val patchEmbedPathUrl: String? = null,
    val vitModelPathUrl: String? = null,
    val vitConfigFilePathUrl: String? = null,
    val audioEncoderHelper0PathUrl: String? = null,
    val audioEncoderHelper1PathUrl: String? = null,
    val audioEncoderModelPathUrl: String? = null,
    val audioEncoderConfigFilePathUrl: String? = null,
) {
    /**
     * 当前设备是否支持该模型
     */
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
)

fun ModelData.allModelFilesExist(modelDir: File): Boolean {
    val files = this.downloadableFiles(modelDir).map { it.file }
    return files.all { it.exists() && it.length() > 0 }
}