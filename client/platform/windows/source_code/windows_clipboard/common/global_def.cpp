#include "global_def.h"
#include "common_utils.h"
#include "common_signals.h"
#include <windows.h>
#include "ddcci/crossShareDefine.h"
#include <QHostAddress>
#include <QDir>
#include <QCoreApplication>

const int g_tagNameLength = strlen(TAG_NAME);
const int g_clientIDLength = 46;
QString g_namedPipeServerName { "CrossSharePipe" };
const QString g_helperServerName { "CrossSharePipe" };

const QString g_drop_table_sql { R"(DROP TABLE IF EXISTS %1;)" };
const QString g_create_opt_record { R"(CREATE TABLE IF NOT EXISTS opt_record (id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL, uuid TEXT UNIQUE, json_data TEXT DEFAULT "" NOT NULL);)" };

void g_globalRegister()
{
    g_globalRegisterForJsonSerialize();
    g_globalRegisterForBoostSerialize();

    qRegisterMetaType<UpdateClientStatusMsgPtr>();
    qRegisterMetaType<SendFileRequestMsgPtr>();
    qRegisterMetaType<SendFileResponseMsgPtr>();
    qRegisterMetaType<UpdateProgressMsgPtr>();
    qRegisterMetaType<UpdateImageProgressMsgPtr>();

    qRegisterMetaType<GetConnStatusRequestMsgPtr>();
    qRegisterMetaType<GetConnStatusResponseMsgPtr>();
    qRegisterMetaType<SystemConfigPtr>();
    qRegisterMetaType<NotifyMessagePtr>();
    qRegisterMetaType<UpdateSystemInfoMsgPtr>();
    qRegisterMetaType<DDCCINotifyMsgPtr>();
    qRegisterMetaType<DragFilesMsgPtr>();
    qRegisterMetaType<StatusInfoNotifyMsgPtr>();
}

GlobalData *g_getGlobalData()
{
    static GlobalData s_data;
    return &s_data;
}

QByteArray RecordDataHash::getHashID() const
{
    std::stringstream str_stream;
    str_stream << QByteArray::fromStdString(clientID).toHex().toUpper().constData() << "_";
    str_stream << ip << "_";
    str_stream << timeStamp;
    //qInfo() << "[RecordDataHash]:" << str_stream.str().c_str();
    return QByteArray::number(std::hash<std::string>{}(str_stream.str()));
}

int MsgHeader::messageLength()
{
    int headerMsgLen = static_cast<int>(g_tagNameLength + sizeof (uint8_t) + sizeof (uint8_t) + sizeof (uint32_t));
    return headerMsgLen;
}

bool g_getCodeFromByteArray(const QByteArray &data, uint8_t &typeValue, uint8_t &codeValue)
{
    if (data.length() < MsgHeader::messageLength()) {
        return false;
    }
    Buffer buffer;
    buffer.append(data);
    QByteArray header = QByteArray(buffer.peek(), g_tagNameLength);
    buffer.retrieve(g_tagNameLength);
    if (header != TAG_NAME) {
        qWarning() << "Illegal message HEADER:" << header.constData();
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

bool g_getCodeFromByteArray(const QByteArray &data, uint8_t &codeValue)
{
    uint8_t typeValue = 0;
    return g_getCodeFromByteArray(data, typeValue, codeValue);
}

QList<QString> g_getPipeServerExePathList()
{
    QList<QString> pathList;
    pathList.append(qApp->applicationDirPath() + "/" + PIPE_SERVER_EXE_NAME);
    // This is for testing purposes
    pathList.append(qApp->applicationDirPath() + "/../test-server/test-server.exe");
    pathList.append(qApp->applicationDirPath() + "/test-server.exe");

    for (auto &pathVal : pathList) {
        pathVal = QDir(pathVal).absolutePath();
    }
    return pathList;
}

//-----------------------------------

QByteArray SendFileRequestMsg::toByteArray(const SendFileRequestMsg &msg)
{
    Buffer data;
    data.append(msg.headerInfo.header.toUtf8());
    data.append(&msg.headerInfo.type, sizeof (msg.headerInfo.type));
    data.append(&msg.headerInfo.code, sizeof (msg.headerInfo.code));

    {
        Buffer tmpBuffer;
        tmpBuffer.append(&msg.flag, sizeof (uint8_t));
        uint32_t ipValue = QHostAddress(msg.ip).toIPv4Address();
        tmpBuffer.appendUInt32(ipValue);
        tmpBuffer.appendUInt16(msg.port);
        Q_ASSERT(msg.clientID.length() == 46);
        tmpBuffer.append(msg.clientID);
        tmpBuffer.appendUInt64(msg.fileSize);
        tmpBuffer.appendUInt64(msg.timeStamp);
        {
            QByteArray fileName_utf16 = CommonUtils::toUtf16LE(msg.fileName);
            tmpBuffer.appendUInt32(static_cast<uint32_t>(fileName_utf16.length()));
            tmpBuffer.append(fileName_utf16);
        }

        do {
            uint32_t filesCount = static_cast<uint32_t>(msg.filePathVec.size());
            tmpBuffer.appendUInt32(filesCount);
            if (filesCount == 0) {
                break;
            }
            for (const auto &filePath : msg.filePathVec) {
                QByteArray filePath_utf16 = CommonUtils::toUtf16LE(filePath);
                tmpBuffer.appendUInt16(static_cast<uint16_t>(filePath_utf16.size()));
                tmpBuffer.append(filePath_utf16);
            }
        } while (false);

        // Processing content length
        data.appendUInt32(static_cast<uint32_t>(tmpBuffer.readableBytes()));
        data.append(tmpBuffer.retrieveAllAsByteArray());
    }
    return data.retrieveAllAsByteArray();
}

bool SendFileRequestMsg::fromByteArray(const QByteArray &data, SendFileRequestMsg &msg)
{
    if (data.length() < MsgHeader::messageLength()) {
        return false;
    }
    Buffer buffer;
    buffer.append(data);
    QByteArray header = QByteArray(buffer.peek(), g_tagNameLength);
    buffer.retrieve(g_tagNameLength);
    if (header != TAG_NAME) {
        qWarning() << "Illegal message HEADER:" << header.constData();
        return false;
    }

    msg.headerInfo.header = header.constData();

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
            msg.ip = QHostAddress(ipValue).toString();
        }

        {
            msg.port = contentBuffer.peekUInt16();
            contentBuffer.retrieveUInt16();
        }

        {
            msg.clientID = QByteArray(contentBuffer.peek(), 46);
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
            QByteArray fileName_utf16(contentBuffer.peek(), static_cast<int>(dataLen));
            msg.fileName = CommonUtils::toUtf8(fileName_utf16).constData();
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
                QByteArray filePathData_utf16(contentBuffer.peek(), dataLen);
                msg.filePathVec.push_back(CommonUtils::toUtf8(filePathData_utf16));
                contentBuffer.retrieve(dataLen);
            }
        } while (false);
    }
    return true;
}

