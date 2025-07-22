#include "plugin.h"
#include <QDateTime>
#include "control_window.h"

StartClipboardMonitorCallback g_StartClipboardMonitor = nullptr;
StopClipboardMonitorCallback g_StopClipboardMonitor = nullptr;
DragFileListNotifyCallback g_DragFileListNotify = nullptr;
MultiFilesDropNotifyCallback g_MultiFilesDropNotify = nullptr;
UpdateMultipleProgressBarCallback g_UpdateMultipleProgressBar = nullptr;
DataTransferCallback g_DataTransfer = nullptr;
UpdateClientStatusCallback g_UpdateClientStatus = nullptr;
UpdateSystemInfoCallback g_UpdateSystemInfo = nullptr;
NotiMessageCallback g_NotiMessage = nullptr;
CleanClipboardCallback g_CleanClipboard = nullptr;
GetDeviceNameCallback g_GetDeviceName = nullptr;
AuthViaIndexCallback g_AuthViaIndex = nullptr;
DIASStatusCallback g_DIASStatus = nullptr;
RequestSourceAndPortCallback g_RequestSourceAndPort = nullptr;
GetDownloadPathCallback g_GetDownloadPath = nullptr;
SetupDstPasteImageCallback g_SetupDstPasteImage = nullptr;

ControlWindow *g_testWindow = nullptr;

void testFunc();

EXPORT_FUNC void InitGoServer()
{
    qInfo() << "----------------------------------[InitGoServer]----------------------------------------";
    QTimer::singleShot(0, qApp, [] {
        g_testWindow = new ControlWindow;
        g_testWindow->show();

        g_UpdateSystemInfo("192.168.6.123:99", L"server_v1.0.9");
    });
}

EXPORT_FUNC void SetClipboardCopyImg(IMAGE_HEADER picHeader, unsigned char* bitmapData, unsigned long dataSize)
{
    Q_UNUSED(bitmapData)
    nlohmann::json jsonInfo;
    jsonInfo["desc"] = "SetClipboardCopyImg";
    jsonInfo["width"] = picHeader.width;
    jsonInfo["height"] = picHeader.height;
    jsonInfo["planes"] = picHeader.planes;
    jsonInfo["bitCount"] = picHeader.bitCount;
    jsonInfo["compression"] = picHeader.compression;
    jsonInfo["dataSize"] = dataSize;
    Q_EMIT CommonSignals::getInstance()->logMessage(jsonInfo.dump(4).c_str());

    std::string imageData(reinterpret_cast<const char*>(bitmapData), dataSize);
    QTimer::singleShot(1000, qApp, [picHeader, imageData] {
        g_CleanClipboard();
        g_SetupDstPasteImage(L"test image paste......", picHeader, imageData.size());
        g_DataTransfer(reinterpret_cast<const unsigned char*>(imageData.data()), imageData.size());
    });
}

EXPORT_FUNC void SetMacAddress(const char* macAddress, int length)
{
    std::string address(macAddress, length);
    Q_EMIT CommonSignals::getInstance()->logMessage("macAddress: " + QByteArray::fromStdString(address).toHex().toUpper());
}

EXPORT_FUNC void SetExtractDIAS()
{
    Q_EMIT CommonSignals::getInstance()->logMessage("[SetExtractDIAS] called......");
}

EXPORT_FUNC void SetAuthStatusCode(unsigned char authResult)
{
    Q_EMIT CommonSignals::getInstance()->logMessage(QString("SetAuthStatusCode: authResult=%1").arg(authResult));
}

EXPORT_FUNC void SetDIASSourceAndPort(unsigned char source, unsigned char port)
{
    Q_EMIT CommonSignals::getInstance()->logMessage(QString("SetDIASSourceAndPort: source=%1, port=%2").arg(source).arg(port));
}

EXPORT_FUNC void SetDragFileListRequest(const wchar_t* filePathArry[], uint32_t arryLength, uint64_t timeStamp)
{
    qInfo() << "[SetDragFileListRequest]\n"
               << "  Count: " << arryLength << L"\n"
               << "  Timestamp: " << timeStamp;
    for (uint32_t i = 0; i < arryLength; ++i) {
        qInfo() << "  File[" << i << "]: "
                << QString::fromStdWString((filePathArry[i] ? filePathArry[i] : L"NULL"));
    }
}

EXPORT_FUNC void SetCancelFileTransfer(const char* ipPort, const char* clientID, uint64_t timeStamp)
{
    g_testWindow->on_SetCancelFileTransfer(ipPort, clientID, timeStamp);
}

EXPORT_FUNC void SetMultiFilesDropRequest(const char* ipPort, const char* clientID, uint64_t timeStamp,
                                          const wchar_t* filePathArry[], uint32_t arryLength)
{
    qInfo() << "[SetMultiFilesDropRequest]\n"
              << "  IP: " << (ipPort ? ipPort : "NULL") << "\n"
              << "  ClientID: " << (clientID ? clientID : "NULL") << "\n"
              << "  Timestamp: " << timeStamp << "\n"
              << "  FileCount: " << arryLength;
    for (uint32_t i = 0; i < arryLength; ++i) {
        qInfo() << "  File[" << i << "]: "
                   << QString::fromStdWString((filePathArry[i] ? filePathArry[i] : L"NULL"));
    }
}


