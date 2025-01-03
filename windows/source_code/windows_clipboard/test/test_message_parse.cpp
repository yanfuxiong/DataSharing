#include <QTest>
#include "common_utils.h"
#include <QHostAddress>
#include <boost/filesystem.hpp>

class TestMessageParse : public QObject
{
    Q_OBJECT

private Q_SLOTS:
    void initTestCase() {}
    void cleanupTestCase() {}

    void test_ip_address()
    {
        {
            QHostAddress address("192.168.30.1");
            uint32_t ip_value = address.toIPv4Address();
            ip_value = qToBigEndian<uint32_t>(ip_value);
            QByteArray data;
            data.append(reinterpret_cast<const char*>(&ip_value), sizeof (ip_value));
            QVERIFY(data.toHex().toUpper() == "C0A81E01");
        }

        {
            QHostAddress address("192.168.30.1");
            uint32_t ip_value = address.toIPv4Address();
            Buffer buffer;
            buffer.appendUInt32(ip_value);
            QVERIFY(buffer.retrieveAllAsByteArray() == QByteArray::fromHex("C0A81E01"));
        }

        {
            uint32_t ip_value = 0;
            {
                Buffer buffer;
                QByteArray data = QByteArray::fromHex("C0A81E01");
                buffer.append(data);
                ip_value = buffer.peekUInt32();
            }
            QVERIFY(QHostAddress(ip_value).toString() == "192.168.30.1");
        }
    }

    void test_server_path()
    {
        bool exists = false;
        for (const auto &val : g_getPipeServerExePathList()) {
            if (QFile::exists(val)) {
                exists = true;
                qInfo() << "[PIPE SERVER PATH]:" << val;
                break;
            }
        }
        QVERIFY(exists == true);
    }

    void test_update_client_status_message()
    {
        {
            UpdateClientStatusMsg msg;
            msg.status = 1;
            msg.ip = "192.168.30.1";
            msg.port = 12345;
            msg.clientID = "QmQ7obXFx1XMFr6hCYXtovn9zREFqSXEtH5hdtpBDLjrAz";
            msg.clientName = R"(abc's 電腦)";

            QByteArray send_data = UpdateClientStatusMsg::toByteArray(msg);
            qInfo() << send_data.toHex().toUpper().constData();
            QVERIFY(send_data.toHex().toUpper()
            == "52544B435302030000004501C0A81E013039516D51376F6258467831584D46723668435958746F766E397A524546715358457448356864747042444C6A72417A610062006300270073002000FB966681");

            UpdateClientStatusMsg newMsg;
            UpdateClientStatusMsg::fromByteArray(send_data, newMsg);
            QVERIFY(msg.clientName == newMsg.clientName);
            QVERIFY(msg.clientID == newMsg.clientID);
            QVERIFY(newMsg.headerInfo.type == PipeMessageType::Notify);
            QVERIFY(static_cast<uint32_t>(send_data.length()) == newMsg.getMessageLength());
        }
    }

    void test_send_file_request_msg()
    {
        {
            SendFileRequestMsg msg;
            msg.ip = "192.168.30.1";
            msg.port = 12345;
            msg.clientID = "QmQ7obXFx1XMFr6hCYXtovn9zREFqSXEtH5hdtpBDLjrAz";
            //msg.fileSize = static_cast<uint64_t>(QFileInfo(__FILE__).size());
            msg.fileSize = 60727169;
            msg.timeStamp = QDateTime::currentDateTime().toUTC().toMSecsSinceEpoch();
            msg.fileName = R"(D:\jack_huang\Downloads\新增資料夾\測試.mp4)";

            QByteArray send_data = SendFileRequestMsg::toByteArray(msg);
            qInfo() << send_data.toHex().toUpper().constData();

            SendFileRequestMsg newMsg;
            SendFileRequestMsg::fromByteArray(send_data, newMsg);
            QVERIFY(msg.fileName == newMsg.fileName);
            QVERIFY(msg.clientID == newMsg.clientID);
            QVERIFY(newMsg.headerInfo.type == PipeMessageType::Request);
            QVERIFY(static_cast<uint32_t>(send_data.length()) == newMsg.getMessageLength());

            qInfo() << QDateTime::fromMSecsSinceEpoch(newMsg.timeStamp, Qt::TimeSpec::UTC).toString("yyyy-MM-dd hh:mm:ss.zzz");
            qInfo() << newMsg.fileName.toUtf8().constData();

            {
                uint8_t typeValue = 99;
                uint8_t codeValue = 66;
                QVERIFY(g_getCodeFromByteArray(send_data, typeValue, codeValue) && typeValue == PipeMessageType::Request);
            }
        }
    }