//------------------------------------------

QByteArray SendFileResponseMsg::toByteArray(const SendFileResponseMsg &msg)
{
    Buffer data;
    data.append(msg.headerInfo.header.toUtf8());
    data.append(&msg.headerInfo.type, sizeof (msg.headerInfo.type));
    data.append(&msg.headerInfo.code, sizeof (msg.headerInfo.code));

    {
        Buffer tmpBuffer;
        Q_ASSERT(sizeof (msg.statusCode) == 1);
        tmpBuffer.append(&msg.statusCode, sizeof (msg.statusCode));
        uint32_t ipValue = QHostAddress(msg.ip).toIPv4Address();
        tmpBuffer.appendUInt32(ipValue);
        tmpBuffer.appendUInt16(msg.port);
        Q_ASSERT(msg.clientID.length() == 46);
        tmpBuffer.append(msg.clientID);
        tmpBuffer.appendUInt64(msg.fileSize);
        tmpBuffer.appendUInt64(msg.timeStamp);
        tmpBuffer.append(CommonUtils::toUtf16LE(msg.fileName));

        // Processing content length
        data.appendUInt32(static_cast<uint32_t>(tmpBuffer.readableBytes()));
        data.append(tmpBuffer.retrieveAllAsByteArray());
    }
    return data.retrieveAllAsByteArray();
}

bool SendFileResponseMsg::fromByteArray(const QByteArray &data, SendFileResponseMsg &msg)
{
    if (data.length() < MsgHeader::messageLength()) {
        return false;
    }
    Buffer buffer;
    buffer.append(data);
    QByteArray header = QByteArray(buffer.peek(), g_tagNameLength);
    buffer.retrieve(g_tagNameLength);
    if (header != TAG_NAME) {
        qWarning() << "Illegal message HEADER:" << header.constData();
        return false;
    }

    msg.headerInfo.header = header.constData();

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
            Q_ASSERT(sizeof (msg.statusCode) == sizeof (uint8_t));
            uint8_t statusCode = 0;
            memcpy(&statusCode, contentBuffer.peek(), sizeof (uint8_t));
            contentBuffer.retrieve(sizeof (uint8_t));
            msg.statusCode = statusCode;
        }

        {
            uint32_t ipValue = contentBuffer.peekUInt32();
            contentBuffer.retrieveUInt32();
            msg.ip = QHostAddress(ipValue).toString();
        }

        {
            msg.port = contentBuffer.peekUInt16();
            contentBuffer.retrieveUInt16();
        }

        {
            msg.clientID = QByteArray(contentBuffer.peek(), 46);
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
            msg.fileName = CommonUtils::toUtf8(contentBuffer.retrieveAllAsByteArray()).constData();
        }
    }
    return true;
}

//---------------------------------------------------

QByteArray UpdateProgressMsg::toByteArray(const UpdateProgressMsg &msg)
{
    Buffer data;
    data.append(msg.headerInfo.header.toUtf8());
    data.append(&msg.headerInfo.type, sizeof (msg.headerInfo.type));
    data.append(&msg.headerInfo.code, sizeof (msg.headerInfo.code));

    {
        Buffer tmpBuffer;
        tmpBuffer.append(&msg.functionCode, sizeof (msg.functionCode));
        uint32_t ipValue = QHostAddress(msg.ip).toIPv4Address();
        tmpBuffer.appendUInt32(ipValue);
        tmpBuffer.appendUInt16(msg.port);
        Q_ASSERT(msg.clientID.length() == g_clientIDLength);
        tmpBuffer.append(msg.clientID);
        tmpBuffer.appendUInt64(msg.timeStamp);
        tmpBuffer.appendUInt64(msg.fileSize);
        tmpBuffer.appendUInt64(msg.sentSize);

        {
            QByteArray fileName_utf16 = CommonUtils::toUtf16LE(msg.fileName);
            tmpBuffer.appendUInt32(static_cast<uint32_t>(fileName_utf16.size()));
            tmpBuffer.append(fileName_utf16);
        }

        {
            QByteArray currentFileName_utf16 = CommonUtils::toUtf16LE(msg.currentFileName);
            tmpBuffer.appendUInt32(static_cast<uint32_t>(currentFileName_utf16.size()));
            tmpBuffer.append(currentFileName_utf16);
        }
        tmpBuffer.appendUInt32(msg.sentFilesCount);
        tmpBuffer.appendUInt32(msg.totalFilesCount);
        tmpBuffer.appendUInt64(msg.currentFileSize);
        tmpBuffer.appendUInt64(msg.totalFilesSize);
        tmpBuffer.appendUInt64(msg.totalSentSize);


        // Processing content length
        data.appendUInt32(static_cast<uint32_t>(tmpBuffer.readableBytes()));
        data.append(tmpBuffer.retrieveAllAsByteArray());
    }
    return data.retrieveAllAsByteArray();
}

