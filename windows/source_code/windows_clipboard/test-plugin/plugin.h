#pragma once
#include <QObject>
#include <QString>
#include <cstdint>
#include <windows.h>
#include "common_utils.h"
#include "common_signals.h"

#define EXPORT_FUNC extern "C" __declspec(dllexport)

extern "C" {

typedef struct IMAGE_HEADER
{
    int width;
    int height;
    unsigned short planes;
    unsigned short bitCount;
    unsigned long compression;
} IMAGE_HEADER;

typedef void (*StartClipboardMonitorCallback)();
typedef void (*StopClipboardMonitorCallback)();

typedef void (*DragFileListNotifyCallback)(const char *ipPort,
                                           const char *clientID,
                                           uint32_t cFileCount,
                                           uint64_t totalSize,
                                           uint64_t timestamp,
                                           const wchar_t *firstFileName,
                                           uint64_t firstFileSize);

typedef void (*MultiFilesDropNotifyCallback)(const char *ipPort,
                                             const char *clientID,
                                             uint32_t cFileCount,
                                             uint64_t totalSize,
                                             uint64_t timestamp,
                                             const wchar_t *firstFileName,
                                             uint64_t firstFileSize);

typedef void (*UpdateMultipleProgressBarCallback)(const char *ipPort,
                                                  const char *clientID,
                                                  const wchar_t *currentFileName,
                                                  uint32_t sentFilesCnt,
                                                  uint32_t totalFilesCnt,
                                                  uint64_t currentFileSize,
                                                  uint64_t totalSize,
                                                  uint64_t sentSize,
                                                  uint64_t timestamp);

typedef void (*DataTransferCallback)(const unsigned char *data, uint32_t size);
typedef void (*UpdateClientStatusCallback)(uint32_t status, const char *ipPort, const char *id, const wchar_t *name, const char *deviceType);
typedef void (*UpdateSystemInfoCallback)(const char *ipPort, const wchar_t *serviceVer);
typedef void (*NotiMessageCallback)(uint64_t timestamp, uint32_t notiCode, const wchar_t *notiParam[], int paramCount);
typedef void (*CleanClipboardCallback)();
typedef const wchar_t* (*GetDeviceNameCallback)();
typedef void (*AuthViaIndexCallback)(uint32_t index);
typedef void (*DIASStatusCallback)(uint32_t statusCode);
typedef void (*RequestSourceAndPortCallback)();
typedef const wchar_t* (*GetDownloadPathCallback)();
typedef void (*SetupDstPasteImageCallback)(const wchar_t* desc, IMAGE_HEADER imgHeader, uint32_t dataSize);

}

extern StartClipboardMonitorCallback g_StartClipboardMonitor;
extern StopClipboardMonitorCallback g_StopClipboardMonitor;
extern DragFileListNotifyCallback g_DragFileListNotify;
extern MultiFilesDropNotifyCallback g_MultiFilesDropNotify;
extern UpdateMultipleProgressBarCallback g_UpdateMultipleProgressBar;
extern DataTransferCallback g_DataTransfer;
extern UpdateClientStatusCallback g_UpdateClientStatus;
extern UpdateSystemInfoCallback g_UpdateSystemInfo;
extern NotiMessageCallback g_NotiMessage;
extern CleanClipboardCallback g_CleanClipboard;
extern GetDeviceNameCallback g_GetDeviceName;
extern AuthViaIndexCallback g_AuthViaIndex;
extern DIASStatusCallback g_DIASStatus;
extern RequestSourceAndPortCallback g_RequestSourceAndPort;
extern GetDownloadPathCallback g_GetDownloadPath;

class ControlWindow;
extern ControlWindow *g_testWindow;