    void test_send_file_response_msg()
    {
        {
            SendFileResponseMsg msg;
            msg.statusCode = 1; // accept
            msg.ip = "192.168.30.1";
            msg.port = 12345;
            msg.clientID = "QmQ7obXFx1XMFr6hCYXtovn9zREFqSXEtH5hdtpBDLjrAz";
            //msg.fileSize = static_cast<uint64_t>(QFileInfo(__FILE__).size());
            msg.fileSize = 60727169;
            msg.timeStamp = QDateTime::currentDateTime().toUTC().toMSecsSinceEpoch();
            msg.fileName = R"(D:\jack_huang\Downloads\新增資料夾\測試.mp4)";

            QByteArray send_data = SendFileResponseMsg::toByteArray(msg);
            qInfo() << send_data.toHex().toUpper().constData();

            SendFileResponseMsg newMsg;
            SendFileResponseMsg::fromByteArray(send_data, newMsg);
            QVERIFY(msg.fileName == newMsg.fileName);
            QVERIFY(msg.clientID == newMsg.clientID);
            QVERIFY(msg.statusCode == newMsg.statusCode);
            QVERIFY(newMsg.headerInfo.type == PipeMessageType::Response);
            QVERIFY(static_cast<uint32_t>(send_data.length()) == newMsg.getMessageLength());

            qInfo() << QDateTime::fromMSecsSinceEpoch(newMsg.timeStamp, Qt::TimeSpec::UTC).toString("yyyy-MM-dd hh:mm:ss.zzz");
            qInfo() << newMsg.fileName.toUtf8().constData();

            {
                uint8_t typeValue = 99;
                uint8_t codeValue = 66;
                QVERIFY(g_getCodeFromByteArray(send_data, typeValue, codeValue) && typeValue == PipeMessageType::Response);
            }
        }
    }

    void test_update_progress_msg()
    {
        {
            UpdateProgressMsg msg;
            msg.ip = "192.168.30.1";
            msg.port = 12345;
            msg.clientID = "QmQ7obXFx1XMFr6hCYXtovn9zREFqSXEtH5hdtpBDLjrAz";
            //msg.fileSize = static_cast<uint64_t>(QFileInfo(__FILE__).size());
            msg.fileSize = 60727169;
            msg.sentSize = 100;
            msg.timeStamp = QDateTime::currentDateTime().toUTC().toMSecsSinceEpoch();
            msg.fileName = R"(D:\jack_huang\Downloads\新增資料夾\測試.mp4)";

            QByteArray send_data = UpdateProgressMsg::toByteArray(msg);
            qInfo() << send_data.toHex().toUpper().constData();

            UpdateProgressMsg newMsg;
            UpdateProgressMsg::fromByteArray(send_data, newMsg);
            QVERIFY(msg.fileName == newMsg.fileName);
            QVERIFY(msg.clientID == newMsg.clientID);
            QVERIFY(msg.sentSize == newMsg.sentSize);
            QVERIFY(newMsg.headerInfo.type == PipeMessageType::Notify);
            QVERIFY(static_cast<uint32_t>(send_data.length()) == newMsg.getMessageLength());

            qInfo() << QDateTime::fromMSecsSinceEpoch(newMsg.timeStamp, Qt::TimeSpec::UTC).toString("yyyy-MM-dd hh:mm:ss.zzz");
            qInfo() << newMsg.fileName.toUtf8().constData();


            {
                uint8_t typeValue = 99;
                uint8_t codeValue = 66;
                QVERIFY(g_getCodeFromByteArray(send_data, typeValue, codeValue) && typeValue == PipeMessageType::Notify);
            }
        }
    }

