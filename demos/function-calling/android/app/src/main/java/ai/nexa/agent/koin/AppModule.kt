package com.nexa.studio.koin

import ai.nexa.agent.model.ChatViewModel
import ai.nexa.agent.model.CommonSettingViewModel
import ai.nexa.agent.repository.CommonSettingRepository
import ai.nexa.agent.util.dataStore
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import okhttp3.OkHttpClient
import org.koin.android.ext.koin.androidContext
import org.koin.androidx.viewmodel.dsl.viewModel
import org.koin.dsl.module
import java.security.SecureRandom
import java.security.cert.CertificateException
import java.security.cert.X509Certificate
import javax.net.ssl.SSLContext
import javax.net.ssl.SSLSession
import javax.net.ssl.SSLSocketFactory
import javax.net.ssl.TrustManager
import javax.net.ssl.X509TrustManager


fun getUnsafeOkHttpClient(): OkHttpClient.Builder {
    try {
        val x509m: X509TrustManager = object : X509TrustManager {
            override fun getAcceptedIssuers(): Array<X509Certificate?>? {
                //注意这里不能返回null，否则会报错
                val x509Certificates = arrayOfNulls<X509Certificate>(0)
                return x509Certificates
            }

            @Throws(CertificateException::class)
            override fun checkServerTrusted(
                chain: Array<X509Certificate?>?,
                authType: String?
            ) {
// 不抛出异常即信任所有服务器证书
            }

            @Throws(CertificateException::class)
            override fun checkClientTrusted(
                chain: Array<X509Certificate?>?,
                authType: String?
            ) {
// 默认信任机制
            }
        }
        // 创建一个信任所有证书的 TrustManager
        val trustAllCerts = arrayOf<TrustManager>(x509m)

        // 初始化 SSLContext
        val sslContext = SSLContext.getInstance("SSL")
        sslContext.init(null, trustAllCerts, SecureRandom())

        // 创建 SSLSocketFactory
        val sslSocketFactory: SSLSocketFactory = sslContext.getSocketFactory()

        // 构建 OkHttpClient
        return OkHttpClient.Builder()
            .sslSocketFactory(
                sslSocketFactory,
                (trustAllCerts[0] as X509TrustManager?)!!
            )
            .hostnameVerifier { hostname: String?, session: SSLSession? -> true }
    } catch (e: Exception) {
        throw RuntimeException(e)
    }
}

val appModule = module {
    single<DataStore<Preferences>> { androidContext().dataStore }
//    single { ChatSessionUseCase(get()) }
    single { OkHttpClient() }
//    single { ModelRepository(androidContext(), get()) }
    single { CommonSettingRepository(get()) }
    viewModel { CommonSettingViewModel(get()) }
//    viewModel { ModelDownloadViewModel(get(), get()) }
    single { ChatViewModel(get()) }
//    single<ConfigRepository> {
//        ConfigRepositoryImpl(get())
//    }
//    single { GenerateChatMessageUseCase(get(), get()) }
//    single<AIModelRepository>(named("LLM")) { LLMRepositoryImpl(androidContext()) }
//    single<AIModelRepository>(named("VLM")) { VLMRepositoryImpl(androidContext()) }
//    single<AIModelRepository>(named("QNN")) { QnnRepositoryImpl(androidContext()) }
//    single<AIModelRepository>(named("QNN-VISION")) { QnnVisionRepositoryImpl(androidContext()) }

//    single {
//        ModelRepositorySelector(
//            mapOf(
//                ModelType.LLM to get<AIModelRepository>(named("LLM")),
//                ModelType.VLM to get<AIModelRepository>(named("VLM")),
//                ModelType.QNN to get<AIModelRepository>(named("QNN")),
//                ModelType.QNN_VISION to get<AIModelRepository>(named("QNN-VISION"))
//            )
//        )
//    }
}


val databaseModule = module {
//    single {
//        Room.databaseBuilder(
//            androidContext(),
//            AppDatabase::class.java,
//            "session_db"
//        ).addMigrations(MIGRATION_3_4).build()
//    }
//    single { get<AppDatabase>().chatSessionDao() }
//    single { ChatSessionRepository(get()) }
}