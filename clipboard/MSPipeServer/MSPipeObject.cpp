#include "MSPipeObject.h"
#include <string>
#include <cstring>

namespace MSPipeObj {
    // ClientStatus
    void CLIENT_STATUS::toByte()
    {
        if (!content.ip || !content.id || !content.name) {
            DEBUG_LOG("[%s %d] Err: Null IP, ID or name", __func__, __LINE__);
            return;
        }
        rawdata = new uint8_t[header.length + LEN_HEADER + LEN_TYPE + LEN_CODE + LEN_LENGTH];

        offset = 0;
        memcpy(rawdata+offset, RTK_PIPE_HEADER, LEN_HEADER);
        offset += LEN_HEADER;

        rawdata[offset] = header.type;
        offset += LEN_TYPE;

        rawdata[offset] = header.code;
        offset += LEN_CODE;

        uint32_t lenContent = LEN_STATUS + LEN_IP + LEN_ID;
        lenContent += (wcslen(content.name) * sizeof(wchar_t));
        MSUtils::intToBigEndianBytes(lenContent, rawdata+offset);
        offset += LEN_LENGTH;

        memcpy(rawdata+offset, &content.status, LEN_STATUS);
        offset += LEN_STATUS;

        unsigned char ipBytes[LEN_IP] = {0};
        MSUtils::ConvertIp2Bytes(content.ip, ipBytes);
        memcpy(rawdata+offset, ipBytes, LEN_IP);
        offset += LEN_IP;

        memcpy(rawdata+offset, content.id, LEN_ID);
        offset += LEN_ID;

        int nameLen = wcslen(content.name) * sizeof(wchar_t);
        memcpy(rawdata+offset, content.name, nameLen);
        offset += nameLen;
    }

    void CLIENT_STATUS::dump()
    {
        DEBUG_LOG("[CLIENT_STATUS] type   = %d", header.type);
        DEBUG_LOG("[CLIENT_STATUS] code   = %d", header.code);
        DEBUG_LOG("[CLIENT_STATUS] content length = %d", header.length);
        DEBUG_LOG("[CLIENT_STATUS] status = %d", content.status);
        DEBUG_LOG("[CLIENT_STATUS] ip     = %s", content.ip);
        DEBUG_LOG("[CLIENT_STATUS] id     = %s", content.id);
        DEBUG_LOG("[CLIENT_STATUS] name   = %ls", content.name);
    }

    CLIENT_STATUS::~CLIENT_STATUS()
    {
        if (rawdata) {
            delete[] rawdata;
        }
        if (content.ip) {
            delete[] content.ip;
        }
        if (content.id) {
            delete[] content.id;
        }
        if (content.name) {
            delete[] content.name;
        }
    }

    // SendFile Request
    void SEND_FILE_REQ::toByte()
    {
        if (!content.ip || !content.id || !content.filePath) {
            DEBUG_LOG("[%s %d] Err: Null IP, ID or filePath", __func__, __LINE__);
            return;
        }
        rawdata = new uint8_t[header.length + LEN_HEADER + LEN_TYPE + LEN_CODE + LEN_LENGTH];

        offset = 0;
        memcpy(rawdata+offset, RTK_PIPE_HEADER, LEN_HEADER);
        offset += LEN_HEADER;

        rawdata[offset] = header.type;
        offset += LEN_TYPE;

        rawdata[offset] = header.code;
        offset += LEN_CODE;

        uint32_t lenContent = LEN_IP + LEN_ID + LEN_FILESIZE + LEN_TIMESTAMP;
        lenContent += (wcslen(content.filePath) * sizeof(wchar_t));
        MSUtils::intToBigEndianBytes(lenContent, rawdata+offset);
        offset += LEN_LENGTH;

        unsigned char ipBytes[LEN_IP] = {0};
        MSUtils::ConvertIp2Bytes(content.ip, ipBytes);
        memcpy(rawdata+offset, ipBytes, LEN_IP);
        offset += LEN_IP;

        memcpy(rawdata+offset, content.id, LEN_ID);
        offset += LEN_ID;

        MSUtils::intToBigEndianBytes(content.fileSize, rawdata+offset);
        offset += LEN_FILESIZE;

        MSUtils::intToBigEndianBytes(content.timestamp, rawdata+offset);
        offset += LEN_TIMESTAMP;

        int filePathLen = wcslen(content.filePath) * sizeof(wchar_t);
        memcpy(rawdata+offset, content.filePath, filePathLen);
        offset += filePathLen;
    }

