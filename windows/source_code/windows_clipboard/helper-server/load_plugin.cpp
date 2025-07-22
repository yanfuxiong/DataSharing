#include "load_plugin.h"
#include "windows_event_monitor.h"
#include <QLibrary>
#include <QDebug>
#include <QDateTime>
#include <QDir>
#include <QSysInfo>
#include <QApplication>
#include <QImage>
#include <windows.h>

namespace {

InitGoServer g_InitGoServer = nullptr;
SetClipboardCopyImg g_SetClipboardCopyImg = nullptr;
SetMacAddress g_SetMacAddress = nullptr;
SetExtractDIAS g_SetExtractDIAS = nullptr;
SetAuthStatusCode g_SetAuthStatusCode = nullptr;
SetDIASSourceAndPort g_SetDIASSourceAndPort = nullptr;
SetDragFileListRequest g_SetDragFileListRequest = nullptr;
SetCancelFileTransfer g_SetCancelFileTransfer = nullptr;
SetMultiFilesDropRequest g_SetMultiFilesDropRequest = nullptr;

SetStartClipboardMonitorCallback g_SetStartClipboardMonitorCallback = nullptr;
SetStopClipboardMonitorCallback g_SetStopClipboardMonitorCallback = nullptr;
SetDragFileListNotifyCallback g_SetDragFileListNotifyCallback = nullptr;
SetMultiFilesDropNotifyCallback g_SetMultiFilesDropNotifyCallback = nullptr;
SetUpdateMultipleProgressBarCallback g_SetUpdateMultipleProgressBarCallback = nullptr;
SetDataTransferCallback g_SetDataTransferCallback = nullptr;
SetUpdateClientStatusCallback g_SetUpdateClientStatusCallback = nullptr;
SetUpdateSystemInfoCallback g_SetUpdateSystemInfoCallback = nullptr;
SetNotiMessageCallback g_SetNotiMessageCallback = nullptr;
SetCleanClipboardCallback g_SetCleanClipboardCallback = nullptr;
SetGetDeviceNameCallback g_SetGetDeviceNameCallback = nullptr;
SetAuthViaIndexCallback g_SetAuthViaIndexCallback = nullptr;
SetDIASStatusCallback g_SetDIASStatusCallback = nullptr;
SetRequestSourceAndPortCallback g_SetRequestSourceAndPortCallback = nullptr;
SetGetDownloadPathCallback g_SetGetDownloadPathCallback = nullptr;
SetSetupDstPasteImageCallback g_SetSetupDstPasteImageCallback = nullptr;

//----------------------------------------------------------------------------
UpdateSystemInfoMsgPtr g_cacheUpdateSystemInfoMsg = nullptr;
StatusInfoNotifyMsgPtr g_cacheStatusInfoNotifyMsg = nullptr;
std::atomic<bool> g_clipboardMonitoringStatus { false };
QString g_getPluginName()
{
#ifndef NDEBUG
    return "libtest-plugin.dll";
#else
    try {
        return g_getGlobalData()->localConfig.at("crossShareServer").at("goServerDllName").get<std::string>().c_str();
    } catch (const std::exception &e) {
        qWarning() << "g_getPluginName:" << e.what();
        return "client_windows.dll";
    }
#endif
}

void getBitmapData(HBITMAP hBitmap)
{
    BITMAP bitmap;
    if (GetObject(hBitmap, sizeof(BITMAP), &bitmap)) {
        HDC hdc = GetDC(NULL);
        BITMAPINFO bmpInfo;
        std::memset(&bmpInfo, 0, sizeof (bmpInfo));
        bmpInfo.bmiHeader.biSize = sizeof(BITMAPINFOHEADER);
        bmpInfo.bmiHeader.biWidth = bitmap.bmWidth;
        bmpInfo.bmiHeader.biHeight = bitmap.bmHeight;
        bmpInfo.bmiHeader.biPlanes = bitmap.bmPlanes;
        bmpInfo.bmiHeader.biBitCount = bitmap.bmBitsPixel;
        bmpInfo.bmiHeader.biCompression = BI_RGB;

        const int dataSize = bitmap.bmWidthBytes * bitmap.bmHeight;
        BYTE* bitmapData = new BYTE[dataSize];
        if (!GetDIBits(hdc, hBitmap, 0, bitmap.bmHeight, bitmapData, &bmpInfo, DIB_RGB_COLORS)) {
            qWarning() << "Get DIBits failed......";
        }

        IMAGE_HEADER picHeader = {
            .width = bmpInfo.bmiHeader.biWidth,
            .height = bmpInfo.bmiHeader.biHeight,
            .planes = bmpInfo.bmiHeader.biPlanes,
            .bitCount = bmpInfo.bmiHeader.biBitCount,
            .compression = bmpInfo.bmiHeader.biCompression
        };
        qInfo("[Copy] Trigger copy image. H=%d,W=%d,Planes=%d,BitCnt=%d,Compress=%lu",
                 picHeader.height, picHeader.width, picHeader.planes, picHeader.bitCount, picHeader.compression);
        g_SetClipboardCopyImg(picHeader, bitmapData, dataSize);

        delete[] bitmapData;
        ReleaseDC(NULL, hdc);
    } else {
        qWarning() << "Failed to get bitmap object details......";
    }
}

void setupDstPasteImage_helper(const wchar_t* desc, IMAGE_HEADER imgHeader, uint32_t dataSize, const BYTE* imageData)
{
    if (dataSize == 0 || imageData == nullptr) {
        qWarning("Invalid image data: size=%u, pointer=%p", dataSize, imageData);
        return;
    }

    // Convert raw data to QImage (need to handle the inverted Y - axis of Windows DIB)
    QImage::Format format = QImage::Format_Invalid;
    switch (imgHeader.bitCount) {
    case 32: format = QImage::Format_ARGB32; break;
    case 24: format = QImage::Format_RGB32; break;
    case 16: format = QImage::Format_RGB16; break;
    case 8:  format = QImage::Format_Grayscale8; break;
    default:
        qWarning("Unsupported bit count: %d", imgHeader.bitCount);
        return;
    }

    // Calculate the number of bytes per line (Windows requires 4 - byte alignment per line)
    const int stride = ((imgHeader.width * imgHeader.bitCount + 31) / 32) * 4;
    const int calculatedSize = stride * imgHeader.height;

    if (static_cast<int>(dataSize) < calculatedSize) {
        qWarning("Data size mismatch: header=%d, calculated=%d", dataSize, calculatedSize);
        return;
    }

    // Create a temporary buffer to flip the image (Windows DIB is stored upside - down).
    QVector<uchar> flippedData(dataSize);
    const uchar* src = imageData;
    uchar* dst = flippedData.data() + dataSize - stride; // Point to the last line

    for (int y = 0; y < imgHeader.height; ++y) {
        memcpy(dst, src, stride);
        src += stride;
        dst -= stride;
    }

    // Create a QImage (using the flipped data)
    QImage img(flippedData.data(),
               imgHeader.width,
               imgHeader.height,
               stride,
               format);

    if (img.isNull()) {
        qWarning("Failed to create QImage from data");
        return;
    }

    // Convert it to QPixmap and set it to the clipboard (using Qt API)
    QPixmap pixmap = QPixmap::fromImage(img);
    if (pixmap.isNull()) {
        qWarning("Failed to convert image to pixmap");
        return;
    }

    if (g_clipboardMonitoringStatus) {
        QObject::disconnect(QApplication::clipboard(), &QClipboard::changed, LoadPlugin::getInstance(), &LoadPlugin::onClipboardChanged);
    }
    QGuiApplication::clipboard()->setPixmap(pixmap);
    qInfo("Clipboard set: %ls | %dx%d %d-bpp",
          desc, imgHeader.width, imgHeader.height, imgHeader.bitCount);
    QTimer::singleShot(0, qApp, [] {
        if (g_clipboardMonitoringStatus) {
            QObject::connect(QApplication::clipboard(), &QClipboard::changed, LoadPlugin::getInstance(), &LoadPlugin::onClipboardChanged);
        }
    });
}

struct RecvImageInfo
{
    QString desc;
    IMAGE_HEADER imgHeader;
    uint32_t dataSize;
    std::string imageData;

