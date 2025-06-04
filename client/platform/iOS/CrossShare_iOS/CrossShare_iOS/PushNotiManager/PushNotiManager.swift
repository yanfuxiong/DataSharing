//
//  PushNotiManager.swift
//  CrossShare_iOS
//
//  Created by jack_huang on 5/28/25.
//
import UserNotifications

class PushNotiManager {
    static let shared = PushNotiManager()

    struct PushNotiMsg {
        private(set) var title: String = ""
        private(set) var body: String = ""
        init(code: PushNotiCode, with params: [String]) {
            self.title = code.title
            self.body = self.fillByParams(in: code.content, with: params)
        }

        private func fillByParams(in content: String, with params: [String]) -> String {
            var result = ""
            var paramIndex = 0
            var startIndex = content.startIndex
            let key = "{$}"

            while let range = content[startIndex...].range(of: key) {
                result += content[startIndex..<range.lowerBound]
                if paramIndex < params.count {
                    result += params[paramIndex]
                    paramIndex += 1
                } else {
                    result += key
                }
                startIndex = range.upperBound
            }

            result += content[startIndex...]
            return result
        }
    }

    private init() {
    }

    func initNoti() {
        UNUserNotificationCenter.current().requestAuthorization(options: [.alert, .sound, .badge]) { granted, error in
            if granted {
                print("Notification permission granted")
            } else {
                print("Notification permission denied: \(error?.localizedDescription ?? "Unknown error")")
            }
        }
    }

    func sendLocalNotification(code: PushNotiCode, with params: [String]) {
        let pushNotiMsg = PushNotiMsg(code: code, with: params)

        let content = UNMutableNotificationContent()
        content.title = pushNotiMsg.title
        content.body = pushNotiMsg.body
        content.sound = UNNotificationSound.default

        let request = UNNotificationRequest(identifier: UUID().uuidString, content: content, trigger: nil)

        UNUserNotificationCenter.current().add(request) { error in
            if let error = error {
                print("[PushNotiManager][Err]: Failed to schedule notification: \(error)")
            }
        }
    }
}
