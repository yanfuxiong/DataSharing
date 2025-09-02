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
#include <QRandomGenerator>
//#include <QColor>
#include <atomic>
#include <memory>
#include <functional>
#include <cmath>
#include <qmath.h>
#include "global_def.h"

extern std::unique_ptr<QFile> g_logFile;

class CommonUtils
{
public:
    static QByteArray getFileContent(const QString &filePath);
    static void commonMessageOutput(QtMsgType type, const QMessageLogContext &context, const QString &msg);
    static QString desktopDirectoryPath();
    static QString downloadDirectoryPath();
    static QString homeDirectoryPath();
    static QString localDataDirectory();
    static QString crossShareRootPath();
    static QString windowsLogFolderPath();
    static QString configFilePath();
    static void ensureDirectoryExists();
    static void runInThreadPool(const std::function<void()> &callback);
    static QString createUuid();
    static QString localIpAddress();

    // utf8 => utf16 LE
    static QByteArray toUtf16LE(const QString &data);
    // utf16 LE => utf8
    static QByteArray toUtf8(const QByteArray &data);
    static QString getFileNameByPath(const QString &filePath);

    static bool processIsRunning(const QString &exePath);
    static int processRunningCount(const QString &exePath);

    static void killServer();
    static void killWindowsClipboard();

    static QString byteCountDisplay(int64_t bytesCount);

    static void setAutoRun(bool status = true);
    static void setAutoRun(const QString &appFilePath, bool status = true);
    static void removeAutoRunByName(const QString &keyName);
    static QString getValueByRegKeyName(const QString &keyName);
    static bool compressFolderToTarGz(const QString &sourceDir, const QString &outputFile);
    static bool compressFileToTarGz(const QString &sourceFilePath, const QString &outputFilePath);
    static void supportForHighDPIDisplays();
    static std::pair<int, int> getMaxMonitorResolution();
    static void startDetachedWithoutInheritance(const QString &program, const QStringList &arguments);
};
