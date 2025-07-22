#pragma once
#include <cstdint>

#ifdef __cplusplus
extern "C" {
#endif

typedef struct IMAGE_HEADER
{
    int width;
    int height;
    unsigned short planes;
    unsigned short bitCount;
    unsigned long compression;
} IMAGE_HEADER;

typedef void (*InitGoServer)();
typedef void (*SetClipboardCopyImg)(IMAGE_HEADER picHeader, unsigned char *bitmapData, unsigned long dataSize);
typedef void (*SetMacAddress)(const char *macAddress, int length);
typedef void (*SetExtractDIAS)();
typedef void (*SetAuthStatusCode)(unsigned char authResult);
typedef void (*SetDIASSourceAndPort)(unsigned char source, unsigned char port);
typedef void (*SetDragFileListRequest)(const wchar_t *filePathArry[], uint32_t arryLength , uint64_t timeStamp);
typedef void (*SetCancelFileTransfer)(const char *ipPort, const char *clientID, uint64_t timeStamp);
typedef void (*SetMultiFilesDropRequest)(const char *ipPort, const char *clientID, uint64_t timeStamp, const wchar_t *filePathArry[], uint32_t arryLength);

typedef void (*StartClipboardMonitorCallback)();
typedef void (*SetStartClipboardMonitorCallback)(StartClipboardMonitorCallback callback);

typedef void (*StopClipboardMonitorCallback)();
typedef void (*SetStopClipboardMonitorCallback)(StopClipboardMonitorCallback callback);

typedef void (*DragFileListNotifyCallback)(const char *ipPort,
                                   const char *clientID,
                                   uint32_t cFileCount,
                                   uint64_t totalSize,
                                   uint64_t timestamp,
                                   const wchar_t *firstFileName,
                                   uint64_t firstFileSize);
typedef void (*SetDragFileListNotifyCallback)(DragFileListNotifyCallback callback);

typedef void (*MultiFilesDropNotifyCallback)(const char *ipPort,
                                     const char *clientID,
                                     uint32_t cFileCount,
                                     uint64_t totalSize,
                                     uint64_t timestamp,
                                     const wchar_t *firstFileName,
                                     uint64_t firstFileSize);
typedef void (*SetMultiFilesDropNotifyCallback)(MultiFilesDropNotifyCallback callback);

typedef void (*UpdateMultipleProgressBarCallback)(const char *ipPort,
                                          const char *clientID,
                                          const wchar_t *currentFileName,
                                          uint32_t sentFilesCnt,
                                          uint32_t totalFilesCnt,
                                          uint64_t currentFileSize,
                                          uint64_t totalSize,
                                          uint64_t sentSize,
                                          uint64_t timestamp);
typedef void (*SetUpdateMultipleProgressBarCallback)(UpdateMultipleProgressBarCallback callback);

typedef void (*DataTransferCallback)(const unsigned char *data, uint32_t size);
typedef void (*SetDataTransferCallback)(DataTransferCallback callback);

typedef void (*UpdateClientStatusCallback)(uint32_t status, const char *ipPort, const char *id, const wchar_t *name, const char *deviceType);
typedef void (*SetUpdateClientStatusCallback)(UpdateClientStatusCallback callback);

typedef void (*UpdateSystemInfoCallback)(const char *ipPort, const wchar_t *serviceVer);
typedef void (*SetUpdateSystemInfoCallback)(UpdateSystemInfoCallback callback);

typedef void (*NotiMessageCallback)(uint64_t timestamp, uint32_t notiCode, const wchar_t *notiParam[], int paramCount);
typedef void (*SetNotiMessageCallback)(NotiMessageCallback callback);

typedef void (*CleanClipboardCallback)();
typedef void (*SetCleanClipboardCallback)(CleanClipboardCallback callback);

typedef const wchar_t* (*GetDeviceNameCallback)();
typedef void (*SetGetDeviceNameCallback)(GetDeviceNameCallback callback);

typedef void (*AuthViaIndexCallback)(uint32_t index);
typedef void (*SetAuthViaIndexCallback)(AuthViaIndexCallback callback);

typedef void (*DIASStatusCallback)(uint32_t statusCode);
typedef void (*SetDIASStatusCallback)(DIASStatusCallback callback);

typedef void (*RequestSourceAndPortCallback)();
typedef void (*SetRequestSourceAndPortCallback)(RequestSourceAndPortCallback callback);

typedef const wchar_t* (*GetDownloadPathCallback)();
typedef void (*SetGetDownloadPathCallback)(GetDownloadPathCallback callback);

typedef void (*SetupDstPasteImageCallback)(const wchar_t* desc, IMAGE_HEADER imgHeader, uint32_t dataSize);
typedef void (*SetSetupDstPasteImageCallback)(SetupDstPasteImageCallback callback);


#ifdef __cplusplus
}
#endif
