#include "global_def.h"
#include "common_utils.h"
#include <QHostAddress>
#include <QDir>
#include <QCoreApplication>

const int g_tagNameLength = strlen(TAG_NAME);
QString g_namedPipeServerName { "CrossSharePipe" };
const QString g_helperServerName { "CrossShareHelperServer" };

const QString g_drop_table_sql { R"(DROP TABLE IF EXISTS %1;)" };
const QString g_create_opt_record { R"(CREATE TABLE IF NOT EXISTS opt_record (id INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE NOT NULL, uuid TEXT UNIQUE, file_name TEXT DEFAULT "" NOT NULL, file_size INTEGER NOT NULL, timestamp INTEGER NOT NULL, progress_value INTEGER NOT NULL DEFAULT (0), client_name TEXT NOT NULL DEFAULT "", client_id TEXT NOT NULL, ip TEXT NOT NULL, direction INTEGER NOT NULL DEFAULT (0));)" };

void g_globalRegister()
{
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
}

GlobalData *g_getGlobalData()
{
    static GlobalData s_data;
    return &s_data;
}

QByteArray RecordDataHash::getHashID() const
{
    std::stringstream str_stream;
    str_stream << CommonUtils::getFileNameByPath(QString::fromStdString(fileName)).toStdString() << "_";
    str_stream << fileSize << "_";
    str_stream << QByteArray::fromStdString(clientID).toHex().toUpper().constData() << "_";
    str_stream << ip;
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
        qWarning() << "非法的消息HEADER:" << header.constData();
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

QByteArray UpdateClientStatusMsg::toByteArray(const UpdateClientStatusMsg &msg)
{
    Buffer data;
    data.append(msg.headerInfo.header.toUtf8());
    data.append(&msg.headerInfo.type, sizeof (msg.headerInfo.type));
    data.append(&msg.headerInfo.code, sizeof (msg.headerInfo.code));

    {
        Buffer tmpBuffer;
        tmpBuffer.append(&msg.status, sizeof (msg.status));
        uint32_t ipValue = QHostAddress(msg.ip).toIPv4Address();
        tmpBuffer.appendUInt32(ipValue);
        tmpBuffer.appendUInt16(msg.port);
        Q_ASSERT(msg.clientID.length() == 46);
        tmpBuffer.append(msg.clientID);
        tmpBuffer.append(CommonUtils::toUtf16LE(msg.clientName));

        // Processing content length
        data.appendUInt32(static_cast<uint32_t>(tmpBuffer.readableBytes()));
        data.append(tmpBuffer.retrieveAllAsByteArray());
    }
    return data.retrieveAllAsByteArray();
}

bool UpdateClientStatusMsg::fromByteArray(const QByteArray &data, UpdateClientStatusMsg &msg)
{
    if (data.length() < MsgHeader::messageLength()) {
        return false;
    }
    Buffer buffer;
    buffer.append(data);
    QByteArray header = QByteArray(buffer.peek(), g_tagNameLength);
    buffer.retrieve(g_tagNameLength);
    if (header != TAG_NAME) {
        qWarning() << "非法的消息HEADER:" << header.constData();
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
            uint8_t status = 0;
            memcpy(&status, contentBuffer.peek(), sizeof (uint8_t));
            contentBuffer.retrieve(sizeof (uint8_t));
            msg.status = status;
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
            msg.clientName = CommonUtils::toUtf8(contentBuffer.retrieveAllAsByteArray()).constData();
        }
    }
    return true;
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
        qWarning() << "非法的消息HEADER:" << header.constData();
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
            msg.timeStamp = contentBuffer.peekUInt64();
            contentBuffer.retrieveUInt64();
        }

        {
            msg.fileName = CommonUtils::toUtf8(contentBuffer.retrieveAllAsByteArray()).constData();
        }
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
        qWarning() << "非法的消息HEADER:" << header.constData();
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
        uint32_t ipValue = QHostAddress(msg.ip).toIPv4Address();
        tmpBuffer.appendUInt32(ipValue);
        tmpBuffer.appendUInt16(msg.port);
        Q_ASSERT(msg.clientID.length() == 46);
        tmpBuffer.append(msg.clientID);
        tmpBuffer.appendUInt64(msg.fileSize);
        tmpBuffer.appendUInt64(msg.sentSize);
        tmpBuffer.appendUInt64(msg.timeStamp);
        tmpBuffer.append(CommonUtils::toUtf16LE(msg.fileName));

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
        qWarning() << "非法的消息HEADER:" << header.constData();
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

        {
            msg.fileName = CommonUtils::toUtf8(contentBuffer.retrieveAllAsByteArray()).constData();
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
        qWarning() << "非法的消息HEADER:" << header.constData();
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
        qWarning() << "非法的消息HEADER:" << header.constData();
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

std::ostream &operator << (std::ostream &os, const FileOperationRecord &record)
{
    nlohmann::json jsonInfo;
    jsonInfo["fileName"] = CommonUtils::getFileNameByPath(record.fileName.c_str()).toStdString();
    jsonInfo["fileSize"] = record.fileSize;
    jsonInfo["timeStamp"] = record.timeStamp;
    jsonInfo["currentTime"] = QDateTime::fromMSecsSinceEpoch(record.timeStamp).toString("yyyy-MM-dd hh:mm:ss").toStdString();
    jsonInfo["progressValue"] = record.progressValue;
    jsonInfo["clientName"] = record.clientName;
    jsonInfo["clientID"] = record.clientID;
    jsonInfo["ip"] = record.ip.toStdString();
    jsonInfo["descInfo"] = QString("%1: %2").arg(record.direction == 0 ? "to" : "from").arg(record.clientName.c_str()).toStdString();
    os << jsonInfo.dump(4);
    return os;
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
        qWarning() << "非法的消息HEADER:" << header.constData();
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
        qWarning() << "非法的消息HEADER:" << header.constData();
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
    case 1: {
        Q_ASSERT(paramInfoVec.size() >= 2);
        QString infoStr = QString("%1 is now online\nDevices online count:%2")
                            .arg(paramInfoVec.at(0).info)
                            .arg(paramInfoVec.at(1).info);
        infoJson["title"] = "Connection status";
        infoJson["content"] = infoStr.toStdString();
        break;
    }
    case 2: {
        Q_ASSERT(paramInfoVec.size() >= 2);
        QString infoStr = QString("%1 has been disconnected.\nDevices online count:%2")
                            .arg(paramInfoVec.at(0).info)
                            .arg(paramInfoVec.at(1).info);
        infoJson["title"] = "Connection status";
        infoJson["content"] = infoStr.toStdString();
        break;
    }
    case 3: {
        Q_ASSERT(paramInfoVec.size() >= 2);
        QString infoStr = QString("%1 transferred to %2 is complete")
                            .arg(paramInfoVec.at(0).info)
                            .arg(paramInfoVec.at(1).info);
        infoJson["title"] = "File transfer";
        infoJson["content"] = infoStr.toStdString();
        break;
    }
    case 4: {
        Q_ASSERT(paramInfoVec.size() >= 2);
        QString infoStr = QString("%1 received from %2 is complete")
                            .arg(paramInfoVec.at(0).info)
                            .arg(paramInfoVec.at(1).info);
        infoJson["title"] = "File transfer";
        infoJson["content"] = infoStr.toStdString();
        break;
    }
    case 5: {
        Q_ASSERT(paramInfoVec.size() >= 2);
        QString infoStr = QString("%1 declined to receive %2")
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

QString g_sqliteDbPath()
{
    return CommonUtils::localDataDirectory() + "/" + SQLITE_DB_NAME;
}

void g_loadDataFromSqliteDB()
{
    // FIXME:
    return;
    QSqlQuery query(g_getGlobalData()->sqlite_db);
    QString sql = QString("SELECT file_name, file_size, timestamp, progress_value, client_name, client_id, ip, direction, uuid FROM opt_record");
    query.exec(sql);
    while (query.next()) {
        const auto &record = query.record();

        FileOperationRecord optRecord;
        optRecord.fileName = record.value(0).toString().toStdString();
        optRecord.fileSize = record.value(1).toULongLong();
        optRecord.timeStamp = record.value(2).toULongLong();
        optRecord.progressValue = record.value(3).toInt();
        optRecord.clientName = record.value(4).toString().toStdString();
        optRecord.clientID = QByteArray::fromHex(record.value(5).toString().toUtf8()).toStdString();
        optRecord.ip = record.value(6).toString();
        optRecord.direction = record.value(7).toInt();

        optRecord.uuid = record.value(8).toString();

        g_getGlobalData()->cacheFileOptRecord.push_back(optRecord);
    }
}

void g_saveDataToSqliteDB()
{
    // FIXME:
    return;
    {
        QSqlQuery query(g_getGlobalData()->sqlite_db);
        query.exec(QString(g_drop_table_sql).arg("opt_record"));
        query.exec(g_create_opt_record);
        //QVERIFY(query.exec(sql_new) == true);
    }
    const auto &cacheFileOptRecord = g_getGlobalData()->cacheFileOptRecord.get<tag_db_timestamp>();
    for (auto itr = cacheFileOptRecord.begin(); itr != cacheFileOptRecord.end(); ++itr) {
        const auto &record = *itr;
        QString sql = QString("INSERT INTO opt_record (file_name, file_size, timestamp, progress_value, client_name, client_id, ip, direction, uuid) "
                              "VALUES('%1', '%2', '%3', '%4', '%5', '%6', '%7', '%8', '%9')")
                              .arg(QDir::fromNativeSeparators(record.fileName.c_str()))
                              .arg(record.fileSize)
                              .arg(record.timeStamp)
                              .arg(record.progressValue)
                              .arg(record.clientName.c_str())
                              .arg(QByteArray::fromStdString(record.clientID).toHex().toUpper().constData())
                              .arg(record.ip)
                              .arg(record.direction)
                              .arg(record.uuid)
                              ;
        QSqlQuery query(g_getGlobalData()->sqlite_db);
        query.exec(sql);
    }
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
        qWarning() << "非法的消息HEADER:" << header.constData();
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
