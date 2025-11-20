#pragma once
#include "common_signals.h"
#include "common_utils.h"
#include "worker_thread.h"
#include "clipboard.h"
#include <QLibrary>
#include <QClipboard>
#include <QMimeData>
#include <QImage>
#include <windows.h>

class WindowsMimeData : public QMimeData
{
    Q_OBJECT
public:
    WindowsMimeData();
    ~WindowsMimeData();
    void setImage(const QImage &image);
    QStringList formats() const override;
    bool hasFormat(const QString &mimeType) const override;

protected:
    QVariant retrieveData(const QString &mimeType, QVariant::Type type) const override;

    bool m_hasBitmap;
};

class LoadPlugin : public QObject
{
    Q_OBJECT
public:
    ~LoadPlugin();
    static LoadPlugin *getInstance();
    void initPlugin();
    bool initDllFunctions();
    void runInLoop(const std::function<void()> &callback);

    void cancelFileTransfer(const char *ipPort, const char *clientID, uint64_t timeStamp);
    void dragFileListRequest(const wchar_t *filePathArry[], uint32_t arryLength , uint64_t timeStamp);
    void multiFilesDropRequest(const char *ipPort, const char *clientID, uint64_t timeStamp, const wchar_t *filePathArry[], uint32_t arryLength);
    void updateDownloadPath(const QString &downloadPath);

private:
    static void onStartClipboardMonitor();
    static void onStopClipboardMonitor();
    static void onDragFileListNotify(const char* ipPort, const char* clientID,
                                 uint32_t cFileCount, uint64_t totalSize,
                                 uint64_t timestamp, const wchar_t* firstFileName,
                                 uint64_t firstFileSize);
    static void onMultiFilesDropNotify(const char* ipPort, const char* clientID,
                                   uint32_t cFileCount, uint64_t totalSize,
                                   uint64_t timestamp, const wchar_t* firstFileName,
                                   uint64_t firstFileSize);
    static void onUpdateMultipleProgressBar(const char* ipPort, const char* clientID,
                                        const wchar_t* currentFileName,
                                        uint32_t sentFilesCnt, uint32_t totalFilesCnt,
                                        uint64_t currentFileSize, uint64_t totalSize,
                                        uint64_t sentSize, uint64_t timestamp);
    static void onUpdateClientStatus(uint32_t status, const char* ipPort,
                                    const char* id, const wchar_t* name,
                                    const char* deviceType);
    static void onUpdateClientStatusEx(const char *clientJson);
    static void onUpdateSystemInfo(const char* ipPort, const wchar_t* serviceVer);
    static void onNotiMessage(uint64_t timestamp, uint32_t notiCode,
                              const wchar_t* notiParam[], int paramCount);
    static void onCleanClipboard();
    static void onAuthViaIndex(uint32_t index);
    static void onDIASStatus(uint32_t statusCode);
    static void onRequestSourceAndPort();
    static void onDataTransfer(const unsigned char* data, uint32_t size);
    static void onSetupDstPasteImage(const wchar_t* desc, IMAGE_HEADER imgHeader, uint32_t dataSize);
    static void onRequestUpdateClientVersion(const char *clientVersion);
    static void onNotifyErrEvent(const char *clientID, uint32_t errorCode, const char *ipPortString, const char *timeStamp, const char *arg3, const char *arg4);
    static void onSetupDstPasteXClipData(const char *textData, const char *imageData, const char *htmlData);

private:
    static std::pair<QString, uint16_t> getIpPort(const QString &ipPortString);

Q_SIGNALS:
    void showWarningIcon();

public Q_SLOTS:
    void onPipeConnected();
    void onClipboardChanged(QClipboard::Mode mode);
    void onDIASStatusChanged(bool status);

private:
    LoadPlugin();

    WorkerThread m_mainThread;
    uint64_t m_copyIndex;
    static LoadPlugin *m_instance;
};
