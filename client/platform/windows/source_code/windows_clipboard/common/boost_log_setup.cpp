#include "boost_log_setup.h"
#include <QCoreApplication>
#include <QDateTime>
#include <QFileInfo>
#include <QDir>
#include <boost/filesystem.hpp>
#include <windows.h>
#include "common_utils.h"

namespace {

uint32_t g_getCurrentThreadID()
{
    thread_local static uint32_t s_thread_id = ::GetCurrentThreadId();
    return s_thread_id;
}

std::string g_getCurrentTimeString()
{
    return QDateTime::currentDateTime().toString("yyyy-MM-dd hh:mm:ss.zzz").toStdString();
}

class FileCollector : public boost::log::sinks::file::collector
{
public:
    explicit FileCollector(const std::string &appName)
        : m_appName(appName)
        , m_logFolderPath(CommonUtils::windowsLogFolderPath().toStdString())
    {
        processAllFiles(true);
    }

    std::string baseLogFileName() const
    {
        return m_appName + ".log";
    }

    void processAllFiles(bool initStatus = false)
    {
        constexpr int maxDays = 3;
        QDateTime currentTime = QDateTime::currentDateTime();
        QDir dir(m_logFolderPath.c_str());
        for (const auto &fileName : dir.entryList({ "*" }, QDir::Files | QDir::Readable)) {
            if (!fileName.contains(m_appName.c_str()) || fileName.toStdString() == baseLogFileName()) {
                continue;
            }
            if (fileName.endsWith(".log") && initStatus) {
                continue;
            }

            if (!fileName.endsWith(".gz")) {
                continue;
            }

            QFileInfo fileInfo(dir.absoluteFilePath(fileName));
            if (fileInfo.birthTime().daysTo(currentTime) > maxDays) {
                std::cout << "remove file(.tar.gz): " << fileInfo.absoluteFilePath().toStdString() << std::endl;
                QFile::remove(fileInfo.absoluteFilePath());
            }
        }

        constexpr int maxFiles = 4; // One log file and three tar.gz files
        std::list<QString> allLogFiles;
        for (const auto &fileName : dir.entryList({ "*" }, QDir::Files | QDir::Readable)) {
            if (!fileName.contains(m_appName.c_str()) || fileName.toStdString() == baseLogFileName()) {
                continue;
            }

            if (!fileName.endsWith(".gz")) {
                continue;
            }
            allLogFiles.push_back(dir.absoluteFilePath(fileName));
        }

        allLogFiles.sort([] (const QString &left, const QString &right) {
            return QFileInfo(left).birthTime().toMSecsSinceEpoch() < QFileInfo(right).birthTime().toMSecsSinceEpoch();
        });

        while (allLogFiles.size() >= maxFiles) {
            QString filePath = allLogFiles.front();
            allLogFiles.pop_front();
            std::cout << "remove file: " << filePath.toStdString() << std::endl;
            QFile::remove(filePath);
        }
    }
    void store_file(const boost::filesystem::path &srcPath) override
    {
        std::cout << "store: " << srcPath << std::endl;
        Q_ASSERT(boost::filesystem::exists(srcPath) == true);
        processAllFiles();
        Q_ASSERT(boost::filesystem::exists(srcPath) == true); // double check
        QString outputFileName = QString::fromStdString(srcPath.string() + ".tar.gz");
        if (CommonUtils::compressFileToTarGz(QString::fromStdString(srcPath.string()), outputFileName)) {
            QFile::remove(srcPath.string().c_str());
            Q_ASSERT(boost::filesystem::exists(srcPath) == false);
            Q_ASSERT(QFile::exists(outputFileName));
        }
        processAllFiles();
    }

    bool is_in_storage(const boost::filesystem::path &srcPath) const override
    {
        std::cout << "is_in_storage: " << srcPath << std::endl;
        return true;
    }

    boost::log::sinks::file::scan_result scan_for_files(boost::log::sinks::file::scan_method method,
                                                        const boost::filesystem::path &pattern) override
    {
        std::cout << "scan_method: " << method << "; pattern:" << pattern << std::endl;
        if (method == boost::log::sinks::file::scan_method::no_scan) {
            return boost::log::sinks::file::scan_result();
        }

        QDir dir(m_logFolderPath.c_str());
        int filesCount = 0;
        for (const auto &fileName : dir.entryList({ "*" }, QDir::Files | QDir::Readable)) {
            if (fileName.contains(m_appName.c_str())) {
                std::cout << "scan_for_files: " << dir.absoluteFilePath(fileName).toStdString() << std::endl;
                ++filesCount;
            }
        }

        boost::log::sinks::file::scan_result result;
        result.found_count = filesCount;
        return result;
    }

private:
    std::string m_appName;
    std::string m_logFolderPath;
};

}