bool UpdateProgressMsg::fromByteArray(const QByteArray &data, UpdateProgressMsg &msg)
{
    if (data.length() < MsgHeader::messageLength()) {
        return false;
    }
    Buffer buffer;
    buffer.append(data);
    QByteArray header = QByteArray(buffer.peek(), g_tagNameLength);
    buffer.retrieve(g_tagNameLength);
    if (header != TAG_NAME) {
        qWarning() << "Illegal message HEADER:" << header.constData();
        return false;
    }

    msg.headerInfo.header = header.constData();

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
            msg.ip = QHostAddress(ipValue).toString();
        }

        {
            msg.port = contentBuffer.peekUInt16();
            contentBuffer.retrieveUInt16();
        }

        {
            msg.clientID = QByteArray(contentBuffer.peek(), 46);
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
            QByteArray fileName_utf16 = QByteArray(contentBuffer.peek(), dataLen);
            contentBuffer.retrieve(dataLen);
            msg.fileName = CommonUtils::toUtf8(fileName_utf16);
        }

        // ----------------- MultiFile

        {
            uint32_t dataLen = contentBuffer.peekUInt32();
            contentBuffer.retrieveUInt32();
            QByteArray currentFileName_utf16 = QByteArray(contentBuffer.peek(), dataLen);
            contentBuffer.retrieve(dataLen);
            msg.currentFileName = CommonUtils::toUtf8(currentFileName_utf16);
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

QByteArray GetConnStatusRequestMsg::toByteArray(const GetConnStatusRequestMsg &msg)
{
    Buffer data;
    data.append(msg.headerInfo.header.toUtf8());
    data.append(&msg.headerInfo.type, sizeof (msg.headerInfo.type));
    data.append(&msg.headerInfo.code, sizeof (msg.headerInfo.code));

    {
        uint32_t contentLength = 0;
        data.appendUInt32(contentLength);
    }
    return data.retrieveAllAsByteArray();
}

bool GetConnStatusRequestMsg::fromByteArray(const QByteArray &data, GetConnStatusRequestMsg &msg)
{
    if (data.length() < MsgHeader::messageLength()) {
        return false;
    }
    Buffer buffer;
    buffer.append(data);
    QByteArray header = QByteArray(buffer.peek(), g_tagNameLength);
    buffer.retrieve(g_tagNameLength);
    if (header != TAG_NAME) {
        qWarning() << "Illegal message HEADER:" << header.constData();
        return false;
    }

    msg.headerInfo.header = header.constData();

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


QByteArray GetConnStatusResponseMsg::toByteArray(const GetConnStatusResponseMsg &msg)
{
    Buffer data;
    data.append(msg.headerInfo.header.toUtf8());
    data.append(&msg.headerInfo.type, sizeof (msg.headerInfo.type));
    data.append(&msg.headerInfo.code, sizeof (msg.headerInfo.code));

    {
        Buffer tmpBuffer;
        tmpBuffer.append(&msg.statusCode, sizeof (msg.statusCode));

        // Processing content length
        data.appendUInt32(static_cast<uint32_t>(tmpBuffer.readableBytes()));
        data.append(tmpBuffer.retrieveAllAsByteArray());
    }
    return data.retrieveAllAsByteArray();
}

bool GetConnStatusResponseMsg::fromByteArray(const QByteArray &data, GetConnStatusResponseMsg &msg)
{
    if (data.length() < MsgHeader::messageLength()) {
        return false;
    }
    Buffer buffer;
    buffer.append(data);
    QByteArray header = QByteArray(buffer.peek(), g_tagNameLength);
    buffer.retrieve(g_tagNameLength);
    if (header != TAG_NAME) {
        qWarning() << "Illegal message HEADER:" << header.constData();
        return false;
    }

    msg.headerInfo.header = header.constData();

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


QByteArray UpdateImageProgressMsg::toByteArray(const UpdateImageProgressMsg &msg)
{
    Buffer data;
    data.append(msg.headerInfo.header.toUtf8());
    data.append(&msg.headerInfo.type, sizeof (msg.headerInfo.type));
    data.append(&msg.headerInfo.code, sizeof (msg.headerInfo.code));

    {
        Buffer tmpBuffer;
        uint32_t ipValue = QHostAddress(msg.ip).toIPv4Address();
        tmpBuffer.appendUInt32(ipValue);
        tmpBuffer.appendUInt16(msg.port);
        Q_ASSERT(msg.clientID.length() == 46);
        tmpBuffer.append(msg.clientID);
        tmpBuffer.appendUInt64(msg.fileSize);
        tmpBuffer.appendUInt64(msg.sentSize);
        tmpBuffer.appendUInt64(msg.timeStamp);

        // Processing content length
        data.appendUInt32(static_cast<uint32_t>(tmpBuffer.readableBytes()));
        data.append(tmpBuffer.retrieveAllAsByteArray());
    }
    return data.retrieveAllAsByteArray();
}

bool UpdateImageProgressMsg::fromByteArray(const QByteArray &data, UpdateImageProgressMsg &msg)
{
    if (data.length() < MsgHeader::messageLength()) {
        return false;
    }
    Buffer buffer;
    buffer.append(data);
    QByteArray header = QByteArray(buffer.peek(), g_tagNameLength);
    buffer.retrieve(g_tagNameLength);
    if (header != TAG_NAME) {
        qWarning() << "Illegal message HEADER:" << header.constData();
        return false;
    }

    msg.headerInfo.header = header.constData();

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
            msg.ip = QHostAddress(ipValue).toString();
        }

        {
            msg.port = contentBuffer.peekUInt16();
            contentBuffer.retrieveUInt16();
        }

        {
            msg.clientID = QByteArray(contentBuffer.peek(), 46);
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
        Q_ASSERT(contentBuffer.readableBytes() == 0);
    }
    return true;
}


//--------------------------------------------------------

QByteArray NotifyMessage::toByteArray(const NotifyMessage &msg)
{
    Buffer data;
    data.append(msg.headerInfo.header.toUtf8());
    data.append(&msg.headerInfo.type, sizeof (msg.headerInfo.type));
    data.append(&msg.headerInfo.code, sizeof (msg.headerInfo.code));

    {
        Buffer tmpBuffer;
        tmpBuffer.appendUInt64(msg.timeStamp);
        tmpBuffer.append(&msg.notiCode, sizeof (uint8_t));
        for (const auto &paramInfo : msg.paramInfoVec) {
            QByteArray messageData = CommonUtils::toUtf16LE(paramInfo.info);
            tmpBuffer.appendUInt32(static_cast<uint32_t>(messageData.size()));
            tmpBuffer.append(messageData);
        }

        // Processing content length
        data.appendUInt32(static_cast<uint32_t>(tmpBuffer.readableBytes()));
        data.append(tmpBuffer.retrieveAllAsByteArray());
    }
    return data.retrieveAllAsByteArray();
}

bool NotifyMessage::fromByteArray(const QByteArray &data, NotifyMessage &msg)
{
    if (data.length() < MsgHeader::messageLength()) {
        return false;
    }
    Buffer buffer;
    buffer.append(data);
    QByteArray header = QByteArray(buffer.peek(), g_tagNameLength);
    buffer.retrieve(g_tagNameLength);
    if (header != TAG_NAME) {
        qWarning() << "Illegal message HEADER:" << header.constData();
        return false;
    }

    msg.headerInfo.header = header.constData();

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
            Q_ASSERT(contentBuffer.readableBytes() >= sizeof (uint32_t));
            NotifyMessage::ParamInfo paramInfo;
            uint32_t notiLength = contentBuffer.readUInt32();
            Q_ASSERT(notiLength <= contentBuffer.readableBytes());
            paramInfo.info = CommonUtils::toUtf8(contentBuffer.retrieveAsByteArray(notiLength)).constData();
            msg.paramInfoVec.push_back(paramInfo);
        }
    }

    return true;
}

