// Copyright 2024-2026 Nexa AI, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

