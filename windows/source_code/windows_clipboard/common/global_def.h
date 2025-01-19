#pragma once
#include <QDebug>
#include <QObject>
#include <QByteArray>
#include <QList>
#include <QString>
#include <QVariant>
#include <QSharedPointer>
#include <QPointer>
#include <QCoreApplication>
#include <QDateTime>
#include <QUuid>
#include <QEvent>
#include <boost/circular_buffer.hpp>
#include <boost/container_hash/hash.hpp>
#include <iostream>
#include "buffer.h"
#include <nlohmann/json.hpp>

#define TAG_NAME "RTKCS"
#define PIPE_SERVER_EXE_NAME "client_windows.exe"
#define STABLE_VERSION_CONTROL 0

typedef std::function<void()> EventCallback;
Q_DECLARE_METATYPE(EventCallback)

extern const int g_tagNameLength;

enum PipeMessageType
{
    Request = 0,
    Response,
    Notify
};

void g_globalRegister();

// 成功返回true, 失败返回false
bool g_getCodeFromByteArray(const QByteArray &data, uint8_t &codeValue);
bool g_getCodeFromByteArray(const QByteArray &data, uint8_t &typeValue, uint8_t &codeValue);
QList<QString> g_getPipeServerExePathList();

struct RecordDataHash
{
    std::string fileName; // 可以包含路径, 函数内部使用的时候会自行处理
    int64_t fileSize = 0;
    std::string clientID;
    std::string ip;

    QByteArray getHashID() const;
};

struct MsgHeader
{
    QString header;
    uint8_t type; // 0, 1, 2
    uint8_t code; // 0 - 5
    uint32_t contentLength;

    MsgHeader(uint8_t typeVal, uint8_t codeVal)
        : header(TAG_NAME)
        , type(typeVal)
        , code(codeVal)
        , contentLength(0)
    {
    }

    static int messageLength();
};

struct UpdateClientStatusMsg
{
    MsgHeader headerInfo {PipeMessageType::Notify, 3};
    // content 部分
    uint8_t status; // 0: 断开状态, 1: 连接状态
    QString ip;
    uint16_t port;
    QByteArray clientID; // 固定46个字节
    QString clientName; // 客户端名称, 设备名称

    uint32_t getMessageLength() const
    { return static_cast<uint32_t>(MsgHeader::messageLength() + headerInfo.contentLength); }

    static QByteArray toByteArray(const UpdateClientStatusMsg &msg);
    static bool fromByteArray(const QByteArray &data, UpdateClientStatusMsg &msg);
};

typedef std::shared_ptr<UpdateClientStatusMsg> UpdateClientStatusMsgPtr;
Q_DECLARE_METATYPE(UpdateClientStatusMsgPtr)

struct SendFileRequestMsg
{
    MsgHeader headerInfo {PipeMessageType::Request, 4};
    QString ip;
    uint16_t port;
    QByteArray clientID; // 固定46个字节
    uint64_t fileSize;
    uint64_t timeStamp;
    QString fileName; // 文件名(包含路径)

    uint32_t getMessageLength() const
    { return static_cast<uint32_t>(MsgHeader::messageLength() + headerInfo.contentLength); }

    static QByteArray toByteArray(const SendFileRequestMsg &msg);
    static bool fromByteArray(const QByteArray &data, SendFileRequestMsg &msg);
};
typedef std::shared_ptr<SendFileRequestMsg> SendFileRequestMsgPtr;
Q_DECLARE_METATYPE(SendFileRequestMsgPtr)

struct SendFileResponseMsg
{
    MsgHeader headerInfo {PipeMessageType::Response, 4};
    uint8_t statusCode; // 0: reject 1: accept
    QString ip;
    uint16_t port;
    QByteArray clientID; // 固定46个字节
    uint64_t fileSize;
    uint64_t timeStamp;
    QString fileName; // 文件名(包含路径)

    uint32_t getMessageLength() const
    { return static_cast<uint32_t>(MsgHeader::messageLength() + headerInfo.contentLength); }

    static QByteArray toByteArray(const SendFileResponseMsg &msg);
    static bool fromByteArray(const QByteArray &data, SendFileResponseMsg &msg);
};
typedef std::shared_ptr<SendFileResponseMsg> SendFileResponseMsgPtr;
Q_DECLARE_METATYPE(SendFileResponseMsgPtr)

