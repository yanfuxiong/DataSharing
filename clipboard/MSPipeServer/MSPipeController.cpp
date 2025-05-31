#include "MSPipeController.h"
#include <string>
#include "MSPipeUtils.h"
#include "MSPipeObject.h"

using namespace std;
using namespace MSPipeObj;

MSPipeController::MSPipeController(const std::atomic<bool>& running_pipe,
                                    FileDropRequestCallback& fileDropReqCb,
                                    FileDropResponseCallback& fileDropRespCb,
                                    PipeConnectedCallback& connCb
                                    ) : IMSPipeConnection(running_pipe),
                                        g_running_pipe(running_pipe),
                                        mFileDropReqCb(fileDropReqCb),
                                        mFileDropRespCb(fileDropRespCb),
                                        mConnCb(connCb)
{
    Connect();
}

MSPipeController::~MSPipeController()
{
}

void MSPipeController::onConnected()
{
    DEBUG_LOG("Connection successfully");
    mConnCb();
}

void MSPipeController::onDisconnected()
{
    DEBUG_LOG("Server disconnected");
}

void MSPipeController::onReadData(unsigned char* data, unsigned int length)
{
    if (!data) {
        DEBUG_LOG("[%s %d] Err: data null", __func__, __LINE__);
        return;
    }
    if (length == 0) {
        DEBUG_LOG("[%s %d] Err: data length=0", __func__, __LINE__);
        return;
    }

    RTK_PIPE_TYPE type = RTK_PIPE_TYPE_UNKNOWN;
    RTK_PIPE_CODE code = RTK_PIPE_CODE_UNKNOWN;
    uint32_t lenContent = 0;
    if (!HandlePipeDataHeader(data, type, code, lenContent)) {
        return;
    }

    uint32_t offsetContent = LEN_HEADER+LEN_TYPE+LEN_CODE+LEN_LENGTH;
    switch (code)
    {
        case RTK_PIPE_CODE_GET_CONN_STATUS:
            break;
        case RTK_PIPE_CODE_SEND_FILE:
            if (type == RTK_PIPE_TYPE_REQ) {
                DEBUG_LOG("SendFile request");
                HandleSendFileReq(data+offsetContent, lenContent);
            } else if (type == RTK_PIPE_TYPE_RESP) {
                DEBUG_LOG("SendFile response");
                HandleSendFileResp(data, length);
            }
            break;
        default:
            DEBUG_LOG("Unhandle code:%d", code);
            break;
    }
}

bool MSPipeController::HandlePipeDataHeader(unsigned char* buffer, RTK_PIPE_TYPE &type, RTK_PIPE_CODE &code, uint32_t &lenContent) {
    if (!buffer) {
        DEBUG_LOG("[%s] Null data", __func__);
        return false;
    }

    uint32_t offset = 0;
    if (strncmp((char*)buffer+offset, RTK_PIPE_HEADER, 5) != 0) {
        DEBUG_LOG("[%s] Unknown header", __func__);
        return false;
    }
    offset += LEN_HEADER;

    unsigned char tmpType = buffer[offset];
    if (tmpType < 0 || tmpType >= RTK_PIPE_TYPE_UNKNOWN) {
        DEBUG_LOG("[%s] Invalid type: code: %d", __func__, tmpType);
        return false;
    }
    offset += LEN_TYPE;

    unsigned char tmpCode = buffer[offset];
    if (tmpCode < 0 || tmpCode >= RTK_PIPE_CODE_UNKNOWN) {
        DEBUG_LOG("[%s] Invalid code: %d", __func__, tmpCode);
        return false;
    }
    offset += LEN_CODE;

    type = static_cast<RTK_PIPE_TYPE>(tmpType);
    code = static_cast<RTK_PIPE_CODE>(tmpCode);
    MSUtils::bigEndianBytesToInt(buffer+offset, &lenContent);
    return true;
}

