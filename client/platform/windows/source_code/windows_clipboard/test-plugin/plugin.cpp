#include "plugin.h"
#include <QDateTime>
#include "control_window.h"

StartClipboardMonitorCallback g_StartClipboardMonitor = nullptr;
StopClipboardMonitorCallback g_StopClipboardMonitor = nullptr;
DragFileListNotifyCallback g_DragFileListNotify = nullptr;
MultiFilesDropNotifyCallback g_MultiFilesDropNotify = nullptr;
UpdateMultipleProgressBarCallback g_UpdateMultipleProgressBar = nullptr;
UpdateClientStatusCallback g_UpdateClientStatus = nullptr;
UpdateClientStatusExCallback g_UpdateClientStatusExCallback = nullptr;
UpdateSystemInfoCallback g_UpdateSystemInfo = nullptr;
NotiMessageCallback g_NotiMessage = nullptr;
CleanClipboardCallback g_CleanClipboard = nullptr;
AuthViaIndexCallback g_AuthViaIndex = nullptr;
DIASStatusCallback g_DIASStatus = nullptr;
RequestSourceAndPortCallback g_RequestSourceAndPort = nullptr;
RequestUpdateClientVersionCallback g_RequestUpdateClientVersionCallback = nullptr;
NotifyErrEventCallback g_NotifyErrEventCallback = nullptr;
SetupDstPasteXClipDataCallback g_SetupDstPasteXClipDataCallback = nullptr;

ControlWindow *g_testWindow = nullptr;
QString g_rootPath;
QString g_downloadPath;
QString g_deviceName;

EXPORT_FUNC void InitGoServer(const wchar_t *rootPath, const wchar_t *downloadPath, const wchar_t *deviceName)
{
    qInfo() << "----------------------------------[InitGoServer]----------------------------------------";
    g_rootPath = QString::fromStdWString(rootPath);
    g_downloadPath = QString::fromStdWString(downloadPath);
    g_deviceName = QString::fromStdWString(deviceName);
    qInfo() << "g_rootPath:" << g_rootPath << ";"
            << "g_downloadPath:" << g_downloadPath << ";"
            << "g_deviceName:" << g_deviceName;

    QTimer::singleShot(0, qApp, [] {
        g_testWindow = new ControlWindow;
        g_testWindow->show();

        g_UpdateSystemInfo("192.168.6.123:99", L"server_v1.0.9");
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

EXPORT_FUNC void RequestUpdateDownloadPath(const wchar_t *downloadPath)
{
    qInfo() << QString::fromStdWString(downloadPath);
}

EXPORT_FUNC void SendXClipData(const char *textData, const char *imageData, const char *htmlData)
{
    g_testWindow->writeClipboardData(textData, imageData, htmlData);
}

EXPORT_FUNC const char *GetClientList()
{
    static std::string s_cacheClientListJsonData;
    static std::mutex s_mutex;
    std::lock_guard<std::mutex> locker(s_mutex);
    return s_cacheClientListJsonData.data();
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

EXPORT_FUNC void SetUpdateClientStatusCallback(UpdateClientStatusCallback callback)
{
    qInfo() << "[SetUpdateClientStatusCallback] Address: "
              << reinterpret_cast<void*>(callback);
    g_UpdateClientStatus = callback;
}

EXPORT_FUNC void SetUpdateClientStatusExCallback(UpdateClientStatusExCallback callback)
{
    qInfo() << "[SetUpdateClientStatusExCallback] Address: "
            << reinterpret_cast<void*>(callback);
    g_UpdateClientStatusExCallback = callback;
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

EXPORT_FUNC void SetRequestUpdateClientVersionCallback(RequestUpdateClientVersionCallback callback)
{
    qInfo() << "[SetRequestUpdateClientVersionCallback] Address: "
            << reinterpret_cast<void*>(callback);
    g_RequestUpdateClientVersionCallback = callback;
}

EXPORT_FUNC void SetNotifyErrEventCallback(NotifyErrEventCallback callback)
{
    qInfo() << "[SetNotifyErrEventCallback] Address: "
            << reinterpret_cast<void*>(callback);
    g_NotifyErrEventCallback = callback;
}

EXPORT_FUNC void SetSetupDstPasteXClipDataCallback(SetupDstPasteXClipDataCallback callback)
{
    qInfo() << "[SetSetupDstPasteXClipDataCallback] Address: "
            << reinterpret_cast<void*>(callback);
    g_SetupDstPasteXClipDataCallback = callback;
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
        std::cerr << "--------------------------DLL Process Detached" << std::endl;
        break;
    case DLL_THREAD_ATTACH:
    case DLL_THREAD_DETACH:
        break;
    }
    return TRUE;
}
