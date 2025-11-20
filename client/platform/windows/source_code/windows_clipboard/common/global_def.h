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
#include <QEvent>
#include <QAxObject>
#include <boost/core/bit.hpp>
#include <boost/core/size.hpp>
#include <boost/core/noncopyable.hpp>
#include <boost/scope_exit.hpp>
#include <boost/scope/scope_exit.hpp>
#include <boost/algorithm/string.hpp>
#include <boost/filesystem.hpp>
#include <boost/regex.h>
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
#include "boost_global_def.h"
#include "json_serialization.h"
#include "boost_serialization.h"
#include <nlohmann/json.hpp>

#define TAG_NAME "RTKCS"
#define PIPE_SERVER_EXE_NAME "client_windows.exe"
#define CROSS_SHARE_SERV_NAME "cross_share_serv.exe"
#define WINDOWS_CLIPBOARD_NAME "windows_clipboard.exe"
#define CROSS_TOOL_NAME "crosstool.exe"
#define STABLE_VERSION_CONTROL 0
#define SQLITE_CONN_NAME "__cross_share_sqlite_conn__"
#define SQLITE_DB_NAME "cross_share_v7.db"
#define LOCAL_CONFIG_VERSION "v1.0.2"
#define LOCAL_CONFIG_NAME "config.json"
#define UPDATE_WINDOW_POS_TAG 1
#define UPDATE_STATUS_TIPS_MSG_TAG 2
#define PR_ADJUST_WINDOW_X_SIZE "__PR_ADJUST_WINDOW_X_SIZE__"
#define PR_ADJUST_WINDOW_Y_SIZE "__PR_ADJUST_WINDOW_Y_SIZE__"
#define PR_ADJUST_WINDOW_SIZE "__PR_ADJUST_WINDOW_SIZE__"
#define PR_TEXT_BROWSER_INTERACTION_DISABLE "__PR_TEXT_BROWSER_INTERACTION_DISABLE__"

namespace mi = boost::multi_index;

typedef std::function<void()> EventCallback;
Q_DECLARE_METATYPE(EventCallback)

typedef std::function<void(QEvent*)> EventCallbackWithEvent;
Q_DECLARE_METATYPE(EventCallbackWithEvent)

extern const int g_tagNameLength;
extern const int g_clientIDLength;
// This variable may be modified, so don't use const.
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

enum PipeMessageCode
{
    GetConnStatus_code = 1,
    GetClientList_code = 2, // Not currently used
    SendFile_code = 4,
    UpdateProgress_code = 5,
    UpdateImageProgress_code = 6,
    NotiMessage_code = 7,
    UpdateSystemInfo_code = 8,
    DDCCIMsg_code = 9,
    DragFilesMsg_code = 12,
    StatusInfoNotifyMsg_code = 13,
    AnyMsg_Code = 14
};

enum AnyMsgFuncCode
{
    UpdateDownloadPath_code = 1,
    UpdateClientVersion_code = 2,
    ShowWindowsClipboard_code = 3,
    NotifyErrorEvent_code = 4,
    UpdateLocalConfigInfo_code = 5,
    UpdateClientStatus_code = 6
};

enum GoErrorCode
{
    ERR_BIZ_FD_FILE_NOT_EXISTS = 5502,
    ERR_BIZ_FD_GET_STREAM_EMPTY = 5504,
    ERR_BIZ_FD_DATA_EMPTY = 5507,
    ERR_BIZ_FD_DATA_INVALID = 5508,
    ERR_BIZ_FD_SRC_OPEN_FILE = 5509,
    ERR_BIZ_FD_SRC_FILE_SEEK = 5510,
    ERR_BIZ_FD_SRC_COPY_FILE = 5511,
    ERR_BIZ_FD_SRC_COPY_FILE_TIMEOUT = 5513,
    ERR_BIZ_FD_SRC_COPY_FILE_CANCEL = 5514,
    ERR_BIZ_FD_DST_OPEN_FILE = 5515,
    ERR_BIZ_FD_DST_COPY_FILE = 5516,
    ERR_BIZ_FD_DST_COPY_FILE_TIMEOUT = 5518,
    ERR_BIZ_FD_DST_COPY_FILE_CANCEL = 5519,
    ERR_BIZ_FD_DST_COPY_FILE_CANCEL_GUI = 5520
};

