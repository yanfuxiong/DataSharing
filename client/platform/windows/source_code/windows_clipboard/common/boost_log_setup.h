#pragma once
#include "boost_global_def.h"
#include <boost/log/core.hpp>
#include <boost/core/null_deleter.hpp>
#include <boost/log/expressions.hpp>
#include <boost/log/attributes.hpp>
#include <boost/log/attributes/attribute.hpp>
#include <boost/log/attributes/attribute_value_set.hpp>
#include <boost/log/attributes/attribute_cast.hpp>
#include <boost/log/attributes/attribute_value.hpp>
#include <boost/log/attributes/function.hpp>
#include <boost/log/utility/manipulators/add_value.hpp>
#include <boost/log/utility/setup/console.hpp>
#include <boost/log/utility/setup/common_attributes.hpp>
#include <boost/log/utility/setup/file.hpp>
#include <boost/log/sources/severity_logger.hpp>
#include <boost/log/sources/global_logger_storage.hpp>
#include <boost/log/keywords/severity.hpp>
#include <boost/log/support/date_time.hpp>
#include <boost/log/sinks/sync_frontend.hpp>
#include <boost/log/sinks/async_frontend.hpp>
#include <boost/log/sinks/unlocked_frontend.hpp>
#include <boost/log/sinks/text_file_backend.hpp>
#include <boost/log/sinks/text_ostream_backend.hpp>
#include <boost/log/sinks/syslog_backend.hpp>

#include <QDebug>
#include <QString>

enum QtSeverity
{
    QLOG_DEBUG,
    QLOG_INFO,
    QLOG_WARN,
    QLOG_CRIT, // CRITICAL
    QLOG_FATAL
};

BOOST_LOG_ATTRIBUTE_KEYWORD(k_severity, "Severity", QtSeverity)
BOOST_LOG_ATTRIBUTE_KEYWORD(k_message, "Message", std::string)
BOOST_LOG_ATTRIBUTE_KEYWORD(k_threadID, "ThreadID", uint32_t)
BOOST_LOG_ATTRIBUTE_KEYWORD(k_timeStamp, "TimeStamp", std::string)
BOOST_LOG_ATTRIBUTE_KEYWORD(k_fileName, "__FILE__", std::string)
BOOST_LOG_ATTRIBUTE_KEYWORD(k_lineValue, "__LINE__", int)

BOOST_LOG_INLINE_GLOBAL_LOGGER_DEFAULT(g_boostLogger, boost::log::sources::severity_logger_mt<QtSeverity>)

template<typename CharT, typename TraitsT>
inline std::basic_ostream<CharT, TraitsT>& operator<< (std::basic_ostream<CharT, TraitsT> &stream, QtSeverity levelVal)
{
    static const char* const s_name[] = {
        "DEBUG",
        "INFO",
        "WARN",
        "CRIT",
        "FATAL"
    };
    if (static_cast<std::size_t>(levelVal) < (sizeof(s_name) / sizeof(*s_name))) {
        stream << s_name[levelVal];
    } else {
        stream << static_cast<int>(levelVal);
    }
    return stream;
}

void g_commonMessageOutput(QtMsgType type, const QMessageLogContext &context, const QString &msg);
void g_boostLogSetup(const std::string &appName);
void g_setBoostLogLevel(QtSeverity levelVal);
void g_setBoostLoggingEnabled(bool status);