nlohmann::json NotifyMessage::toString() const
{
    nlohmann::json infoJson;
    infoJson["timestamp"] = QDateTime::fromMSecsSinceEpoch(timeStamp).toString("yyyy-MM-dd hh:mm:ss.zzz").toStdString();
    infoJson["notiCode"] = notiCode;

    switch (notiCode) {
    case ConnSuccess: {
        Q_ASSERT(paramInfoVec.size() >= 2);
        if (paramInfoVec.size() < 2) {
            break;
        }
        QString infoStr = QString("%1 is now online\nDevices online count:%2")
                            .arg(paramInfoVec.at(0).info)
                            .arg(paramInfoVec.at(1).info);
        infoJson["title"] = "Connection status";
        infoJson["content"] = infoStr.toStdString();
        break;
    }
    case ConnFailed: {
        Q_ASSERT(paramInfoVec.size() >= 2);
        if (paramInfoVec.size() < 2) {
            break;
        }
        QString infoStr = QString("%1 has been disconnected.\nDevices online count:%2")
                            .arg(paramInfoVec.at(0).info)
                            .arg(paramInfoVec.at(1).info);
        infoJson["title"] = "Connection status";
        infoJson["content"] = infoStr.toStdString();
        break;
    }
    case TransferSuccess: {
        Q_ASSERT(paramInfoVec.size() >= 2);
        if (paramInfoVec.size() < 2) {
            break;
        }
        QString infoStr = QString("%1 transferred to %2 is complete")
                            .arg(paramInfoVec.at(0).info)
                            .arg(paramInfoVec.at(1).info);
        infoJson["title"] = "File transfer";
        infoJson["content"] = infoStr.toStdString();
        break;
    }
    case RecvSuccess: {
        Q_ASSERT(paramInfoVec.size() >= 2);
        if (paramInfoVec.size() < 2) {
            break;
        }
        QString infoStr = QString("%1 received from %2 is complete")
                            .arg(paramInfoVec.at(0).info)
                            .arg(paramInfoVec.at(1).info);
        infoJson["title"] = "File transfer";
        infoJson["content"] = infoStr.toStdString();
        break;
    }
    case RefuseRecv: {
        Q_ASSERT(paramInfoVec.size() >= 2);
        if (paramInfoVec.size() < 2) {
            break;
        }
        QString infoStr = QString("%1 declined to receive %2")
                            .arg(paramInfoVec.at(0).info)
                            .arg(paramInfoVec.at(1).info);
        infoJson["title"] = "File transfer";
        infoJson["content"] = infoStr.toStdString();
        break;
    }
    case StartTransferNoti: {
        Q_ASSERT(paramInfoVec.size() >= 2);
        if (paramInfoVec.size() < 2) {
            break;
        }
        QString infoStr = QString("Start receiving %1 from %2")
                              .arg(paramInfoVec.at(0).info)
                              .arg(paramInfoVec.at(1).info);
        infoJson["title"] = "File transfer";
        infoJson["content"] = infoStr.toStdString();
        break;
    }
    default: {
        Q_ASSERT(false);
        break;
    }
    }
    return infoJson;
}

