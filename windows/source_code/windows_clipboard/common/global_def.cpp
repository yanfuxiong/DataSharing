#include "global_def.h"
#include "common_utils.h"
#include <QHostAddress>
#include <QDir>
#include <QCoreApplication>

const int g_tagNameLength = strlen(TAG_NAME);

void g_globalRegister()
{
    qRegisterMetaType<UpdateClientStatusMsgPtr>();
    qRegisterMetaType<SendFileRequestMsgPtr>();
    qRegisterMetaType<SendFileResponseMsgPtr>();
    qRegisterMetaType<UpdateProgressMsgPtr>();

    qRegisterMetaType<GetConnStatusRequestMsgPtr>();
    qRegisterMetaType<GetConnStatusResponseMsgPtr>();
    qRegisterMetaType<SystemConfigPtr>();
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
    // 这个用于测试
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

        // 处理content长度
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
            return false; // 此时说明数据不是完整的, 需要继续等待
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

        // 处理content长度
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
            return false; // 此时说明数据不是完整的, 需要继续等待
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

        // 处理content长度
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
            return false; // 此时说明数据不是完整的, 需要继续等待
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

        // 处理content长度
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
            return false; // 此时说明数据不是完整的, 需要继续等待
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

    // 这里直接赋值为0
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

        // 处理content长度
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
            return false; // 此时说明数据不是完整的, 需要继续等待
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