    void test_update_image_progress_msg()
    {
        {
            UpdateImageProgressMsg msg;
            msg.ip = "192.168.30.1";
            msg.port = 12345;
            msg.clientID = "QmQ7obXFx1XMFr6hCYXtovn9zREFqSXEtH5hdtpBDLjrAz";
            //msg.fileSize = static_cast<uint64_t>(QFileInfo(__FILE__).size());
            msg.fileSize = 60727169;
            msg.sentSize = 100;
            msg.timeStamp = QDateTime::currentDateTime().toUTC().toMSecsSinceEpoch();

            QByteArray send_data = UpdateImageProgressMsg::toByteArray(msg);
            qInfo() << send_data.toHex().toUpper().constData();

            UpdateImageProgressMsg newMsg;
            UpdateImageProgressMsg::fromByteArray(send_data, newMsg);
            QVERIFY(msg.clientID == newMsg.clientID);
            QVERIFY(msg.sentSize == newMsg.sentSize);
            QVERIFY(newMsg.headerInfo.type == PipeMessageType::Notify);
            QVERIFY(static_cast<uint32_t>(send_data.length()) == newMsg.getMessageLength());

            qInfo() << QDateTime::fromMSecsSinceEpoch(newMsg.timeStamp, Qt::TimeSpec::UTC).toString("yyyy-MM-dd hh:mm:ss.zzz");


            {
                uint8_t typeValue = 99;
                uint8_t codeValue = 66;
                QVERIFY(g_getCodeFromByteArray(send_data, typeValue, codeValue) && typeValue == PipeMessageType::Notify);
            }
        }
    }

    void test_get_conn_status()
    {
        {
            GetConnStatusRequestMsg msg;
            QVERIFY(msg.headerInfo.type == PipeMessageType::Request);
            QVERIFY(msg.headerInfo.contentLength == 0);
            QByteArray sendData = GetConnStatusRequestMsg::toByteArray(msg);

            GetConnStatusRequestMsg newMsg;
            {
                //Data with assignment confusion is used for testing
                newMsg.headerInfo.header = "hello world";
                newMsg.headerInfo.code = 7;
                newMsg.headerInfo.type = 6;
                newMsg.headerInfo.contentLength = 9999;
            }
            GetConnStatusRequestMsg::fromByteArray(sendData, newMsg);
            QVERIFY(newMsg.headerInfo.header == TAG_NAME);
            QVERIFY(newMsg.headerInfo.type == msg.headerInfo.type);
            QVERIFY(newMsg.headerInfo.code == msg.headerInfo.code);
            QVERIFY(newMsg.headerInfo.contentLength == 0);
            QVERIFY(newMsg.headerInfo.type == PipeMessageType::Request);
        }

        {
            GetConnStatusResponseMsg msg;
            QVERIFY(msg.headerInfo.type == PipeMessageType::Response);
            msg.statusCode = 1;
            QByteArray sendData = GetConnStatusResponseMsg::toByteArray(msg);

            GetConnStatusResponseMsg newMsg;
            {
                //Data with assignment confusion is used for testing
                newMsg.headerInfo.header = "hello world";
                newMsg.headerInfo.code = 7;
                newMsg.headerInfo.type = 6;
                newMsg.headerInfo.contentLength = 9999;
            }
            GetConnStatusResponseMsg::fromByteArray(sendData, newMsg);

            QVERIFY(newMsg.headerInfo.header == TAG_NAME);
            QVERIFY(newMsg.headerInfo.type == msg.headerInfo.type);
            QVERIFY(newMsg.headerInfo.code == msg.headerInfo.code);
            QVERIFY(newMsg.headerInfo.contentLength == 1);
            QVERIFY(newMsg.statusCode == msg.statusCode && newMsg.statusCode == 1);
            QVERIFY(newMsg.headerInfo.type == PipeMessageType::Response);
        }
    }

