#ifndef __INCLUDED_MS_COMMON__
#define __INCLUDED_MS_COMMON__

#include <string>
#include "MSCommonExt.h"
#include <ctime>

#define DEBUG 1

#if DEBUG
#ifdef DEBUG_LOG
#undef DEBUG_LOG
#endif
#define DEBUG_LOG(format, ...) printf(format "\n", ##__VA_ARGS__);
// #define DEBUG_LOG(format, ...) do { \
//     const char *file = "cpp.log"; \
//     FILE *logFile = fopen(file, "a"); /* Open file in append mode */ \
//     if (logFile != NULL) { \
//         time_t now = time(NULL); /* Get current time */ \
//         struct tm *t = localtime(&now); /* Convert to local time */ \
//         char timeStr[64]; \
//         strftime(timeStr, sizeof(timeStr), "%Y-%m-%d %H:%M:%S", t); /* Format time */ \
//         fprintf(logFile, "[%s] ", timeStr); /* Write timestamp */ \
//         fprintf(logFile, format, ##__VA_ARGS__); /* Write formatted log */ \
//         fprintf(logFile, "\n"); \
//         fflush(logFile); \
//         fclose(logFile); /* Close the file */ \
//     } else { \
//         perror("Error opening log file"); /* Print error if file can't be opened */ \
//     } \
// } while (0)
#else
#define DEBUG_LOG(format, ...) printf("not debug\n")
#endif

class IStream;

enum PASTE_TYPE
{
    PASTE_TYPE_FILE,
    PASTE_TYPE_DIB,
    PASTE_TYPE_UNKNOWN = -1,
};

struct FILE_INFO
{
    std::wstring desc; // TODO: P2P ID, or other info
    std::wstring fileName;
    unsigned long fileSizeHigh;
    unsigned long fileSizeLow;
    IStream* pStream = nullptr;
};

struct IMAGE_INFO
{
    std::wstring desc; // TODO: P2P ID, or other info
    IMAGE_HEADER picHeader;
    unsigned long dataSize;
};

typedef void (*ClipboardPasteFileCallback)(char*);
typedef void (*FileDropCmdCallback)(char*, unsigned long, wchar_t*);

class IStreamObserver {
public:
    virtual void OnStreamEOF() = 0;
};

#endif //__INCLUDED_MS_COMMON__