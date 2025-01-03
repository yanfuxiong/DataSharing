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
#include <QSqlDatabase>
#include <QSqlQuery>
#include <QSqlRecord>
#include <boost/circular_buffer.hpp>
#include <boost/container_hash/hash.hpp>
#include <boost/multi_index_container.hpp>
#include <boost/multi_index/identity.hpp>
#include <boost/multi_index/key.hpp>
#include <boost/multi_index/global_fun.hpp>
#include <boost/multi_index/ordered_index.hpp>
#include <boost/multi_index/sequenced_index.hpp>
#include <boost/multi_index/random_access_index.hpp>
#include <boost/multi_index/ranked_index.hpp>
#include <iostream>
#include "buffer.h"
#include <nlohmann/json.hpp>

using namespace boost::multi_index;

#define TAG_NAME "RTKCS"
#define PIPE_SERVER_EXE_NAME "client_windows.exe"
#define STABLE_VERSION_CONTROL 0
#define SQLITE_CONN_NAME "__cross_share_sqlite_conn__"
#define SQLITE_DB_NAME "cross_share_v1.db"
#define VERSION_STR "v1.0.2"

typedef std::function<void()> EventCallback;
Q_DECLARE_METATYPE(EventCallback)

extern const int g_tagNameLength;
// 这个变量可能修改, 不使用const
extern QString g_namedPipeServerName;
extern const QString g_helperServerName;
extern const QString g_drop_table_sql;
extern const QString g_create_opt_record;

enum PipeMessageType
{
    Request = 0,
    Response,
    Notify
};

void g_globalRegister();

// Success returns true, failure returns false
bool g_getCodeFromByteArray(const QByteArray &data, uint8_t &codeValue);
bool g_getCodeFromByteArray(const QByteArray &data, uint8_t &typeValue, uint8_t &codeValue);
QList<QString> g_getPipeServerExePathList();

struct RecordDataHash
{
    std::string fileName; // It can include paths, and the function will handle them internally
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
    // Content section
    uint8_t status; // 0: Disconnected state, 1: Connected state
    QString ip;
    uint16_t port;
    QByteArray clientID; // Fixed 46 bytes
    QString clientName; // Client name, device name

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
    QByteArray clientID; // Fixed 46 bytes
    uint64_t fileSize;
    uint64_t timeStamp;
    QString fileName; // File name (including path)

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
    QByteArray clientID; // Fixed 46 bytes
    uint64_t fileSize;
    uint64_t timeStamp;
    QString fileName; // File name (including path)

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
    QByteArray clientID; // Fixed 46 bytes
    uint64_t fileSize;
    uint64_t sentSize; // Sent data size
    uint64_t timeStamp;
    QString fileName; // File name (including path)

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
    uint8_t statusCode; // 0: Disconnected state, 1: Connected state

    uint32_t getMessageLength() const
    { return static_cast<uint32_t>(MsgHeader::messageLength() + sizeof (uint8_t)); }

    static QByteArray toByteArray(const GetConnStatusResponseMsg &msg);
    static bool fromByteArray(const QByteArray &data, GetConnStatusResponseMsg &msg);

};
typedef std::shared_ptr<GetConnStatusResponseMsg> GetConnStatusResponseMsgPtr;
Q_DECLARE_METATYPE(GetConnStatusResponseMsgPtr)

struct UpdateImageProgressMsg
{
    MsgHeader headerInfo {PipeMessageType::Notify, 6};
    QString ip;
    uint16_t port;
    QByteArray clientID; // Fixed 46 bytes
    uint64_t fileSize;
    uint64_t sentSize; // Sent data size
    uint64_t timeStamp;

    RecordDataHash toRecordData() const
    {
        RecordDataHash data;
        data.fileSize = fileSize;
        data.clientID = clientID.toStdString();
        data.ip = ip.toStdString();
        return data;
    }

    uint32_t getMessageLength() const
    { return static_cast<uint32_t>(MsgHeader::messageLength() + headerInfo.contentLength); }

