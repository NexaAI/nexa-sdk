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

plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.android)
    alias(libs.plugins.kotlin.compose)
    id("org.jetbrains.kotlin.plugin.serialization") version "1.9.23"
}

android {
    namespace = "com.nexa.demo"
    compileSdk = 36

    signingConfigs {
        create("release") {
            // Note: For production builds, use environment variables or local.properties
            // Example: storePassword = System.getenv("KEYSTORE_PASSWORD") ?: ""
            storeFile = file("test")
            storePassword = project.findProperty("KEYSTORE_PASSWORD")?.toString() ?: "123456"
            keyAlias = "test"
            keyPassword = project.findProperty("KEY_PASSWORD")?.toString() ?: "123456"
        }
    }

    defaultConfig {
        applicationId = "com.nexa.demo"
        minSdk = 27
        targetSdk = 36
        versionCode = 1
        versionName = "1.0"

        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"
    }

    buildTypes {
        release {
            isMinifyEnabled = false
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
            signingConfig = signingConfigs.getByName("release")
        }

        debug {
            isMinifyEnabled = false
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
            signingConfig = signingConfigs.getByName("release")
        }
    }
    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_11
        targetCompatibility = JavaVersion.VERSION_11
    }
    kotlinOptions {
        jvmTarget = "11"
    }
//    sourceSets {
//        getByName("main") {
//            jniLibs.srcDirs("src/main/jniLibs")
//        }
//    }
    packagingOptions {
        jniLibs.useLegacyPackaging = true
    }

    buildFeatures {
        viewBinding = true
        dataBinding = true
        compose = true
        buildConfig = true
    }
}

val bridgePathExist = gradle.extra["bridgePathExist"] as Boolean
print("bridgePathExist: $bridgePathExist\n")

dependencies {

    // ===== NEXA CLOUD SDK =====
    // Using cloud SDK instead of local bridge - latest version
    implementation("ai.nexa:core:+")
    // ===== NEXA CLOUD SDK END =====
    implementation(project(":transform"))
    implementation(":okdownload-core@aar")
    implementation(":okdownload-sqlite@aar")
    implementation(":okdownload-okhttp@aar")
    implementation(":okdownload-ktx@aar")
    implementation(kotlin("reflect"))
    implementation(libs.glide)
    implementation(libs.gson)
    implementation(libs.markwon.core)
    implementation(libs.markwon.strikethrough)
    implementation(libs.markwon.tables)
    implementation(libs.markwon.linkify)
    implementation(libs.recyclerview)
    implementation(libs.toaster)
    implementation(libs.material)
    implementation(libs.imm.bar)
    implementation(libs.imm.bar.ktx)
    implementation(libs.auto.size)
    implementation(libs.okhttp)
    implementation(libs.kotlinx.serialization.json)
    implementation(libs.androidx.core.ktx)
    implementation(libs.androidx.lifecycle.runtime.ktx)
    implementation(libs.androidx.activity.compose)
    implementation(platform(libs.androidx.compose.bom))
    implementation(libs.androidx.ui)
    implementation(libs.androidx.ui.graphics)
    implementation(libs.androidx.ui.tooling.preview)
    implementation(libs.androidx.material3)
    implementation(libs.androidx.appcompat)
    implementation(libs.androidx.activity)
    implementation(libs.androidx.constraintlayout)
    testImplementation(libs.junit)
    androidTestImplementation(libs.androidx.junit)
    androidTestImplementation(libs.androidx.espresso.core)
    androidTestImplementation(platform(libs.androidx.compose.bom))
    androidTestImplementation(libs.androidx.ui.test.junit4)
    debugImplementation(libs.androidx.ui.tooling)
    debugImplementation(libs.androidx.ui.test.manifest)
}