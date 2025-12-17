package ai.nexa.agent.retrofit

import ai.nexa.agent.bean.GoogleRequestData
import ai.nexa.agent.bean.GoogleResponseBean
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.POST

interface ApiService {

    @POST("posts")
    suspend fun getGoogleCalendarResult(@Body googleRequestData: GoogleRequestData): GoogleResponseBean
}