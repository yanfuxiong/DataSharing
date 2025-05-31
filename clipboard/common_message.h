#pragma once
#include "global.h"
#include "global_utils.h"
#include "buffer_x.h"

namespace sunkang {

constexpr const char *TAG_NAME_X = "RTKCS";
constexpr int g_tagNameLength = 5;
constexpr int g_clientIDLength = 46;

enum PipeMessageType
{
    Request = 0,
    Response,
    Notify
};

enum PipeMessageCode
{
    GetConnStatus_code = 1,
    //GetClientList_code = 2, // Not currently used
    UpdateClientStatus_code = 3,
    SendFile_code = 4,
    UpdateProgress_code = 5,
    UpdateImageProgress_code = 6,
    NotiMessage_code = 7,
    UpdateSystemInfo_code = 8,
    DDCCIMsg_code = 9,
    DragFilesMsg_code = 12,
    StatusInfoNotifyMsg_code = 13
};

// Success returns true, failure returns false
bool g_getCodeFromByteArray(const std::string &data, uint8_t &codeValue);
bool g_getCodeFromByteArray(const std::string &data, uint8_t &typeValue, uint8_t &codeValue);

struct RecordDataHash
{
    std::string fileName; // It can include paths, and the function will handle them internally
    int64_t fileSize { 0 };
    std::string clientID;
    std::string ip;

    std::string getHashID() const;
};

struct MsgHeader
{
    std::string header;
    uint8_t type; // 0, 1, 2
    uint8_t code; // 0 - 5
    uint32_t contentLength;

    MsgHeader(uint8_t typeVal, uint8_t codeVal)
        : header(TAG_NAME_X)
        , type(typeVal)
        , code(codeVal)
        , contentLength(0)
    {
    }

    static int messageLength();
};

struct UpdateClientStatusMsg
{
    MsgHeader headerInfo {PipeMessageType::Notify, UpdateClientStatus_code};
    // Content section
    uint8_t status { 0 }; // 0: Disconnected state, 1: Connected state
    std::string ip;
    uint16_t port { 0 };
    std::string clientID; // Fixed 46 bytes
    std::string clientName; // Client name, device name
    std::string deviceType;

    uint32_t getMessageLength() const
    { return static_cast<uint32_t>(MsgHeader::messageLength() + headerInfo.contentLength); }

    static std::string toByteArray(const UpdateClientStatusMsg &msg);
    static bool fromByteArray(const std::string &data, UpdateClientStatusMsg &msg);
};

typedef std::shared_ptr<UpdateClientStatusMsg> UpdateClientStatusMsgPtr;


struct SendFileRequestMsg
{
    enum FlagType {
        NoneFlag = 0,
        SendFlag = 1,
        DragFlag = 2
    };

    MsgHeader headerInfo {PipeMessageType::Request, SendFile_code};

    uint8_t flag { SendFlag };
    std::string ip;
    uint16_t port { 0 };
    std::string clientID; // Fixed 46 bytes
    uint64_t fileSize { 0 };
    uint64_t timeStamp { 0 };
    std::string fileName; // File name (including path)
    std::vector<std::string> filePathVec;

    uint32_t getMessageLength() const
    { return static_cast<uint32_t>(MsgHeader::messageLength() + headerInfo.contentLength); }

    static std::string toByteArray(const SendFileRequestMsg &msg);
    static bool fromByteArray(const std::string &data, SendFileRequestMsg &msg);
};
typedef std::shared_ptr<SendFileRequestMsg> SendFileRequestMsgPtr;

struct SendFileResponseMsg
{
    MsgHeader headerInfo {PipeMessageType::Response, SendFile_code};
    uint8_t statusCode { 0 }; // 0: reject 1: accept
    std::string ip;
    uint16_t port { 0 };
    std::string clientID; // Fixed 46 bytes
    uint64_t fileSize { 0 };
    uint64_t timeStamp { 0 };
    std::string fileName; // File name (including path)

    uint32_t getMessageLength() const
    { return static_cast<uint32_t>(MsgHeader::messageLength() + headerInfo.contentLength); }

    static std::string toByteArray(const SendFileResponseMsg &msg);
    static bool fromByteArray(const std::string &data, SendFileResponseMsg &msg);
};
typedef std::shared_ptr<SendFileResponseMsg> SendFileResponseMsgPtr;

struct UpdateProgressMsg
{
    MsgHeader headerInfo {PipeMessageType::Notify, UpdateProgress_code};
    enum FuncCode {
        NoneFuncCode = 0,
        SingleFile = 1,
        MultiFile = 2
    };

    uint8_t functionCode { SingleFile };
    std::string ip;
    uint16_t port { 0 };
    std::string clientID; // Fixed 46 bytes
    uint64_t timeStamp { 0 };

    // ------------------- SingleFile
    uint64_t fileSize { 0 };
    uint64_t sentSize { 0 }; // Sent data size
    std::string fileName; // File name (including path)

    // -------------------MultiFile
    std::string currentFileName;
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
        data.fileName = fileName;
        data.fileSize = fileSize;
        data.clientID = clientID;
        data.ip = ip;
        return data;
    }

    static std::string toByteArray(const UpdateProgressMsg &msg);
    static bool fromByteArray(const std::string &data, UpdateProgressMsg &msg);
};