    RecvImageInfo(const QString &descInfo, const IMAGE_HEADER &imgHeaderInfo, uint32_t dataSizeInfo)
    {
        desc = descInfo;
        imgHeader = imgHeaderInfo;
        dataSize = dataSizeInfo;
        imageData.reserve(dataSize);
    }

    ~RecvImageInfo() = default;
};

std::unique_ptr<RecvImageInfo> g_cacheImageInfo { nullptr };

}

LoadPlugin *LoadPlugin::m_instance { nullptr };
LoadPlugin::LoadPlugin()
    : QObject(nullptr)
    , m_mainThread(nullptr)
    , m_copyIndex(0)
{
#ifndef NDEBUG
    QString customPath = qApp->applicationDirPath() + "/../test-plugin";
    QByteArray path = qgetenv("PATH");
    path.prepend(QDir::toNativeSeparators(customPath + ";").toLocal8Bit());
    qputenv("PATH", path);
#endif
    {
        connect(CommonSignals::getInstance(), &CommonSignals::pipeConnected, this, &LoadPlugin::onPipeConnected);
        connect(MonitorPlugEvent::getInstance(), &MonitorPlugEvent::statusChanged, this, &LoadPlugin::onDIASStatusChanged);
    }
}

LoadPlugin::~LoadPlugin()
{

}

