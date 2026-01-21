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
            storeFile = file("test")
            storePassword = "123456"
            keyAlias = "test"
            keyPassword = "123456"
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
            applicationIdSuffix = ".rag"
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
            signingConfig = signingConfigs.getByName("release")
        }

        debug {
            isMinifyEnabled = false
            applicationIdSuffix = ".rag"
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
    // NexaAI SDK from Maven - pinned version for stability
    implementation("ai.nexa:core:0.0.17")
    // ===== NEXA CLOUD SDK END =====
    implementation(project(":transform"))
    // Local AAR dependencies from libs folder
    implementation(files("libs/okdownload-core.aar"))
    implementation(files("libs/okdownload-sqlite.aar"))
    implementation(files("libs/okdownload-okhttp.aar"))
    implementation(files("libs/okdownload-ktx.aar"))
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