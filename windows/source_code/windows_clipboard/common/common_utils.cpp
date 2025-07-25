#include "common_utils.h"
#include <QStandardPaths>
#include <QDir>
#include <QProcess>
#include <QSettings>
#include <QUuid>
#include <QHostInfo>
#ifdef Q_OS_WINDOWS
#include <windows.h>
#include <psapi.h>
#else
#include <signal.h>
#include <unistd.h>
#include <sys/syscall.h>
#endif
#include <iostream>
#include "common_signals.h"

std::unique_ptr<QFile> g_logFile;

QByteArray CommonUtils::getFileContent(const QString &filePath)
{
    QFile file(filePath);
    if (file.open(QFile::ReadOnly)) {
        return file.readAll();
    }
    return {};
}


void CommonUtils::commonMessageOutput(QtMsgType type, const QMessageLogContext &context, const QString &msg)
{
#ifdef Q_OS_WINDOWS
    thread_local static uint32_t s_thread_id = ::GetCurrentThreadId();
#else
    thread_local static uint32_t s_thread_id = static_cast<uint32_t>(::syscall(SYS_gettid));
#endif

    QString msg_str;
    QTextStream str_stream(&msg_str);
    str_stream << QDateTime::currentDateTime().toString("yyyy-MM-dd hh:mm:ss.zzz");
    str_stream << " " << s_thread_id << " ";

    switch (type) {
    case QtDebugMsg:
        str_stream << "DEBUG";
        break;
    case QtInfoMsg:
        str_stream << "INFO";
        break;
    case QtWarningMsg:
        str_stream << "WARN";
        break;
    case QtCriticalMsg:
        str_stream << "CRITICAL";
        break;
    case QtFatalMsg:
        str_stream << "FATAL";
        break;
    }

    str_stream << " "
               << msg
               << " - "
               << QFileInfo(context.file).fileName()
               << ":"
               << context.line
#if (QT_VERSION >= QT_VERSION_CHECK(5, 15, 0))
               << Qt::endl;
#else
               << endl;
#endif

    str_stream.flush();

    {
        static std::mutex s_mutex;
        std::lock_guard<std::mutex> locker(s_mutex);
        std::cerr << msg_str.toLocal8Bit().constData();
        if (g_logFile) {
            g_logFile->write(msg_str.toLocal8Bit().constData());
            g_logFile->flush();
        }
    }
}

QString CommonUtils::desktopDirectoryPath()
{
    return QStandardPaths::writableLocation(QStandardPaths::StandardLocation::DesktopLocation);
}

QString CommonUtils::downloadDirectoryPath()
{
    return QStandardPaths::writableLocation(QStandardPaths::StandardLocation::DownloadLocation);
}

QString CommonUtils::homeDirectoryPath()
{
    return QStandardPaths::writableLocation(QStandardPaths::StandardLocation::HomeLocation);
}

QString CommonUtils::localDataDirectory()
{
    auto path = QStandardPaths::writableLocation(QStandardPaths::StandardLocation::AppLocalDataLocation);
    if (QFile::exists(path) == false) {
        QDir().mkpath(path);
    }
    Q_ASSERT(QFile::exists(path));
    return path;
}

void CommonUtils::runInThreadPool(const std::function<void()> &callback)
{
    class RunnableEx : public QRunnable
    {
    public:
        RunnableEx(const std::function<void()> &callback) : m_callback(callback){}
        void run() override
        {
            if (m_callback) {
                m_callback();
            }
        }
    private:
        std::function<void()> m_callback;
    };

    QThreadPool::globalInstance()->start(new RunnableEx(callback));
}

QString CommonUtils::createUuid()
{
    return QUuid::createUuid().toString();
}

QString CommonUtils::localIpAddress()
{
    QHostInfo info = QHostInfo::fromName(QHostInfo::localHostName());
    for (const auto &address : info.addresses()) {
        if (address.protocol() == QAbstractSocket::IPv4Protocol) {
            return address.toString();
        }
    }
    return "127.0.0.1";
}

QByteArray CommonUtils::toUtf16LE(const QString &data)
{
    QByteArray newData(reinterpret_cast<const char*>(data.utf16()), data.length() * 2);
    return newData;
}

QByteArray CommonUtils::toUtf8(const QByteArray &data)
{
    Q_ASSERT(sizeof (QChar) == 2);
    Q_ASSERT(data.size() % 2 == 0);
    QString newData(reinterpret_cast<const QChar*>(data.data()), data.length() / 2);
    return newData.toUtf8();
}

QString CommonUtils::getFileNameByPath(const QString &filePath)
{
    std::string newFilePath = filePath.toStdString();
    do {
        auto pos = newFilePath.find_last_of('/');
        if (pos == std::string::npos) {
            break;
        }
        return newFilePath.substr(pos + 1).c_str();
    } while (false);

    do {
        auto pos = newFilePath.find_last_of('\\');
        if (pos == std::string::npos) {
            break;
        }
        return newFilePath.substr(pos + 1).c_str();
    } while (false);

    return filePath;
}

