enableFeaturePreview("TYPESAFE_PROJECT_ACCESSORS")

pluginManagement {
    repositories {
        google {
            content {
                includeGroupByRegex("com\\.android.*")
                includeGroupByRegex("com\\.google.*")
                includeGroupByRegex("androidx.*")
            }
        }
        mavenCentral()
        gradlePluginPortal()
    }
}
dependencyResolutionManagement {
    repositoriesMode.set(RepositoriesMode.FAIL_ON_PROJECT_REPOS)
    repositories {
        google()
        mavenCentral()
        maven { url = uri("https://jitpack.io") } // Added JitPack for AndroidAutoSize
//        maven {
//            url = uri("https://raw.githubusercontent.com/NexaAI/core/main")
//        }
        flatDir {
            dirs("app/libs")
        }
    }
}

rootProject.name = "NexaDemo"

// Path to nexasdk-bridge library
val defaultBridgeLibDir = File(rootDir, "../../nexasdk-bridge/bindings/android/app")
val bridgeLibDirPath: String = System.getenv("NEXA_BRIDGE_ANDROID")
    ?: defaultBridgeLibDir.absolutePath

print("bridgeLibDirPath:${bridgeLibDirPath} exist? ${File(bridgeLibDirPath).exists()}\n")
if (File(bridgeLibDirPath).exists()) {
    gradle.extra["bridgePathExist"] = true
    include(":bridgeLib")
    project(":bridgeLib").projectDir = File(bridgeLibDirPath)
} else {
    gradle.extra["bridgePathExist"] = false
}

include(":transform")
include(":app")