typedef std::shared_ptr<UpdateProgressMsg> UpdateProgressMsgPtr;

struct GetConnStatusRequestMsg
{
    MsgHeader headerInfo {PipeMessageType::Request, GetConnStatus_code};

    uint32_t getMessageLength() const
    { return static_cast<uint32_t>(MsgHeader::messageLength()); }

    static std::string toByteArray(const GetConnStatusRequestMsg &msg);
    static bool fromByteArray(const std::string &data, GetConnStatusRequestMsg &msg);

};
typedef std::shared_ptr<GetConnStatusRequestMsg> GetConnStatusRequestMsgPtr;

struct GetConnStatusResponseMsg
{
    MsgHeader headerInfo {PipeMessageType::Response, GetConnStatus_code};
    uint8_t statusCode; // 0: Disconnected state, 1: Connected state

    uint32_t getMessageLength() const
    { return static_cast<uint32_t>(MsgHeader::messageLength() + sizeof (uint8_t)); }

    static std::string toByteArray(const GetConnStatusResponseMsg &msg);
    static bool fromByteArray(const std::string &data, GetConnStatusResponseMsg &msg);

};
typedef std::shared_ptr<GetConnStatusResponseMsg> GetConnStatusResponseMsgPtr;

struct UpdateImageProgressMsg
{
    MsgHeader headerInfo {PipeMessageType::Notify, UpdateImageProgress_code};
    std::string ip;
    uint16_t port { 0 };
    std::string clientID; // Fixed 46 bytes
    uint64_t fileSize { 0 };
    uint64_t sentSize { 0 }; // Sent data size
    uint64_t timeStamp { 0 };

    RecordDataHash toRecordData() const
    {
        RecordDataHash data;
        data.fileSize = fileSize;
        data.clientID = clientID;
        data.ip = ip;
        return data;
    }

    uint32_t getMessageLength() const
    { return static_cast<uint32_t>(MsgHeader::messageLength() + headerInfo.contentLength); }

    static std::string toByteArray(const UpdateImageProgressMsg &msg);
    static bool fromByteArray(const std::string &data, UpdateImageProgressMsg &msg);
};
typedef std::shared_ptr<UpdateImageProgressMsg> UpdateImageProgressMsgPtr;

struct NotifyMessage
{
    struct ParamInfo
    {
        /*uint32_t notiLength;*/
        std::string info;
    };

    MsgHeader headerInfo {PipeMessageType::Notify, NotiMessage_code};
    uint64_t timeStamp { 0 };
    uint8_t notiCode { 0 };
    std::vector<ParamInfo> paramInfoVec;

    uint32_t getMessageLength() const
    { return static_cast<uint32_t>(MsgHeader::messageLength() + headerInfo.contentLength); }

    // Specific information display
    //nlohmann::json toString() const;
    static std::string toByteArray(const NotifyMessage &msg);
    static bool fromByteArray(const std::string &data, NotifyMessage &msg);
};
typedef std::shared_ptr<NotifyMessage> NotifyMessagePtr;

struct UpdateSystemInfoMsg
{
    MsgHeader headerInfo {PipeMessageType::Notify, UpdateSystemInfo_code};
    uint8_t statusCode { 0 }; // 0: reject 1: accept
    std::string ip;
    uint16_t port { 0 };
    std::string serverVersion;

    uint32_t getMessageLength() const
    { return static_cast<uint32_t>(MsgHeader::messageLength() + headerInfo.contentLength); }

    static std::string toByteArray(const UpdateSystemInfoMsg &msg);
    static bool fromByteArray(const std::string &data, UpdateSystemInfoMsg &msg);
};
typedef std::shared_ptr<UpdateSystemInfoMsg> UpdateSystemInfoMsgPtr;


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

    static std::string toByteArray(const DDCCINotifyMsg &msg);
    static bool fromByteArray(const std::string &data, DDCCINotifyMsg &msg);
};
typedef std::shared_ptr<DDCCINotifyMsg> DDCCINotifyMsgPtr;

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
    std::string rootPath;
    std::vector<std::string> filePathVec;

    //--------------- ReceiveFileInfo
    std::string ip;
    uint16_t port { 0 };
    std::string clientID; // Fixed 46 bytes
    uint32_t fileCount { 0 };
    uint64_t totalFileSize { 0 };
    std::string firstTransferFileName;
    uint64_t firstTransferFileSize { 0 };

    uint32_t getMessageLength() const
    { return static_cast<uint32_t>(MsgHeader::messageLength() + headerInfo.contentLength); }

    static std::string toByteArray(const DragFilesMsg &msg);
    static bool fromByteArray(const std::string &data, DragFilesMsg &msg);
};
typedef std::shared_ptr<DragFilesMsg> DragFilesMsgPtr;

struct StatusInfoNotifyMsg
{
    MsgHeader headerInfo {PipeMessageType::Notify, StatusInfoNotifyMsg_code};
    uint32_t statusCode { 0 };

    uint32_t getMessageLength() const
    { return static_cast<uint32_t>(MsgHeader::messageLength() + headerInfo.contentLength); }

    static std::string toByteArray(const StatusInfoNotifyMsg &msg);
    static bool fromByteArray(const std::string &data, StatusInfoNotifyMsg &msg);
};
typedef std::shared_ptr<StatusInfoNotifyMsg> StatusInfoNotifyMsgPtr;

}
