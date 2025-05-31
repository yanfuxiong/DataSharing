#include "pipeserver_control.h"
#include "ddcci/windows_event_monitor.h"
#include "clipboard.h"
#include <vector>
#include <shlobj.h>

HWND g_hWnd = NULL;
std::atomic<bool> g_running;
std::atomic<bool> g_running_pipe;
HANDLE g_event;
HANDLE g_thread = NULL;
bool g_isOleClipboardOperation = false;
std::mutex clipboardMutex;
MSPasteImpl *msPasteImpl = NULL;
MSFileDrop *msFileDrop = NULL;

PipeServerControl *msPipeCtrl = NULL;

ClipboardCopyFileCallback 	g_cpFilecallback 			= nullptr;
ClipboardPasteFileCallback 	g_pscallback 				= nullptr;
ClipboardCopyImgCallback 	g_cpImgCallback 			= nullptr;
FileDropRequestCallback 	g_fdReqCallback 			= nullptr;
MultiFilesDropRequestCallback g_multiFilesReqCallback   = nullptr;
FileDropResponseCallback 	g_fdRespCallback 			= nullptr;
PipeConnectedCallback 		g_pipeConnCallback 			= nullptr;
GetMacAddressCallback 		g_getMacAddressCallback 	= nullptr;
ExtractDIASCallback 		g_extractDIASCallback 		= nullptr;
AuthStatusCodeCallback 		g_authStatusCodeCallback 	= nullptr;
DIASSourceAndPortCallback 	g_DIASSourceAndPortCallback = nullptr;
DragFileCallback            g_dragFileCallback          = nullptr;
DragFileListRequestCallback g_dragFileListCallback      = nullptr;
CancelFileTransferCallback  g_cancelFileTransferCallback = nullptr;

void GetFileSizeW(wchar_t filePath[MAX_PATH], unsigned long& fileSizeHigh, unsigned long& fileSizeLow)
{
    HANDLE hFile = CreateFileW(filePath, GENERIC_READ, FILE_SHARE_READ, NULL, OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL, NULL);
    if (hFile == INVALID_HANDLE_VALUE) {
        DEBUG_LOG("[%s %d] Failed to open file", __func__, __LINE__);
        return;
    }

    LARGE_INTEGER fileSize;
    if (!GetFileSizeEx(hFile, &fileSize)) {
        DEBUG_LOG("[%s %d] Failed to get file size", __func__, __LINE__);
        return;
    }

    fileSizeHigh = fileSize.HighPart;
    fileSizeLow = fileSize.LowPart;
    CloseHandle(hFile);
}

void GetBitmapData(HBITMAP hBitmap)
{
    BITMAP bitmap;
    if (GetObject(hBitmap, sizeof(BITMAP), &bitmap)) {
        HDC hdc = GetDC(NULL);
        BITMAPINFO bmpInfo = { 0 };
        bmpInfo.bmiHeader.biSize = sizeof(BITMAPINFOHEADER);
        bmpInfo.bmiHeader.biWidth = bitmap.bmWidth;
        bmpInfo.bmiHeader.biHeight = bitmap.bmHeight;
        bmpInfo.bmiHeader.biPlanes = bitmap.bmPlanes;
        bmpInfo.bmiHeader.biBitCount = bitmap.bmBitsPixel;
        bmpInfo.bmiHeader.biCompression = BI_RGB;

        const int dataSize = bitmap.bmWidthBytes * bitmap.bmHeight;
        BYTE* bitmapData = new BYTE[dataSize];
        if (!GetDIBits(hdc, hBitmap, 0, bitmap.bmHeight, bitmapData, &bmpInfo, DIB_RGB_COLORS)) {
            DEBUG_LOG("[%s %d] Get DIBits failed", __func__, __LINE__);
        }

        IMAGE_HEADER picHeader = {
            .width = bmpInfo.bmiHeader.biWidth,
            .height = bmpInfo.bmiHeader.biHeight,
            .planes = bmpInfo.bmiHeader.biPlanes,
            .bitCount = bmpInfo.bmiHeader.biBitCount,
            .compression = bmpInfo.bmiHeader.biCompression
        };
        DEBUG_LOG("[Copy] Trigger copy image. H=%d,W=%d,Planes=%d,BitCnt=%d,Compress=%d",
            picHeader.height, picHeader.width, picHeader.planes, picHeader.bitCount, picHeader.compression)
        g_cpImgCallback(picHeader, bitmapData, dataSize);

        delete[] bitmapData;
        ReleaseDC(NULL, hdc);
    } else {
        DEBUG_LOG("[%s %d] Failed to get bitmap object details", __func__, __LINE__);
    }
}

