import SwiftUI

struct CalendarEventView: View {
    var event: CalendarEventModel
    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            titleView
                .padding(.bottom, 12)

            VStack(alignment: .leading) {
                subtitleView("Event Name")
                Text(event.eventName)
                    .padding(.leading, 22)
            }

            if !event.startDate.isEmpty {
                VStack(alignment: .leading) {
                    subtitleView("Event Time")
                    HStack {
                        Text("Date:")
                            .padding(.leading, 22)
                        buildDateTime(event.startDate)
                    }

                    HStack {
                        Text("Start:")
                            .padding(.leading, 22)
                        buildDateTime(event.startDateTime)

                        if !event.endDateTime.isEmpty {
                            Text("End:")
                                .padding(.leading, 12)
                            buildDateTime(event.endDateTime)
                        }
                    }
                }
            }

            if !event.location.isEmpty {
                VStack(alignment: .leading) {
                    subtitleView("Event Location")
                    Text(event.location)
                        .padding(.leading, 22)
                }
            }

            if !event.description.isEmpty {
                VStack(alignment: .leading) {
                    subtitleView("Event Description")
                    Text(event.description)
                        .padding(.leading, 22)
                }
            }
        }
        .font(.system(size: 14))
        .foregroundStyle(Color.Text.primary)
        .padding(12)
        .background(
            RoundedRectangle(cornerRadius: 12)
                .fill(Color.Background.primary)
                .stroke(Color.Thinkingbox.border)
        )
        .frame(maxWidth: .infinity, alignment: .leading)
    }

    @ViewBuilder
    var titleView: some View {
        Group {
            Text("Event added to ")
                .foregroundStyle(Color.Text.primary)
            +
            Text("@Google Calendar")
                .foregroundStyle(Color.Icon.brand)
        }
        .font(.system(size: 20, weight: .medium))
    }

    @ViewBuilder
    func subtitleView(_ title: String) -> some View {
        HStack(spacing: 8) {
            ZStack {
                Circle()
                    .fill(Color.Thinkingbox.dotBack)
                    .frame(width: 14)
                Circle()
                    .fill(Color.Thinkingbox.dotFront)
                    .frame(width: 8)
            }

            Text(title)
                .font(.system(size: 16, weight: .medium))
        }
    }

    @ViewBuilder
    func buildDateTime(_ content: String) -> some View {
        Text(content)
            .font(.system(size: 12))
            .foregroundStyle(Color.Text.secondary)
            .padding(.horizontal, 12)
            .padding(.vertical, 4)
            .background(
                RoundedRectangle(cornerRadius: 6)
                    .fill(Color.Background.primary)
                    .stroke(Color.Component.Border.secondary)
            )
    }

}

#Preview {
    CalendarEventView(event: .mock)
}