// ------------------------------- FileOperationRecord

std::ostream &operator << (std::ostream &os, const FileOperationRecord &record)
{
    nlohmann::json jsonInfo;
    jsonInfo["fileName"] = record.fileName.c_str();
    jsonInfo["fileSize"] = record.fileSize;
    jsonInfo["timeStamp"] = record.timeStamp;
    jsonInfo["progressValue"] = record.progressValue;
    jsonInfo["clientName"] = record.clientName;
    jsonInfo["clientID"] = record.clientID;
    jsonInfo["ip"] = record.ip.toStdString();
    jsonInfo["port"] = record.port;
    jsonInfo["direction"] = record.direction;
    jsonInfo["optStatus"] = record.optStatus;
    jsonInfo["uuid"] = record.uuid.toStdString();

    jsonInfo["sentFileCount"] = record.sentFileCount;
    jsonInfo["totalFileCount"] = record.totalFileCount;
    jsonInfo["sentFileSize"] = record.sentFileSize;
    jsonInfo["totalFileSize"] = record.totalFileSize;
    jsonInfo["currentTransferFileName"] = record.currentTransferFileName.toStdString();
    jsonInfo["currentTransferFileSize"] = record.currentTransferFileSize;

    jsonInfo["cacheFileName"] = record.cacheFileName.toStdString();
    jsonInfo["cacheFileSize"] = record.cacheFileSize;

    os << jsonInfo.dump();
    return os;
}

std::string FileOperationRecord::toJsonString() const
{
    std::stringstream str_stream;
    str_stream << *this;
    return str_stream.str();
}

void FileOperationRecord::fromJsonString(const std::string &jsonData)
{
    try {
        nlohmann::json jsonInfo = nlohmann::json::parse(jsonData);
        fileName = jsonInfo["fileName"].get<std::string>();
        fileSize = jsonInfo["fileSize"].get<uint64_t>();
        timeStamp = jsonInfo["timeStamp"].get<uint64_t>();
        progressValue = jsonInfo["progressValue"].get<int>();
        clientName = jsonInfo["clientName"].get<std::string>();
        clientID = jsonInfo["clientID"].get<std::string>();
        ip = jsonInfo["ip"].get<std::string>().c_str();
        port = jsonInfo["port"].get<uint16_t>();
        direction = jsonInfo["direction"].get<int>();
        optStatus = jsonInfo["optStatus"].get<int>();
        uuid = jsonInfo["uuid"].get<std::string>().c_str();

        sentFileCount = jsonInfo["sentFileCount"].get<uint32_t>();
        totalFileCount = jsonInfo["totalFileCount"].get<uint32_t>();
        sentFileSize = jsonInfo["sentFileSize"].get<uint64_t>();
        totalFileSize = jsonInfo["totalFileSize"].get<uint64_t>();
        currentTransferFileName = jsonInfo["currentTransferFileName"].get<std::string>().c_str();
        currentTransferFileSize = jsonInfo["currentTransferFileSize"].get<uint64_t>();

        cacheFileName = jsonInfo["cacheFileName"].get<std::string>().c_str();
        cacheFileSize = jsonInfo["cacheFileSize"].get<uint64_t>();
    } catch (const std::exception &e) {
        qWarning() << "FileOperationRecord::fromJsonString ERROR:" << e.what();
    }
}

QString g_sqliteDbPath()
{
    return CommonUtils::localDataDirectory() + "/" + SQLITE_DB_NAME;
}

void g_loadDataFromSqliteDB()
{
    QSqlQuery query(g_getGlobalData()->sqlite_db);
    QString sql = QString("SELECT json_data, uuid FROM opt_record");
    query.exec(sql);
    while (query.next()) {
        const auto &record = query.record();

        FileOperationRecord optRecord;
        optRecord.fromJsonString(record.value(0).toString().toStdString());

        g_getGlobalData()->cacheFileOptRecord.push_back(optRecord);
    }
}

// Perform data synchronization processing
void g_updateCacheFileOptRecord()
{
    auto &cacheFileOptRecord = g_getGlobalData()->cacheFileOptRecord.get<tag_db_timestamp>();
    for (auto itr = cacheFileOptRecord.begin(); itr != cacheFileOptRecord.end(); ++itr) {
        auto &record = *itr;
        if (record.progressValue < 100) {
            cacheFileOptRecord.modify(itr, [] (FileOperationRecord &data) {
                if (data.optStatus == FileOperationRecord::InitStatus) {
                    data.optStatus = FileOperationRecord::TransferFileCancelStatus;
                }
            });
        }
    }
}