LRESULT CALLBACK WndProc(HWND hWnd, UINT message, WPARAM wParam, LPARAM lParam) {
    switch (message) {
    case WM_CLIPBOARDUPDATE:
        {
            bool isOleOperation = false;
            {
                std::lock_guard<std::mutex> lock(clipboardMutex);
                isOleOperation = g_isOleClipboardOperation;
            }

            if (isOleOperation)
            {
                DEBUG_LOG("[Detect Copy] Wait for lock and skip")
                return 0;
            }
            DEBUG_LOG("[Detect Copy] Clipboard updated")

            if (OpenClipboard(NULL)) {
                // if (IsClipboardFormatAvailable(CF_HDROP)) {
                //     HANDLE hData = GetClipboardData(CF_HDROP);
                //     if (hData) {
                //         HDROP hDrop = static_cast<HDROP>(hData);
                //         UINT fileCount = DragQueryFile(hDrop, 0xFFFFFFFF, NULL, 0);
                //         if (fileCount > 0) {
                //             wchar_t filePath[MAX_PATH] = {};
                //             DragQueryFileW(hDrop, 0, filePath, _countof(filePath));

                //             unsigned long fileSizeHigh, fileSizeLow;
                //             GetFileSizeW(filePath, fileSizeHigh, fileSizeLow);
                //             if (g_cpFilecallback) {
                //                 if (fileSizeHigh > 0 || fileSizeLow > 0) {
                //                     g_cpFilecallback(filePath, fileSizeHigh, fileSizeLow);
                //                 } else {
                //                     DEBUG_LOG("[%s %d] Skip copy file: Empty fil", __func__, __LINE__);
                //                 }
                //             }
                //             // DragFinish(hDrop);
                //         }
                //     }
                // }
                // else if (IsClipboardFormatAvailable(CF_BITMAP)) {
                if (IsClipboardFormatAvailable(CF_BITMAP)) {
                    DEBUG_LOG("[%s %d] Get clipboard event: CF_BITMAP", __func__, __LINE__);
                    HANDLE hData = GetClipboardData(CF_BITMAP);
                    if (hData) {
                        HBITMAP hBitmap = static_cast<HBITMAP>(hData);
                        GetBitmapData(hBitmap);
                    }
                }

                CloseClipboard();
            }
        }
        break;
    case WM_DESTROY:
        RemoveClipboardFormatListener(hWnd);
        PostQuitMessage(0);
        break;
    default:
        return DefWindowProc(hWnd, message, wParam, lParam);
    }
    return 0;
}

DWORD WINAPI ClipboardMonitorThread(LPVOID lpParam)
{
    (void)lpParam;
    HINSTANCE hInstance = GetModuleHandle(NULL);
    WNDCLASS wc = { 0 };

    wc.lpfnWndProc = WndProc;
    wc.hInstance = hInstance;
    wc.lpszClassName = TEXT("ClipboardListener");

    if (!RegisterClass(&wc)) {
        SetEvent(g_event);
        return 1;
    }

    g_hWnd = CreateWindow(
        wc.lpszClassName,
        TEXT("Clipboard Listener"),
        0,
        0, 0,
        0, 0,
        NULL, NULL, hInstance, NULL
    );

    if (g_hWnd == NULL) {
        SetEvent(g_event);
        return 1;
    }

    if (!AddClipboardFormatListener(g_hWnd)) {
        DestroyWindow(g_hWnd);
        SetEvent(g_event);
        return 1;
    }

    SetEvent(g_event);

    MSG msg;
    while (g_running.load()) {
        BOOL res = GetMessage(&msg, NULL, 0, 0);
        if (res > 0) {
            TranslateMessage(&msg);
            DispatchMessage(&msg);
        } else if (res == 0) {
            break;
        } else {
            break;
        }
    }

    DestroyWindow(g_hWnd);
    UnregisterClass(wc.lpszClassName, hInstance);
    return 0;
}