void g_setBoostLogLevel(QtSeverity levelVal)
{
    boost::log::core::get()->set_filter([levelVal] (const boost::log::attribute_value_set &attrValueSet) {
        auto itr = attrValueSet.find(k_severity.get_name());
        if (itr == attrValueSet.end()) {
            return false;
        }
        return *itr->second.extract<QtSeverity>() >= levelVal;
    });
}

void g_setBoostLoggingEnabled(bool status)
{
    boost::log::core::get()->set_logging_enabled(status);
}

void g_commonMessageOutput(QtMsgType type, const QMessageLogContext &context, const QString &msg)
{
    QtSeverity levelVal = QtSeverity::QLOG_DEBUG;
    switch (type) {
    case QtDebugMsg:
        levelVal = QtSeverity::QLOG_DEBUG;
        break;
    case QtInfoMsg:
        levelVal = QtSeverity::QLOG_INFO;
        break;
    case QtWarningMsg:
        levelVal = QtSeverity::QLOG_WARN;
        break;
    case QtCriticalMsg:
        levelVal = QtSeverity::QLOG_CRIT;
        break;
    case QtFatalMsg:
        levelVal = QtSeverity::QLOG_FATAL;
        break;
    }

    BOOST_LOG_SEV(g_boostLogger::get(), levelVal)
            << boost::log::add_value(k_fileName.get_name(), QFileInfo(context.file).fileName().toStdString())
            << boost::log::add_value(k_lineValue.get_name(), context.line)
            << msg.toStdString();
}

void g_boostLogSetup(const std::string &appName)
{
    namespace attrs = boost::log::attributes;
    namespace sinks = boost::log::sinks;
    namespace keywords = boost::log::keywords;

    boost::log::core::get()->remove_all_sinks();
    boost::log::core::get()->add_global_attribute(k_threadID.get_name(), attrs::make_function(&g_getCurrentThreadID));
    boost::log::core::get()->add_global_attribute(k_timeStamp.get_name(), attrs::make_function(&g_getCurrentTimeString));

    auto formatterFunc = [] (const boost::log::record_view &rec, boost::log::formatting_ostream &stream, bool utf8) {
            stream << rec[k_timeStamp]
                    << " " << rec[k_threadID]
                    << " " << rec[k_severity];

            if (utf8) {
                stream << " " << QString::fromStdString(*rec[k_message]).toStdString();
            } else {
                stream << " " << QString::fromStdString(*rec[k_message]).toLocal8Bit().constData();
            }

            stream << " - " << rec[k_fileName] << ":" << rec[k_lineValue];
    };

    {
        auto consoleSink = boost::log::add_console_log(std::cerr);
        consoleSink->set_formatter(std::bind(formatterFunc, std::placeholders::_1, std::placeholders::_2, false));
    }

    {
        QDir().mkpath(CommonUtils::windowsLogFolderPath());
        std::string target = CommonUtils::windowsLogFolderPath().toStdString();
        std::string fileName = target + "/" + appName + ".log";
        std::string targetFileName = target + "/" + appName +
                                 "_" +
                                 QString::number(::GetCurrentProcessId()).toStdString() +
                                 "_%Y%m%d_%H%M%S_%2N.log";

        auto textFileBackend = boost::make_shared<sinks::text_file_backend>(
            keywords::file_name = fileName,
            keywords::open_mode = std::ios_base::out | std::ios_base::app,
            keywords::target_file_name = targetFileName,
            keywords::rotation_size = 5 * 1024 * 1024, // 5M
            keywords::enable_final_rotation = false,
            keywords::time_based_rotation = sinks::file::rotation_at_time_point(0, 0, 0),
            keywords::auto_flush = true
        );

        using streamType = sinks::text_file_backend::stream_type;
        textFileBackend->set_open_handler([] (streamType &file) {
            file << std::string(75, '-') << "[LOG START]" << std::string(75, '-') << "\n";
        });

        textFileBackend->set_close_handler([] (streamType &file) {
            Q_UNUSED(file)
        });

        textFileBackend->set_file_collector(boost::make_shared<FileCollector>(appName));
        textFileBackend->scan_for_files();

        using sink_t = sinks::synchronous_sink<sinks::text_file_backend>;
        auto textFileSink = boost::make_shared<sink_t>(textFileBackend);
        textFileSink->set_formatter(std::bind(formatterFunc, std::placeholders::_1, std::placeholders::_2, true));
        boost::log::core::get()->add_sink(textFileSink);
    }
}
