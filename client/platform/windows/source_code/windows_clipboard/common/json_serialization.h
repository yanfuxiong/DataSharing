#pragma once
#include <QObject>
#include <QString>
#include <QByteArray>
#include <QDebug>
#include <iostream>
#include <sstream>
#include "boost_global_def.h"
#include <nlohmann/json.hpp>


void g_globalRegisterForJsonSerialize();

namespace nlohmann {
template <>
struct adl_serializer<QString>
{
    static inline void to_json(json &j, const QString &qstr)
    {
        j = qstr.toStdString();
    }

    static inline void from_json(const json &j, QString &qstr)
    {
        qstr = QString::fromStdString(j.get<std::string>());
    }
};

template <>
struct adl_serializer<QByteArray>
{
    static inline void to_json(json &j, const QByteArray &qstr)
    {
        j = qstr.toStdString();
    }

    static inline void from_json(const json &j, QByteArray &qstr)
    {
        qstr = QByteArray::fromStdString(j.get<std::string>());
    }
};

}

#define JSON_SERIALIZE(className, ...)\
NLOHMANN_DEFINE_TYPE_INTRUSIVE(className, __VA_ARGS__)\
static inline QByteArray toByteArray(const className &msg)\
{\
    try {\
        return QByteArray::fromStdString(nlohmann::json(msg).dump());\
    } catch (const std::exception &e) {\
        std::cerr << e.what() << std::endl;\
        return {};\
    }\
}\
static inline bool fromByteArray(const QByteArray &data, className &msg)\
{\
    try {\
        nlohmann::json::parse(data).get_to(msg);\
        return true;\
    } catch (const std::exception &e) {\
        std::cerr << e.what() << std::endl;\
        return false;\
    }\
}\
friend inline QDebug operator << (QDebug stream, const className &msg)\
{\
    try {\
        stream << nlohmann::json(msg).dump(4).c_str();\
        return stream;\
    } catch (const std::exception &e) {\
        std::cerr << e.what() << std::endl;\
        return stream;\
    }\
}\
friend inline std::ostream &operator << (std::ostream &stream, const className &msg)\
{\
    try {\
        stream << nlohmann::json(msg).dump(4).c_str();\
        return stream;\
    } catch (const std::exception &e) {\
        std::cerr << e.what() << std::endl;\
        return stream;\
    }\
}

struct UpdateDownloadPathMsg
{
    QString downloadPath;

    JSON_SERIALIZE(UpdateDownloadPathMsg,
                   downloadPath
    )
};
typedef std::shared_ptr<UpdateDownloadPathMsg> UpdateDownloadPathMsgPtr;
Q_DECLARE_METATYPE(UpdateDownloadPathMsgPtr)


struct UpdateClientVersionMsg
{
    QString clientVersion;

    JSON_SERIALIZE(UpdateClientVersionMsg,
                   clientVersion
    )
};
typedef std::shared_ptr<UpdateClientVersionMsg> UpdateClientVersionMsgPtr;
Q_DECLARE_METATYPE(UpdateClientVersionMsgPtr)

struct ShowWindowsClipboardMsg
{
    QString desc;

    JSON_SERIALIZE(ShowWindowsClipboardMsg,
                   desc
    )
};
typedef std::shared_ptr<ShowWindowsClipboardMsg> ShowWindowsClipboardMsgPtr;
Q_DECLARE_METATYPE(ShowWindowsClipboardMsgPtr)

struct NotifyErrorEventMsg
{
    QByteArray clientID;
    uint32_t errorCode;
    QByteArray ipPortString;
    QByteArray timeStamp;
    QByteArray argValue3;
    QByteArray argValue4;

    JSON_SERIALIZE(NotifyErrorEventMsg,
                   clientID,
                   errorCode,
                   ipPortString,
                   timeStamp,
                   argValue3,
                   argValue4
    )
};
typedef std::shared_ptr<NotifyErrorEventMsg> NotifyErrorEventMsgPtr;
Q_DECLARE_METATYPE(NotifyErrorEventMsgPtr)
