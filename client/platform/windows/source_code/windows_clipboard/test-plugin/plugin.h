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

typedef void (*UpdateClientStatusCallback)(uint32_t status, const char *ipPort, const char *id, const wchar_t *name, const char *deviceType);
typedef void (*UpdateClientStatusExCallback)(const char *clientJson);
typedef void (*UpdateSystemInfoCallback)(const char *ipPort, const wchar_t *serviceVer);
typedef void (*NotiMessageCallback)(uint64_t timestamp, uint32_t notiCode, const wchar_t *notiParam[], int paramCount);
typedef void (*CleanClipboardCallback)();
typedef void (*AuthViaIndexCallback)(uint32_t index);
typedef void (*DIASStatusCallback)(uint32_t statusCode);
typedef void (*RequestSourceAndPortCallback)();
typedef void (*RequestUpdateClientVersionCallback)(const char *clientVersion);
typedef void (*NotifyErrEventCallback)(const char *clientID, uint32_t errorCode, const char *ipPortString, const char *timeStamp, const char *arg3, const char *arg4);
typedef void (*SetupDstPasteXClipDataCallback)(const char *textData, const char *imageData, const char *htmlData);

}

extern StartClipboardMonitorCallback g_StartClipboardMonitor;
extern StopClipboardMonitorCallback g_StopClipboardMonitor;
extern DragFileListNotifyCallback g_DragFileListNotify;
extern MultiFilesDropNotifyCallback g_MultiFilesDropNotify;
extern UpdateMultipleProgressBarCallback g_UpdateMultipleProgressBar;
extern UpdateClientStatusCallback g_UpdateClientStatus;
extern UpdateClientStatusExCallback g_UpdateClientStatusExCallback;
extern UpdateSystemInfoCallback g_UpdateSystemInfo;
extern NotiMessageCallback g_NotiMessage;
extern CleanClipboardCallback g_CleanClipboard;
extern AuthViaIndexCallback g_AuthViaIndex;
extern DIASStatusCallback g_DIASStatus;
extern RequestSourceAndPortCallback g_RequestSourceAndPort;
extern RequestUpdateClientVersionCallback g_RequestUpdateClientVersionCallback;
extern NotifyErrEventCallback g_NotifyErrEventCallback;
extern SetupDstPasteXClipDataCallback g_SetupDstPasteXClipDataCallback;

class ControlWindow;
extern ControlWindow *g_testWindow;
extern QString g_rootPath;
extern QString g_downloadPath;
extern QString g_deviceName;


