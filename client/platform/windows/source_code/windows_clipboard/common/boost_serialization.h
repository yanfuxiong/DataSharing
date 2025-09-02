#pragma once
#include <boost/serialization/access.hpp>
#include <boost/serialization/utility.hpp>
#include <boost/archive/tmpdir.hpp>
#include <boost/serialization/base_object.hpp>
#include <boost/serialization/string.hpp>
#include <boost/serialization/vector.hpp>
#include <boost/serialization/map.hpp>
#include <boost/serialization/unordered_map.hpp>
#include <boost/serialization/list.hpp>
#include <boost/serialization/array.hpp>
#include <boost/serialization/deque.hpp>
#include <boost/serialization/shared_ptr.hpp>
#include <boost/serialization/unique_ptr.hpp>
#include <boost/archive/text_oarchive.hpp>
#include <boost/archive/text_iarchive.hpp>
#include <boost/archive/binary_iarchive.hpp>
#include <boost/archive/binary_oarchive.hpp>
#include <boost/archive/xml_iarchive.hpp>
#include <boost/archive/xml_oarchive.hpp>
#include <boost/serialization/split_free.hpp>

#include <QObject>
#include <QString>
#include <QByteArray>
#include <QDebug>
#include <iostream>
#include <sstream>
#include "boost_global_def.h"

void g_globalRegisterForBoostSerialize();

enum BoostSerializeVersion
{
    BSV_V1 = 0x01
};

BOOST_SERIALIZATION_SPLIT_FREE(QString)
BOOST_SERIALIZATION_SPLIT_FREE(QByteArray)

namespace boost {
namespace serialization {

template<class Archive>
inline void save(Archive &ar, const QString &str, const unsigned int version)
{
    Q_UNUSED(version)
    std::string data = str.toStdString();
    ar << BOOST_SERIALIZATION_NVP(data);
}

template<class Archive>
inline void load(Archive &ar, QString &str, const unsigned int version)
{
    Q_UNUSED(version)
    std::string data;
    ar >> BOOST_SERIALIZATION_NVP(data);
    str = QString::fromStdString(data);
}

template<class Archive>
inline void save(Archive &ar, const QByteArray &str, const unsigned int version)
{
    Q_UNUSED(version)
    std::string data = str.toStdString();
    ar << BOOST_SERIALIZATION_NVP(data);
}

template<class Archive>
inline void load(Archive &ar, QByteArray &str, const unsigned int version)
{
    Q_UNUSED(version)
    std::string data;
    ar >> BOOST_SERIALIZATION_NVP(data);
    str = QByteArray::fromStdString(data);
}

}}

#if 0
#define BOOST_SERIALIZE(className) BOOST_XML_SERIALIZE(className)
#else
#define BOOST_SERIALIZE(className) BOOST_BINARY_SERIALIZE(className)
#endif


#define BS_ADD_ITEM(name)\
archive & BOOST_SERIALIZATION_NVP(name)

#define BOOST_XML_SERIALIZE(className)\
public:\
    static inline QByteArray toByteArray(const className &msg)\
    {\
        std::stringstream stream;\
        {\
            boost::archive::xml_oarchive archive(stream);\
            archive << boost::serialization::make_nvp(#className, msg);\
        }\
        return QByteArray::fromStdString(stream.str());\
    }\
    static inline bool fromByteArray(const QByteArray &data, className &msg)\
    {\
        std::stringstream stream;\
        stream << data.toStdString();\
        boost::archive::xml_iarchive archive(stream);\
        archive >> boost::serialization::make_nvp(#className, msg);\
        return true;\
    }\
    friend inline QDebug operator << (QDebug stream, const className &msg)\
    {\
        stream << className::toByteArray(msg).constData();\
        return stream;\
    }\
    friend inline std::ostream &operator << (std::ostream &stream, const className &msg)\
    {\
        stream << className::toByteArray(msg).constData();\
        return stream;\
    }\
private:\
    friend class boost::serialization::access;\
    template<class Archive>\
    void serialize(Archive &archive, const unsigned int version)


#define BOOST_BINARY_SERIALIZE(className)\
public:\
    static inline QByteArray toByteArray(const className &msg)\
    {\
        std::stringstream stream;\
        {\
            boost::archive::binary_oarchive archive(stream);\
            archive << msg;\
        }\
        return QByteArray::fromStdString(stream.str());\
    }\
    static inline bool fromByteArray(const QByteArray &data, className &msg)\
    {\
        std::stringstream stream;\
        stream << data.toStdString();\
        boost::archive::binary_iarchive archive(stream);\
        archive >> msg;\
        return true;\
    }\
    friend inline QDebug operator << (QDebug stream, const className &msg)\
    {\
        stream << "[" #className "]: binary data size =" << className::toByteArray(msg).size() << "bytes";\
        return stream;\
    }\
    friend inline std::ostream &operator << (std::ostream &stream, const className &msg)\
    {\
        stream << "[" #className "]: binary data size = " << className::toByteArray(msg).size() << " bytes";\
        return stream;\
    }\
private:\
    friend class boost::serialization::access;\
    template<class Archive>\
    void serialize(Archive &archive, const unsigned int version)

struct SystemConfig
{
    QString serverVersionStr;
    QString clientVersionStr;
    QString localIpAddress;
    uint16_t port;

    BOOST_XML_SERIALIZE(SystemConfig)
    {
        Q_UNUSED(version)
        BS_ADD_ITEM(serverVersionStr);
        BS_ADD_ITEM(clientVersionStr);
        BS_ADD_ITEM(localIpAddress);
        BS_ADD_ITEM(port);
    }
};
typedef std::shared_ptr<SystemConfig> SystemConfigPtr;
Q_DECLARE_METATYPE(SystemConfigPtr)
BOOST_CLASS_VERSION(SystemConfig, BSV_V1)

struct UpdateLocalConfigInfoMsg
{
    QByteArray configData;
    QString appFilePath;
    QString appVersion;
    uint64_t timeStamp;

    BOOST_SERIALIZE(UpdateLocalConfigInfoMsg)
    {
        Q_UNUSED(version)
        BS_ADD_ITEM(configData);
        BS_ADD_ITEM(appFilePath);
        BS_ADD_ITEM(appVersion);
        BS_ADD_ITEM(timeStamp);
    }
};
typedef std::shared_ptr<UpdateLocalConfigInfoMsg> UpdateLocalConfigInfoMsgPtr;
Q_DECLARE_METATYPE(UpdateLocalConfigInfoMsgPtr)
BOOST_CLASS_VERSION(UpdateLocalConfigInfoMsg, BSV_V1)