//--------------------------------------------------------------------------------
void SetClipboardCopyFileCallback(ClipboardCopyFileCallback callback) {
    g_cpFilecallback = callback;
}

void SetClipboardPasteFileCallback(ClipboardPasteFileCallback callback) {
    g_pscallback = callback;
}

void SetFileDropRequestCallback(FileDropRequestCallback callback) {
    g_fdReqCallback = callback;
}

void SetMultiFilesDropRequestCallback(MultiFilesDropRequestCallback callback) {
    g_multiFilesReqCallback = callback;
}

void SetFileDropResponseCallback(FileDropResponseCallback callback) {
    g_fdRespCallback = callback;
}

void SetClipboardCopyImgCallback(ClipboardCopyImgCallback callback) {
    g_cpImgCallback = callback;
}

void StartClipboardMonitor() {
    if (g_running.exchange(true)) {
        return;
    }

    g_event = CreateEvent(NULL, TRUE, FALSE, NULL);
    g_thread = CreateThread(NULL, 0, ClipboardMonitorThread, NULL, 0, NULL);

    WaitForSingleObject(g_event, INFINITE);
    CloseHandle(g_event);
}

void StopClipboardMonitor() {
    if (!g_running.exchange(false)) {
        return;
    }

    PostMessage(g_hWnd, WM_CLOSE, 0, 0);
    WaitForSingleObject(g_thread, INFINITE);
    CloseHandle(g_thread);
    g_thread = NULL;
}

void SetPipeConnectedCallback(PipeConnectedCallback callback) {
    g_pipeConnCallback = callback;
}

void SetGetMacAddressCallback(GetMacAddressCallback callback) {
    g_getMacAddressCallback = callback;
}

void SetExtractDIASCallback(ExtractDIASCallback callback) {
    g_extractDIASCallback = callback;
}

void SetAuthStatusCodeCallback(AuthStatusCodeCallback callback) {
    g_authStatusCodeCallback = callback;
}

void SetDIASSourceAndPortCallback(DIASSourceAndPortCallback callback) {
    g_DIASSourceAndPortCallback = callback;
}

void SetDragFileCallback(DragFileCallback callback) {
    g_dragFileCallback = callback;
}

void SetDragFileListRequestCallback(DragFileListRequestCallback callback) {
    g_dragFileListCallback = callback;
}

void SetCancelFileTransferCallback(CancelFileTransferCallback callback) {
    g_cancelFileTransferCallback = callback;
}

// TODO: consider that transfering multiple files
void SetupDstPasteFile(wchar_t* desc,
                        wchar_t* fileName,
                        unsigned long fileSizeHigh,
                        unsigned long fileSizeLow) {
    if (msPasteImpl) {
        delete msPasteImpl;
        msPasteImpl = NULL;
    }

    msPasteImpl = new MSPasteImpl(g_pscallback);
    FILE_INFO fileInfo = {std::wstring(desc), std::wstring(fileName), fileSizeHigh, fileSizeLow};
    std::vector<FILE_INFO> fileList;
    fileList.push_back(fileInfo);
    msPasteImpl->SetupPasteFile(fileList, clipboardMutex, g_isOleClipboardOperation);
}

void SetupDstPasteImage(wchar_t* desc,
                        IMAGE_HEADER imgHeader,
                        unsigned long dataSize) {
    if (msPasteImpl) {
        delete msPasteImpl;
        msPasteImpl = NULL;
    }

    msPasteImpl = new MSPasteImpl(g_pscallback);
    IMAGE_INFO imgInfo = {desc, imgHeader, dataSize};
    msPasteImpl->SetupPasteImage(imgInfo, clipboardMutex, g_isOleClipboardOperation);
}

