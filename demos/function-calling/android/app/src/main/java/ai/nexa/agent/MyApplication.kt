package ai.nexa.agent

import ai.nexa.agent.constant.Configs
import android.app.Application
import android.content.Context
import android.os.Build
import com.nexa.studio.koin.appModule
import org.koin.android.ext.koin.androidContext
import org.koin.core.context.GlobalContext.startKoin

class MyApplication: Application() {
    override fun onCreate() {
        super.onCreate()
        if (Configs.USE_UNSAFE_HTTP) {
            okdownload()
        }
        startKoin {
            androidContext(this@MyApplication)
            modules(appModule)
//            modules(databaseModule)
        }
    }

    override fun attachBaseContext(base: Context) {
        super.attachBaseContext(base)
    }

    private fun okdownload() {
//        val okDownloadBuilder = OkDownload.Builder(this)
//        val factory = DownloadOkHttp3Connection.Factory()
//        factory.setBuilder(getUnsafeOkHttpClient())
//        okDownloadBuilder.connectionFactory(factory)
//        try {
//            OkDownload.setSingletonInstance(okDownloadBuilder.build())
//        } catch (e: java.lang.Exception) {
//            L.e("download", "download init failed")
//        }
    }
}