bool CommonUtils::processIsRunning(const QString &exePath)
{
    DWORD aProcesses[1024];
    DWORD cbNeeded;
    DWORD cProcesses;
    if (!::EnumProcesses(aProcesses, sizeof(aProcesses), &cbNeeded)) {
        return false;
    }
    cProcesses = cbNeeded / sizeof(DWORD);
    for (unsigned int index = 0; index < cProcesses; ++index) {
        if (aProcesses[index] != 0) {
            auto processID = aProcesses[index];
            TCHAR szProcessName[MAX_PATH] { 0 };
            HANDLE hProcess = ::OpenProcess(PROCESS_QUERY_INFORMATION | PROCESS_VM_READ, FALSE, processID);
            if (NULL != hProcess) {
                HMODULE hMod;
                DWORD cbNeeded;
                if (::EnumProcessModules(hProcess, &hMod, sizeof(hMod), &cbNeeded)) {
                    ::GetModuleBaseName(hProcess, hMod, szProcessName, sizeof(szProcessName)/sizeof(TCHAR));
                    QString tmpName = QString::fromStdString(reinterpret_cast<const char*>(szProcessName));
                    //qInfo() << tmpName.toUtf8().constData();
                    if (exePath.endsWith(tmpName)) {
                        return true;
                    }
                }
            }
            CloseHandle(hProcess);
        }
    }
    return false;
}

int CommonUtils::processRunningCount(const QString &exePath)
{
    int runningCount = 0;
    DWORD aProcesses[1024];
    DWORD cbNeeded;
    DWORD cProcesses;
    if (!::EnumProcesses(aProcesses, sizeof(aProcesses), &cbNeeded)) {
        return runningCount;
    }
    cProcesses = cbNeeded / sizeof(DWORD);
    for (unsigned int index = 0; index < cProcesses; ++index) {
        if (aProcesses[index] != 0) {
            auto processID = aProcesses[index];
            TCHAR szProcessName[MAX_PATH] { 0 };
            HANDLE hProcess = ::OpenProcess(PROCESS_QUERY_INFORMATION | PROCESS_VM_READ, FALSE, processID);
            if (NULL != hProcess) {
                HMODULE hMod;
                DWORD cbNeeded;
                if (::EnumProcessModules(hProcess, &hMod, sizeof(hMod), &cbNeeded)) {
                    ::GetModuleBaseName(hProcess, hMod, szProcessName, sizeof(szProcessName)/sizeof(TCHAR));
                    QString tmpName = QString::fromStdString(reinterpret_cast<const char*>(szProcessName));
                    //qInfo() << tmpName.toUtf8().constData();
                    if (exePath.endsWith(tmpName)) {
                        ++runningCount;
                    }
                }
            }
            CloseHandle(hProcess);
        }
    }
    return runningCount;
}

void CommonUtils::killServer()
{
    {
        QProcess process;
        process.setProcessChannelMode(QProcess::ProcessChannelMode::MergedChannels);
        process.start(QString("taskkill /F /IM %1 /T").arg(PIPE_SERVER_EXE_NAME));
        process.waitForFinished();
        qInfo() << process.readAll().constData();
    }

    {
        QProcess process;
        process.setProcessChannelMode(QProcess::ProcessChannelMode::MergedChannels);
        // 测试服务器进程
        process.start(QString("taskkill /F /IM %1 /T").arg("test-server.exe"));
        process.waitForFinished();
        qInfo() << process.readAll().constData();
    }
}

QString CommonUtils::byteCountDisplay(int64_t bytesCount)
{
    if (bytesCount < 1024) {
        return QString::asprintf("%lld B", bytesCount);
    } else if (bytesCount < 1024 * 1024) {
        return QString::asprintf("%.2lf KB", static_cast<double>(bytesCount) / std::pow(1024.0, 1.0));
    } else {
        return QString::asprintf("%.2lf MB", static_cast<double>(bytesCount) / std::pow(1024.0, 2.0));
    }
}

void CommonUtils::setAutoRun(bool status)
{
    setAutoRun(qApp->applicationFilePath(), status);
}

void CommonUtils::removeAutoRunByName(const QString &keyName)
{
    const char *auto_run = R"(HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run)";
    QScopedPointer<QSettings> settings(new QSettings(auto_run, QSettings::NativeFormat));
    if (settings->contains(keyName)) {
        settings->remove(keyName);
    }
}

QString CommonUtils::getValueByRegKeyName(const QString &keyName)
{
    const char *auto_run = R"(HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run)";
    QScopedPointer<QSettings> settings(new QSettings(auto_run, QSettings::NativeFormat));
    const QString path = settings->value(keyName).toString();
    return path;
}

void CommonUtils::setAutoRun(const QString &appFilePath, bool status)
{
    const char *auto_run = R"(HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run)";
    QScopedPointer<QSettings> settings(new QSettings(auto_run, QSettings::NativeFormat));
    const QString appName = QFileInfo(appFilePath).fileName();
    const QString path = settings->value(appName).toString();

    qInfo() << "appName:" << appName.toUtf8().constData() << "; path:" << path.toUtf8().constData();

    if (status) {
        QString newPath = QDir::toNativeSeparators(appFilePath);
        if (path != newPath) {
            settings->setValue(appName, newPath);
        }
    } else {
        settings->remove(appName);
    }
}