void g_globalRegister();

// Success returns true, failure returns false
bool g_getCodeFromByteArray(const QByteArray &data, uint8_t &codeValue);
bool g_getCodeFromByteArray(const QByteArray &data, uint8_t &typeValue, uint8_t &codeValue);
QList<QString> g_getPipeServerExePathList();

struct RecordDataHash
{
    std::string clientID;
    std::string ip;
    uint64_t timeStamp { 0 };

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

struct SendFileRequestMsg
{
    enum FlagType {
        NoneFlag = 0,
        SendFlag = 1,
        DragFlag = 2
    };
    MsgHeader headerInfo {PipeMessageType::Request, SendFile_code};

    uint8_t flag { SendFlag };
    QString ip;
    uint16_t port { 0 };
    QByteArray clientID; // Fixed 46 bytes
    uint64_t fileSize { 0 };
    uint64_t timeStamp { 0 };
    QString fileName; // File name (including path)
    std::vector<QString> filePathVec;

    uint32_t getMessageLength() const
    { return static_cast<uint32_t>(MsgHeader::messageLength() + headerInfo.contentLength); }

    static QByteArray toByteArray(const SendFileRequestMsg &msg);
    static bool fromByteArray(const QByteArray &data, SendFileRequestMsg &msg);
};
typedef std::shared_ptr<SendFileRequestMsg> SendFileRequestMsgPtr;
Q_DECLARE_METATYPE(SendFileRequestMsgPtr)

struct SendFileResponseMsg
{
    MsgHeader headerInfo {PipeMessageType::Response, SendFile_code};
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
    MsgHeader headerInfo {PipeMessageType::Notify, UpdateProgress_code};
    enum FuncCode {
        NoneFuncCode = 0,
        SingleFile = 1,
        MultiFile = 2
    };

    uint8_t functionCode { SingleFile };
    QString ip;
    uint16_t port { 0 };
    QByteArray clientID; // Fixed 46 bytes
    uint64_t timeStamp { 0 };

    // ------------------- SingleFile
    uint64_t fileSize { 0 };
    uint64_t sentSize { 0 }; // Sent data size
    QString fileName; // File name (including path)

    // -------------------MultiFile
    QString currentFileName;
    uint32_t sentFilesCount { 0 };
    uint32_t totalFilesCount { 0 };
    uint64_t currentFileSize { 0 };
    uint64_t totalFilesSize { 0 };
    uint64_t totalSentSize { 0 };

    uint32_t getMessageLength() const
    { return static_cast<uint32_t>(MsgHeader::messageLength() + headerInfo.contentLength); }

    RecordDataHash toRecordData() const
    {
        RecordDataHash data;
        data.clientID = clientID.toStdString();
        data.ip = ip.toStdString();
        data.timeStamp = timeStamp;
        return data;
    }

    static QByteArray toByteArray(const UpdateProgressMsg &msg);
    static bool fromByteArray(const QByteArray &data, UpdateProgressMsg &msg);
};

typedef std::shared_ptr<UpdateProgressMsg> UpdateProgressMsgPtr;
Q_DECLARE_METATYPE(UpdateProgressMsgPtr)

struct GetConnStatusRequestMsg
{
    MsgHeader headerInfo {PipeMessageType::Request, GetConnStatus_code};

    uint32_t getMessageLength() const
    { return static_cast<uint32_t>(MsgHeader::messageLength()); }

    static QByteArray toByteArray(const GetConnStatusRequestMsg &msg);
    static bool fromByteArray(const QByteArray &data, GetConnStatusRequestMsg &msg);

};
typedef std::shared_ptr<GetConnStatusRequestMsg> GetConnStatusRequestMsgPtr;
Q_DECLARE_METATYPE(GetConnStatusRequestMsgPtr)

struct GetConnStatusResponseMsg
{
    MsgHeader headerInfo {PipeMessageType::Response, GetConnStatus_code};
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
    MsgHeader headerInfo {PipeMessageType::Notify, UpdateImageProgress_code};
    QString ip;
    uint16_t port;
    QByteArray clientID; // Fixed 46 bytes
    uint64_t fileSize;
    uint64_t sentSize; // Sent data size
    uint64_t timeStamp;

