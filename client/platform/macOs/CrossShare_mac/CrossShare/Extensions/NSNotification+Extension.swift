import Cocoa

extension NSNotification.Name {
    static let accessibilityApi = NSNotification.Name(rawValue: "com.apple.accessibility.api")
    static let displayConnectedNotification = Notification.Name("CrossShareDisplayConnected")
    static let displayDisconnectedNotification = Notification.Name("CrossShareDisplayDisconnected")
    static let displayConfigurationChangedNotification = Notification.Name("CrossShareDisplayConfigurationChanged")
    static let deviceDataReceived = Notification.Name("CrossShareDeviceDataReceived")
    static let screenCountChanged = Notification.Name("CrossShareScreenCountChanged")
    static let fileTransferSessionUpdated = Notification.Name("FileTransferSessionUpdated")
    static let fileTransferSessionStarted = Notification.Name("FileTransferSessionStarted")
    static let fileTransferSessionCompleted = Notification.Name("FileTransferSessionCompleted")
    static let fileTransferSessionFailed = Notification.Name("FileTransferSessionFailed")
    //
    static let deviceDiasStatusNotification = Notification.Name("deviceDiasStatusNotification")
    static let didReceiveErrorEventNotification = Notification.Name("didReceiveErrorEventNotification")

}