    void SEND_FILE_REQ::dump()
    {
        DEBUG_LOG("[SEND_FILE_REQ] type      = %d", header.type);
        DEBUG_LOG("[SEND_FILE_REQ] code      = %d", header.code);
        DEBUG_LOG("[SEND_FILE_REQ] content length = %d", header.length);
        DEBUG_LOG("[SEND_FILE_REQ] ip        = %s", content.ip);
        DEBUG_LOG("[SEND_FILE_REQ] id        = %s", content.id);
        DEBUG_LOG("[SEND_FILE_REQ] fileSize  = %llu", content.fileSize);
        DEBUG_LOG("[SEND_FILE_REQ] timestamp = %llu", content.timestamp);
        DEBUG_LOG("[SEND_FILE_REQ] filePath  = %ls", content.filePath);
    }

    SEND_FILE_REQ::~SEND_FILE_REQ()
    {
        if (rawdata) {
            delete[] rawdata;
        }
        if (content.ip) {
            delete[] content.ip;
        }
        if (content.id) {
            delete[] content.id;
        }
        if (content.filePath) {
            delete[] content.filePath;
        }
    }

    // SendFile Response
    void SEND_FILE_RESP::toStruct()
    {
        if (!rawdata) {
            DEBUG_LOG("[%s %d] Err: Null rawdata", __func__, __LINE__);
            return;
        }

        if (offset == 0) {
            DEBUG_LOG("[%s %d] Err: Invalid offset:%d", __func__, __LINE__, offset);
            return;
        }

        int tmpOffset = LEN_HEADER+LEN_TYPE+LEN_CODE+LEN_LENGTH;
        content.status = rawdata[tmpOffset];
        tmpOffset += LEN_STATUS;

        std::string ipStr = const_cast<char*>(MSUtils::ConvertIp2Str(rawdata+tmpOffset).c_str());
        content.ip = new char[ipStr.size() + 1];
        std::strcpy(content.ip, ipStr.c_str());
        tmpOffset += LEN_IP;

        std::string idStr = const_cast<char*>(MSUtils::ConvertId2Str(rawdata+tmpOffset).c_str());
        content.id = new char[idStr.size() + 1];
        std::strcpy(content.id, idStr.c_str());
        tmpOffset += LEN_ID;

        uint64_t fileSize = 0;
        MSUtils::bigEndianBytesToInt(rawdata+tmpOffset, &fileSize);
        content.fileSize = fileSize;
        tmpOffset += LEN_FILESIZE;

        uint64_t timestamp = 0;
        MSUtils::bigEndianBytesToInt(rawdata+tmpOffset, &timestamp);
        content.timestamp = timestamp;
        tmpOffset += LEN_TIMESTAMP;

        uint32_t lenFilePath = (offset - tmpOffset);
        uint32_t sizeFilePath = (lenFilePath / sizeof(wchar_t));
        content.filePath = new wchar_t[sizeFilePath+1];
        memset(content.filePath, 0, (sizeFilePath + 1) * sizeof(wchar_t));
        for (int i=0; i<lenFilePath; i+=sizeof(wchar_t)) {
            content.filePath[i/sizeof(wchar_t)] = (rawdata[tmpOffset+i] | (rawdata[tmpOffset+i+1] << 8));
        }
        content.filePath[sizeFilePath] = L'\0';
    }

