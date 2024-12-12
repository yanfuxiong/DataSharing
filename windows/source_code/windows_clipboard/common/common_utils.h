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
#include <qmath.h> // windows环境下的 M_PI定义
#include "global_def.h"


class CommonUtils
{
public:
    static QByteArray getFileContent(const QString &filePath);
    static void commonMessageOutput(QtMsgType type, const QMessageLogContext &context, const QString &msg);
    static QString desktopDirectoryPath();
    static QString homeDirectoryPath();
    static void runInThreadPool(const std::function<void()> &callback);

    // 查找文件夹中的所有文件
    static void findAllFiles(const QString &directoryPath, std::vector<QString> &filesVec,
                             const std::function<bool(const QFileInfo&)> &filterFunc = isFileHelper);
    static std::vector<QString> findAllFiles(const QString &directoryPath, const std::function<bool(const QFileInfo&)> &filterFunc = isFileHelper);

    // utf8 => utf16 LE
    static QByteArray toUtf16LE(const QString &data);
    // utf16 LE 转 utf8
    static QByteArray toUtf8(const QByteArray &data);
    // 通过文件路径获取文件名 (只做解析, 不判断文件是否存在)
    static QString getFileNameByPath(const QString &filePath);

    static bool processIsRunning(const QString &exePath);

    static void killServer();

    static QString byteCountDisplay(int64_t bytesCount);

    // 控制软件开机自动
    static void setAutoRun(bool status = true);
    static void setAutoRun(const QString &appFilePath, bool status = true);

private:
    inline static bool isFileHelper(const QFileInfo &fileInfo) { return fileInfo.isFile(); }
};
