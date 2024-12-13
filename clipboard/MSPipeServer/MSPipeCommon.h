#ifndef __INCLUDED_MS_PIPE_COMMON__
#define __INCLUDED_MS_PIPE_COMMON__

#include <cstdio>
#include <ctime>

#define DEBUG 1

#if DEBUG
#ifdef DEBUG_LOG
#undef DEBUG_LOG
#endif
#define DEBUG_LOG(format, ...) do { \
    const char *file = "cpp.log"; \
    FILE *logFile = fopen(file, "a"); /* Open file in append mode */ \
    if (logFile != NULL) { \
        time_t now = time(NULL); /* Get current time */ \
        struct tm *t = localtime(&now); /* Convert to local time */ \
        char timeStr[64]; \
        strftime(timeStr, sizeof(timeStr), "%Y-%m-%d %H:%M:%S", t); /* Format time */ \
        fprintf(logFile, "[%s] ", timeStr); /* Write timestamp */ \
        fprintf(logFile, format, ##__VA_ARGS__); /* Write formatted log */ \
        fprintf(logFile, "\n"); \
        fflush(logFile); \
        fclose(logFile); /* Close the file */ \
    } else { \
        perror("Error opening log file"); /* Print error if file can't be opened */ \
    } \
} while (0)
#else
#define DEBUG_LOG(format, ...) printf("[Pipe] not debug\n")
#endif

typedef void (*FileDropRequestCallback)(char*, char*, unsigned long long, unsigned long long, wchar_t*);
typedef void (*FileDropResponseCallback)(int, char*, char*, unsigned long long, unsigned long long, wchar_t*);
typedef void (*PipeConnectedCallback)(void);

inline const char* RTK_PIPE_NAME = "\\\\.\\pipe\\CrossSharePipe";
inline const char* RTK_PIPE_HEADER = "RTKCS";
inline const unsigned int RTK_PIPE_BUFF_SIZE = 1024;

enum RTK_PIPE_LENGTH {
    LEN_HEADER      = 5,
    LEN_TYPE        = 1,
    LEN_CODE        = 1,
    LEN_LENGTH      = 4,
    LEN_IP          = 6,
    LEN_ID          = 46,
    LEN_FILESIZE    = 8,
    LEN_SENTSIZE    = 8,
    LEN_TIMESTAMP   = 8,
    LEN_STATUS      = 1,
};

enum RTK_PIPE_TYPE {
    RTK_PIPE_TYPE_REQ = 0,
    RTK_PIPE_TYPE_RESP = 1,
    RTK_PIPE_TYPE_NOTI = 2,
    RTK_PIPE_TYPE_UNKNOWN,
};

enum RTK_PIPE_CODE {
    RTK_PIPE_CODE_RESERVED              = 0,
    RTK_PIPE_CODE_GET_CONN_STATUS       = 1,
    RTK_PIPE_CODE_GET_CLIENT_LIST       = 2,
    RTK_PIPE_CODE_UPDATE_CLIENT_STATUS  = 3,
    RTK_PIPE_CODE_SEND_FILE             = 4,
    RTK_PIPE_CODE_UPDATE_PROGRESS       = 5,
    RTK_PIPE_CODE_UNKNOWN,
};

#endif //__INCLUDED_MS_PIPE_COMMON__