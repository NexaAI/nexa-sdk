package ai.nexa.agent.bean

import kotlinx.serialization.Serializable

@Serializable
class GoogleCalendarEvent {
    var event: Event? = null

    @Serializable
    class Event {
        var id: String? = null
        var summary: String? = null
        var description: String? = null
        var location: String? = null
        var start: Start? = null
        var end: End? = null
        var status: String? = null
        var htmlLink: String? = null
        var created: String? = null
        var updated: String? = null
        var creator: Creator? = null
        var organizer: Organizer? = null
        var iCalUID: String? = null
        var sequence: Int = 0
        var reminders: Reminders? = null
        var eventType: String? = null
        var calendarId: String? = null
        var accountId: String? = null

        @Serializable
        class Start {
            var dateTime: String? = null
            var timeZone: String? = null
        }

        @Serializable
        class End {
            var dateTime: String? = null
            var timeZone: String? = null
        }

        @Serializable
        class Creator {
            var email: String? = null
            var isSelf: Boolean = false
        }

        @Serializable
        class Organizer {
            var email: String? = null
            var isSelf: Boolean = false
        }

        @Serializable
        class Reminders {
            var isUseDefault: Boolean = false
        }
    }
}