    void test_notify_message()
    {
        {
            NotifyMessage msg;
            msg.timeStamp = QDateTime::currentDateTime().toUTC().toMSecsSinceEpoch();
            msg.notiCode = 2;

            {
                NotifyMessage::ParamInfo paramInfo;
                paramInfo.info = "测试设备_1";
                msg.paramInfoVec.push_back(paramInfo);
            }

            {
                NotifyMessage::ParamInfo paramInfo;
                paramInfo.info = "10";
                msg.paramInfoVec.push_back(paramInfo);
            }

            QByteArray send_data = NotifyMessage::toByteArray(msg);
            qInfo() << send_data.toHex().toUpper().constData();

            NotifyMessage newMsg;
            NotifyMessage::fromByteArray(send_data, newMsg);
            QVERIFY(newMsg.headerInfo.type == PipeMessageType::Notify);
            QVERIFY(static_cast<uint32_t>(send_data.length()) == newMsg.getMessageLength());


            QVERIFY(newMsg.paramInfoVec.size() == 2);
            QVERIFY(newMsg.paramInfoVec.front().info == "测试设备_1");
            QVERIFY(newMsg.paramInfoVec.back().info == "10");

            qInfo()<< newMsg.toString().dump(4).c_str();
        }
    }

    void test_update_system_info()
    {
        {
            UpdateSystemInfoMsg msg;
            msg.ip = "192.168.30.1";
            msg.port = 12345;
            msg.serverVersion = R"(v1.0.1)";

            QByteArray send_data = UpdateSystemInfoMsg::toByteArray(msg);
            qInfo() << send_data.toHex().toUpper().constData();

            UpdateSystemInfoMsg newMsg;
            UpdateSystemInfoMsg::fromByteArray(send_data, newMsg);
            QVERIFY(newMsg.headerInfo.type == PipeMessageType::Notify);
            QVERIFY(newMsg.ip == "192.168.30.1");
            QVERIFY(newMsg.port == 12345);
            QVERIFY(newMsg.serverVersion == "v1.0.1");
            QVERIFY(static_cast<uint32_t>(send_data.length()) == newMsg.getMessageLength());
            qInfo() << newMsg.serverVersion.toUtf8().constData();
        }
    }

    void test_file_opt_record()
    {
        boost::circular_buffer<FileOperationRecord> test_record { 100 };
        QVERIFY(test_record.capacity() == 100);
        QVERIFY(test_record.size() == 0 && test_record.empty() == true);
        {
            FileOperationRecord record;
            record.fileName = __FILE__;
            record.fileSize = boost::filesystem::file_size(__FILE__);
            record.timeStamp = QDateTime::currentDateTime().toMSecsSinceEpoch();
            record.progressValue = 80;
            record.clientID = "aaaaaaaaaaaa";
            record.clientName = "Device A";
            record.direction = 0;

            test_record.push_back(record);
        }

        {
            FileOperationRecord record;
            record.fileName = __FILE__;
            record.fileSize = boost::filesystem::file_size(__FILE__);
            record.timeStamp = QDateTime::currentDateTime().toMSecsSinceEpoch();
            record.progressValue = 100;

            record.clientID = "bbbbbbbbbbbbb";
            record.clientName = "Device C";
            record.direction = 1;

            test_record.push_back(record);
        }

        QVERIFY(test_record.size() == 2 && test_record.empty() == false);

        QVERIFY(test_record.front().fileSize == QFileInfo(test_record.front().fileName.c_str()).size());
        qInfo() << QUuid::createUuid().toString();
    }
};

QTEST_GUILESS_MAIN(TestMessageParse)

#include "test_message_parse.moc"

