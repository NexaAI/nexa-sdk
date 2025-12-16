import Foundation

//"{\"event\":{\"id\":\"fv0v268gfidanp762iva9r5tko\",\"summary\":\"The Voice of AGI\",\"description\":\"The voice interface from sci-fi's of old like the Hitchikeer's Guide to the Galaxy to Ironman, the Voice Interface lays out a futurre form of connection, command and interaction.\",\"location\":\"AGI House SF: 170 St. Germain Ave, San Francsoco CA 94114\",\"start\":{\"dateTime\":\"2025-12-20T09:00:00-11:00\",\"timeZone\":\"Pacific/Pago_Pago\"},\"end\":{\"dateTime\":\"2025-12-20T22:00:00-11:00\",\"timeZone\":\"Pacific/Pago_Pago\"},\"status\":\"confirmed\",\"htmlLink\":\"https://www.google.com/calendar/event?eid=ZnYwdjI2OGdmaWRhbnA3NjJpdmE5cjV0a28geWFuZ3hpYW5kYTAwN0AxNjMuY29t\",\"created\":\"2025-12-11T03:48:38.000Z\",\"updated\":\"2025-12-11T03:48:38.842Z\",\"creator\":{\"email\":\"xxx007@163.com\",\"self\":true},\"organizer\":{\"email\":\"xxx007@163.com\",\"self\":true},\"iCalUID\":\"fv0v268gfidanp762iva9r5tko@google.com\",\"sequence\":0,\"reminders\":{\"useDefault\":true},\"eventType\":\"default\",\"calendarId\":\"primary\",\"accountId\":\"normal\"}}"

struct CalendarEventModel: Codable, Identifiable, Equatable {
    var id: String
    var eventName: String
    var description: String
    var summary: String
    var location: String
    var startDate: String
    var startDateTime: String
    var endDateTime: String

    init(
        id: String = "",
        eventName: String = "",
        description: String = "",
        summary: String = "",
        location: String = "",
        startDate: String = "",
        startDateTime: String = "",
        endDateTime: String = ""
    ) {
        self.id = id
        self.eventName = eventName
        self.description = description
        self.summary = summary
        self.location = location
        self.startDate = startDate
        self.startDateTime = startDateTime
        self.endDateTime = endDateTime
    }
}

extension CalendarEventModel {
    static let mock = CalendarEventModel(
        id: "fv0v268gfidanp762iva9r5tko",
        eventName: "The Voice of AGI",
        description: """
The voice interface from sci-fi's of old like the Hitchikeer's Guide to the Galaxy to Ironman, the Voice Interface lays out a futurre form of connection, command and interaction.
""",
        summary: "The Voice of AGI",
        location: "AGI House SF: 170 St. Germain Ave, San Francsoco CA 94114",
        startDate: "2025-12-20",
        startDateTime: "09:00:00",
        endDateTime: "22:00:00"
    )
}