struct UpdateProgressMsg
{
    MsgHeader headerInfo {PipeMessageType::Notify, 5};
    QString ip;
    uint16_t port;
    QByteArray clientID; // 固定46个字节
    uint64_t fileSize;
    uint64_t sentSize; // 已发送数据大小
    uint64_t timeStamp;
    QString fileName; // 文件名(包含路径)

    uint32_t getMessageLength() const
    { return static_cast<uint32_t>(MsgHeader::messageLength() + headerInfo.contentLength); }

    RecordDataHash toRecordData() const
    {
        RecordDataHash data;
        data.fileName = fileName.toStdString();
        data.fileSize = fileSize;
        data.clientID = clientID.toStdString();
        data.ip = ip.toStdString();
        return data;
    }

    static QByteArray toByteArray(const UpdateProgressMsg &msg);
    static bool fromByteArray(const QByteArray &data, UpdateProgressMsg &msg);
};

typedef std::shared_ptr<UpdateProgressMsg> UpdateProgressMsgPtr;
Q_DECLARE_METATYPE(UpdateProgressMsgPtr)

struct GetConnStatusRequestMsg
{
    MsgHeader headerInfo {PipeMessageType::Request, 1};

    uint32_t getMessageLength() const
    { return static_cast<uint32_t>(MsgHeader::messageLength()); }

    static QByteArray toByteArray(const GetConnStatusRequestMsg &msg);
    static bool fromByteArray(const QByteArray &data, GetConnStatusRequestMsg &msg);

};
typedef std::shared_ptr<GetConnStatusRequestMsg> GetConnStatusRequestMsgPtr;
Q_DECLARE_METATYPE(GetConnStatusRequestMsgPtr)

struct GetConnStatusResponseMsg
{
    MsgHeader headerInfo {PipeMessageType::Response, 1};
    uint8_t statusCode; // 0: 断开状态 1:连接状态

    uint32_t getMessageLength() const
    { return static_cast<uint32_t>(MsgHeader::messageLength() + sizeof (uint8_t)); }

    static QByteArray toByteArray(const GetConnStatusResponseMsg &msg);
    static bool fromByteArray(const QByteArray &data, GetConnStatusResponseMsg &msg);

};
typedef std::shared_ptr<GetConnStatusResponseMsg> GetConnStatusResponseMsgPtr;
Q_DECLARE_METATYPE(GetConnStatusResponseMsgPtr)

struct FileOperationRecord
{
    std::string fileName; // 包含路径, 只需要文件名的时候调用函数转换
    int64_t fileSize = 0;
    uint64_t timeStamp = 0;
    int progressValue = 0; // 进度值 [0,100], 如果为 -1, 表示传送失败
    std::string clientName; // 设备名称
    std::string clientID;
    QString ip;
    uint8_t direction = 0; // 0: 发送 1: 接收

    friend std::ostream &operator << (std::ostream &os, const FileOperationRecord &record);
    std::string toString() const
    {
        std::stringstream str_stream;
        str_stream << *this;
        return str_stream.str();
    }

    RecordDataHash toRecordData() const
    {
        RecordDataHash data;
        data.fileName = fileName;
        data.fileSize = fileSize;
        data.clientID = clientID;
        data.ip = ip.toStdString();
        return data;
    }
};

struct SystemConfig
{
    bool displayLogSwitch = false;
};
typedef std::shared_ptr<SystemConfig> SystemConfigPtr;
Q_DECLARE_METATYPE(SystemConfigPtr)


struct GlobalData
{
    std::atomic<bool> namedPipeConnected { false };
    std::vector<UpdateClientStatusMsgPtr> m_clientVec;
    QList<UpdateClientStatusMsgPtr> m_selectedClientVec; // 用户选中的设备
    QString selectedFileName; // 当前选中的文件名(包含路径)
    SystemConfig systemConfig;

    boost::circular_buffer<FileOperationRecord> cacheFileOptRecord { 500 }; // 文件操作记录, 最多500条记录
};

GlobalData *g_getGlobalData();
