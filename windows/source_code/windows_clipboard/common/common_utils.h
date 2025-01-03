#pragma once
#include <QString>
#include <QFile>
#include <QDebug>
#include <QDateTime>
#include <QTimer>
#include <QFileInfo>
#include <QPointer>
#include <QThread>
#include <QThreadPool>
#include <QDateTime>
#include <QJsonDocument>
#include <QJsonArray>
#include <QJsonObject>
#include <QJsonParseError>
#include <QPointF>
#include <QAtomicInteger>
//#include <QColor>
#include <atomic>
#include <memory>
#include <functional>
#include <cmath>
#include <qmath.h>
#include "global_def.h"


class CommonUtils
{
public:
    static QByteArray getFileContent(const QString &filePath);
    static void commonMessageOutput(QtMsgType type, const QMessageLogContext &context, const QString &msg);
    static QString desktopDirectoryPath();
    static QString homeDirectoryPath();
    static QString localDataDirectory();
    static void runInThreadPool(const std::function<void()> &callback);
    static QString createUuid();
    static QString localIpAddress();

    static void findAllFiles(const QString &directoryPath, std::vector<QString> &filesVec,
                             const std::function<bool(const QFileInfo&)> &filterFunc = isFileHelper);
    static std::vector<QString> findAllFiles(const QString &directoryPath, const std::function<bool(const QFileInfo&)> &filterFunc = isFileHelper);

    // utf8 => utf16 LE
    static QByteArray toUtf16LE(const QString &data);
    // utf16 LE => utf8
    static QByteArray toUtf8(const QByteArray &data);
    static QString getFileNameByPath(const QString &filePath);

    static bool processIsRunning(const QString &exePath);
    static int processRunningCount(const QString &exePath);

    static void killServer();

    static QString byteCountDisplay(int64_t bytesCount);

    static void setAutoRun(bool status = true);
    static void setAutoRun(const QString &appFilePath, bool status = true);

private:
    inline static bool isFileHelper(const QFileInfo &fileInfo) { return fileInfo.isFile(); }
};