    static QByteArray toByteArray(const UpdateImageProgressMsg &msg);
    static bool fromByteArray(const QByteArray &data, UpdateImageProgressMsg &msg);
};
typedef std::shared_ptr<UpdateImageProgressMsg> UpdateImageProgressMsgPtr;
Q_DECLARE_METATYPE(UpdateImageProgressMsgPtr)

struct NotifyMessage
{
    struct ParamInfo
    {
        /*uint32_t notiLength;*/
        QString info;
    };

    MsgHeader headerInfo {PipeMessageType::Notify, 7};
    uint64_t timeStamp;
    uint8_t notiCode;
    std::vector<ParamInfo> paramInfoVec;

    uint32_t getMessageLength() const
    { return static_cast<uint32_t>(MsgHeader::messageLength() + headerInfo.contentLength); }

    // Specific information display
    nlohmann::json toString() const;
    static QByteArray toByteArray(const NotifyMessage &msg);
    static bool fromByteArray(const QByteArray &data, NotifyMessage &msg);
};
typedef std::shared_ptr<NotifyMessage> NotifyMessagePtr;
Q_DECLARE_METATYPE(NotifyMessagePtr)

struct FileOperationRecord
{
    std::string fileName; // Include path, only call function conversion when file name is needed
    int64_t fileSize = 0;
    uint64_t timeStamp = 0;
    int progressValue = 0; // Progress value [0, 100], if it is -1, it indicates a transmission failure
    std::string clientName;
    std::string clientID;
    QString ip;
    uint8_t direction = 0; // 0: Send 1: Receive

    QString uuid;

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

    std::size_t toStdHashID() const { return toRecordData().getHashID().toULongLong(); }
};

struct UpdateSystemInfoMsg
{
    MsgHeader headerInfo {PipeMessageType::Notify, 8};
    uint8_t statusCode; // 0: reject 1: accept
    QString ip;
    uint16_t port;
    QString serverVersion;

    uint32_t getMessageLength() const
    { return static_cast<uint32_t>(MsgHeader::messageLength() + headerInfo.contentLength); }

    static QByteArray toByteArray(const UpdateSystemInfoMsg &msg);
    static bool fromByteArray(const QByteArray &data, UpdateSystemInfoMsg &msg);
};
typedef std::shared_ptr<UpdateSystemInfoMsg> UpdateSystemInfoMsgPtr;
Q_DECLARE_METATYPE(UpdateSystemInfoMsgPtr)

struct SystemConfig
{
    bool displayLogSwitch = false; // This field is currently invalid
    QString serverVersionStr;
    QString clientVersionStr;
    QString localIpAddress;
    uint16_t port;
};
typedef std::shared_ptr<SystemConfig> SystemConfigPtr;
Q_DECLARE_METATYPE(SystemConfigPtr)

struct tag_db_main{};
struct tag_db_timestamp{};
struct tag_db_filename{};
struct tag_db_uuid{};

using FileOperationRecordContainer = multi_index_container<
    FileOperationRecord,
    indexed_by<
        sequenced<tag<tag_db_main> >,
        ordered_unique<tag<tag_db_uuid>, key<&FileOperationRecord::uuid> >,
        ordered_non_unique<tag<tag_db_timestamp>, key<&FileOperationRecord::timeStamp>, std::greater<uint64_t> >,
        ordered_non_unique<tag<tag_db_filename>, key<&FileOperationRecord::fileName> >
    >
>;


struct GlobalData
{
    std::atomic<bool> namedPipeConnected { false };
    std::vector<UpdateClientStatusMsgPtr> m_clientVec;
    QList<UpdateClientStatusMsgPtr> m_selectedClientVec; // The device selected by the user
    QString selectedFileName; // The currently selected file name (including path)
    SystemConfig systemConfig;

    //boost::circular_buffer<FileOperationRecord> cacheFileOptRecord { 500 };
    FileOperationRecordContainer cacheFileOptRecord;
    QSqlDatabase sqlite_db;
};

GlobalData *g_getGlobalData();

QString g_sqliteDbPath();
void g_loadDataFromSqliteDB();
void g_saveDataToSqliteDB();
