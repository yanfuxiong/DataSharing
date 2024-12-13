// g++ -shared -o clipboard.dll clipboard/clipboard.cpp -lgdi32
#include <string>
#include <atomic>
#include <iostream>
#include <mutex>
#include "MSPaste/MSPasteImpl.h"
#include "MSPaste/MSFileDrop.h"
#include "MSPipeServer/MSPipeController.h"

HWND g_hWnd = NULL;
std::atomic<bool> g_running;
std::atomic<bool> g_running_pipe;
HANDLE g_event;
HANDLE g_thread = NULL;
bool g_isOleClipboardOperation = false;
std::mutex clipboardMutex;
MSPasteImpl *msPasteImpl = NULL;
MSFileDrop *msFileDrop = NULL;
MSPipeController *msPipeCtrl = NULL;

typedef void (*ClipboardCopyFileCallback)(wchar_t*, unsigned long, unsigned long);
ClipboardCopyFileCallback g_cpFilecallback = nullptr;
typedef void (*ClipboardPasteFileCallback)(char*);
ClipboardPasteFileCallback g_pscallback = nullptr;
typedef void (*ClipboardCopyImgCallback)(IMAGE_HEADER, unsigned char*, unsigned long);
ClipboardCopyImgCallback g_cpImgCallback = nullptr;
typedef void (*FileDropRequestCallback)(char*, char*, unsigned long long, unsigned long long, wchar_t*);
FileDropRequestCallback g_fdReqCallback = nullptr;
typedef void (*FileDropResponseCallback)(int, char*, char*, unsigned long long, unsigned long long, wchar_t*);
FileDropResponseCallback g_fdRespCallback = nullptr;
typedef void (*PipeConnectedCallback)(void);
PipeConnectedCallback g_pipeConnCallback = nullptr;

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
        BITMAPINFO bmpInfo = {0};
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

            if (isOleOperation) { return 0; }

            if (OpenClipboard(NULL)) {
                if (IsClipboardFormatAvailable(CF_HDROP)) {
                    HANDLE hData = GetClipboardData(CF_HDROP);
                    if (hData) {
                        HDROP hDrop = static_cast<HDROP>(hData);
                        UINT fileCount = DragQueryFile(hDrop, 0xFFFFFFFF, NULL, 0);
                        if (fileCount > 0) {
                            wchar_t filePath[MAX_PATH] = {};
                            DragQueryFileW(hDrop, 0, filePath, _countof(filePath));

                            unsigned long fileSizeHigh, fileSizeLow;
                            GetFileSizeW(filePath, fileSizeHigh, fileSizeLow);
                            if (g_cpFilecallback) {
                                if (fileSizeHigh > 0 || fileSizeLow > 0) {
                                    g_cpFilecallback(filePath, fileSizeHigh, fileSizeLow);
                                } else {
                                    DEBUG_LOG("[%s %d] Skip copy file: Empty fil", __func__, __LINE__);
                                }
                            }
                            // DragFinish(hDrop);
                        }
                    }
                }
                else if (IsClipboardFormatAvailable(CF_BITMAP)) {
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

DWORD WINAPI ClipboardMonitorThread(LPVOID lpParam) {
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

extern "C" __declspec(dllexport) void SetClipboardCopyFileCallback(ClipboardCopyFileCallback callback) {
    g_cpFilecallback = callback;
}

extern "C" __declspec(dllexport) void SetClipboardPasteFileCallback(ClipboardPasteFileCallback callback) {
    g_pscallback = callback;
}

extern "C" __declspec(dllexport) void SetFileDropRequestCallback(FileDropRequestCallback callback) {
    g_fdReqCallback = callback;
}

extern "C" __declspec(dllexport) void SetFileDropResponseCallback(FileDropResponseCallback callback) {
    g_fdRespCallback = callback;
}

extern "C" __declspec(dllexport) void SetClipboardCopyImgCallback(ClipboardCopyImgCallback callback) {
    g_cpImgCallback = callback;
}

extern "C" __declspec(dllexport) void StartClipboardMonitor() {
    if (g_running.exchange(true)) {
        return;
    }

    g_event = CreateEvent(NULL, TRUE, FALSE, NULL);
    g_thread = CreateThread(NULL, 0, ClipboardMonitorThread, NULL, 0, NULL);

    WaitForSingleObject(g_event, INFINITE);
    CloseHandle(g_event);
}

extern "C" __declspec(dllexport) void StopClipboardMonitor() {
    if (!g_running.exchange(false)) {
        return;
    }

    PostMessage(g_hWnd, WM_CLOSE, 0, 0);
    WaitForSingleObject(g_thread, INFINITE);
    CloseHandle(g_thread);
    g_thread = NULL;
}

extern "C" __declspec(dllexport) void SetPipeConnectedCallback(PipeConnectedCallback callback) {
    g_pipeConnCallback = callback;
}

// TODO: consider that transfering multiple files
extern "C" __declspec(dllexport) void SetupDstPasteFile(wchar_t* desc,
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

extern "C" __declspec(dllexport) void SetupDstPasteImage(wchar_t* desc,
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

extern "C" __declspec(dllexport) void SetupFileDrop(char* ip,
                                                        char* id,
                                                        unsigned long long fileSize,
                                                        unsigned long long timestamp,
                                                        wchar_t* fileName) {
    if (!msPipeCtrl) {
        return;
    }

    msPipeCtrl->SendFileReq(ip, id, fileSize, timestamp, fileName);
}

extern "C" __declspec(dllexport) void DataTransfer(unsigned char* data, unsigned int size) {
    if (!msPasteImpl) {
        return;
    }

    msPasteImpl->WriteFile(data, size);
}

extern "C" __declspec(dllexport) void UpdateProgressBar(char* ip, char* id, uint64_t fileSize, uint64_t sentSize, uint64_t timestamp, wchar_t* fileName) {
    if (!msPipeCtrl) {
        return;
    }

    msPipeCtrl->UpdateProgress(ip, id, fileSize, sentSize, timestamp, fileName);
}

extern "C" __declspec(dllexport) void DeinitProgressBar() {
    if (!msFileDrop) {
        return;
    }

    msFileDrop->DeinitProgressBar();
}

extern "C" __declspec(dllexport) void EventHandle(EVENT_TYPE event) {
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

extern "C" __declspec(dllexport) void StartPipeMonitor() {
    if (g_running_pipe.exchange(true)) {
        return;
    }

    if (msPipeCtrl) {
        delete msPipeCtrl;
        msPipeCtrl = NULL;
    }
    msPipeCtrl = new MSPipeController(g_running_pipe, g_fdReqCallback, g_fdRespCallback, g_pipeConnCallback);
}

extern "C" __declspec(dllexport) void StopPipeMonitor() {
    if (!g_running_pipe.exchange(false)) {
        return;
    }

    if (msPipeCtrl) {
        delete msPipeCtrl;
        msPipeCtrl = NULL;
    }
}

extern "C" __declspec(dllexport) void UpdateClientStatus(unsigned int status, char* ip, char* id, wchar_t* name) {
    if (!msPipeCtrl) {
        return;
    }

    msPipeCtrl->UpdateClientStatus(status, ip, id, name);
}

extern "C" __declspec(dllexport) void CleanClipboard() {
    msPasteImpl->CleanClipboard();
}
