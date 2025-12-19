package ai.nexa.agent.retrofit

import ai.nexa.agent.bean.GoogleRequestData
import ai.nexa.agent.bean.GoogleRequestData2
import ai.nexa.agent.bean.GoogleResponseBean
import ai.nexa.agent.bean.GoogleResponseBean2
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.POST

interface ApiService {
    @POST("api/calendar/create")
    suspend fun getGoogleCalendarResult(@Body googleRequestData: GoogleRequestData2): GoogleResponseBean2
}