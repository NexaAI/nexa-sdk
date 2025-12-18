package ai.nexa.agent.retrofit

import ai.nexa.agent.constant.Configs
import ai.nexa.agent.koin.getUnsafeOkHttpClient
import okhttp3.Interceptor
import okhttp3.Response
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory

object RetrofitClient {

    private const val BASE_URL = "http://${Configs.DEFAULT_SERVER_IP}/"

    val instance: ApiService by lazy {
        Retrofit.Builder()
            .baseUrl(BASE_URL)
            .client(okHttpClient)
            .addConverterFactory(GsonConverterFactory.create())
            .build()
            .create(ApiService::class.java)
    }

    private class GlobalBaseUrlInterceptor(private val baseUrlProvider: () -> String) :
        Interceptor {
        override fun intercept(chain: Interceptor.Chain): Response {
            var request = chain.request()
            val baseUrl = baseUrlProvider()

            if (baseUrl.isNotEmpty()) {
                val newUrl = request.url.newBuilder()
                    .scheme(getScheme(baseUrl))
                    .port(getPort(baseUrl))
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
                .split("/").first().split(":").first()
        }

        private fun getPort(url: String): Int {
            var port = 443
            url.split(":").let {
                if (it.first() == "http") {
                    port = 80
                }
                if (it.size == 3) {
                    port = it[2].replace("/", "").toInt()
                }
            }
            return port
        }
    }

    var currentBaseUrl = BASE_URL

    val okHttpClient = getUnsafeOkHttpClient()
        .addInterceptor(GlobalBaseUrlInterceptor { currentBaseUrl })
        .build()

    fun switchApiServer(newBaseUrl: String) {
        currentBaseUrl = if (newBaseUrl.startsWith("http")) {
            newBaseUrl
        } else {
            "http://$newBaseUrl"
        }
    }
}
