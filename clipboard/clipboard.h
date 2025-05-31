#pragma once
#include <string>
#include <atomic>
#include <iostream>
#include <mutex>
#include <cstdlib>
#include <cstring>
#include "MSPaste/MSPasteImpl.h"
#include "MSPaste/MSFileDrop.h"
#include "MSPipeServer/MSPipeController.h"

typedef void (*ClipboardCopyFileCallback)(wchar_t*, unsigned long, unsigned long);
typedef void (*ClipboardPasteFileCallback)(char*);
typedef void (*ClipboardCopyImgCallback)(IMAGE_HEADER, unsigned char*, unsigned long);
typedef void (*FileDropRequestCallback)(char*, char*, unsigned long long, unsigned long long, wchar_t*);
typedef void (*FileDropResponseCallback)(int, char*, char*, unsigned long long, unsigned long long, wchar_t*);
typedef void (*PipeConnectedCallback)(void);
typedef void (*GetMacAddressCallback)(char*, int);
typedef void (*ExtractDIASCallback)();
typedef void (*AuthStatusCodeCallback)(unsigned char);
typedef void (*DIASSourceAndPortCallback)(unsigned char, unsigned char);
typedef void (*DragFileCallback)(unsigned long long timeStamp, wchar_t *filePath);
typedef void (*DragFileListRequestCallback)(wchar_t *filePathArry[], unsigned int arryLength , unsigned long long timeStamp);
typedef void (*CancelFileTransferCallback)(char *ipPort, char *clientID, unsigned long long timeStamp);
typedef void (*MultiFilesDropRequestCallback)(char *ipPort, char *clientID, unsigned long long timeStamp, wchar_t *filePathArry[], unsigned int arryLength);

#ifdef __cplusplus
extern "C" {
#endif

__declspec(dllexport) void SetClipboardCopyFileCallback(ClipboardCopyFileCallback callback);

__declspec(dllexport) void SetClipboardPasteFileCallback(ClipboardPasteFileCallback callback);

__declspec(dllexport) void SetFileDropRequestCallback(FileDropRequestCallback callback);

__declspec(dllexport) void SetMultiFilesDropRequestCallback(MultiFilesDropRequestCallback callback);

__declspec(dllexport) void SetFileDropResponseCallback(FileDropResponseCallback callback);

__declspec(dllexport) void SetClipboardCopyImgCallback(ClipboardCopyImgCallback callback);

__declspec(dllexport) void StartClipboardMonitor();

__declspec(dllexport) void StopClipboardMonitor();

__declspec(dllexport) void SetPipeConnectedCallback(PipeConnectedCallback callback);

__declspec(dllexport) void SetGetMacAddressCallback(GetMacAddressCallback callback);

__declspec(dllexport) void SetExtractDIASCallback(ExtractDIASCallback callback);

__declspec(dllexport) void SetAuthStatusCodeCallback(AuthStatusCodeCallback callback);

__declspec(dllexport) void SetDIASSourceAndPortCallback(DIASSourceAndPortCallback callback);

__declspec(dllexport) void SetDragFileCallback(DragFileCallback callback);
__declspec(dllexport) void SetDragFileListRequestCallback(DragFileListRequestCallback callback);
__declspec(dllexport) void SetCancelFileTransferCallback(CancelFileTransferCallback callback);


// TODO: consider that transfering multiple files
__declspec(dllexport) void SetupDstPasteFile(wchar_t* desc,
                            wchar_t* fileName,
                            unsigned long fileSizeHigh,
                            unsigned long fileSizeLow);

__declspec(dllexport) void SetupDstPasteImage(wchar_t* desc,
                                                        IMAGE_HEADER imgHeader,
                                                        unsigned long dataSize);

__declspec(dllexport) void SetupFileDrop(char* ip,
                                        char* id,
                                        unsigned long long fileSize,
                                        unsigned long long timestamp,
                                        wchar_t* fileName);

__declspec(dllexport) void DragFileNotify(char *ipPort,
                                        char *clientID,
                                        unsigned long long fileSize,
                                        unsigned long long timestamp,
                                        wchar_t *fileName);

__declspec(dllexport) void DragFileListNotify(char *ipPort,
                                              char *clientID,
                                              unsigned int cFileCount,
                                              unsigned long long totalSize,
                                              unsigned long long timestamp,
                                              wchar_t *firstFileName,
                                              unsigned long long firstFileSize);

__declspec(dllexport) void MultiFilesDropNotify(char *ipPort,
                                              char *clientID,
                                              unsigned int cFileCount,
                                              unsigned long long totalSize,
                                              unsigned long long timestamp,
                                              wchar_t *firstFileName,
                                              unsigned long long firstFileSize);

__declspec(dllexport) void UpdateMultipleProgressBar(char *ipPort,
                                   char *clientID,
                                   wchar_t *currentFileName,
                                   unsigned int sentFilesCnt,
                                   unsigned int totalFilesCnt,
                                   unsigned long long currentFileSize,
                                   unsigned long long totalSize,
                                   unsigned long long sentSize,
                                   unsigned long long timestamp);

__declspec(dllexport) void DataTransfer(unsigned char* data, unsigned int size);

__declspec(dllexport) void UpdateProgressBar(char* ip, char* id, uint64_t fileSize, uint64_t sentSize, uint64_t timestamp, wchar_t* fileName);
__declspec(dllexport) void DeinitProgressBar();

__declspec(dllexport) void UpdateImageProgressBar(char *ip, char *id, uint64_t fileSize, uint64_t sentSize, uint64_t timestamp);

__declspec(dllexport) void EventHandle(EVENT_TYPE event);

__declspec(dllexport) void StartPipeMonitor();

__declspec(dllexport) void StopPipeMonitor();

__declspec(dllexport) void UpdateClientStatus(unsigned int status, char *ip, char *id, wchar_t *name, char *deviceType);

__declspec(dllexport) void UpdateSystemInfo(char *ip, wchar_t *serviceVer);

// Please refer to the protocol document for the specific meanings of the parameters
__declspec(dllexport) void NotiMessage(uint64_t timestamp, unsigned int notiCode, wchar_t *notiParam[], int paramCount);

__declspec(dllexport) void CleanClipboard();

__declspec(dllexport) const wchar_t* GetDeviceName();

__declspec(dllexport) void AuthViaIndex(uint32_t index);

__declspec(dllexport) void DIASStatus(unsigned int statusCode);

__declspec(dllexport) void RequestSourceAndPort();

__declspec(dllexport) const wchar_t* GetDownloadPath();

#ifdef __cplusplus
}
#endif

