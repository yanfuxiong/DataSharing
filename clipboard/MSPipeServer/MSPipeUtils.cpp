#include "MSPipeUtils.h"
#include "../MSPaste/MSCommon.h"
#include "MSPipeCommon.h"
#include <cstring>
#include <cstdlib>
#include <cstdint>

namespace MSUtils
{
    template <typename T>
    void intToBigEndianBytes(T value, unsigned char* buffer)
    {
        if (!buffer) {
            return;
        }

        static_assert(std::is_integral<T>::value, "T must be an integral type.");
        constexpr int size = sizeof(T);

        for (int i=0; i<size; i++) {
            buffer[(size - 1 - i)] = (value & 0xFF);
            value >>= 8; // Shift right by 8 bits
        }
    }
    template void intToBigEndianBytes<uint32_t>(uint32_t value, unsigned char* buffer);
    template void intToBigEndianBytes<uint64_t>(uint64_t value, unsigned char* buffer);

    template <typename T>
    void bigEndianBytesToInt(unsigned char* buffer, T* value)
    {
        if (!buffer || !value) {
            return;
        }

        static_assert(std::is_integral<T>::value, "T must be an integral type.");
        constexpr int size = sizeof(T);

        for (int i=0; i<size; i++) {
            *value = (*value << 8) | buffer[i];
        }
    }
    template void bigEndianBytesToInt<uint32_t>(unsigned char* buffer, uint32_t* value);
    template void bigEndianBytesToInt<uint64_t>(unsigned char* buffer, uint64_t* value);

    std::string ConvertIp2Str(unsigned char* content) {
        if (!content) {
            DEBUG_LOG("[%s] Null data", __func__);
            return "";
        }

        char tmpbuf[RTK_PIPE_BUFF_SIZE] = {0};
        snprintf(tmpbuf, RTK_PIPE_BUFF_SIZE, "%d.%d.%d.%d:%d", content[0], content[1], content[2], content[3], ((content[4] << 8) | content[5]));
        return std::string(tmpbuf);
    }

    void ConvertIp2Bytes(char* ip, unsigned char* ipBytes) {
        if (!ip || !ipBytes) {
            DEBUG_LOG("[%s] Null data", __func__);
            return;
        }

        const char *delim = ".:";
        char *savePtr = nullptr;
        char * const dupstr = strdup(ip);
        if (dupstr == nullptr) {
            DEBUG_LOG("RtErr : [%s %d], cannot reserve storage for %s", __func__, __LINE__, ip);
            return;
        }

        char *ip1 = strtok_r(dupstr, delim, &savePtr);
        char *ip2 = strtok_r(nullptr, delim, &savePtr);
        char *ip3 = strtok_r(nullptr, delim, &savePtr);
        char *ip4 = strtok_r(nullptr, delim, &savePtr);
        char *port = strtok_r(nullptr, delim, &savePtr);
        if (!ip1 || !ip2 || !ip3 || !ip4 || !port) {
            DEBUG_LOG("[%s %d] Invalid IP", __func__, __LINE__);
            return;
        }

        auto getStr2Byte = [](char* val, unsigned char* ret) {
            int intVal = std::strtol(val, nullptr, 10);
            if (intVal < 0 || intVal > 255) {
                DEBUG_LOG("[%s %d] Invalid IP: %s", __func__, __LINE__, val);
                return;
            }

            *ret = static_cast<unsigned char>(intVal);
        };
        getStr2Byte(ip1, ipBytes);
        getStr2Byte(ip2, ipBytes+1);
        getStr2Byte(ip3, ipBytes+2);
        getStr2Byte(ip4, ipBytes+3);

        uint32_t portVal = std::strtol(port, nullptr, 10);
        if (portVal < 0 || portVal > 65535) {
            DEBUG_LOG("[%s %d] Invalid Port: %s", __func__, __LINE__, port);
            return;
        }
        ipBytes[4] = portVal >> 8;
        ipBytes[5] = portVal & 0xFF;

        free(dupstr);
    }

    std::string ConvertId2Str(unsigned char* content) {
        if (!content) {
            DEBUG_LOG("[%s] Null data", __func__);
            return "";
        }

        unsigned char tmpbuf[LEN_ID] = {0};
        memcpy(tmpbuf, content, LEN_ID);
        return std::string(reinterpret_cast<char*>(tmpbuf), LEN_ID);
    }
};