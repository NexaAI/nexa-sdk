package ai.nexa.agent.retrofit

import ai.nexa.agent.koin.getUnsafeOkHttpClient
import okhttp3.Interceptor
import okhttp3.Response
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory

object RetrofitClient {

    private const val BASE_URL = "https://192.168.1.107/"

    val instance: ApiService by lazy {
        Retrofit.Builder()
            .baseUrl(BASE_URL)
            .client(okHttpClient)
            .addConverterFactory(GsonConverterFactory.create())
            .build()
            .create(ApiService::class.java)
    }

    class GlobalBaseUrlInterceptor(private val baseUrlProvider: () -> String) : Interceptor {
        override fun intercept(chain: Interceptor.Chain): Response {
            var request = chain.request()
            val baseUrl = baseUrlProvider()

            if (baseUrl.isNotEmpty()) {
                val newUrl = request.url.newBuilder()
                    .scheme(getScheme(baseUrl))
                    .host(getHost(baseUrl))
                    .build()

                request = request.newBuilder()
                    .url(newUrl)
                    .build()
            }

            return chain.proceed(request)
        }

        private fun getScheme(url: String): String {
            return if (url.startsWith("https://")) "https" else "http"
        }

        private fun getHost(url: String): String {
            return url.removePrefix("http://")
                .removePrefix("https://")
                .split("/").first()
        }
    }

    // 使用
    var currentBaseUrl = "https://default-api.com"

    val okHttpClient = getUnsafeOkHttpClient()
        .addInterceptor(GlobalBaseUrlInterceptor { currentBaseUrl })
        .build()

    // 动态修改 baseUrl
    fun switchApiServer(newBaseUrl: String) {
        currentBaseUrl = if (newBaseUrl.startsWith("http")) {
            newBaseUrl
        } else {
            "http://$newBaseUrl"
        }
    }
}