void g_saveDataToSqliteDB()
{
    {
        QSqlQuery query(g_getGlobalData()->sqlite_db);
        query.exec(QString(g_drop_table_sql).arg("opt_record"));
        query.exec(g_create_opt_record);
        //QVERIFY(query.exec(sql_new) == true);
    }
    const auto &cacheFileOptRecord = g_getGlobalData()->cacheFileOptRecord.get<tag_db_timestamp>();
    for (auto itr = cacheFileOptRecord.begin(); itr != cacheFileOptRecord.end(); ++itr) {
        const auto &record = *itr;
        QString sql = QString("INSERT INTO opt_record (json_data, uuid) "
                              "VALUES('%1', '%2')")
                              .arg(record.toJsonString().c_str())
                              .arg(record.uuid)
                              ;
        QSqlQuery query(g_getGlobalData()->sqlite_db);
        query.exec(sql);
    }
}

bool g_loadLocalConfig()
{
    QString localConfigPath = CommonUtils::configFilePath();
    do {
        if (QFile::exists(localConfigPath) == false) {
            break;
        }
        try {
            nlohmann::json configJson = nlohmann::json::parse(CommonUtils::getFileContent(localConfigPath).toStdString());
            if (configJson.at("version").get<std::string>() == LOCAL_CONFIG_VERSION) {
                break;
            }
            QFile::remove(localConfigPath);
        } catch (const std::exception &e) {
            Q_EMIT CommonSignals::getInstance()->showWarningMessageBox("warning", e.what());
            return false;
        }
    } while (false);

    if (QFile::exists(localConfigPath) == false) {
        nlohmann::json configJson;
        configJson["version"] = LOCAL_CONFIG_VERSION;
        configJson["fileExplorer"]["mousePressTimeout"] = 300; // 300ms
        configJson["fileExplorer"]["movingPixels"] = 30;
        configJson["filesRecords"]["enableCancelTransfer"] = true;
        configJson["crossShareServer"]["notifyMessageDuration"] = 3000; // 3000ms
        configJson["crossShareServer"]["goServerDllName"] = "client_windows.dll";
        configJson["crossShareServer"]["downloadPath"] = QDir::toNativeSeparators(CommonUtils::downloadDirectoryPath()).toStdString();
        configJson["UITheme"]["customerID"] = 0;
        configJson["UITheme"]["isInited"] = false;
        configJson["UITheme"]["timerThreshold"] = 3000; // ms
        g_getGlobalData()->localConfig = std::move(configJson);

        QFile writeConfigFile(localConfigPath);
        if (writeConfigFile.open(QFile::WriteOnly)) {
            writeConfigFile.write(g_getGlobalData()->localConfig.dump(4).c_str());
            return true;
        } else {
            qWarning() << "write config file failed:" << writeConfigFile.errorString();
            return false;
        }
    }

    try {
        nlohmann::json configJson = nlohmann::json::parse(CommonUtils::getFileContent(localConfigPath).toStdString());
        g_getGlobalData()->localConfig = std::move(configJson);
        return true;
    } catch (const std::exception &e) {
        Q_EMIT CommonSignals::getInstance()->showWarningMessageBox("warning", e.what());
        return false;
    }
}

bool g_updateLocalConfig()
{
    QString localConfigPath = CommonUtils::configFilePath();
    QFile writeConfigFile(localConfigPath);
    if (writeConfigFile.open(QFile::WriteOnly)) {
        writeConfigFile.write(g_getGlobalData()->localConfig.dump(4).c_str());
        return true;
    } else {
        qWarning() << "write config file failed:" << writeConfigFile.errorString();
        return false;
    }
}

void g_sendDataToServer(int msgCode, const QByteArray &data)
{
    AnyMsg message;
    message.funcCode = static_cast<uint32_t>(msgCode);
    message.msgData = data;
    Q_EMIT CommonSignals::getInstance()->sendDataToServer(AnyMsg::toByteArray(message));
}

void g_broadcastData(int msgCode, const QByteArray &data)
{
    AnyMsg message;
    message.funcCode = static_cast<uint32_t>(msgCode);
    message.msgData = data;
    Q_EMIT CommonSignals::getInstance()->broadcastData(AnyMsg::toByteArray(message));
}

//---------------------------------------UpdateSystemInfoMsg----------------

QByteArray UpdateSystemInfoMsg::toByteArray(const UpdateSystemInfoMsg &msg)
{
    Buffer data;
    data.append(msg.headerInfo.header.toUtf8());
    data.append(&msg.headerInfo.type, sizeof (msg.headerInfo.type));
    data.append(&msg.headerInfo.code, sizeof (msg.headerInfo.code));

    {
        Buffer tmpBuffer;
        uint32_t ipValue = QHostAddress(msg.ip).toIPv4Address();
        tmpBuffer.appendUInt32(ipValue);
        tmpBuffer.appendUInt16(msg.port);
        tmpBuffer.append(CommonUtils::toUtf16LE(msg.serverVersion));

        // Processing content length
        data.appendUInt32(static_cast<uint32_t>(tmpBuffer.readableBytes()));
        data.append(tmpBuffer.retrieveAllAsByteArray());
    }
    return data.retrieveAllAsByteArray();
}