LoadPlugin *LoadPlugin::getInstance()
{
    if (m_instance == nullptr) {
        m_instance = new LoadPlugin;
    }
    return m_instance;
}

void LoadPlugin::initPlugin()
{
    if (initDllFunctions() == false) {
        return;
    }
    std::thread([&] {
        qInfo() << "---------------------------g_InitGoServer()-----------------------";
        g_InitGoServer();
    }).detach();

    QTimer::singleShot(1000, this, [] {
        MonitorPlugEvent::getInstance()->initData();
    });
}

void LoadPlugin::runInLoop(const std::function<void()> &callback)
{
    m_mainThread.runInThread(callback);
}

void LoadPlugin::cancelFileTransfer(const char *ipPort, const char *clientID, uint64_t timeStamp)
{
    g_SetCancelFileTransfer(ipPort, clientID, timeStamp);
}

void LoadPlugin::dragFileListRequest(const wchar_t *filePathArry[], uint32_t arryLength , uint64_t timeStamp)
{
    g_SetDragFileListRequest(filePathArry, arryLength, timeStamp);
}

void LoadPlugin::multiFilesDropRequest(const char *ipPort, const char *clientID, uint64_t timeStamp, const wchar_t *filePathArry[], uint32_t arryLength)
{
    g_SetMultiFilesDropRequest(ipPort, clientID, timeStamp, filePathArry, arryLength);
}