//--------------------------------------callback----------------------------------------
EXPORT_FUNC void SetStartClipboardMonitorCallback(StartClipboardMonitorCallback callback)
{
    qInfo() << "[SetStartClipboardMonitorCallback] Address: "
              << reinterpret_cast<void*>(callback);
    g_StartClipboardMonitor = callback;
}

EXPORT_FUNC void SetStopClipboardMonitorCallback(StopClipboardMonitorCallback callback)
{
    qInfo() << "[SetStopClipboardMonitorCallback] Address: "
              << reinterpret_cast<void*>(callback);
    g_StopClipboardMonitor = callback;
}

EXPORT_FUNC void SetDragFileListNotifyCallback(DragFileListNotifyCallback callback)
{
    qInfo() << "[SetDragFileListNotifyCallback] Address: "
              << reinterpret_cast<void*>(callback);
    g_DragFileListNotify = callback;
}

EXPORT_FUNC void SetMultiFilesDropNotifyCallback(MultiFilesDropNotifyCallback callback)
{
    qInfo() << "[SetMultiFilesDropNotifyCallback] Address: "
              << reinterpret_cast<void*>(callback);
    g_MultiFilesDropNotify = callback;
}

EXPORT_FUNC void SetUpdateMultipleProgressBarCallback(UpdateMultipleProgressBarCallback callback)
{
    qInfo() << "[SetUpdateMultipleProgressBarCallback] Address: "
              << reinterpret_cast<void*>(callback);
    g_UpdateMultipleProgressBar = callback;
}

EXPORT_FUNC void SetDataTransferCallback(DataTransferCallback callback)
{
    qInfo() << "[SetDataTransferCallback] Address: "
              << reinterpret_cast<void*>(callback);
    g_DataTransfer = callback;
}

EXPORT_FUNC void SetUpdateClientStatusCallback(UpdateClientStatusCallback callback)
{
    qInfo() << "[SetUpdateClientStatusCallback] Address: "
              << reinterpret_cast<void*>(callback);
    g_UpdateClientStatus = callback;
}

EXPORT_FUNC void SetUpdateSystemInfoCallback(UpdateSystemInfoCallback callback)
{
    qInfo() << "[SetUpdateSystemInfoCallback] Address: "
              << reinterpret_cast<void*>(callback);
    g_UpdateSystemInfo = callback;
}

EXPORT_FUNC void SetNotiMessageCallback(NotiMessageCallback callback)
{
    qInfo() << "[SetNotiMessageCallback] Address: "
              << reinterpret_cast<void*>(callback);
    g_NotiMessage = callback;
}

EXPORT_FUNC void SetCleanClipboardCallback(CleanClipboardCallback callback)
{
    qInfo() << "[SetCleanClipboardCallback] Address: "
              << reinterpret_cast<void*>(callback);
    g_CleanClipboard = callback;
}

EXPORT_FUNC void SetGetDeviceNameCallback(GetDeviceNameCallback callback)
{
    qInfo() << "[SetGetDeviceNameCallback] Address: "
              << reinterpret_cast<void*>(callback);
    g_GetDeviceName = callback;
}

EXPORT_FUNC void SetAuthViaIndexCallback(AuthViaIndexCallback callback)
{
    qInfo() << "[SetAuthViaIndexCallback] Address: "
              << reinterpret_cast<void*>(callback);
    g_AuthViaIndex = callback;
}

EXPORT_FUNC void SetDIASStatusCallback(DIASStatusCallback callback)
{
    qInfo() << "[SetDIASStatusCallback] Address: "
              << reinterpret_cast<void*>(callback);
    g_DIASStatus = callback;
}

EXPORT_FUNC void SetRequestSourceAndPortCallback(RequestSourceAndPortCallback callback)
{
    qInfo() << "[SetRequestSourceAndPortCallback] Address: "
              << reinterpret_cast<void*>(callback);
    g_RequestSourceAndPort = callback;
}

EXPORT_FUNC void SetGetDownloadPathCallback(GetDownloadPathCallback callback)
{
    qInfo() << "[SetGetDownloadPathCallback] Address: "
              << reinterpret_cast<void*>(callback);
    g_GetDownloadPath = callback;
}

EXPORT_FUNC void SetSetupDstPasteImageCallback(SetupDstPasteImageCallback callback)
{
    qInfo() << "[SetSetupDstPasteImageCallback] Address: "
              << reinterpret_cast<void*>(callback);
    g_SetupDstPasteImage = callback;
}

BOOL APIENTRY DllMain(HMODULE hModule, DWORD  ul_reason_for_call, LPVOID lpReserved)
{
    Q_UNUSED(hModule)
    Q_UNUSED(lpReserved)
    switch (ul_reason_for_call) {
    case DLL_PROCESS_ATTACH:
        qInfo() << "--------------------------DLL Process Attached";
        break;
    case DLL_PROCESS_DETACH:
        qInfo() << "--------------------------DLL Process Detached";
        break;
    case DLL_THREAD_ATTACH:
    case DLL_THREAD_DETACH:
        break;
    }
    return TRUE;
}
