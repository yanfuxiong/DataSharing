//
//  PushNotiCode.swift
//  CrossShare_iOS
//
//  Created by jack_huang on 5/28/25.
//
enum PushNotiCode {
    case sendStart
    case sendDone
    case sendError
    case receiveStart
    case receiveDone
    case receiveError
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
        case .sendError:
            return "File transfer"
        case .receiveError:
            return "File transfer"
        }
    }
    var content: String {
        switch self {
        case .sendStart:
            return "Starting to transfer {$} to {$}..." // fileName, clientName
        case .sendDone:
            return "{$} transferred to {$} is complete" // fileName, clientName
        case .receiveStart:
            return "Starting to receive {$} from {$}" // fileName, clientName
        case .receiveDone:
            return "{$} received from {$} is complete" // fileName, clientName
        case .sendError:
            return "{$} transferred to {$} failed" // fileName, clientName
        case .receiveError:
            return "{$} received from {$} failed" // fileName, clientName
        }
    }
}