bool LoadPlugin::initDllFunctions()
{
    QLibrary lib(g_getPluginName());
    if (!lib.load()) {
        qFatal("Failed to load DLL: %s", lib.errorString().toUtf8().constData());
        return false;
    }
#define RESOLVE_SYMBOL(name) \
    [&]() -> name { \
        auto ptr = reinterpret_cast<name>(lib.resolve(#name)); \
        if (!ptr) { \
            qFatal("Failed to resolve symbol: %s, Error: %s", #name, lib.errorString().toUtf8().constData()); \
        } \
        return ptr; \
    }()

    g_InitGoServer = RESOLVE_SYMBOL(InitGoServer);
    g_SetClipboardCopyImg = RESOLVE_SYMBOL(SetClipboardCopyImg);
    g_SetMacAddress = RESOLVE_SYMBOL(SetMacAddress);
    g_SetExtractDIAS = RESOLVE_SYMBOL(SetExtractDIAS);
    g_SetAuthStatusCode = RESOLVE_SYMBOL(SetAuthStatusCode);
    g_SetDIASSourceAndPort = RESOLVE_SYMBOL(SetDIASSourceAndPort);
    g_SetDragFileListRequest = RESOLVE_SYMBOL(SetDragFileListRequest);
    g_SetCancelFileTransfer = RESOLVE_SYMBOL(SetCancelFileTransfer);
    g_SetMultiFilesDropRequest = RESOLVE_SYMBOL(SetMultiFilesDropRequest);

    g_SetStartClipboardMonitorCallback = RESOLVE_SYMBOL(SetStartClipboardMonitorCallback);
    g_SetStopClipboardMonitorCallback = RESOLVE_SYMBOL(SetStopClipboardMonitorCallback);
    g_SetDragFileListNotifyCallback = RESOLVE_SYMBOL(SetDragFileListNotifyCallback);
    g_SetMultiFilesDropNotifyCallback = RESOLVE_SYMBOL(SetMultiFilesDropNotifyCallback);
    g_SetUpdateMultipleProgressBarCallback = RESOLVE_SYMBOL(SetUpdateMultipleProgressBarCallback);
    g_SetDataTransferCallback = RESOLVE_SYMBOL(SetDataTransferCallback);
    g_SetUpdateClientStatusCallback = RESOLVE_SYMBOL(SetUpdateClientStatusCallback);
    g_SetUpdateSystemInfoCallback = RESOLVE_SYMBOL(SetUpdateSystemInfoCallback);
    g_SetNotiMessageCallback = RESOLVE_SYMBOL(SetNotiMessageCallback);
    g_SetCleanClipboardCallback = RESOLVE_SYMBOL(SetCleanClipboardCallback);
    g_SetGetDeviceNameCallback = RESOLVE_SYMBOL(SetGetDeviceNameCallback);
    g_SetAuthViaIndexCallback = RESOLVE_SYMBOL(SetAuthViaIndexCallback);
    g_SetDIASStatusCallback = RESOLVE_SYMBOL(SetDIASStatusCallback);
    g_SetRequestSourceAndPortCallback = RESOLVE_SYMBOL(SetRequestSourceAndPortCallback);
    g_SetGetDownloadPathCallback = RESOLVE_SYMBOL(SetGetDownloadPathCallback);
    g_SetSetupDstPasteImageCallback = RESOLVE_SYMBOL(SetSetupDstPasteImageCallback);

    if (g_SetStartClipboardMonitorCallback) {
        g_SetStartClipboardMonitorCallback(onStartClipboardMonitor);
    }

    if (g_SetStopClipboardMonitorCallback) {
        g_SetStopClipboardMonitorCallback(onStopClipboardMonitor);
    }

    if (g_SetDragFileListNotifyCallback) {
        g_SetDragFileListNotifyCallback(onDragFileListNotify);
    }

    if (g_SetMultiFilesDropNotifyCallback) {
        g_SetMultiFilesDropNotifyCallback(onMultiFilesDropNotify);
    }

    if (g_SetUpdateMultipleProgressBarCallback) {
        g_SetUpdateMultipleProgressBarCallback(onUpdateMultipleProgressBar);
    }

    if (g_SetDataTransferCallback) {
        g_SetDataTransferCallback(onDataTransfer);
    }

    if (g_SetUpdateClientStatusCallback) {
        g_SetUpdateClientStatusCallback(onUpdateClientStatus);
    }

    if (g_SetUpdateSystemInfoCallback) {
        g_SetUpdateSystemInfoCallback(onUpdateSystemInfo);
    }

    if (g_SetNotiMessageCallback) {
        g_SetNotiMessageCallback(onNotiMessage);
    }

    if (g_SetCleanClipboardCallback) {
        g_SetCleanClipboardCallback(onCleanClipboard);
    }

    if (g_SetGetDeviceNameCallback) {
        g_SetGetDeviceNameCallback(onGetDeviceName);
    }

    if (g_SetAuthViaIndexCallback) {
        g_SetAuthViaIndexCallback(onAuthViaIndex);
    }

    if (g_SetDIASStatusCallback) {
        g_SetDIASStatusCallback(onDIASStatus);
    }

    if (g_SetRequestSourceAndPortCallback) {
        g_SetRequestSourceAndPortCallback(onRequestSourceAndPort);
    }
    if (g_SetGetDownloadPathCallback) {
        g_SetGetDownloadPathCallback(onGetDownloadPath);
    }

    if (g_SetSetupDstPasteImageCallback) {
        g_SetSetupDstPasteImageCallback(onSetupDstPasteImage);
    }

    qInfo() << "DLL functions initialized successfully......";
    return true;
}

//---------------------------------------------------------------------------------------
void LoadPlugin::onStartClipboardMonitor()
{
    if (g_clipboardMonitoringStatus.load() == true) {
        return;
    }
    qInfo() << "[Callback] Clipboard monitor started";
    connect(QApplication::clipboard(), &QClipboard::changed, LoadPlugin::getInstance(), &LoadPlugin::onClipboardChanged);
    g_clipboardMonitoringStatus.store(true);
}

void LoadPlugin::onStopClipboardMonitor()
{
    if (g_clipboardMonitoringStatus.load() == false) {
        return;
    }
    qInfo() << "[Callback] Clipboard monitor stopped";
    disconnect(QApplication::clipboard(), &QClipboard::changed, LoadPlugin::getInstance(), &LoadPlugin::onClipboardChanged);
    g_clipboardMonitoringStatus.store(false);
}

void LoadPlugin::onDragFileListNotify(const char* ipPortString, const char* clientID,
                          uint32_t cFileCount, uint64_t totalSize,
                          uint64_t timestamp, const wchar_t* firstFileName,
                          uint64_t firstFileSize)
{
    qInfo() << "[Callback] onDragFileListNotify or onMultiFilesDropNotify:"
             << "IP:Port:" << ipPortString
             << "Client ID:" << clientID
             << "File count:" << cFileCount
             << "Total size:" << totalSize
             << "First file:" << QString::fromStdWString(firstFileName)
             << "First file size:" << firstFileSize;

    DragFilesMsg message;
    message.functionCode = DragFilesMsg::FuncCode::ReceiveFileInfo;
    message.timeStamp = timestamp;
    const auto &ipPort = getIpPort(ipPortString);
    message.ip = ipPort.first;
    message.port = ipPort.second;
    message.clientID = clientID;
    message.fileCount = cFileCount;
    message.totalFileSize = totalSize;
    message.firstTransferFileName = QString::fromStdWString(firstFileName);
    message.firstTransferFileSize = firstFileSize;

    auto data = DragFilesMsg::toByteArray(message);
    Q_EMIT CommonSignals::getInstance()->broadcastData(data);
}

void LoadPlugin::onMultiFilesDropNotify(const char* ipPortString, const char* clientID,
                            uint32_t cFileCount, uint64_t totalSize,
                            uint64_t timestamp, const wchar_t* firstFileName,
                            uint64_t firstFileSize)
{
    onDragFileListNotify(ipPortString, clientID, cFileCount, totalSize, timestamp, firstFileName, firstFileSize);
}

void LoadPlugin::onUpdateMultipleProgressBar(const char* ipPortString, const char* clientID,
                                 const wchar_t* currentFileName,
                                 uint32_t sentFilesCnt, uint32_t totalFilesCnt,
                                 uint64_t currentFileSize, uint64_t totalSize,
                                 uint64_t sentSize, uint64_t timestamp)
{
    UpdateProgressMsg message;
    message.functionCode = UpdateProgressMsg::FuncCode::MultiFile;
    const auto &ipPort = getIpPort(ipPortString);
    message.ip = ipPort.first;
    message.port = ipPort.second;
    message.clientID = clientID;
    message.timeStamp = timestamp;

    message.currentFileName = QString::fromStdWString(currentFileName);
    message.sentFilesCount = sentFilesCnt;
    message.totalFilesCount = totalFilesCnt;
    message.currentFileSize = currentFileSize;
    message.totalFilesSize = totalSize;
    message.totalSentSize = sentSize;

    auto data = UpdateProgressMsg::toByteArray(message);
    Q_EMIT CommonSignals::getInstance()->broadcastData(data);
}

void LoadPlugin::onUpdateClientStatus(uint32_t status, const char* ipPortString,
                          const char* id, const wchar_t* name,
                          const char* deviceType) {
    UpdateClientStatusMsg message;
    message.status = static_cast<uint8_t>(status);
    const auto &ipPort = getIpPort(ipPortString);
    message.ip = ipPort.first;
    message.port = ipPort.second;
    message.clientID = id;
    message.clientName = QString::fromStdWString(name);
    message.deviceType = deviceType;

    UpdateClientStatusMsgPtr ptr_msg = std::make_shared<UpdateClientStatusMsg>(message);

    LoadPlugin::getInstance()->runInLoop([ptr_msg] {
        auto &clientVec = g_getGlobalData()->m_clientVec;
        bool exists = false;
        for (auto itr = clientVec.begin(); itr != clientVec.end(); ++itr) {
            if ((*itr)->clientID == ptr_msg->clientID) {
                exists = true;
                if (ptr_msg->status == 0) {
                    clientVec.erase(itr);
                } else {
                    *itr = ptr_msg;
                }
                break;
            }
        }
        if (exists == false && ptr_msg->status == 1) {
            clientVec.push_back(ptr_msg);
        }

        Q_EMIT CommonSignals::getInstance()->broadcastData(UpdateClientStatusMsg::toByteArray(*ptr_msg));

        for (const auto &clientStatus : g_getGlobalData()->m_clientVec) {
            Q_EMIT CommonSignals::getInstance()->broadcastData(UpdateClientStatusMsg::toByteArray(*clientStatus));
        }
    });
}

void LoadPlugin::onUpdateSystemInfo(const char* ipPortString, const wchar_t* serviceVer)
{
    qInfo() << "[Callback] System info update:"
             << "IP:Port:" << ipPortString
             << "Service version:" << QString::fromWCharArray(serviceVer);

    UpdateSystemInfoMsg message;
    const auto &ipPort = getIpPort(ipPortString);
    message.ip = ipPort.first;
    message.port = ipPort.second;
    message.serverVersion = QString::fromStdWString(serviceVer);

    LoadPlugin::getInstance()->runInLoop([message] {
        g_cacheUpdateSystemInfoMsg = std::make_shared<UpdateSystemInfoMsg>(message);

        auto data = UpdateSystemInfoMsg::toByteArray(*g_cacheUpdateSystemInfoMsg);
        Q_EMIT CommonSignals::getInstance()->broadcastData(data);
    });
}

void LoadPlugin::onNotiMessage(uint64_t timestamp, uint32_t notiCode,
                   const wchar_t* notiParam[], int paramCount)
{
    NotifyMessage message;
    message.timeStamp = timestamp;
    message.notiCode = static_cast<uint8_t>(notiCode);
    for (int index = 0; index < paramCount; ++index) {
        message.paramInfoVec.push_back({ QString::fromStdWString(notiParam[index]) });
    }

    auto ptrMessage = std::make_shared<NotifyMessage>(std::move(message));
    Q_EMIT CommonSignals::getInstance()->dispatchMessage(QVariant::fromValue<NotifyMessagePtr>(ptrMessage));
}

void LoadPlugin::onCleanClipboard()
{
    LoadPlugin::getInstance()->runInLoop([] {
        QApplication::clipboard()->clear();
    });
}

const wchar_t* LoadPlugin::onGetDeviceName()
{
    static std::wstring s_deviceName = QSysInfo::machineHostName().toStdWString();
    qInfo() << "----------------------------onGetDeviceName:" << QString::fromStdWString(s_deviceName);
    return s_deviceName.c_str();
}

void LoadPlugin::onAuthViaIndex(uint32_t indexValue)
{
    LoadPlugin::getInstance()->runInLoop([indexValue] {
        qInfo() << "---------------onAuthViaIndex:" << indexValue;
        auto cacheData = MonitorPlugEvent::getInstance()->getCacheMonitorData();
        uint8_t authResult = 0;
        if (MonitorPlugEvent::querySmartMonitorAuthStatus(cacheData.hPhysicalMonitor, indexValue, authResult) == false) {
            qWarning() << "querySmartMonitorAuthStatus failed ......";
            return;
        }
        g_SetAuthStatusCode(static_cast<unsigned char>(authResult));
    });
}

void LoadPlugin::onDIASStatus(uint32_t statusCode)
{
    LoadPlugin::getInstance()->runInLoop([statusCode] {
        qInfo() << "---------------onDIASStatus:" << statusCode;
        StatusInfoNotifyMsg message;
        message.statusCode = statusCode;

        g_cacheStatusInfoNotifyMsg = std::make_shared<StatusInfoNotifyMsg>(message);

        auto data = StatusInfoNotifyMsg::toByteArray(*g_cacheStatusInfoNotifyMsg);
        Q_EMIT CommonSignals::getInstance()->broadcastData(data);
    });
}

void LoadPlugin::onRequestSourceAndPort()
{
    LoadPlugin::getInstance()->runInLoop([] {
        qInfo() << "---------------onRequestSourceAndPort";
        auto cacheData = MonitorPlugEvent::getInstance()->getCacheMonitorData();
        uint8_t source = 0;
        uint8_t port = 0;
        if (MonitorPlugEvent::getConnectedPortInfo(cacheData.hPhysicalMonitor, source, port) == false) {
            qWarning() << "getConnectedPortInfo failed ......";
            return;
        }
        g_SetDIASSourceAndPort(static_cast<unsigned char>(source), static_cast<unsigned char>(port));
    });
}

const wchar_t* LoadPlugin::onGetDownloadPath()
{
    static std::wstring s_downloadPath = QDir::toNativeSeparators(CommonUtils::downloadDirectoryPath()).toStdWString();
    qInfo() << "----------------------------onGetDownloadPath:" << QString::fromStdWString(s_downloadPath);
    return s_downloadPath.c_str();
}

void LoadPlugin::onDataTransfer(const unsigned char* data, uint32_t size)
{
    std::string tmpData(reinterpret_cast<const char*>(data), size);
    LoadPlugin::getInstance()->runInLoop([tmpData] {
        if (g_cacheImageInfo == nullptr) {
            return;
        }
        g_cacheImageInfo->imageData += tmpData;
        if (g_cacheImageInfo->imageData.size() >= g_cacheImageInfo->dataSize) {
            std::wstring descInfo = g_cacheImageInfo->desc.toStdWString();
            setupDstPasteImage_helper(descInfo.c_str(),
                                      g_cacheImageInfo->imgHeader,
                                      g_cacheImageInfo->dataSize,
                                      reinterpret_cast<const BYTE*>(g_cacheImageInfo->imageData.data()));
            g_cacheImageInfo.reset();
        }
    });
}

void LoadPlugin::onSetupDstPasteImage(const wchar_t* desc, IMAGE_HEADER imgHeader, uint32_t dataSize)
{
    qInfo() << "[Callback] Setup destination paste image:"
             << "Description:" << QString::fromWCharArray(desc)
             << "Image header:"
             << imgHeader.width << "x" << imgHeader.height
             << "planes:" << imgHeader.planes
             << "bitCount:" << imgHeader.bitCount
             << "compression:" << imgHeader.compression
             << "Data size:" << dataSize;
    QString descInfo = QString::fromStdWString(desc);
    LoadPlugin::getInstance()->runInLoop([descInfo, imgHeader, dataSize] {
        g_cacheImageInfo.reset(new RecvImageInfo(descInfo, imgHeader, dataSize));
    });
}

std::pair<QString, uint16_t> LoadPlugin::getIpPort(const QString &ipPortString)
{
    auto pos = ipPortString.indexOf(':');
    Q_ASSERT(pos != -1);
    QString ip = ipPortString.left(pos);
    uint16_t port = ipPortString.mid(pos + 1).toUShort();
    return { ip, port };
}

void LoadPlugin::onPipeConnected()
{
    QTimer::singleShot(200, this, [] {
        if (g_cacheUpdateSystemInfoMsg) {
            auto data = UpdateSystemInfoMsg::toByteArray(*g_cacheUpdateSystemInfoMsg);
            Q_EMIT CommonSignals::getInstance()->broadcastData(data);
        }

        if (g_cacheStatusInfoNotifyMsg) {
            auto data = StatusInfoNotifyMsg::toByteArray(*g_cacheStatusInfoNotifyMsg);
            Q_EMIT CommonSignals::getInstance()->broadcastData(data);
        }

        {
            for (const auto &clientStatus : g_getGlobalData()->m_clientVec) {
                Q_EMIT CommonSignals::getInstance()->broadcastData(UpdateClientStatusMsg::toByteArray(*clientStatus));
            }
        }
    });
}

void LoadPlugin::onClipboardChanged(QClipboard::Mode mode)
{
    Q_ASSERT(mode == QClipboard::Mode::Clipboard);
    uint64_t currentIndex = ++m_copyIndex;
    QTimer::singleShot(100, qApp, [currentIndex, this] {
        if (currentIndex != m_copyIndex) {
            return;
        }
        if (::OpenClipboard(nullptr)) {
            if (::IsClipboardFormatAvailable(CF_BITMAP)) {
                HBITMAP hBitmap = static_cast<HBITMAP>(GetClipboardData(CF_BITMAP));
                if (hBitmap) {
                    getBitmapData(hBitmap);
                } else {
                    qWarning() << "GetClipboardData failed:" << ::GetLastError();
                }
            }
            ::CloseClipboard();
        } else {
            qWarning() << "OpenClipboard failed:" << ::GetLastError();
        }
    });
}

void LoadPlugin::onDIASStatusChanged(bool status)
{
    if (status) {
        auto cacheData = MonitorPlugEvent::getInstance()->getCacheMonitorData();
        if (cacheData.macAddress.empty() == false) {
            qInfo() << "g_SetMacAddress:" << QByteArray::fromStdString(cacheData.macAddress).toHex().toUpper().constData();
            g_SetMacAddress(cacheData.macAddress.data(), static_cast<int>(cacheData.macAddress.size()));
        }
    } else {
        auto cacheData = MonitorPlugEvent::getInstance()->getCacheMonitorData();
        if (cacheData.isDIAS) {
            g_SetExtractDIAS();
            MonitorPlugEvent::getInstance()->clearData();
        }
    }
}