// TODO: Refactor SendFileRequest from client
void MSPipeController::HandleSendFileReq(unsigned char* content, uint32_t length)
{
    if (!content) {
        DEBUG_LOG("[%s] Null data", __func__);
        return;
    }

    uint32_t minLenContent = (LEN_IP+LEN_ID+LEN_FILESIZE+LEN_TIMESTAMP);
    if (length < minLenContent) {
        DEBUG_LOG("[%s] Invalid content length:%u", __func__, length);
        return;
    }

    int offset = 0;
    std::string strIp = MSUtils::ConvertIp2Str(content+offset);
    offset += LEN_IP;

    std::string strId = MSUtils::ConvertId2Str(content+offset);
    offset += LEN_ID;

    uint64_t fileSize = 0;
    MSUtils::bigEndianBytesToInt(content+offset, &fileSize);
    offset += LEN_FILESIZE;

    uint64_t timestamp = 0;
    MSUtils::bigEndianBytesToInt(content+offset, &timestamp);
    offset += LEN_TIMESTAMP;

    uint32_t lenFilePath = length - offset;
    wchar_t filePath[lenFilePath/2+1];
    memset(filePath, 0, lenFilePath/2+1);
    for (int i=0; i<lenFilePath; i+=2) {
        filePath[i/2] = (content[offset+i] | (content[offset+i+1] << 8));
    }
    filePath[lenFilePath / 2] = L'\0';

    DEBUG_LOG("Sending file request...");
    mFileDropReqCb(const_cast<char*>(strIp.c_str()), const_cast<char*>(strId.c_str()), fileSize, timestamp, filePath);
}

void MSPipeController::HandleSendFileResp(unsigned char* rawdata, uint32_t length)
{
    if (!rawdata) {
        DEBUG_LOG("[%s] Null data", __func__);
        return;
    }

    SEND_FILE_RESP sendFile;
    int contentLen = length-LEN_HEADER-LEN_TYPE-LEN_CODE-LEN_LENGTH;
    uint32_t minContentLen = (LEN_STATUS+LEN_IP+LEN_ID+LEN_FILESIZE+LEN_TIMESTAMP);
    if (contentLen < minContentLen) {
        DEBUG_LOG("[%s] Invalid content length:%u", __func__, contentLen);
        return;
    }

    sendFile.rawdata = new uint8_t[length];
    memcpy(sendFile.rawdata, rawdata, length);
    sendFile.offset = length;

    sendFile.toStruct();
    sendFile.dump();

    DEBUG_LOG("Sending file response...");
    mFileDropRespCb(sendFile.content.status,
                        sendFile.content.ip,
                        sendFile.content.id,
                        sendFile.content.fileSize,
                        sendFile.content.timestamp,
                        sendFile.content.filePath);
}

void MSPipeController::UpdateClientStatus(unsigned int status, char* ip, char* id, wchar_t* name)
{
    if (!ip || !id || !name) {
        DEBUG_LOG("[%s] Err: Null IP, ID or Name", __func__);
        return;
    }

    CLIENT_STATUS clientStatus;
    uint32_t length = LEN_STATUS + LEN_IP + LEN_ID;
    uint32_t lengthName = (wcslen(name) * sizeof(wchar_t));
    length += lengthName;
    clientStatus.header.length = length;

    clientStatus.content.status = status;

    if (clientStatus.content.ip) {
        delete[] clientStatus.content.ip;
    }
    clientStatus.content.ip = new char[strlen(ip)+1];
    memcpy(clientStatus.content.ip, ip, strlen(ip));
    clientStatus.content.ip[strlen(ip)] = '\0';

    if (clientStatus.content.id) {
        delete[] clientStatus.content.id;
    }
    clientStatus.content.id = new char[strlen(id)+1];
    memcpy(clientStatus.content.id, id, strlen(id));
    clientStatus.content.id[strlen(id)] = '\0';

    if (clientStatus.content.name) {
        delete[] clientStatus.content.name;
    }
    clientStatus.content.name = new wchar_t[wcslen(name)+1];
    memcpy(clientStatus.content.name, name, wcslen(name)*sizeof(wchar_t));
    clientStatus.content.name[wcslen(name)] = L'\0';

    clientStatus.toByte();
    if (!clientStatus.rawdata) {
        DEBUG_LOG("[%s %d] Err: Rawdata is null", __func__, __LINE__);
        return;
    }
    clientStatus.dump();
    SendData(clientStatus.rawdata, clientStatus.offset);
}

