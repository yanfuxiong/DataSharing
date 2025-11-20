#include "common_utils.h"
#include <QStandardPaths>
#include <QDir>
#include <QProcess>
#include <QSettings>
#include <QUuid>
#include <QHostInfo>
#include <QStack>
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
#include <archive.h>
#include <archive_entry.h>

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
        str_stream << "CRIT";
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

QString CommonUtils::crossShareRootPath()
{
    QString newPath = QDir::cleanPath(localDataDirectory() + "/../CrossShare");
    return newPath;
}

QString CommonUtils::windowsLogFolderPath()
{
    return crossShareRootPath() + "/Log";
}

QString CommonUtils::configFilePath()
{
    return crossShareRootPath() + "/" + LOCAL_CONFIG_NAME;
}

void CommonUtils::ensureDirectoryExists()
{
    QDir().mkpath(localDataDirectory());
    QDir().mkpath(crossShareRootPath());
    QDir().mkpath(windowsLogFolderPath());
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

bool CommonUtils::crossShareServerIsRunning()
{
    QProcess process;
    process.setProcessChannelMode(QProcess::ProcessChannelMode::MergedChannels);
    process.start(QString("tasklist"));
    process.waitForFinished();
    QString data = QString::fromLocal8Bit(process.readAll());
    //qDebug() << data.toUtf8().constData();
    return data.contains(CROSS_SHARE_SERV_NAME, Qt::CaseInsensitive);
}

bool CommonUtils::crosstoolIsRunning()
{
    QProcess process;
    process.setProcessChannelMode(QProcess::ProcessChannelMode::MergedChannels);
    process.start(QString("tasklist"));
    process.waitForFinished();
    QString data = QString::fromLocal8Bit(process.readAll());
    //qDebug() << data.toUtf8().constData();
    return data.contains(CROSS_TOOL_NAME, Qt::CaseInsensitive);
}

void CommonUtils::killServer()
{
    QProcess process;
    process.setProcessChannelMode(QProcess::ProcessChannelMode::MergedChannels);
    process.start(QString("taskkill /F /IM %1").arg(CROSS_SHARE_SERV_NAME));
    process.waitForFinished();
    qDebug() << QString::fromLocal8Bit(process.readAll()).toUtf8().constData();
}

void CommonUtils::killWindowsClipboard()
{
    QProcess process;
    process.setProcessChannelMode(QProcess::ProcessChannelMode::MergedChannels);
    process.start(QString("taskkill /F /IM %1").arg(WINDOWS_CLIPBOARD_NAME));
    process.waitForFinished();
    qDebug() << QString::fromLocal8Bit(process.readAll()).toUtf8().constData();
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

void CommonUtils::supportForHighDPIDisplays()
{
    const auto &screenData = getMaxMonitorResolution();
    const int width = screenData.first;
    const int height = screenData.second;
    std::cerr << "-----------screen info; width=" << width << "; height=" << height << std::endl;
    {
        using SetProcessDpiAwarenessContextFn = BOOL(WINAPI*)(DPI_AWARENESS_CONTEXT);
        HMODULE hUser32 = ::LoadLibraryW(L"user32.dll");
        if (hUser32) {
            auto setContext = reinterpret_cast<SetProcessDpiAwarenessContextFn>(::GetProcAddress(hUser32, "SetProcessDpiAwarenessContext"));
            if (setContext) {
                setContext(DPI_AWARENESS_CONTEXT_UNAWARE);
            }
            ::FreeLibrary(hUser32);
        }
    }
    qputenv("QT_AUTO_SCREEN_SCALE_FACTOR", QByteArray::number(0));
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

bool CommonUtils::compressFolderToTarGz(const QString &sourceDir, const QString &outputFile)
{
    // Validate source directory exists
    QDir source(sourceDir);
    if (!source.exists()) {
        qCritical() << "Source directory does not exist:" << sourceDir;
        return false;
    }

    // Create archive object
    struct archive *a = archive_write_new();
    if (!a) {
        qCritical() << "Failed to create archive object";
        return false;
    }

    // Set compression format to tar.gz
    archive_write_add_filter_gzip(a);
    archive_write_set_format_ustar(a);

    // Set compression level (optional)
    if (archive_write_set_options(a, "compression-level=6") != ARCHIVE_OK) {
        qWarning() << "Warning setting compression level:" << archive_error_string(a);
    }

    // Open output file using wide character API for Unicode support
    int rc = archive_write_open_filename_w(a, outputFile.toStdWString().c_str());
    if (rc != ARCHIVE_OK) {
        qCritical() << "Failed to open output file:" << archive_error_string(a);
        archive_write_free(a);
        return false;
    }

    // Collect all files and directories using stack-based traversal
    QStringList filePaths;
    QStringList dirPaths;
    QStack<QString> dirStack;
    dirStack.push(sourceDir);

    while (!dirStack.isEmpty()) {
        QString currentDirPath = dirStack.pop();
        QDir currentDir(currentDirPath);

        // Add this directory to the list
        dirPaths.append(currentDirPath);

        // Get all entries in the current directory
        QFileInfoList entries = currentDir.entryInfoList(
            QDir::Files | QDir::Dirs | QDir::NoDotAndDotDot | QDir::Hidden | QDir::System,
            QDir::DirsFirst
            );

        for (const QFileInfo &entry : entries) {
            if (entry.isDir()) {
                // Push subdirectory to stack for later processing
                dirStack.push(entry.filePath());
            } else {
                // Add file to the list
                filePaths.append(entry.filePath());
            }
        }
    }

    int successCount = 0;
    const int BUFFER_SIZE = 4 * 1024 * 1024; // 4MB buffer
    char* buffer = new char[BUFFER_SIZE];

    // First, add all directories (including empty ones)
    for (const QString &dirPath : dirPaths) {
        QFileInfo dirInfo(dirPath);

        // Calculate relative path
        QString relPath = source.relativeFilePath(dirPath).replace('\\', '/');
        if (!relPath.endsWith('/')) {
            relPath += '/'; // Ensure trailing slash
        }

        //qDebug() << "Adding directory:" << relPath;

        // Create archive entry for directory
        struct archive_entry *entry = archive_entry_new();
        if (!entry) {
            qWarning() << "Failed to create entry for directory:" << relPath;
            continue;
        }

        // Convert path to UTF-8 for libarchive
        QByteArray utf8Path = relPath.toUtf8();
        archive_entry_set_pathname(entry, utf8Path.constData());

        // Directory properties
        archive_entry_set_size(entry, 0);
        archive_entry_set_filetype(entry, AE_IFDIR);
        archive_entry_set_perm(entry, 0755); // drwxr-xr-x
        archive_entry_set_mtime(entry, dirInfo.lastModified().toSecsSinceEpoch(), 0);

        // Write directory header
        rc = archive_write_header(a, entry);
        if (rc < ARCHIVE_OK) {
            qWarning() << "Directory header write failed for:" << relPath << "-" << archive_error_string(a);
            archive_entry_free(entry);
            continue;
        }

        successCount++;
        archive_entry_free(entry);
        //qDebug() << "Added directory:" << relPath;
    }

    // Then, add all files
    for (const QString &filePath : filePaths) {
        QFileInfo fileInfo(filePath);

        // Calculate relative path
        QString relPath = source.relativeFilePath(filePath).replace('\\', '/');
        //qDebug() << "Processing file:" << relPath << "(Size:" << fileInfo.size() << "bytes)";

        // Create archive entry
        struct archive_entry *entry = archive_entry_new();
        if (!entry) {
            qWarning() << "Failed to create entry for file:" << relPath;
            continue;
        }

        // Convert path to UTF-8 for libarchive
        QByteArray utf8Path = relPath.toUtf8();
        archive_entry_set_pathname(entry, utf8Path.constData());

        archive_entry_set_size(entry, fileInfo.size());
        archive_entry_set_filetype(entry, AE_IFREG);
        archive_entry_set_perm(entry, 0644); // -rw-r--r--
        archive_entry_set_mtime(entry, fileInfo.lastModified().toSecsSinceEpoch(), 0);

        // Write file header
        rc = archive_write_header(a, entry);
        if (rc < ARCHIVE_OK) {
            qWarning() << "File header write failed for:" << relPath << "-" << archive_error_string(a);
            archive_entry_free(entry);
            continue;
        }

        // Handle zero-byte files - no need to write content
        if (fileInfo.size() == 0) {
            successCount++;
            archive_entry_free(entry);
            //qDebug() << "Added zero-byte file:" << relPath;
            continue;
        }

        // Open file with Windows API (allow shared access)
        HANDLE hFile = CreateFileW(
            filePath.toStdWString().c_str(),
            GENERIC_READ,
            FILE_SHARE_READ | FILE_SHARE_WRITE,
            NULL,
            OPEN_EXISTING,
            FILE_ATTRIBUTE_NORMAL | FILE_FLAG_SEQUENTIAL_SCAN,
            NULL
        );

        if (hFile == INVALID_HANDLE_VALUE) {
            DWORD error = GetLastError();
            qWarning() << "Failed to open file:" << relPath << "- Error code:" << error;
            archive_entry_free(entry);
            continue;
        }

        // Read and compress file content
        DWORD bytesRead = 0;
        bool fileSuccess = true;
        LARGE_INTEGER fileSize;
        GetFileSizeEx(hFile, &fileSize);
        qint64 remaining = fileSize.QuadPart;

        while (remaining > 0) {
            DWORD bytesToRead = static_cast<DWORD>(qMin(static_cast<qint64>(BUFFER_SIZE), remaining));

            if (!ReadFile(hFile, buffer, bytesToRead, &bytesRead, NULL)) {
                DWORD error = GetLastError();
                qWarning() << "Read failed for:" << relPath << "- Error code:" << error;
                fileSuccess = false;
                break;
            }

            if (bytesRead == 0) {
                qWarning() << "Read 0 bytes for:" << relPath;
                break;
            }

            ssize_t bytesWritten = archive_write_data(a, buffer, bytesRead);
            if (bytesWritten != static_cast<ssize_t>(bytesRead)) {
                qDebug() << "Write failed for:" << relPath << "-" << archive_error_string(a);
                fileSuccess = false;
                break;
            }

            remaining -= bytesRead;
        }

        CloseHandle(hFile);
        archive_entry_free(entry);

        if (fileSuccess) {
            successCount++;
            //qDebug() << "Compressed file:" << relPath << "(" << fileInfo.size() << "bytes)";
        }
    }

    delete[] buffer;

    // Close archive
    rc = archive_write_close(a);
    if (rc != ARCHIVE_OK) {
        qWarning() << "Archive close error:" << archive_error_string(a);
    }

    archive_write_free(a);

    int totalItems = dirPaths.size() + filePaths.size();
    qDebug() << "Compression complete: Success" << successCount << "/" << totalItems << "items";
    return (successCount == totalItems);
}

bool CommonUtils::compressFileToTarGz(const QString &sourceFilePath, const QString &outputFilePath)
{
    QFile sourceFile(sourceFilePath);
    if (!sourceFile.exists()) {
        std::cerr << "Error: Source file not found: " << sourceFilePath.toStdString() << std::endl;
        return false;
    }

    struct archive *a = archive_write_new();
    if (!a) {
        std::cerr << "Error: Failed to create archive object" << std::endl;
        return false;
    }

    // Set tar.gz format
    archive_write_add_filter_gzip(a);
    archive_write_set_format_pax_restricted(a);

    // Open output file
    QByteArray outPath = outputFilePath.toUtf8();
    int ret = archive_write_open_filename(a, outPath.constData());
    if (ret != ARCHIVE_OK) {
        std::cerr << "Error: Failed to open output file (" << archive_error_string(a) << ")" << std::endl;
        archive_write_free(a);
        return false;
    }

    // Prepare archive entry
    struct archive_entry *entry = archive_entry_new();
    QFileInfo fi(sourceFilePath);
    QByteArray filename = fi.fileName().toUtf8();
    archive_entry_set_pathname(entry, filename.constData());
    archive_entry_set_size(entry, fi.size());
    archive_entry_set_filetype(entry, AE_IFREG);
    archive_entry_set_perm(entry, 0644);

    // Write header
    ret = archive_write_header(a, entry);
    if (ret != ARCHIVE_OK) {
        std::cerr << "Error: Failed to write header (" << archive_error_string(a) << ")" << std::endl;
        archive_entry_free(entry);
        archive_write_free(a);
        return false;
    }

    // Write file content
    if (sourceFile.open(QIODevice::ReadOnly)) {
        char buffer[10240];
        qint64 bytesRead;
        while ((bytesRead = sourceFile.read(buffer, sizeof(buffer))) > 0) {
            if (archive_write_data(a, buffer, static_cast<size_t>(bytesRead)) < 0) {
                std::cerr << "Error: Write failure (" << archive_error_string(a) << ")" << std::endl;
                sourceFile.close();
                archive_entry_free(entry);
                archive_write_free(a);
                return false;
            }
        }
        sourceFile.close();
    } else {
        std::cerr << "Error: Failed to open source file (" << sourceFile.errorString().toStdString() << ")" << std::endl;
        archive_entry_free(entry);
        archive_write_free(a);
        return false;
    }

    // Cleanup
    archive_entry_free(entry);
    archive_write_close(a);
    archive_write_free(a);

    return true;
}

std::pair<int, int> CommonUtils::getMaxMonitorResolution()
{
    int maxWidth = 0;
    int maxHeight = 0;

    DISPLAY_DEVICE displayDevice = { 0 };
    displayDevice.cb = sizeof(DISPLAY_DEVICE);

    for (DWORD deviceIndex = 0; EnumDisplayDevices(NULL, deviceIndex, &displayDevice, 0); deviceIndex++) {
        if (!(displayDevice.StateFlags & DISPLAY_DEVICE_ACTIVE) ||
            !(displayDevice.StateFlags & DISPLAY_DEVICE_ATTACHED_TO_DESKTOP)) {
            continue;
        }

        DEVMODE devMode = { 0 };
        devMode.dmSize = sizeof(DEVMODE);

        if (EnumDisplaySettings(displayDevice.DeviceName, ENUM_CURRENT_SETTINGS, &devMode)) {
            const int width = devMode.dmPelsWidth;
            const int height = devMode.dmPelsHeight;

            if ((static_cast<long long>(width) * height) >
                (static_cast<long long>(maxWidth) * maxHeight)) {
                maxWidth = width;
                maxHeight = height;
            }
        }
    }

    if (maxWidth == 0 || maxHeight == 0) {
        maxWidth = GetSystemMetrics(SM_CXSCREEN);
        maxHeight = GetSystemMetrics(SM_CYSCREEN);
    }

    return { maxWidth, maxHeight };
}

void CommonUtils::startDetachedWithoutInheritance(const QString &program, const QStringList &arguments)
{
    STARTUPINFOW si;
    std::memset(&si, 0, sizeof (si));
    si.cb = sizeof (si);

    PROCESS_INFORMATION pi;
    DWORD flags = CREATE_NO_WINDOW | DETACHED_PROCESS;

    QString commandLine = program;
    if (!arguments.isEmpty()) {
        commandLine += " " + arguments.join(" ");
    }

    std::wstring wCommandLine = commandLine.toStdWString();

    bool success = CreateProcessW(
        nullptr,
        const_cast<LPWSTR>(wCommandLine.c_str()),
        nullptr,
        nullptr,
        FALSE,
        flags,
        nullptr,
        nullptr,
        &si,
        &pi
    );

    if (success) {
        CloseHandle(pi.hThread);
        CloseHandle(pi.hProcess);
    }
}