    void SEND_FILE_RESP::dump()
    {
        DEBUG_LOG("[SEND_FILE_RESP] type      = %d", header.type);
        DEBUG_LOG("[SEND_FILE_RESP] code      = %d", header.code);
        DEBUG_LOG("[SEND_FILE_RESP] content length = %d", header.length);
        DEBUG_LOG("[SEND_FILE_RESP] status    = %d", content.status);
        DEBUG_LOG("[SEND_FILE_RESP] ip        = %s", content.ip);
        DEBUG_LOG("[SEND_FILE_RESP] id        = %s", content.id);
        DEBUG_LOG("[SEND_FILE_RESP] fileSize  = %llu", content.fileSize);
        DEBUG_LOG("[SEND_FILE_RESP] timestamp = %llu", content.timestamp);
        DEBUG_LOG("[SEND_FILE_RESP] filePath  = %ls", content.filePath);
    }

    SEND_FILE_RESP::~SEND_FILE_RESP()
    {
        if (rawdata) {
            delete[] rawdata;
        }
        if (content.ip) {
            delete[] content.ip;
        }
        if (content.id) {
            delete[] content.id;
        }
        if (content.filePath) {
            delete[] content.filePath;
        }
    }

    // Update Progress
    void UPDATE_PROGRESS::toByte()
    {
        if (!content.ip || !content.id || !content.filePath) {
            DEBUG_LOG("[%s %d] Err: Null IP, ID or fileName", __func__, __LINE__);
            return;
        }
        rawdata = new uint8_t[header.length + LEN_HEADER + LEN_TYPE + LEN_CODE + LEN_LENGTH];

        offset = 0;
        memcpy(rawdata+offset, RTK_PIPE_HEADER, LEN_HEADER);
        offset += LEN_HEADER;

        rawdata[offset] = header.type;
        offset += LEN_TYPE;

        rawdata[offset] = header.code;
        offset += LEN_CODE;

        uint32_t lenContent = LEN_IP + LEN_ID + LEN_FILESIZE + LEN_SENTSIZE + LEN_TIMESTAMP;
        lenContent += (wcslen(content.filePath) * sizeof(wchar_t));
        MSUtils::intToBigEndianBytes(lenContent, rawdata+offset);
        offset += LEN_LENGTH;

        unsigned char ipBytes[LEN_IP] = {0};
        MSUtils::ConvertIp2Bytes(content.ip, ipBytes);
        memcpy(rawdata+offset, ipBytes, LEN_IP);
        offset += LEN_IP;

        memcpy(rawdata+offset, content.id, LEN_ID);
        offset += LEN_ID;

        MSUtils::intToBigEndianBytes(content.fileSize, rawdata+offset);
        offset += LEN_FILESIZE;

        MSUtils::intToBigEndianBytes(content.sentSize, rawdata+offset);
        offset += LEN_SENTSIZE;

        MSUtils::intToBigEndianBytes(content.timestamp, rawdata+offset);
        offset += LEN_TIMESTAMP;

        int filePathLen = wcslen(content.filePath) * sizeof(wchar_t);
        memcpy(rawdata+offset, content.filePath, filePathLen);
        offset += filePathLen;
    }

    void UPDATE_PROGRESS::dump()
    {
        DEBUG_LOG("[UPDATE_PROGRESS] type      = %d", header.type);
        DEBUG_LOG("[UPDATE_PROGRESS] code      = %d", header.code);
        DEBUG_LOG("[UPDATE_PROGRESS] content length = %d", header.length);
        DEBUG_LOG("[UPDATE_PROGRESS] ip        = %s", content.ip);
        DEBUG_LOG("[UPDATE_PROGRESS] id        = %s", content.id);
        DEBUG_LOG("[UPDATE_PROGRESS] fileSize  = %llu", content.fileSize);
        DEBUG_LOG("[UPDATE_PROGRESS] sentSize  = %llu", content.sentSize);
        DEBUG_LOG("[UPDATE_PROGRESS] timestamp = %llu", content.timestamp);
        DEBUG_LOG("[UPDATE_PROGRESS] filePath  = %ls", content.filePath);
    }

    UPDATE_PROGRESS::~UPDATE_PROGRESS()
    {
        if (rawdata) {
            delete[] rawdata;
        }
        if (content.ip) {
            delete[] content.ip;
        }
        if (content.id) {
            delete[] content.id;
        }
        if (content.filePath) {
            delete[] content.filePath;
        }
    }
}; // MSPipeObj