    RecordDataHash toRecordData() const
    {
        RecordDataHash data;
        data.clientID = clientID.toStdString();
        data.ip = ip.toStdString();
        data.timeStamp = timeStamp;
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
    enum NotiCodeType {
        ConnSuccess = 1,
        ConnFailed = 2,
        TransferSuccess = 3,
        RecvSuccess = 4,
        RefuseRecv = 5,
        StartTransferNoti = 6
    };

    struct ParamInfo
    {
        /*uint32_t notiLength;*/
        QString info;
    };

    MsgHeader headerInfo {PipeMessageType::Notify, NotiMessage_code};
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
    enum DirectionType {
        SendType = 0,
        ReceiveType = 1,
        DragSingleFileType = 2,
        DragMultiFileType = 3,
    };

    enum OptStatusType {
        InitStatus = 0,
        TransferFileCancelStatus = 1,
        TransferFileErrorStatus = 2
    };

    std::string fileName; // Include path, only call function conversion when file name is needed
    uint64_t fileSize { 0 };
    uint64_t timeStamp { 0 };
    int progressValue { 0 }; // Progress value [0, 100], if it is -1, it indicates a transmission failure
    std::string clientName;
    std::string clientID;
    QString ip;
    uint16_t port;
    int direction { SendType };
    int optStatus { InitStatus };

    QString uuid;

    // ---------------------- multiFileTransfer
    uint32_t sentFileCount { 0 };
    uint32_t totalFileCount { 0 };
    uint64_t sentFileSize { 0 };
    uint64_t totalFileSize { 0 };
    QString currentTransferFileName;
    uint64_t currentTransferFileSize { 0 };

    // cache data
    QString cacheFileName;
    uint64_t cacheFileSize { 0 };

    friend std::ostream &operator << (std::ostream &os, const FileOperationRecord &record);
    std::string toJsonString() const;
    void fromJsonString(const std::string &jsonData);

    RecordDataHash toRecordData() const
    {
        RecordDataHash data;
        data.clientID = clientID;
        data.ip = ip.toStdString();
        data.timeStamp = timeStamp;
        return data;
    }

    std::size_t toStdHashID() const { return toRecordData().getHashID().toULongLong(); }
};

struct UpdateSystemInfoMsg
{
    MsgHeader headerInfo {PipeMessageType::Notify, UpdateSystemInfo_code};
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

struct DDCCINotifyMsg
{
    MsgHeader headerInfo {PipeMessageType::Notify, DDCCIMsg_code};
    enum FuncCode {
        NoneFuncCode = 0,
        MacAddress = 1, // MacAddress (windows => Go)
        AuthViaIndex = 2, // Auth Via Index (Go => windows)
        ReturnAuthStatus = 3, // Return Auth Status (windows => Go)
        RequestSourcePort = 4, // Request source and port (Go => windows)
        ReturnSourcePort = 5, // Return source and port (windows => Go)
        ExtractDIASMonitor = 6 // (windows => Go)
    };

    uint8_t functionCode { NoneFuncCode };

    std::string macAddress;
    uint16_t source { 0 };
    uint16_t port { 0 };
    uint32_t authResult { 0 };
    uint32_t indexValue { 0 };

    uint32_t getMessageLength() const
    { return static_cast<uint32_t>(MsgHeader::messageLength() + headerInfo.contentLength); }

    static QByteArray toByteArray(const DDCCINotifyMsg &msg);
    static bool fromByteArray(const QByteArray &data, DDCCINotifyMsg &msg);
};
typedef std::shared_ptr<DDCCINotifyMsg> DDCCINotifyMsgPtr;
Q_DECLARE_METATYPE(DDCCINotifyMsgPtr)

struct DragFilesMsg
{
    MsgHeader headerInfo {PipeMessageType::Notify, DragFilesMsg_code};
    enum FuncCode {
        NoneFuncCode = 0,
        MultiFiles = 1,
        ReceiveFileInfo = 2,
        CancelFileTransfer = 3
    };
    uint8_t functionCode { NoneFuncCode };
    uint64_t timeStamp;

