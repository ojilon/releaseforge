plugins {
    id("com.android.application")
    id("org.jetbrains.kotlin.android")
}

android {
    namespace = "com.example.kts"
    compileSdk = 35

    defaultConfig {
        applicationId = "com.example.kts"
        minSdk = 28
        targetSdk = 35
        versionCode = 3
        versionName = "0.3.0"
        ndk {
            abiFilters += listOf("arm64-v8a", "x86_64")
        }
    }
}