bool UpdateSystemInfoMsg::fromByteArray(const QByteArray &data, UpdateSystemInfoMsg &msg)
{
    if (data.length() < MsgHeader::messageLength()) {
        return false;
    }
    Buffer buffer;
    buffer.append(data);
    QByteArray header = QByteArray(buffer.peek(), g_tagNameLength);
    buffer.retrieve(g_tagNameLength);
    if (header != TAG_NAME) {
        qWarning() << "Illegal message HEADER:" << header.constData();
        return false;
    }

    msg.headerInfo.header = header.constData();

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
            msg.ip = QHostAddress(ipValue).toString();
        }

        {
            msg.port = contentBuffer.peekUInt16();
            contentBuffer.retrieveUInt16();
        }

        {
            msg.serverVersion = CommonUtils::toUtf8(contentBuffer.retrieveAllAsByteArray()).constData();
        }
    }
    return true;
}

//---------------------------------------DDCCINotifyMsg----------------

QByteArray DDCCINotifyMsg::toByteArray(const DDCCINotifyMsg &msg)
{
    Buffer data;
    data.append(msg.headerInfo.header.toUtf8());
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
        data.append(tmpBuffer.retrieveAllAsByteArray());
    }
    return data.retrieveAllAsByteArray();
}

bool DDCCINotifyMsg::fromByteArray(const QByteArray &data, DDCCINotifyMsg &msg)
{
    if (data.length() < MsgHeader::messageLength()) {
        return false;
    }
    Buffer buffer;
    buffer.append(data);
    QByteArray header = QByteArray(buffer.peek(), g_tagNameLength);
    buffer.retrieve(g_tagNameLength);
    if (header != TAG_NAME) {
        qWarning() << "Illegal message HEADER:" << header.constData();
        return false;
    }

    msg.headerInfo.header = header.constData();

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
QByteArray DragFilesMsg::toByteArray(const DragFilesMsg &msg)
{
    Buffer data;
    data.append(msg.headerInfo.header.toUtf8());
    data.append(&msg.headerInfo.type, sizeof (msg.headerInfo.type));
    data.append(&msg.headerInfo.code, sizeof (msg.headerInfo.code));

    {
        Buffer tmpBuffer;
        tmpBuffer.append(&msg.functionCode, sizeof (msg.functionCode));
        tmpBuffer.appendUInt64(msg.timeStamp);

        {
            QByteArray filePath_utf16 = CommonUtils::toUtf16LE(msg.rootPath);
            tmpBuffer.appendUInt16(static_cast<uint16_t>(filePath_utf16.size()));
            tmpBuffer.append(filePath_utf16);
        }

        do {
            uint32_t filesCount = static_cast<uint32_t>(msg.filePathVec.size());
            tmpBuffer.appendUInt32(filesCount);
            if (filesCount == 0) {
                break;
            }
            for (const auto &filePath : msg.filePathVec) {
                QByteArray filePath_utf16 = CommonUtils::toUtf16LE(filePath);
                tmpBuffer.appendUInt16(static_cast<uint16_t>(filePath_utf16.size()));
                tmpBuffer.append(filePath_utf16);
            }
        } while (false);

        do {
            if (msg.functionCode != DragFilesMsg::FuncCode::ReceiveFileInfo && msg.functionCode != DragFilesMsg::FuncCode::CancelFileTransfer) {
                break;
            }

            uint32_t ipValue = QHostAddress(msg.ip).toIPv4Address();
            tmpBuffer.appendUInt32(ipValue);
            tmpBuffer.appendUInt16(msg.port);
            Q_ASSERT(msg.clientID.length() == g_clientIDLength);
            tmpBuffer.append(msg.clientID);
            tmpBuffer.appendUInt32(msg.fileCount);
            tmpBuffer.appendUInt64(msg.totalFileSize);
            {
                QByteArray fileName_utf16 = CommonUtils::toUtf16LE(msg.firstTransferFileName);
                tmpBuffer.appendUInt32(static_cast<uint32_t>(fileName_utf16.length()));
                tmpBuffer.append(fileName_utf16);
            }
            tmpBuffer.appendUInt64(msg.firstTransferFileSize);
        } while (false);

        // Processing content length
        data.appendUInt32(static_cast<uint32_t>(tmpBuffer.readableBytes()));
        data.append(tmpBuffer.retrieveAllAsByteArray());
    }
    return data.retrieveAllAsByteArray();
}

bool DragFilesMsg::fromByteArray(const QByteArray &data, DragFilesMsg &msg)
{
    if (data.length() < MsgHeader::messageLength()) {
        return false;
    }
    Buffer buffer;
    buffer.append(data);
    QByteArray header = QByteArray(buffer.peek(), g_tagNameLength);
    buffer.retrieve(g_tagNameLength);
    if (header != TAG_NAME) {
        qWarning() << "Illegal message HEADER:" << header.constData();
        return false;
    }

    msg.headerInfo.header = header.constData();

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
            QByteArray filePathData_utf16(contentBuffer.peek(), dataLen);
            msg.rootPath = CommonUtils::toUtf8(filePathData_utf16);
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
                QByteArray filePathData_utf16(contentBuffer.peek(), dataLen);
                msg.filePathVec.push_back(CommonUtils::toUtf8(filePathData_utf16));
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
                msg.ip = QHostAddress(ipValue).toString();
            }

            {
                msg.port = contentBuffer.peekUInt16();
                contentBuffer.retrieveUInt16();
            }

            {
                msg.clientID = QByteArray(contentBuffer.peek(), g_clientIDLength);
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
                QByteArray fileName_utf16 = QByteArray(contentBuffer.peek(), dataLen);
                contentBuffer.retrieve(dataLen);
                msg.firstTransferFileName = CommonUtils::toUtf8(fileName_utf16);
            }

            {
                msg.firstTransferFileSize = contentBuffer.peekUInt64();
                contentBuffer.retrieveUInt64();
            }
        } while (false);
    }
    return true;
}

