//
//  PushNotiCode.swift
//  CrossShare_iOS
//
//  Created by jack_huang on 5/28/25.
//
enum PushNotiCode {
    case sendStart
    case sendDone
    case receiveStart
    case receiveDone
}

extension PushNotiCode {
    var title: String {
        switch self {
        case .sendStart:
            return "File transfer"
        case .sendDone:
            return "File transfer"
        case .receiveStart:
            return "File transfer"
        case .receiveDone:
            return "File transfer"
        }
    }
    var content: String {
        switch self {
        case .sendStart:
            return "Starting to transfer {$} to {$}..."
        case .sendDone:
            return "{$} transferred to {$} is complete"
        case .receiveStart:
            return "Starting to receive {$} from {$}"
        case .receiveDone:
            return "{$} received from {$} is complete"
        }
    }
}
