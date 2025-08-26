
import SwiftUI

struct AnyAlert: Sendable {
    
    let title: String
    let subtitle: String
    let buttons: (@Sendable () -> AnyView)?
    
    init(
        title: String = "",
        subtitle: String = "",
        buttons: (@Sendable () -> AnyView)? = nil
    ) {
        self.title = title
        self.subtitle = subtitle
        self.buttons = buttons ?? {
            AnyView(Button("OK", action: {}))
        }
    }
    
    init(_ error: Error) {
        self.init(title: error.localizedDescription, subtitle: "", buttons: nil)
    }
}

enum AlertType {
    case alert
    case sheet
}

extension View {
    
    @ViewBuilder
    func anyAlert(_ type: AlertType = .alert, alert: Binding<AnyAlert?>) -> some View {
        switch type {
        case .alert:
            self
                .alert(alert.wrappedValue?.title ?? "", isPresented: .init(alert)) {
                    alert.wrappedValue?.buttons?()
                }  message: {
                    Text(alert.wrappedValue?.subtitle ?? "")
                }
        case .sheet:
            let showTitle = alert.wrappedValue?.title != nil && !(alert.wrappedValue?.title)!.isEmpty
            self.confirmationDialog(alert.wrappedValue?.title ?? "", isPresented: .init(alert), titleVisibility: showTitle ? .visible : .hidden) {
                alert.wrappedValue?.buttons?()
            } message: {
                Text(alert.wrappedValue?.subtitle ?? "")
            }
        }
        
    }
    
}