    //--------------- MultiFiles
    QString rootPath;
    std::vector<QString> filePathVec;

    //--------------- ReceiveFileInfo
    QString ip;
    uint16_t port { 0 };
    QByteArray clientID; // Fixed 46 bytes
    uint32_t fileCount { 0 };
    uint64_t totalFileSize { 0 };
    QString firstTransferFileName;
    uint64_t firstTransferFileSize { 0 };

    uint32_t getMessageLength() const
    { return static_cast<uint32_t>(MsgHeader::messageLength() + headerInfo.contentLength); }

    static QByteArray toByteArray(const DragFilesMsg &msg);
    static bool fromByteArray(const QByteArray &data, DragFilesMsg &msg);
};
typedef std::shared_ptr<DragFilesMsg> DragFilesMsgPtr;
Q_DECLARE_METATYPE(DragFilesMsgPtr)

struct StatusInfoNotifyMsg
{
    MsgHeader headerInfo {PipeMessageType::Notify, StatusInfoNotifyMsg_code};
    uint32_t statusCode { 0 };

    uint32_t getMessageLength() const
    { return static_cast<uint32_t>(MsgHeader::messageLength() + headerInfo.contentLength); }

    static QByteArray toByteArray(const StatusInfoNotifyMsg &msg);
    static bool fromByteArray(const QByteArray &data, StatusInfoNotifyMsg &msg);
};
typedef std::shared_ptr<StatusInfoNotifyMsg> StatusInfoNotifyMsgPtr;
Q_DECLARE_METATYPE(StatusInfoNotifyMsgPtr)

struct AnyMsg
{
    MsgHeader headerInfo {PipeMessageType::Notify, AnyMsg_Code};
    uint32_t funcCode { 0 };

    QByteArray msgData;

    uint32_t getMessageLength() const
    { return static_cast<uint32_t>(MsgHeader::messageLength() + headerInfo.contentLength); }

    static QByteArray toByteArray(const AnyMsg &msg);
    static bool fromByteArray(const QByteArray &data, AnyMsg &msg);
};
typedef std::shared_ptr<AnyMsg> AnyMsgPtr;
Q_DECLARE_METATYPE(AnyMsgPtr)

struct tag_db_main{};
struct tag_db_timestamp{};
struct tag_db_filename{};
struct tag_db_uuid{};

using FileOperationRecordContainer = mi::multi_index_container<
    FileOperationRecord,
    mi::indexed_by<
        mi::sequenced<mi::tag<tag_db_main> >,
        mi::ordered_unique<mi::tag<tag_db_uuid>, mi::key<&FileOperationRecord::uuid> >,
        mi::ordered_non_unique<mi::tag<tag_db_timestamp>, mi::key<&FileOperationRecord::timeStamp>, std::greater<uint64_t> >,
        mi::ordered_non_unique<mi::tag<tag_db_filename>, mi::key<&FileOperationRecord::fileName> >
    >
>;


struct GlobalData
{
    std::atomic<bool> namedPipeConnected { false };
    std::vector<UpdateClientStatusMsgPtr> m_clientVec;
    QList<UpdateClientStatusMsgPtr> m_selectedClientVec; // The device selected by the user
    QString selectedFileName; // The currently selected file name (including path)
    SystemConfig systemConfig;
    nlohmann::json localConfig;

    //boost::circular_buffer<FileOperationRecord> cacheFileOptRecord { 500 };
    FileOperationRecordContainer cacheFileOptRecord;
    QSqlDatabase sqlite_db;
};

GlobalData *g_getGlobalData();

QString g_sqliteDbPath();
void g_loadDataFromSqliteDB();
void g_updateCacheFileOptRecord();
void g_saveDataToSqliteDB();
bool g_loadLocalConfig();
bool g_updateLocalConfig();
void g_sendDataToServer(int msgCode, const QByteArray &data);
void g_broadcastData(int msgCode, const QByteArray &data);
const char *g_goErrorCodeToString(uint32_t errorCode);
uint32_t g_getCustomerIDForUITheme(bool &isInited);
bool g_is_ROG_Theme();