void MSPipeController::UpdateProgress(char* ip, char* id, uint64_t fileSize, uint64_t sentSize, uint64_t timestamp, wchar_t* fileName)
{
    if (!ip || !id || !fileName) {
        DEBUG_LOG("[%s] Err: Null IP, ID or Fie name", __func__);
        return;
    }

    UPDATE_PROGRESS updateProgress;
    uint32_t length = LEN_IP + LEN_ID + LEN_FILESIZE + LEN_SENTSIZE + LEN_TIMESTAMP;
    uint32_t lengthFilePath = (wcslen(fileName) * sizeof(wchar_t));
    length += lengthFilePath;
    updateProgress.header.length = length;

    if (updateProgress.content.ip) {
        delete[] updateProgress.content.ip;
    }
    updateProgress.content.ip = new char[strlen(ip)+1];
    memcpy(updateProgress.content.ip, ip, strlen(ip));
    updateProgress.content.ip[strlen(ip)] = '\0';

    if (updateProgress.content.id) {
        delete[] updateProgress.content.id;
    }
    updateProgress.content.id = new char[strlen(id)+1];
    memcpy(updateProgress.content.id, id, strlen(id));
    updateProgress.content.id[strlen(id)] = '\0';

    updateProgress.content.fileSize = fileSize;

    updateProgress.content.sentSize = sentSize;

    updateProgress.content.timestamp = timestamp;

    if (updateProgress.content.filePath) {
        delete[] updateProgress.content.filePath;
    }
    updateProgress.content.filePath = new wchar_t[wcslen(fileName)+1];
    memcpy(updateProgress.content.filePath, fileName, wcslen(fileName) * sizeof(wchar_t));
    updateProgress.content.filePath[wcslen(fileName)] = L'\0';

    updateProgress.toByte();
    if (!updateProgress.rawdata) {
        DEBUG_LOG("[%s %d] Err: Rawdata is null", __func__, __LINE__);
        return;
    }
    updateProgress.dump();
    SendData(updateProgress.rawdata, updateProgress.offset);
}

void MSPipeController::UpdateSystemInfo(char* ip, wchar_t* serviceVer)
{
    if (!ip || !serviceVer) {
        DEBUG_LOG("[%s] Err: NULL IP or serviceVer", __func__);
        return;
    }

    UPDATE_SYSTEM_INFO updateSysInfo;
    uint32_t length = LEN_IP;
    uint32_t lengthServiceVer = (wcslen(serviceVer) * sizeof(wchar_t));
    length += lengthServiceVer;
    updateSysInfo.header.length = length;

    if (updateSysInfo.content.ip) {
        delete[] updateSysInfo.content.ip;
    }
    updateSysInfo.content.ip = new char[strlen(ip)+1];
    memcpy(updateSysInfo.content.ip, ip, strlen(ip));
    updateSysInfo.content.ip[strlen(ip)] = '\0';

    if (updateSysInfo.content.serviceVer) {
        delete[] updateSysInfo.content.serviceVer;
    }
    updateSysInfo.content.serviceVer = new wchar_t[wcslen(serviceVer)+1];
    memcpy(updateSysInfo.content.serviceVer, serviceVer, wcslen(serviceVer) * sizeof(wchar_t));
    updateSysInfo.content.serviceVer[wcslen(serviceVer)] = L'\0';

    updateSysInfo.toByte();
    if (!updateSysInfo.rawdata) {
        DEBUG_LOG("[%s %d] Err: Rawdata is null", __func__, __LINE__);
        return;
    }
    updateSysInfo.dump();
    SendData(updateSysInfo.rawdata, updateSysInfo.offset);
}

void MSPipeController::SendFileReq(char* ip, char* id, uint64_t fileSize, uint64_t timestamp, wchar_t* fileName)
{
    if (!ip || !id || !fileName) {
        DEBUG_LOG("[%s] Err: Null IP, ID or Fie name", __func__);
        return;
    }

    SEND_FILE_REQ sendFile;
    uint32_t length = LEN_IP + LEN_ID + LEN_FILESIZE + LEN_TIMESTAMP;
    uint32_t lengthFilePath = (wcslen(fileName) * sizeof(wchar_t));
    length += lengthFilePath;
    sendFile.header.length = length;

    if (sendFile.content.ip) {
        delete[] sendFile.content.ip;
    }
    sendFile.content.ip = new char[strlen(ip)+1];
    memcpy(sendFile.content.ip, ip, strlen(ip));
    sendFile.content.ip[strlen(ip)] = '\0';

    if (sendFile.content.id) {
        delete[] sendFile.content.id;
    }
    sendFile.content.id = new char[strlen(id)+1];
    memcpy(sendFile.content.id, id, strlen(id));
    sendFile.content.id[strlen(id)] = '\0';

    sendFile.content.fileSize = fileSize;

    sendFile.content.timestamp = timestamp;

    if (sendFile.content.filePath) {
        delete[] sendFile.content.filePath;
    }
    sendFile.content.filePath = new wchar_t[wcslen(fileName)+1];
    memcpy(sendFile.content.filePath, fileName, wcslen(fileName) * sizeof(wchar_t));
    sendFile.content.filePath[wcslen(fileName)] = L'\0';

    sendFile.toByte();
    if (!sendFile.rawdata) {
        DEBUG_LOG("[%s %d] Err: Rawdata is null", __func__, __LINE__);
        return;
    }
    sendFile.dump();
    SendData(sendFile.rawdata, sendFile.offset);
}
