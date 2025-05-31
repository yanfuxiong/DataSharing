#include "common_message.h"

namespace sunkang {

std::string RecordDataHash::getHashID() const
{
    std::stringstream str_stream;
    str_stream << Utils::getFileNameByPath(fileName) << "_";
    str_stream << fileSize << "_";
    str_stream << Utils::toHex(clientID) << "_";
    str_stream << ip;
    //qInfo() << "[RecordDataHash]:" << str_stream.str().c_str();
    return std::to_string(std::hash<std::string>{}(str_stream.str()));
}

int MsgHeader::messageLength()
{
    int headerMsgLen = static_cast<int>(g_tagNameLength + sizeof (uint8_t) + sizeof (uint8_t) + sizeof (uint32_t));
    return headerMsgLen;
}

bool g_getCodeFromByteArray(const std::string &data, uint8_t &typeValue, uint8_t &codeValue)
{
    if (static_cast<int>(data.length()) < MsgHeader::messageLength()) {
        return false;
    }
    Buffer buffer;
    buffer.append(data);
    std::string header = std::string(buffer.peek(), g_tagNameLength);
    buffer.retrieve(g_tagNameLength);
    if (header != TAG_NAME_X) {
        LOG_WARN << "Illegal Message HEADER:" << header;
        return false;
    }

    {
        uint8_t type = 0;
        memcpy(&type, buffer.peek(), sizeof (uint8_t));
        buffer.retrieve(sizeof (uint8_t));

        typeValue = type;
    }

    {
        uint8_t code = 0;
        memcpy(&code, buffer.peek(), sizeof (uint8_t));
        buffer.retrieve(sizeof (uint8_t));

        codeValue = code;
    }

    if (buffer.readableBytes() < sizeof (uint32_t)) {
        return false;
    }

    {
        uint32_t contentLength = buffer.peekUInt32();
        buffer.retrieveUInt32();
        if (contentLength > buffer.readableBytes()) {
            return false; // At this point, it indicates that the data is not complete and we need to continue waiting
        }
    }

    return true;
}

bool g_getCodeFromByteArray(const std::string &data, uint8_t &codeValue)
{
    uint8_t typeValue = 0;
    return g_getCodeFromByteArray(data, typeValue, codeValue);
}

std::string UpdateClientStatusMsg::toByteArray(const UpdateClientStatusMsg &msg)
{
    Buffer data;
    data.append(msg.headerInfo.header);
    data.append(&msg.headerInfo.type, sizeof (msg.headerInfo.type));
    data.append(&msg.headerInfo.code, sizeof (msg.headerInfo.code));

    {
        Buffer tmpBuffer;
        tmpBuffer.append(&msg.status, sizeof (msg.status));
        uint32_t ipValue = Utils::toIPv4Address(msg.ip);
        tmpBuffer.appendUInt32(ipValue);
        tmpBuffer.appendUInt16(msg.port);
        assert(msg.clientID.length() == 46);
        tmpBuffer.append(msg.clientID);
        {
            std::wstring clientName_utf16 = Utils::toUtf16LE(msg.clientName);
            tmpBuffer.appendUInt32(static_cast<uint32_t>(clientName_utf16.length() * sizeof (wchar_t)));
            tmpBuffer.append(clientName_utf16);
        }

        {
            tmpBuffer.appendUInt32(static_cast<uint32_t>(msg.deviceType.length()));
            tmpBuffer.append(msg.deviceType);
        }

        // Processing content length
        data.appendUInt32(static_cast<uint32_t>(tmpBuffer.readableBytes()));
        data.append(tmpBuffer.retrieveAllAsString());
    }
    return data.retrieveAllAsString();
}

bool UpdateClientStatusMsg::fromByteArray(const std::string &data, UpdateClientStatusMsg &msg)
{
    if (static_cast<int>(data.length()) < MsgHeader::messageLength()) {
        return false;
    }
    Buffer buffer;
    buffer.append(data);
    std::string header = std::string(buffer.peek(), g_tagNameLength);
    buffer.retrieve(g_tagNameLength);
    if (header != TAG_NAME_X) {
        LOG_WARN << "Illegal Message HEADER:" << header;
        return false;
    }

    msg.headerInfo.header = header;

    {
        uint8_t type = 0;
        memcpy(&type, buffer.peek(), sizeof (uint8_t));
        buffer.retrieve(sizeof (uint8_t));
        msg.headerInfo.type = type;
    }

    {
        uint8_t code = 0;
        memcpy(&code, buffer.peek(), sizeof (uint8_t));
        buffer.retrieve(sizeof (uint8_t));
        msg.headerInfo.code = code;
    }

    {
        uint32_t contentLength = buffer.peekUInt32();
        buffer.retrieveUInt32();
        if (contentLength > buffer.readableBytes()) {
            return false; // At this point, it indicates that the data is not complete and we need to continue waiting
        }
        msg.headerInfo.contentLength = contentLength;
    }

    {
        Buffer contentBuffer;
        contentBuffer.append(buffer.peek(), msg.headerInfo.contentLength);
        buffer.retrieve(msg.headerInfo.contentLength);

        {
            uint8_t status = 0;
            memcpy(&status, contentBuffer.peek(), sizeof (uint8_t));
            contentBuffer.retrieve(sizeof (uint8_t));
            msg.status = status;
        }

        {
            uint32_t ipValue = contentBuffer.peekUInt32();
            contentBuffer.retrieveUInt32();
            msg.ip = Utils::fromIPv4Address(ipValue);
        }

        {
            msg.port = contentBuffer.peekUInt16();
            contentBuffer.retrieveUInt16();
        }

        {
            msg.clientID = std::string(contentBuffer.peek(), 46);
            contentBuffer.retrieve(46);
        }

        {
            uint32_t dataLen = contentBuffer.peekUInt32();
            contentBuffer.retrieveUInt32();
            std::string clientName_utf16(contentBuffer.peek(), dataLen);
            contentBuffer.retrieve(dataLen);
            msg.clientName = Utils::toUtf8(sunkang::Utils::stringToWString(clientName_utf16));
        }

        {
            uint32_t dataLen = contentBuffer.peekUInt32();
            contentBuffer.retrieveUInt32();
            msg.deviceType = std::string(contentBuffer.peek(), dataLen);
            contentBuffer.retrieve(dataLen);
        }
    }
    return true;
}

//-----------------------------------

std::string SendFileRequestMsg::toByteArray(const SendFileRequestMsg &msg)
{
    Buffer data;
    data.append(msg.headerInfo.header);
    data.append(&msg.headerInfo.type, sizeof (msg.headerInfo.type));
    data.append(&msg.headerInfo.code, sizeof (msg.headerInfo.code));

    {
        Buffer tmpBuffer;
        tmpBuffer.append(&msg.flag, sizeof (uint8_t));
        uint32_t ipValue = Utils::toIPv4Address(msg.ip);
        tmpBuffer.appendUInt32(ipValue);
        tmpBuffer.appendUInt16(msg.port);
        assert(msg.clientID.length() == 46);
        tmpBuffer.append(msg.clientID);
        tmpBuffer.appendUInt64(msg.fileSize);
        tmpBuffer.appendUInt64(msg.timeStamp);
        {
            std::wstring fileName_utf16 = Utils::toUtf16LE(msg.fileName);
            tmpBuffer.appendUInt32(static_cast<uint32_t>(fileName_utf16.size() * sizeof (wchar_t)));
            tmpBuffer.append(fileName_utf16);
        }

        do {
            uint32_t filesCount = static_cast<uint32_t>(msg.filePathVec.size());
            tmpBuffer.appendUInt32(filesCount);
            if (filesCount == 0) {
                break;
            }
            for (const auto &filePath : msg.filePathVec) {
                std::wstring filePath_utf16 = Utils::toUtf16LE(filePath);
                tmpBuffer.appendUInt16(static_cast<uint16_t>(filePath_utf16.size() * sizeof (wchar_t)));
                tmpBuffer.append(filePath_utf16);
            }
        } while (false);

        // Processing content length
        data.appendUInt32(static_cast<uint32_t>(tmpBuffer.readableBytes()));
        data.append(tmpBuffer.retrieveAllAsString());
    }
    return data.retrieveAllAsString();
}

bool SendFileRequestMsg::fromByteArray(const std::string &data, SendFileRequestMsg &msg)
{
    if (static_cast<int>(data.length()) < MsgHeader::messageLength()) {
        return false;
    }
    Buffer buffer;
    buffer.append(data);
    std::string header = std::string(buffer.peek(), g_tagNameLength);
    buffer.retrieve(g_tagNameLength);
    if (header != TAG_NAME_X) {
        LOG_WARN << "Illegal Message HEADER:" << header;
        return false;
    }

    msg.headerInfo.header = header;

    {
        uint8_t type = 0;
        memcpy(&type, buffer.peek(), sizeof (uint8_t));
        buffer.retrieve(sizeof (uint8_t));
        msg.headerInfo.type = type;
    }

    {
        uint8_t code = 0;
        memcpy(&code, buffer.peek(), sizeof (uint8_t));
        buffer.retrieve(sizeof (uint8_t));
        msg.headerInfo.code = code;
    }

    {
        uint32_t contentLength = buffer.peekUInt32();
        buffer.retrieveUInt32();
        if (contentLength > buffer.readableBytes()) {
            return false; // At this point, it indicates that the data is not complete and we need to continue waiting
        }
        msg.headerInfo.contentLength = contentLength;
    }

    {
        Buffer contentBuffer;
        contentBuffer.append(buffer.peek(), msg.headerInfo.contentLength);
        buffer.retrieve(msg.headerInfo.contentLength);

        {
            uint8_t flagVal = 0;
            memcpy(&flagVal, contentBuffer.peek(), sizeof (uint8_t));
            msg.flag = flagVal;
            contentBuffer.retrieve(sizeof (uint8_t));
        }

        {
            uint32_t ipValue = contentBuffer.peekUInt32();
            contentBuffer.retrieveUInt32();
            msg.ip = Utils::fromIPv4Address(ipValue);
        }

        {
            msg.port = contentBuffer.peekUInt16();
            contentBuffer.retrieveUInt16();
        }

        {
            msg.clientID = std::string(contentBuffer.peek(), 46);
            contentBuffer.retrieve(46);
        }

        {
            msg.fileSize = contentBuffer.peekUInt64();
            contentBuffer.retrieveUInt64();
        }

        {
            msg.timeStamp = contentBuffer.peekUInt64();
            contentBuffer.retrieveUInt64();
        }

        {
            uint32_t dataLen = contentBuffer.peekUInt32();
            contentBuffer.retrieveUInt32();
            std::string fileName_utf16(contentBuffer.peek(), dataLen);
            contentBuffer.retrieve(dataLen);
            msg.fileName = Utils::toUtf8(Utils::stringToWString(fileName_utf16));
        }

        do {
            uint32_t filesCount = contentBuffer.peekUInt32();
            contentBuffer.retrieveUInt32();
            if (filesCount == 0) {
                break;
            }
            for (uint32_t index = 0; index < filesCount; ++index) {
                uint16_t dataLen = contentBuffer.peekUInt16();
                contentBuffer.retrieveUInt16();
                std::string filePathData_utf16(contentBuffer.peek(), dataLen);
                msg.filePathVec.push_back(Utils::toUtf8(Utils::stringToWString(filePathData_utf16)));
                contentBuffer.retrieve(dataLen);
            }
        } while (false);
    }
    return true;
}

//------------------------------------------

std::string SendFileResponseMsg::toByteArray(const SendFileResponseMsg &msg)
{
    Buffer data;
    data.append(msg.headerInfo.header);
    data.append(&msg.headerInfo.type, sizeof (msg.headerInfo.type));
    data.append(&msg.headerInfo.code, sizeof (msg.headerInfo.code));

    {
        Buffer tmpBuffer;
        assert(sizeof (msg.statusCode) == 1);
        tmpBuffer.append(&msg.statusCode, sizeof (msg.statusCode));
        uint32_t ipValue = Utils::toIPv4Address(msg.ip);
        tmpBuffer.appendUInt32(ipValue);
        tmpBuffer.appendUInt16(msg.port);
        assert(msg.clientID.length() == 46);
        tmpBuffer.append(msg.clientID);
        tmpBuffer.appendUInt64(msg.fileSize);
        tmpBuffer.appendUInt64(msg.timeStamp);
        tmpBuffer.append(Utils::toUtf16LE(msg.fileName));

        // Processing content length
        data.appendUInt32(static_cast<uint32_t>(tmpBuffer.readableBytes()));
        data.append(tmpBuffer.retrieveAllAsString());
    }
    return data.retrieveAllAsString();
}

bool SendFileResponseMsg::fromByteArray(const std::string &data, SendFileResponseMsg &msg)
{
    if (static_cast<int>(data.length()) < MsgHeader::messageLength()) {
        return false;
    }
    Buffer buffer;
    buffer.append(data);
    std::string header = std::string(buffer.peek(), g_tagNameLength);
    buffer.retrieve(g_tagNameLength);
    if (header != TAG_NAME_X) {
        LOG_WARN << "Illegal Message HEADER:" << header;
        return false;
    }

    msg.headerInfo.header = header;

    {
        uint8_t type = 0;
        memcpy(&type, buffer.peek(), sizeof (uint8_t));
        buffer.retrieve(sizeof (uint8_t));
        msg.headerInfo.type = type;
    }

    {
        uint8_t code = 0;
        memcpy(&code, buffer.peek(), sizeof (uint8_t));
        buffer.retrieve(sizeof (uint8_t));
        msg.headerInfo.code = code;
    }

    {
        uint32_t contentLength = buffer.peekUInt32();
        buffer.retrieveUInt32();
        if (contentLength > buffer.readableBytes()) {
            return false; // At this point, it indicates that the data is not complete and we need to continue waiting
        }
        msg.headerInfo.contentLength = contentLength;
    }

    {
        Buffer contentBuffer;
        contentBuffer.append(buffer.peek(), msg.headerInfo.contentLength);
        buffer.retrieve(msg.headerInfo.contentLength);

        {
            assert(sizeof (msg.statusCode) == sizeof (uint8_t));
            uint8_t statusCode = 0;
            memcpy(&statusCode, contentBuffer.peek(), sizeof (uint8_t));
            contentBuffer.retrieve(sizeof (uint8_t));
            msg.statusCode = statusCode;
        }

        {
            uint32_t ipValue = contentBuffer.peekUInt32();
            contentBuffer.retrieveUInt32();
            msg.ip = Utils::fromIPv4Address(ipValue);
        }

        {
            msg.port = contentBuffer.peekUInt16();
            contentBuffer.retrieveUInt16();
        }

        {
            msg.clientID = std::string(contentBuffer.peek(), 46);
            contentBuffer.retrieve(46);
        }

        {
            msg.fileSize = contentBuffer.peekUInt64();
            contentBuffer.retrieveUInt64();
        }

        {
            msg.timeStamp = contentBuffer.peekUInt64();
            contentBuffer.retrieveUInt64();
        }

        {
            msg.fileName = Utils::toUtf8(Utils::stringToWString(contentBuffer.retrieveAllAsString()));
        }
    }
    return true;
}

//---------------------------------------------------

std::string UpdateProgressMsg::toByteArray(const UpdateProgressMsg &msg)
{
    Buffer data;
    data.append(msg.headerInfo.header);
    data.append(&msg.headerInfo.type, sizeof (msg.headerInfo.type));
    data.append(&msg.headerInfo.code, sizeof (msg.headerInfo.code));

    {
        Buffer tmpBuffer;
        tmpBuffer.append(&msg.functionCode, sizeof (msg.functionCode));
        uint32_t ipValue = Utils::toIPv4Address(msg.ip);
        tmpBuffer.appendUInt32(ipValue);
        tmpBuffer.appendUInt16(msg.port);
        assert(msg.clientID.length() == g_clientIDLength);
        tmpBuffer.append(msg.clientID);
        tmpBuffer.appendUInt64(msg.timeStamp);
        tmpBuffer.appendUInt64(msg.fileSize);
        tmpBuffer.appendUInt64(msg.sentSize);

        {
            std::wstring fileName_utf16 = Utils::toUtf16LE(msg.fileName);
            tmpBuffer.appendUInt32(static_cast<uint32_t>(fileName_utf16.size() * sizeof (wchar_t)));
            tmpBuffer.append(fileName_utf16);
        }

        {
            std::wstring currentFileName_utf16 = Utils::toUtf16LE(msg.currentFileName);
            tmpBuffer.appendUInt32(static_cast<uint32_t>(currentFileName_utf16.size() * sizeof (wchar_t)));
            tmpBuffer.append(currentFileName_utf16);
        }
        tmpBuffer.appendUInt32(msg.sentFilesCount);
        tmpBuffer.appendUInt32(msg.totalFilesCount);
        tmpBuffer.appendUInt64(msg.currentFileSize);
        tmpBuffer.appendUInt64(msg.totalFilesSize);
        tmpBuffer.appendUInt64(msg.totalSentSize);

        // Processing content length
        data.appendUInt32(static_cast<uint32_t>(tmpBuffer.readableBytes()));
        data.append(tmpBuffer.retrieveAllAsString());
    }
    return data.retrieveAllAsString();
}

bool UpdateProgressMsg::fromByteArray(const std::string &data, UpdateProgressMsg &msg)
{
    if (static_cast<int>(data.length()) < MsgHeader::messageLength()) {
        return false;
    }
    Buffer buffer;
    buffer.append(data);
    std::string header = std::string(buffer.peek(), g_tagNameLength);
    buffer.retrieve(g_tagNameLength);
    if (header != TAG_NAME_X) {
        LOG_WARN << "Illegal Message HEADER:" << header;
        return false;
    }

    msg.headerInfo.header = header;

    {
        uint8_t type = 0;
        memcpy(&type, buffer.peek(), sizeof (uint8_t));
        buffer.retrieve(sizeof (uint8_t));
        msg.headerInfo.type = type;
    }

    {
        uint8_t code = 0;
        memcpy(&code, buffer.peek(), sizeof (uint8_t));
        buffer.retrieve(sizeof (uint8_t));
        msg.headerInfo.code = code;
    }

    {
        uint32_t contentLength = buffer.peekUInt32();
        buffer.retrieveUInt32();
        if (contentLength > buffer.readableBytes()) {
            return false; // At this point, it indicates that the data is not complete and we need to continue waiting
        }
        msg.headerInfo.contentLength = contentLength;
    }

    {
        Buffer contentBuffer;
        contentBuffer.append(buffer.peek(), msg.headerInfo.contentLength);
        buffer.retrieve(msg.headerInfo.contentLength);

        {
            uint8_t funcCode = 0;
            memcpy(&funcCode, contentBuffer.peek(), sizeof (uint8_t));
            msg.functionCode = funcCode;
            contentBuffer.retrieve(sizeof (uint8_t));
        }

        {
            uint32_t ipValue = contentBuffer.peekUInt32();
            contentBuffer.retrieveUInt32();
            msg.ip = Utils::fromIPv4Address(ipValue);
        }

        {
            msg.port = contentBuffer.peekUInt16();
            contentBuffer.retrieveUInt16();
        }

        {
            msg.clientID = std::string(contentBuffer.peek(), 46);
            contentBuffer.retrieve(46);
        }

        {
            msg.timeStamp = contentBuffer.peekUInt64();
            contentBuffer.retrieveUInt64();
        }

        {
            msg.fileSize = contentBuffer.peekUInt64();
            contentBuffer.retrieveUInt64();
        }

        {
            msg.sentSize = contentBuffer.peekUInt64();
            contentBuffer.retrieveUInt64();
        }

        {
            uint32_t dataLen = contentBuffer.peekUInt32();
            contentBuffer.retrieveUInt32();
            std::string fileName_utf16 = std::string(contentBuffer.peek(), dataLen);
            contentBuffer.retrieve(dataLen);
            msg.fileName = Utils::toUtf8(Utils::stringToWString(fileName_utf16));
        }

        // ----------------- MultiFile

        {
            uint32_t dataLen = contentBuffer.peekUInt32();
            contentBuffer.retrieveUInt32();
            std::string currentFileName_utf16 = std::string(contentBuffer.peek(), dataLen);
            contentBuffer.retrieve(dataLen);
            msg.currentFileName = Utils::toUtf8(Utils::stringToWString(currentFileName_utf16));
        }

        {
            msg.sentFilesCount = contentBuffer.peekUInt32();
            contentBuffer.retrieveUInt32();
        }

        {
            msg.totalFilesCount = contentBuffer.peekUInt32();
            contentBuffer.retrieveUInt32();
        }

        {
            msg.currentFileSize = contentBuffer.peekUInt64();
            contentBuffer.retrieveUInt64();
        }

        {
            msg.totalFilesSize = contentBuffer.peekUInt64();
            contentBuffer.retrieveUInt64();
        }

        {
            msg.totalSentSize = contentBuffer.peekUInt64();
            contentBuffer.retrieveUInt64();
        }
    }
    return true;
}

//-------------------------------------

std::string GetConnStatusRequestMsg::toByteArray(const GetConnStatusRequestMsg &msg)
{
    Buffer data;
    data.append(msg.headerInfo.header);
    data.append(&msg.headerInfo.type, sizeof (msg.headerInfo.type));
    data.append(&msg.headerInfo.code, sizeof (msg.headerInfo.code));

    {
        uint32_t contentLength = 0;
        data.appendUInt32(contentLength);
    }
    return data.retrieveAllAsString();
}

bool GetConnStatusRequestMsg::fromByteArray(const std::string &data, GetConnStatusRequestMsg &msg)
{
    if (static_cast<int>(data.length()) < MsgHeader::messageLength()) {
        return false;
    }
    Buffer buffer;
    buffer.append(data);
    std::string header = std::string(buffer.peek(), g_tagNameLength);
    buffer.retrieve(g_tagNameLength);
    if (header != TAG_NAME_X) {
        LOG_WARN << "Illegal Message HEADER:" << header;
        return false;
    }

    msg.headerInfo.header = header;

    {
        uint8_t type = 0;
        memcpy(&type, buffer.peek(), sizeof (uint8_t));
        buffer.retrieve(sizeof (uint8_t));
        msg.headerInfo.type = type;
    }

    {
        uint8_t code = 0;
        memcpy(&code, buffer.peek(), sizeof (uint8_t));
        buffer.retrieve(sizeof (uint8_t));
        msg.headerInfo.code = code;
    }

    // Assign a value of 0 directly here
    msg.headerInfo.contentLength = 0;

    return true;
}

// ----------------


std::string GetConnStatusResponseMsg::toByteArray(const GetConnStatusResponseMsg &msg)
{
    Buffer data;
    data.append(msg.headerInfo.header);
    data.append(&msg.headerInfo.type, sizeof (msg.headerInfo.type));
    data.append(&msg.headerInfo.code, sizeof (msg.headerInfo.code));

    {
        Buffer tmpBuffer;
        tmpBuffer.append(&msg.statusCode, sizeof (msg.statusCode));

        // Processing content length
        data.appendUInt32(static_cast<uint32_t>(tmpBuffer.readableBytes()));
        data.append(tmpBuffer.retrieveAllAsString());
    }
    return data.retrieveAllAsString();
}

bool GetConnStatusResponseMsg::fromByteArray(const std::string &data, GetConnStatusResponseMsg &msg)
{
    if (static_cast<int>(data.length()) < MsgHeader::messageLength()) {
        return false;
    }
    Buffer buffer;
    buffer.append(data);
    std::string header = std::string(buffer.peek(), g_tagNameLength);
    buffer.retrieve(g_tagNameLength);
    if (header != TAG_NAME_X) {
        LOG_WARN << "Illegal Message HEADER:" << header;
        return false;
    }

    msg.headerInfo.header = header;

    {
        uint8_t type = 0;
        memcpy(&type, buffer.peek(), sizeof (uint8_t));
        buffer.retrieve(sizeof (uint8_t));
        msg.headerInfo.type = type;
    }

    {
        uint8_t code = 0;
        memcpy(&code, buffer.peek(), sizeof (uint8_t));
        buffer.retrieve(sizeof (uint8_t));
        msg.headerInfo.code = code;
    }

    {
        uint32_t contentLength = buffer.peekUInt32();
        buffer.retrieveUInt32();
        if (contentLength > buffer.readableBytes()) {
            return false; // At this point, it indicates that the data is not complete and we need to continue waiting
        }
        msg.headerInfo.contentLength = contentLength;
    }

    {
        Buffer contentBuffer;
        contentBuffer.append(buffer.peek(), msg.headerInfo.contentLength);
        buffer.retrieve(msg.headerInfo.contentLength);

        {
            uint8_t statusCode = 0;
            memcpy(&statusCode, contentBuffer.peek(), sizeof (uint8_t));
            contentBuffer.retrieve(sizeof (uint8_t));
            msg.statusCode = statusCode;
        }
    }

    return true;
}

// ----------------------------------------------------------------


std::string UpdateImageProgressMsg::toByteArray(const UpdateImageProgressMsg &msg)
{
    Buffer data;
    data.append(msg.headerInfo.header);
    data.append(&msg.headerInfo.type, sizeof (msg.headerInfo.type));
    data.append(&msg.headerInfo.code, sizeof (msg.headerInfo.code));

    {
        Buffer tmpBuffer;
        uint32_t ipValue = Utils::toIPv4Address(msg.ip);
        tmpBuffer.appendUInt32(ipValue);
        tmpBuffer.appendUInt16(msg.port);
        assert(msg.clientID.length() == 46);
        tmpBuffer.append(msg.clientID);
        tmpBuffer.appendUInt64(msg.fileSize);
        tmpBuffer.appendUInt64(msg.sentSize);
        tmpBuffer.appendUInt64(msg.timeStamp);

        // Processing content length
        data.appendUInt32(static_cast<uint32_t>(tmpBuffer.readableBytes()));
        data.append(tmpBuffer.retrieveAllAsString());
    }
    return data.retrieveAllAsString();
}

bool UpdateImageProgressMsg::fromByteArray(const std::string &data, UpdateImageProgressMsg &msg)
{
    if (static_cast<int>(data.length()) < MsgHeader::messageLength()) {
        return false;
    }
    Buffer buffer;
    buffer.append(data);
    std::string header = std::string(buffer.peek(), g_tagNameLength);
    buffer.retrieve(g_tagNameLength);
    if (header != TAG_NAME_X) {
        LOG_WARN << "Illegal Message HEADER:" << header;
        return false;
    }

    msg.headerInfo.header = header;

    {
        uint8_t type = 0;
        memcpy(&type, buffer.peek(), sizeof (uint8_t));
        buffer.retrieve(sizeof (uint8_t));
        msg.headerInfo.type = type;
    }

    {
        uint8_t code = 0;
        memcpy(&code, buffer.peek(), sizeof (uint8_t));
        buffer.retrieve(sizeof (uint8_t));
        msg.headerInfo.code = code;
    }

    {
        uint32_t contentLength = buffer.peekUInt32();
        buffer.retrieveUInt32();
        if (contentLength > buffer.readableBytes()) {
            return false; // At this point, it indicates that the data is not complete and we need to continue waiting
        }
        msg.headerInfo.contentLength = contentLength;
    }

    {
        Buffer contentBuffer;
        contentBuffer.append(buffer.peek(), msg.headerInfo.contentLength);
        buffer.retrieve(msg.headerInfo.contentLength);

        {
            uint32_t ipValue = contentBuffer.peekUInt32();
            contentBuffer.retrieveUInt32();
            msg.ip = Utils::fromIPv4Address(ipValue);
        }

        {
            msg.port = contentBuffer.peekUInt16();
            contentBuffer.retrieveUInt16();
        }

        {
            msg.clientID = std::string(contentBuffer.peek(), 46);
            contentBuffer.retrieve(46);
        }

        {
            msg.fileSize = contentBuffer.peekUInt64();
            contentBuffer.retrieveUInt64();
        }

        {
            msg.sentSize = contentBuffer.peekUInt64();
            contentBuffer.retrieveUInt64();
        }

        {
            msg.timeStamp = contentBuffer.peekUInt64();
            contentBuffer.retrieveUInt64();
        }
        assert(contentBuffer.readableBytes() == 0);
    }
    return true;
}


//--------------------------------------------------------

std::string NotifyMessage::toByteArray(const NotifyMessage &msg)
{
    Buffer data;
    data.append(msg.headerInfo.header);
    data.append(&msg.headerInfo.type, sizeof (msg.headerInfo.type));
    data.append(&msg.headerInfo.code, sizeof (msg.headerInfo.code));

    {
        Buffer tmpBuffer;
        tmpBuffer.appendUInt64(msg.timeStamp);
        tmpBuffer.append(&msg.notiCode, sizeof (uint8_t));
        for (const auto &paramInfo : msg.paramInfoVec) {
            std::wstring messageData = Utils::toUtf16LE(paramInfo.info);
            tmpBuffer.appendUInt32(static_cast<uint32_t>(messageData.size() * sizeof (wchar_t)));
            tmpBuffer.append(messageData);
        }

        // Processing content length
        data.appendUInt32(static_cast<uint32_t>(tmpBuffer.readableBytes()));
        data.append(tmpBuffer.retrieveAllAsString());
    }
    return data.retrieveAllAsString();
}

bool NotifyMessage::fromByteArray(const std::string &data, NotifyMessage &msg)
{
    if (static_cast<int>(data.length()) < MsgHeader::messageLength()) {
        return false;
    }
    Buffer buffer;
    buffer.append(data);
    std::string header = std::string(buffer.peek(), g_tagNameLength);
    buffer.retrieve(g_tagNameLength);
    if (header != TAG_NAME_X) {
        LOG_WARN << "Illegal Message HEADER:" << header;
        return false;
    }

    msg.headerInfo.header = header;

    {
        uint8_t type = 0;
        memcpy(&type, buffer.peek(), sizeof (uint8_t));
        buffer.retrieve(sizeof (uint8_t));
        msg.headerInfo.type = type;
    }

    {
        uint8_t code = 0;
        memcpy(&code, buffer.peek(), sizeof (uint8_t));
        buffer.retrieve(sizeof (uint8_t));
        msg.headerInfo.code = code;
    }

    {
        uint32_t contentLength = buffer.peekUInt32();
        buffer.retrieveUInt32();
        if (contentLength > buffer.readableBytes()) {
            return false; // At this point, it indicates that the data is not complete and we need to continue waiting
        }
        msg.headerInfo.contentLength = contentLength;
    }

    {
        Buffer contentBuffer;
        contentBuffer.append(buffer.peek(), msg.headerInfo.contentLength);
        buffer.retrieve(msg.headerInfo.contentLength);

        {
            msg.timeStamp = contentBuffer.peekUInt64();
            contentBuffer.retrieveUInt64();
        }

        {
            memcpy(&msg.notiCode, contentBuffer.peek(), sizeof (uint8_t));
            contentBuffer.retrieve(sizeof (uint8_t));
        }

        while (contentBuffer.readableBytes() > 0) {
            assert(contentBuffer.readableBytes() >= sizeof (uint32_t));
            NotifyMessage::ParamInfo paramInfo;
            uint32_t notiLength = contentBuffer.readUInt32();
            assert(notiLength <= contentBuffer.readableBytes());
            paramInfo.info = Utils::toUtf8(Utils::stringToWString(contentBuffer.retrieveAsString(notiLength)));
            msg.paramInfoVec.push_back(paramInfo);
        }
    }

    return true;
}

//---------------------------------------UpdateSystemInfoMsg----------------

std::string UpdateSystemInfoMsg::toByteArray(const UpdateSystemInfoMsg &msg)
{
    Buffer data;
    data.append(msg.headerInfo.header);
    data.append(&msg.headerInfo.type, sizeof (msg.headerInfo.type));
    data.append(&msg.headerInfo.code, sizeof (msg.headerInfo.code));

    {
        Buffer tmpBuffer;
        uint32_t ipValue = Utils::toIPv4Address(msg.ip);
        tmpBuffer.appendUInt32(ipValue);
        tmpBuffer.appendUInt16(msg.port);
        tmpBuffer.append(Utils::toUtf16LE(msg.serverVersion));

        // Processing content length
        data.appendUInt32(static_cast<uint32_t>(tmpBuffer.readableBytes()));
        data.append(tmpBuffer.retrieveAllAsString());
    }
    return data.retrieveAllAsString();
}

bool UpdateSystemInfoMsg::fromByteArray(const std::string &data, UpdateSystemInfoMsg &msg)
{
    if (static_cast<int>(data.length()) < MsgHeader::messageLength()) {
        return false;
    }
    Buffer buffer;
    buffer.append(data);
    std::string header = std::string(buffer.peek(), g_tagNameLength);
    buffer.retrieve(g_tagNameLength);
    if (header != TAG_NAME_X) {
        LOG_WARN << "Illegal Message HEADER:" << header;
        return false;
    }

    msg.headerInfo.header = header;

    {
        uint8_t type = 0;
        memcpy(&type, buffer.peek(), sizeof (uint8_t));
        buffer.retrieve(sizeof (uint8_t));
        msg.headerInfo.type = type;
    }

    {
        uint8_t code = 0;
        memcpy(&code, buffer.peek(), sizeof (uint8_t));
        buffer.retrieve(sizeof (uint8_t));
        msg.headerInfo.code = code;
    }

    {
        uint32_t contentLength = buffer.peekUInt32();
        buffer.retrieveUInt32();
        if (contentLength > buffer.readableBytes()) {
            return false; // At this point, it indicates that the data is not complete and we need to continue waiting
        }
        msg.headerInfo.contentLength = contentLength;
    }

    {
        Buffer contentBuffer;
        contentBuffer.append(buffer.peek(), msg.headerInfo.contentLength);
        buffer.retrieve(msg.headerInfo.contentLength);

        {
            uint32_t ipValue = contentBuffer.peekUInt32();
            contentBuffer.retrieveUInt32();
            msg.ip = Utils::fromIPv4Address(ipValue);
        }

        {
            msg.port = contentBuffer.peekUInt16();
            contentBuffer.retrieveUInt16();
        }

        {
            msg.serverVersion = Utils::toUtf8(Utils::stringToWString(contentBuffer.retrieveAllAsString()));
        }
    }
    return true;
}

//---------------------------------------DDCCINotifyMsg----------------

std::string DDCCINotifyMsg::toByteArray(const DDCCINotifyMsg &msg)
{
    Buffer data;
    data.append(msg.headerInfo.header);
    data.append(&msg.headerInfo.type, sizeof (msg.headerInfo.type));
    data.append(&msg.headerInfo.code, sizeof (msg.headerInfo.code));

    {
        Buffer tmpBuffer;
        tmpBuffer.append(&msg.functionCode, sizeof (msg.functionCode));
        {
            tmpBuffer.appendUInt32(static_cast<uint32_t>(msg.macAddress.size()));
            tmpBuffer.append(msg.macAddress.data(), msg.macAddress.size());
        }
        tmpBuffer.appendUInt16(msg.source);
        tmpBuffer.appendUInt16(msg.port);
        tmpBuffer.appendUInt32(msg.authResult);
        tmpBuffer.appendUInt32(msg.indexValue);

        // Processing content length
        data.appendUInt32(static_cast<uint32_t>(tmpBuffer.readableBytes()));
        data.append(tmpBuffer.retrieveAllAsString());
    }
    return data.retrieveAllAsString();
}

bool DDCCINotifyMsg::fromByteArray(const std::string &data, DDCCINotifyMsg &msg)
{
    if (static_cast<int>(data.length()) < MsgHeader::messageLength()) {
        return false;
    }
    Buffer buffer;
    buffer.append(data);
    std::string header = std::string(buffer.peek(), g_tagNameLength);
    buffer.retrieve(g_tagNameLength);
    if (header != TAG_NAME_X) {
        LOG_WARN << "Illegal message HEADER:" << header;
        return false;
    }

    msg.headerInfo.header = header;

    {
        uint8_t type = 0;
        memcpy(&type, buffer.peek(), sizeof (uint8_t));
        buffer.retrieve(sizeof (uint8_t));
        msg.headerInfo.type = type;
    }

    {
        uint8_t code = 0;
        memcpy(&code, buffer.peek(), sizeof (uint8_t));
        buffer.retrieve(sizeof (uint8_t));
        msg.headerInfo.code = code;
    }

    {
        uint32_t contentLength = buffer.peekUInt32();
        buffer.retrieveUInt32();
        if (contentLength > buffer.readableBytes()) {
            return false; // At this point, it indicates that the data is not complete and we need to continue waiting
        }
        msg.headerInfo.contentLength = contentLength;
    }

    {
        Buffer contentBuffer;
        contentBuffer.append(buffer.peek(), msg.headerInfo.contentLength);
        buffer.retrieve(msg.headerInfo.contentLength);

        {
            uint8_t funcCode = 0;
            memcpy(&funcCode, contentBuffer.peek(), sizeof (uint8_t));
            msg.functionCode = funcCode;
            contentBuffer.retrieve(sizeof (uint8_t));
        }

        {
            uint32_t macAddressLen = contentBuffer.peekUInt32();
            contentBuffer.retrieveUInt32();
            msg.macAddress = std::string(contentBuffer.peek(), macAddressLen);
            contentBuffer.retrieve(macAddressLen);
        }

        {
            msg.source = contentBuffer.peekUInt16();
            contentBuffer.retrieveUInt16();
        }

        {
            msg.port = contentBuffer.peekUInt16();
            contentBuffer.retrieveUInt16();
        }

        {
            msg.authResult = contentBuffer.peekUInt32();
            contentBuffer.retrieveUInt32();
        }

        {
            msg.indexValue = contentBuffer.peekUInt32();
            contentBuffer.retrieveUInt32();
        }
    }
    return true;
}

//---------------------------------------DragFilesMsg----------------
std::string DragFilesMsg::toByteArray(const DragFilesMsg &msg)
{
    Buffer data;
    data.append(msg.headerInfo.header);
    data.append(&msg.headerInfo.type, sizeof (msg.headerInfo.type));
    data.append(&msg.headerInfo.code, sizeof (msg.headerInfo.code));

    {
        Buffer tmpBuffer;
        tmpBuffer.append(&msg.functionCode, sizeof (msg.functionCode));
        tmpBuffer.appendUInt64(msg.timeStamp);

        {
            std::wstring filePath_utf16 = Utils::toUtf16LE(msg.rootPath);
            tmpBuffer.appendUInt16(static_cast<uint16_t>(filePath_utf16.size() * sizeof (wchar_t)));
            tmpBuffer.append(filePath_utf16);
        }

        do {
            uint32_t filesCount = static_cast<uint32_t>(msg.filePathVec.size());
            tmpBuffer.appendUInt32(filesCount);
            if (filesCount == 0) {
                break;
            }
            for (const auto &filePath : msg.filePathVec) {
                std::wstring filePath_utf16 = Utils::toUtf16LE(filePath);
                tmpBuffer.appendUInt16(static_cast<uint16_t>(filePath_utf16.size() * sizeof (wchar_t)));
                tmpBuffer.append(filePath_utf16);
            }
        } while (false);


        do {
            if (msg.functionCode != DragFilesMsg::FuncCode::ReceiveFileInfo && msg.functionCode != DragFilesMsg::FuncCode::CancelFileTransfer) {
                break;
            }

            uint32_t ipValue = Utils::toIPv4Address(msg.ip);
            tmpBuffer.appendUInt32(ipValue);
            tmpBuffer.appendUInt16(msg.port);
            assert(msg.clientID.length() == g_clientIDLength);
            tmpBuffer.append(msg.clientID);
            tmpBuffer.appendUInt32(msg.fileCount);
            tmpBuffer.appendUInt64(msg.totalFileSize);
            {
                std::wstring fileName_utf16 = Utils::toUtf16LE(msg.firstTransferFileName);
                tmpBuffer.appendUInt32(static_cast<uint32_t>(fileName_utf16.length() * sizeof (wchar_t)));
                tmpBuffer.append(fileName_utf16);
            }
            tmpBuffer.appendUInt64(msg.firstTransferFileSize);
        } while (false);

        // Processing content length
        data.appendUInt32(static_cast<uint32_t>(tmpBuffer.readableBytes()));
        data.append(tmpBuffer.retrieveAllAsString());
    }
    return data.retrieveAllAsString();
}

bool DragFilesMsg::fromByteArray(const std::string &data, DragFilesMsg &msg)
{
    if (static_cast<int>(data.length()) < MsgHeader::messageLength()) {
        return false;
    }
    Buffer buffer;
    buffer.append(data);
    std::string header = std::string(buffer.peek(), g_tagNameLength);
    buffer.retrieve(g_tagNameLength);
    if (header != TAG_NAME_X) {
        LOG_WARN << "Illegal message HEADER:" << header;
        return false;
    }

    msg.headerInfo.header = header;

    {
        uint8_t type = 0;
        memcpy(&type, buffer.peek(), sizeof (uint8_t));
        buffer.retrieve(sizeof (uint8_t));
        msg.headerInfo.type = type;
    }

    {
        uint8_t code = 0;
        memcpy(&code, buffer.peek(), sizeof (uint8_t));
        buffer.retrieve(sizeof (uint8_t));
        msg.headerInfo.code = code;
    }

    {
        uint32_t contentLength = buffer.peekUInt32();
        buffer.retrieveUInt32();
        if (contentLength > buffer.readableBytes()) {
            return false; // At this point, it indicates that the data is not complete and we need to continue waiting
        }
        msg.headerInfo.contentLength = contentLength;
    }

    {
        Buffer contentBuffer;
        contentBuffer.append(buffer.peek(), msg.headerInfo.contentLength);
        buffer.retrieve(msg.headerInfo.contentLength);

        {
            uint8_t funcCode = 0;
            memcpy(&funcCode, contentBuffer.peek(), sizeof (uint8_t));
            msg.functionCode = funcCode;
            contentBuffer.retrieve(sizeof (uint8_t));
        }

        {
            msg.timeStamp = contentBuffer.peekUInt64();
            contentBuffer.retrieveUInt64();
        }

        {
            uint16_t dataLen = contentBuffer.peekUInt16();
            contentBuffer.retrieveUInt16();
            std::string filePathData_utf16(contentBuffer.peek(), dataLen);
            msg.rootPath = Utils::toUtf8(Utils::stringToWString(filePathData_utf16));
            contentBuffer.retrieve(dataLen);
        }

        do {
            uint32_t filesCount = contentBuffer.peekUInt32();
            contentBuffer.retrieveUInt32();
            if (filesCount == 0) {
                break;
            }
            for (uint32_t index = 0; index < filesCount; ++index) {
                uint16_t dataLen = contentBuffer.peekUInt16();
                contentBuffer.retrieveUInt16();
                std::string filePathData_utf16(contentBuffer.peek(), dataLen);
                msg.filePathVec.push_back(Utils::toUtf8(Utils::stringToWString(filePathData_utf16)));
                contentBuffer.retrieve(dataLen);
            }
        } while (false);


        do {
            if (msg.functionCode != DragFilesMsg::FuncCode::ReceiveFileInfo && msg.functionCode != DragFilesMsg::FuncCode::CancelFileTransfer) {
                break;
            }

            {
                uint32_t ipValue = contentBuffer.peekUInt32();
                contentBuffer.retrieveUInt32();
                msg.ip = Utils::fromIPv4Address(ipValue);
            }

            {
                msg.port = contentBuffer.peekUInt16();
                contentBuffer.retrieveUInt16();
            }

            {
                msg.clientID = std::string(contentBuffer.peek(), g_clientIDLength);
                contentBuffer.retrieve(g_clientIDLength);
            }

            {
                msg.fileCount = contentBuffer.peekUInt32();
                contentBuffer.retrieveUInt32();
            }

            {
                msg.totalFileSize = contentBuffer.peekUInt64();
                contentBuffer.retrieveUInt64();
            }

            {
                uint32_t dataLen = contentBuffer.peekUInt32();
                contentBuffer.retrieveUInt32();
                std::string fileName_utf16 = std::string(contentBuffer.peek(), dataLen);
                contentBuffer.retrieve(dataLen);
                msg.firstTransferFileName = Utils::toUtf8(Utils::stringToWString(fileName_utf16));
            }

            {
                msg.firstTransferFileSize = contentBuffer.peekUInt64();
                contentBuffer.retrieveUInt64();
            }
        } while (false);
    }
    return true;
}

//---------------------------------------------------------------

std::string StatusInfoNotifyMsg::toByteArray(const StatusInfoNotifyMsg &msg)
{
    Buffer data;
    data.append(msg.headerInfo.header);
    data.append(&msg.headerInfo.type, sizeof (msg.headerInfo.type));
    data.append(&msg.headerInfo.code, sizeof (msg.headerInfo.code));

    {
        Buffer tmpBuffer;
        tmpBuffer.appendUInt32(msg.statusCode);

        // Processing content length
        data.appendUInt32(static_cast<uint32_t>(tmpBuffer.readableBytes()));
        data.append(tmpBuffer.retrieveAllAsString());
    }
    return data.retrieveAllAsString();
}

bool StatusInfoNotifyMsg::fromByteArray(const std::string &data, StatusInfoNotifyMsg &msg)
{
    if (static_cast<int>(data.length()) < MsgHeader::messageLength()) {
        return false;
    }
    Buffer buffer;
    buffer.append(data);
    std::string header = std::string(buffer.peek(), g_tagNameLength);
    buffer.retrieve(g_tagNameLength);
    if (header != TAG_NAME_X) {
        LOG_WARN << "Illegal message HEADER:" << header;
        return false;
    }

    msg.headerInfo.header = header;

    {
        uint8_t type = 0;
        memcpy(&type, buffer.peek(), sizeof (uint8_t));
        buffer.retrieve(sizeof (uint8_t));
        msg.headerInfo.type = type;
    }

    {
        uint8_t code = 0;
        memcpy(&code, buffer.peek(), sizeof (uint8_t));
        buffer.retrieve(sizeof (uint8_t));
        msg.headerInfo.code = code;
    }

    {
        uint32_t contentLength = buffer.peekUInt32();
        buffer.retrieveUInt32();
        if (contentLength > buffer.readableBytes()) {
            return false; // At this point, it indicates that the data is not complete and we need to continue waiting
        }
        msg.headerInfo.contentLength = contentLength;
    }

    {
        Buffer contentBuffer;
        contentBuffer.append(buffer.peek(), msg.headerInfo.contentLength);
        buffer.retrieve(msg.headerInfo.contentLength);

        {
            msg.statusCode = contentBuffer.peekUInt32();
            contentBuffer.retrieveUInt32();
        }
        assert(contentBuffer.readableBytes() == 0);
    }
    return true;
}

}