void SetupFileDrop(char* ip,
                    char* id,
                    unsigned long long fileSize,
                    unsigned long long timestamp,
                    wchar_t* fileName) {
    if (!msPipeCtrl) {
        return;
    }
    LOG_INFO << "------SetupFileDrop: timeStamp = " << timestamp;
    msPipeCtrl->sendFileRequest(ip, id, fileSize, timestamp, fileName);
}

void DragFileNotify(char *ipPort,
                    char *clientID,
                    unsigned long long fileSize,
                    unsigned long long timestamp,
                    wchar_t *fileName) {
    if (!msPipeCtrl) {
        return;
    }

    msPipeCtrl->dragFileNotify(ipPort, clientID, fileSize, timestamp, fileName);
}


void DragFileListNotify(char *ipPort,
                        char *clientID,
                        unsigned int cFileCount,
                        unsigned long long totalSize,
                        unsigned long long timestamp,
                        wchar_t *firstFileName,
                        unsigned long long firstFileSize) {
    if (!msPipeCtrl) {
        return;
    }
    msPipeCtrl->dragFileListNotify(ipPort, clientID, cFileCount, totalSize, timestamp, firstFileName, firstFileSize);
}

void MultiFilesDropNotify(char *ipPort,
                          char *clientID,
                          unsigned int cFileCount,
                          unsigned long long totalSize,
                          unsigned long long timestamp,
                          wchar_t *firstFileName,
                          unsigned long long firstFileSize)
{
    if (!msPipeCtrl) {
        return;
    }
    // Reuse the dragFileListNotify interface
    msPipeCtrl->dragFileListNotify(ipPort, clientID, cFileCount, totalSize, timestamp, firstFileName, firstFileSize);
}

void UpdateMultipleProgressBar(char *ipPort,
                           char *clientID,
                           wchar_t *currentFileName,
                           unsigned int sentFilesCnt,
                           unsigned int totalFilesCnt,
                           unsigned long long currentFileSize,
                           unsigned long long totalSize,
                           unsigned long long sentSize,
                           unsigned long long timestamp) {
    if (!msPipeCtrl) {
        return;
    }
    msPipeCtrl->updateProgressForMultiFileTransfers(ipPort,
                                                    clientID,
                                                    currentFileName,
                                                    sentFilesCnt,
                                                    totalFilesCnt,
                                                    currentFileSize,
                                                    totalSize,
                                                    sentSize,
                                                    timestamp);
}

void DataTransfer(unsigned char* data, unsigned int size) {
    if (!msPasteImpl) {
        return;
    }

    msPasteImpl->WriteFile(data, size);
}

void UpdateProgressBar(char* ip, char* id, uint64_t fileSize, uint64_t sentSize, uint64_t timestamp, wchar_t* fileName) {
    if (!msPipeCtrl) {
        return;
    }

    msPipeCtrl->updateProgress(ip, id, fileSize, sentSize, timestamp, fileName);
}

void DeinitProgressBar() {
    if (!msFileDrop) {
        return;
    }

    msFileDrop->DeinitProgressBar();
}

void UpdateImageProgressBar(char *ip, char *id, uint64_t fileSize, uint64_t sentSize, uint64_t timestamp)
{
    if (!msPipeCtrl) {
        return;
    }

    msPipeCtrl->updateImageProgress(ip, id, fileSize, sentSize, timestamp);
}

void EventHandle(EVENT_TYPE event) {
    if (!msPasteImpl) {
        return;
    }

    msPasteImpl->EventHandle(event);
}

BOOL APIENTRY DllMain(HMODULE hModule, DWORD ul_reason_for_call, LPVOID lpReserved) {
    switch (ul_reason_for_call) {
    case DLL_PROCESS_ATTACH:
        g_running = false;
        g_running_pipe = false;
        DisableThreadLibraryCalls(hModule);
        break;
    case DLL_PROCESS_DETACH:
        StopClipboardMonitor();
        break;
    }
    return TRUE;
}

