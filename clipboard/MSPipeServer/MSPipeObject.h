#ifndef __INCLUDED_MS_PIPE_OBJECT__
#define __INCLUDED_MS_PIPE_OBJECT__

#include "MSPipeUtils.h"
#include "MSPipeCommon.h"
#include <stdint.h>

namespace MSPipeObj
{
    struct HEADER
    {
        uint8_t magic[LEN_HEADER] = {'R','T','K','C','S'};
        uint8_t type = 0;
        uint8_t code = 0;
        uint32_t length = 0;
    };

    struct CLIENT_STATUS
    {
        HEADER header = {
            .type = RTK_PIPE_TYPE_NOTI,
            .code = RTK_PIPE_CODE_UPDATE_CLIENT_STATUS,
        };

        struct CONTENT {
            uint8_t status  = 0;
            char *ip        = nullptr;
            char *id        = nullptr;
            wchar_t* name   = nullptr;
        } content;

        uint8_t* rawdata    = nullptr;
        uint32_t offset     = 0;

        void toByte();
        void dump();
        ~CLIENT_STATUS();
    };

    struct SEND_FILE_REQ
    {
        HEADER header = {
            .type = RTK_PIPE_TYPE_REQ,
            .code = RTK_PIPE_CODE_SEND_FILE,
        };

        struct CONTENT {
            char *ip            = nullptr;
            char *id            = nullptr;
            uint64_t fileSize   = 0;
            uint64_t timestamp  = 0;
            wchar_t* filePath   = nullptr;
        } content;

        uint8_t* rawdata        = nullptr;
        uint32_t offset         = 0;

        void toByte();
        void dump();
        ~SEND_FILE_REQ();
    };

    struct SEND_FILE_RESP
    {
        HEADER header = {
            .type = RTK_PIPE_TYPE_RESP,
            .code = RTK_PIPE_CODE_SEND_FILE,
        };

        struct CONTENT {
            int8_t status       = 0;
            char *ip            = nullptr;
            char *id            = nullptr;
            uint64_t fileSize   = 0;
            uint64_t timestamp  = 0;
            wchar_t* filePath   = nullptr;
        } content;

        uint8_t* rawdata        = nullptr;
        uint32_t offset         = 0;

        void toStruct();
        void dump();
        ~SEND_FILE_RESP();
    };

    struct UPDATE_PROGRESS
    {
        HEADER header = {
            .type = RTK_PIPE_TYPE_NOTI,
            .code = RTK_PIPE_CODE_UPDATE_PROGRESS,
        };

        struct CONTENT {
            char *ip            = nullptr;
            char *id            = nullptr;
            uint64_t fileSize   = 0;
            uint64_t sentSize   = 0;
            uint64_t timestamp  = 0;
            wchar_t* filePath   = nullptr;
        } content;

        uint8_t* rawdata        = nullptr;
        uint32_t offset         = 0;

        void toByte();
        void dump();
        ~UPDATE_PROGRESS();
    };

    struct UPDATE_SYSTEM_INFO
    {
        HEADER header = {
            .type = RTK_PIPE_TYPE_NOTI,
            .code = RTK_PIPE_CODE_UPDATE_SYSTEM_INFO,
        };

        struct CONTENT {
            char *ip            = nullptr;
            wchar_t* serviceVer = nullptr;
        } content;

        uint8_t* rawdata        = nullptr;
        uint32_t offset         = 0;

        void toByte();
        void dump();
        ~UPDATE_SYSTEM_INFO();
    };
};

#endif //__INCLUDED_MS_PIPE_OBJECT__