//---------------------------------------StatusInfoNotifyMsg----------------

QByteArray StatusInfoNotifyMsg::toByteArray(const StatusInfoNotifyMsg &msg)
{
    Buffer data;
    data.append(msg.headerInfo.header.toUtf8());
    data.append(&msg.headerInfo.type, sizeof (msg.headerInfo.type));
    data.append(&msg.headerInfo.code, sizeof (msg.headerInfo.code));

    {
        Buffer tmpBuffer;
        tmpBuffer.appendUInt32(msg.statusCode);

        // Processing content length
        data.appendUInt32(static_cast<uint32_t>(tmpBuffer.readableBytes()));
        data.append(tmpBuffer.retrieveAllAsByteArray());
    }
    return data.retrieveAllAsByteArray();
}

bool StatusInfoNotifyMsg::fromByteArray(const QByteArray &data, StatusInfoNotifyMsg &msg)
{
    if (data.length() < MsgHeader::messageLength()) {
        return false;
    }
    Buffer buffer;
    buffer.append(data);
    QByteArray header = QByteArray(buffer.peek(), g_tagNameLength);
    buffer.retrieve(g_tagNameLength);
    if (header != TAG_NAME) {
        qWarning() << "Illegal message HEADER:" << header.constData();
        return false;
    }

    msg.headerInfo.header = header.constData();

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
        Q_ASSERT(contentBuffer.readableBytes() == 0);
    }
    return true;
}

//---------------------------------------AnyMsg----------------
QByteArray AnyMsg::toByteArray(const AnyMsg &msg)
{
    Buffer data;
    data.append(msg.headerInfo.header.toUtf8());
    data.append(&msg.headerInfo.type, sizeof (msg.headerInfo.type));
    data.append(&msg.headerInfo.code, sizeof (msg.headerInfo.code));

    {
        Buffer tmpBuffer;
        tmpBuffer.appendUInt32(msg.funcCode);
        {
            tmpBuffer.appendUInt32(static_cast<uint32_t>(msg.msgData.length()));
            tmpBuffer.append(msg.msgData);
        }

        // Processing content length
        data.appendUInt32(static_cast<uint32_t>(tmpBuffer.readableBytes()));
        data.append(tmpBuffer.retrieveAllAsByteArray());
    }
    return data.retrieveAllAsByteArray();
}

bool AnyMsg::fromByteArray(const QByteArray &data, AnyMsg &msg)
{
    if (data.length() < MsgHeader::messageLength()) {
        return false;
    }
    Buffer buffer;
    buffer.append(data);
    QByteArray header = QByteArray(buffer.peek(), g_tagNameLength);
    buffer.retrieve(g_tagNameLength);
    if (header != TAG_NAME) {
        qWarning() << "Illegal message HEADER:" << header.constData();
        return false;
    }

    msg.headerInfo.header = header.constData();

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
            msg.funcCode = contentBuffer.peekUInt32();
            contentBuffer.retrieveUInt32();
        }

        {
            uint32_t msgLen = contentBuffer.peekUInt32();
            contentBuffer.retrieveUInt32();
            msg.msgData = contentBuffer.retrieveAsByteArray(msgLen);
            Q_ASSERT(contentBuffer.readableBytes() == 0);
        }
    }
    return true;
}

const char *g_goErrorCodeToString(uint32_t errorCode)
{
    switch (static_cast<int>(errorCode)) {
    case ERR_BIZ_FD_FILE_NOT_EXISTS:
        return "file not exists";
    case ERR_BIZ_FD_GET_STREAM_EMPTY:
        return "can not get file transfer stream";
    case ERR_BIZ_FD_DATA_EMPTY:
        return "get file data empty";
    case ERR_BIZ_FD_DATA_INVALID:
        return "file data invalid";
    case ERR_BIZ_FD_SRC_OPEN_FILE:
        return "open file err";
    case ERR_BIZ_FD_SRC_FILE_SEEK:
        return "seek file err";
    case ERR_BIZ_FD_SRC_COPY_FILE:
        return "file sending err";
    case ERR_BIZ_FD_SRC_COPY_FILE_TIMEOUT:
        return "tcp connect time out!";
    case ERR_BIZ_FD_SRC_COPY_FILE_CANCEL:
        return "file sending cancel by dst err";
    case ERR_BIZ_FD_DST_OPEN_FILE:
        return "open dst file err";
    case ERR_BIZ_FD_DST_COPY_FILE:
        return "file receive err";
    case ERR_BIZ_FD_DST_COPY_FILE_TIMEOUT:
        return "tcp connect time out!";
    case ERR_BIZ_FD_DST_COPY_FILE_CANCEL:
        return "file receive cancel by src err";
    case ERR_BIZ_FD_DST_COPY_FILE_CANCEL_GUI:
        return "file sending interrupt by dst GUI cancel";
    default:
        return "unknown error";
    }
}

uint32_t g_getCustomerIDForUITheme(bool &isInited)
{
    try {
        uint32_t customerID = g_getGlobalData()->localConfig.at("UITheme").at("customerID").get<uint32_t>();
        isInited = g_getGlobalData()->localConfig.at("UITheme").at("isInited").get<bool>();
        return customerID;
    } catch (const std::exception &e) {
        qWarning() << e.what();
        isInited = false;
        return 0;
    }
}

bool g_is_ROG_Theme()
{
    bool isInited = false;
    uint32_t customerID = g_getCustomerIDForUITheme(isInited);
    return isInited && (customerID == CUSTOMER_ID);
}