void StartPipeMonitor() {
    if (g_running_pipe.exchange(true)) {
        return;
    }

    if (msPipeCtrl == nullptr) {
        msPipeCtrl = new PipeServerControl("CrossSharePipe");
        msPipeCtrl->setPipeConnectedCallback(g_pipeConnCallback);
        msPipeCtrl->setFileDropRequestCallback(g_fdReqCallback);
        msPipeCtrl->setMultiFilesDropRequestCallback(g_multiFilesReqCallback);
        msPipeCtrl->setFileDropResponseCallback(g_fdRespCallback);
        msPipeCtrl->setDragFileCallback(g_dragFileCallback);
        msPipeCtrl->setDragFileListCallback(g_dragFileListCallback);
        msPipeCtrl->setCancelFileTransferCallback(g_cancelFileTransferCallback);

        std::thread([] {
            WindowsEventMonitor::getInstance()->setGetMacAddressCallback(g_getMacAddressCallback);
            WindowsEventMonitor::getInstance()->setExtractDIASCallback(g_extractDIASCallback);
            WindowsEventMonitor::getInstance()->setAuthStatusCodeCallback(g_authStatusCodeCallback);
            WindowsEventMonitor::getInstance()->setDIASSourceAndPortCallback(g_DIASSourceAndPortCallback);

            if (WindowsEventMonitor::getInstance()->initialize() == false) {
                LOG_WARN << "initialize failed ......";
                return;
            }

            LOG_INFO << "WindowsEventMonitor start ......";
            WindowsEventMonitor::getInstance()->start();
        }).detach();

        msPipeCtrl->startServer();
    }
}

void StopPipeMonitor() {
//    if (!g_running_pipe.exchange(false)) {
//        return;
//    }

    if (msPipeCtrl) {
        //The service remains running, but all connections are forcibly closed
        msPipeCtrl->closeAllConnection();
    }
}

void UpdateClientStatus(unsigned int status, char *ip, char *id, wchar_t *name, char *deviceType) {
    if (!msPipeCtrl) {
        return;
    }

    msPipeCtrl->updateClientStatus(status, ip, id, name, deviceType);
}

void UpdateSystemInfo(char *ip, wchar_t *serviceVer) {
    if (!msPipeCtrl) {
        return;
    }

    msPipeCtrl->updateSystemInfo(ip, serviceVer);
}

void NotiMessage(uint64_t timestamp, unsigned int notiCode, wchar_t *notiParam[], int paramCount) {
    if (!msPipeCtrl) {
        return;
    }
    std::vector<std::wstring> paramInfoVec;
    paramInfoVec.reserve(paramCount);
    for (int index = 0; index < paramCount; ++index) {
        paramInfoVec.push_back(notiParam[index]);
    }
    msPipeCtrl->notiMessage(timestamp, notiCode, paramInfoVec);
}

void CleanClipboard() {
    msPasteImpl->CleanClipboard();
}

const wchar_t* GetDeviceName() {
    static std::wstring deviceName;
    wchar_t name[MAX_COMPUTERNAME_LENGTH+1];
    DWORD size = MAX_COMPUTERNAME_LENGTH+1;
    if (GetComputerNameW(name, &size)) {
        deviceName = std::wstring(name, size);
        return deviceName.c_str();
    } else {
        return L"";
    }
}

void AuthViaIndex(uint32_t index) {
    WindowsEventMonitor::getInstance()->authViaIndex(index);
}

void DIASStatus(unsigned int statusCode) {
    if (!msPipeCtrl) {
        return;
    }
    msPipeCtrl->statusInfoNotify(statusCode);
}

void RequestSourceAndPort() {
    WindowsEventMonitor::getInstance()->requestSourcePort();
}

const wchar_t* GetDownloadPath() {
    const GUID FOLDERID_Downloads = {
        0x374DE290, 0x123F, 0x4565,
        { 0x91, 0x64, 0x39, 0xC4, 0x92, 0x5E, 0x46, 0x7B }
    };
    PWSTR path = nullptr;
    HRESULT hr = SHGetKnownFolderPath(FOLDERID_Downloads, 0, NULL, &path);
    if (SUCCEEDED(hr)) {
        return path;
    } else {
        return L"";
    }
}
