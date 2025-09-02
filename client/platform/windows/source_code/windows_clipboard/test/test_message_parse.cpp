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

            UpdateClientStatusMsg newMsg;
            UpdateClientStatusMsg::fromByteArray(send_data, newMsg);
            QVERIFY(msg.clientName == newMsg.clientName);
            QVERIFY(msg.clientID == newMsg.clientID);
            QVERIFY(msg.deviceType == newMsg.deviceType);
            QVERIFY(newMsg.headerInfo.type == PipeMessageType::Notify);
            QVERIFY(static_cast<uint32_t>(send_data.length()) == newMsg.getMessageLength());
        }
    }

    void test_send_file_request_msg()
    {
        {
            SendFileRequestMsg msg;
            msg.flag = SendFileRequestMsg::FlagType::SendFlag;
            msg.ip = "192.168.30.1";
            msg.port = 12345;
            msg.clientID = "QmQ7obXFx1XMFr6hCYXtovn9zREFqSXEtH5hdtpBDLjrAz";
            //msg.fileSize = static_cast<uint64_t>(QFileInfo(__FILE__).size());
            msg.fileSize = 60727169;
            msg.timeStamp = QDateTime::currentDateTime().toUTC().toMSecsSinceEpoch();
            msg.fileName = R"(D:\jack_huang\Downloads\新增資料夾\測試.mp4)";
            msg.filePathVec.push_back(R"(D:\test_folder\test_1.log)");
            msg.filePathVec.push_back(R"(D:\test_folder\test_2.log)");
            msg.filePathVec.push_back(R"(D:\test_folder\test_3.log)");

            QByteArray send_data = SendFileRequestMsg::toByteArray(msg);
            qInfo() << send_data.toHex().toUpper().constData();

            SendFileRequestMsg newMsg;
            SendFileRequestMsg::fromByteArray(send_data, newMsg);
            QVERIFY(msg.flag == newMsg.flag);
            QVERIFY(msg.fileName == newMsg.fileName);
            QVERIFY(msg.clientID == newMsg.clientID);
            QVERIFY(msg.filePathVec == newMsg.filePathVec);
            QVERIFY(newMsg.headerInfo.type == PipeMessageType::Request);
            QVERIFY(static_cast<uint32_t>(send_data.length()) == newMsg.getMessageLength());

            qInfo() << QDateTime::fromMSecsSinceEpoch(newMsg.timeStamp, Qt::TimeSpec::UTC).toString("yyyy-MM-dd hh:mm:ss.zzz");
            qInfo() << newMsg.fileName.toUtf8().constData();
            qInfo() << newMsg.filePathVec.front().toUtf8().constData();

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

        {
            UpdateProgressMsg msg;
            msg.functionCode = UpdateProgressMsg::FuncCode::MultiFile;
            msg.ip = "192.168.30.1";
            msg.port = 12345;
            msg.clientID = "QmQ7obXFx1XMFr6hCYXtovn9zREFqSXEtH5hdtpBDLjrAz";
            msg.timeStamp = QDateTime::currentDateTime().toUTC().toMSecsSinceEpoch();

            msg.currentFileName = R"(D:/文件夹_root/文件夹_1/files_1.log)";
            msg.sentFilesCount = 50;
            msg.totalFilesCount = 100;
            msg.currentFileSize = 123456;
            msg.totalFilesSize = 999999;
            msg.totalSentSize = 333333;

            QByteArray send_data = UpdateProgressMsg::toByteArray(msg);
            qInfo() << send_data.toHex().toUpper().constData();

            UpdateProgressMsg newMsg;
            UpdateProgressMsg::fromByteArray(send_data, newMsg);
            QVERIFY(msg.functionCode == newMsg.functionCode);
            QVERIFY(msg.ip == newMsg.ip);
            QVERIFY(msg.port == newMsg.port);
            QVERIFY(msg.clientID == newMsg.clientID);
            QVERIFY(msg.timeStamp == newMsg.timeStamp);

            QVERIFY(msg.currentFileName == newMsg.currentFileName);
            QVERIFY(msg.sentFilesCount == newMsg.sentFilesCount);
            QVERIFY(msg.totalFilesCount == newMsg.totalFilesCount);
            QVERIFY(msg.currentFileSize == newMsg.currentFileSize);
            QVERIFY(msg.totalFilesSize == newMsg.totalFilesSize);
            QVERIFY(msg.totalSentSize == newMsg.totalSentSize);

            QVERIFY(newMsg.headerInfo.type == PipeMessageType::Notify);
            QVERIFY(static_cast<uint32_t>(send_data.length()) == newMsg.getMessageLength());

            qInfo() << QDateTime::fromMSecsSinceEpoch(newMsg.timeStamp, Qt::TimeSpec::UTC).toString("yyyy-MM-dd hh:mm:ss.zzz");
            qInfo() << newMsg.fileName.toUtf8().constData();

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

    void test_ddcci_msg()
    {
        {
            DDCCINotifyMsg msg;
            msg.macAddress = "abcdefg";
            msg.functionCode = 2;
            msg.source = 2;
            msg.port = 8;
            msg.authResult = 6;
            msg.indexValue = 3;

            QByteArray send_data = DDCCINotifyMsg::toByteArray(msg);
            qInfo() << send_data.toHex().toUpper().constData();

            DDCCINotifyMsg newMsg;
            DDCCINotifyMsg::fromByteArray(send_data, newMsg);
            QVERIFY(newMsg.headerInfo.code == DDCCIMsg_code);
            QVERIFY(newMsg.headerInfo.type == PipeMessageType::Notify);
            QVERIFY(newMsg.macAddress == "abcdefg");
            QVERIFY(newMsg.functionCode == 2);
            QVERIFY(newMsg.source == 2);
            QVERIFY(newMsg.port == 8);
            QVERIFY(newMsg.authResult == 6);
            QVERIFY(newMsg.indexValue == 3);
            QVERIFY(static_cast<uint32_t>(send_data.length()) == newMsg.getMessageLength());
        }

        {
            DDCCINotifyMsg msg;
            msg.functionCode = DDCCINotifyMsg::FuncCode::MacAddress;
            msg.macAddress = "aabbccdd";

            QByteArray send_data = DDCCINotifyMsg::toByteArray(msg);
            qInfo() << send_data.toHex().toUpper().constData();

            DDCCINotifyMsg newMsg;
            DDCCINotifyMsg::fromByteArray(send_data, newMsg);
            QVERIFY(newMsg.headerInfo.code == DDCCIMsg_code);
            QVERIFY(newMsg.headerInfo.type == PipeMessageType::Notify);
            QVERIFY(newMsg.functionCode == DDCCINotifyMsg::FuncCode::MacAddress);
            QVERIFY(newMsg.macAddress == "aabbccdd");
            QVERIFY(static_cast<uint32_t>(send_data.length()) == newMsg.getMessageLength());
        }

        {
            DDCCINotifyMsg msg;
            msg.functionCode = DDCCINotifyMsg::FuncCode::AuthViaIndex;
            msg.indexValue = 0x1234;

            QByteArray send_data = DDCCINotifyMsg::toByteArray(msg);
            qInfo() << send_data.toHex().toUpper().constData();

            DDCCINotifyMsg newMsg;
            DDCCINotifyMsg::fromByteArray(send_data, newMsg);
            QVERIFY(newMsg.headerInfo.code == DDCCIMsg_code);
            QVERIFY(newMsg.headerInfo.type == PipeMessageType::Notify);
            QVERIFY(newMsg.functionCode == DDCCINotifyMsg::FuncCode::AuthViaIndex);
            QVERIFY(newMsg.indexValue == 0x1234);
            QVERIFY(static_cast<uint32_t>(send_data.length()) == newMsg.getMessageLength());
        }

        {
            DDCCINotifyMsg msg;
            msg.functionCode = DDCCINotifyMsg::FuncCode::ReturnAuthStatus;
            msg.authResult = 3;

            QByteArray send_data = DDCCINotifyMsg::toByteArray(msg);
            qInfo() << send_data.toHex().toUpper().constData();

            DDCCINotifyMsg newMsg;
            DDCCINotifyMsg::fromByteArray(send_data, newMsg);
            QVERIFY(newMsg.headerInfo.code == DDCCIMsg_code);
            QVERIFY(newMsg.headerInfo.type == PipeMessageType::Notify);
            QVERIFY(newMsg.functionCode == DDCCINotifyMsg::FuncCode::ReturnAuthStatus);
            QVERIFY(newMsg.authResult == 3);
            QVERIFY(static_cast<uint32_t>(send_data.length()) == newMsg.getMessageLength());
        }

        {
            DDCCINotifyMsg msg;
            msg.functionCode = DDCCINotifyMsg::FuncCode::RequestSourcePort;

            QByteArray send_data = DDCCINotifyMsg::toByteArray(msg);
            qInfo() << send_data.toHex().toUpper().constData();

            DDCCINotifyMsg newMsg;
            DDCCINotifyMsg::fromByteArray(send_data, newMsg);
            QVERIFY(newMsg.headerInfo.code == DDCCIMsg_code);
            QVERIFY(newMsg.headerInfo.type == PipeMessageType::Notify);
            QVERIFY(newMsg.functionCode == DDCCINotifyMsg::FuncCode::RequestSourcePort);
            {
                QVERIFY(newMsg.macAddress.empty() == true);
                QVERIFY(newMsg.source == 0 && newMsg.port == 0);
            }
            QVERIFY(static_cast<uint32_t>(send_data.length()) == newMsg.getMessageLength());
        }

        {
            DDCCINotifyMsg msg;
            msg.functionCode = DDCCINotifyMsg::FuncCode::ReturnSourcePort;
            msg.source = 0x1122;
            msg.port = 0x2233;

            QByteArray send_data = DDCCINotifyMsg::toByteArray(msg);
            qInfo() << send_data.toHex().toUpper().constData();

            DDCCINotifyMsg newMsg;
            DDCCINotifyMsg::fromByteArray(send_data, newMsg);
            QVERIFY(newMsg.headerInfo.code == DDCCIMsg_code);
            QVERIFY(newMsg.headerInfo.type == PipeMessageType::Notify);
            QVERIFY(newMsg.functionCode == DDCCINotifyMsg::FuncCode::ReturnSourcePort);
            QVERIFY(static_cast<uint32_t>(send_data.length()) == newMsg.getMessageLength());
            QVERIFY(newMsg.source == 0x1122);
            QVERIFY(newMsg.port == 0x2233);
        }

        {
            DDCCINotifyMsg msg;
            msg.functionCode = DDCCINotifyMsg::FuncCode::ExtractDIASMonitor;

            QByteArray send_data = DDCCINotifyMsg::toByteArray(msg);
            qInfo() << send_data.toHex().toUpper().constData();

            DDCCINotifyMsg newMsg;
            DDCCINotifyMsg::fromByteArray(send_data, newMsg);
            QVERIFY(newMsg.headerInfo.code == DDCCIMsg_code);
            QVERIFY(newMsg.headerInfo.type == PipeMessageType::Notify);
            QVERIFY(newMsg.functionCode == DDCCINotifyMsg::FuncCode::ExtractDIASMonitor);
            QVERIFY(static_cast<uint32_t>(send_data.length()) == newMsg.getMessageLength());
        }
    }

    void test_drag_files_msg()
    {
        {
            DragFilesMsg msg;
            msg.functionCode = DragFilesMsg::FuncCode::MultiFiles;
            msg.timeStamp = QDateTime::currentDateTime().toMSecsSinceEpoch();
            msg.rootPath = "D:/文件夹_root";
            msg.filePathVec.push_back("D:/文件夹_root/文件夹_1/files_1.log");
            msg.filePathVec.push_back("D:/文件夹_root/文件夹_2/files_2.log");
            msg.filePathVec.push_back("D:/文件夹_root/文件夹_3/files_3.log");

            QByteArray send_data = DragFilesMsg::toByteArray(msg);
            qInfo() << send_data.toHex().toUpper().constData();

            DragFilesMsg newMsg;
            DragFilesMsg::fromByteArray(send_data, newMsg);
            QVERIFY(newMsg.headerInfo.type == PipeMessageType::Notify);
            QVERIFY(static_cast<uint32_t>(send_data.length()) == newMsg.getMessageLength());
            QVERIFY(msg.functionCode == newMsg.functionCode);
            QVERIFY(msg.filePathVec == newMsg.filePathVec);
            QVERIFY(msg.timeStamp == newMsg.timeStamp);
            QVERIFY(msg.rootPath == newMsg.rootPath);
        }

        {
            DragFilesMsg msg;
            msg.functionCode = DragFilesMsg::FuncCode::ReceiveFileInfo;
            msg.timeStamp = QDateTime::currentDateTime().toMSecsSinceEpoch();
            msg.ip = "192.168.1.123";
            msg.port = 6789;
            msg.clientID = "QmQ7obXFx1XMFr6hCYXtovn9zREFqSXEtH5hdtpBDLjrAz";
            msg.fileCount = 50;
            msg.totalFileSize = 123456;
            msg.firstTransferFileName = "D:/文件夹_root/文件夹_1/files_1.log";
            msg.firstTransferFileSize = 9999;

            QByteArray send_data = DragFilesMsg::toByteArray(msg);
            qInfo() << send_data.toHex().toUpper().constData();

            DragFilesMsg newMsg;
            DragFilesMsg::fromByteArray(send_data, newMsg);
            QVERIFY(newMsg.headerInfo.type == PipeMessageType::Notify);
            QVERIFY(static_cast<uint32_t>(send_data.length()) == newMsg.getMessageLength());
            QVERIFY(msg.functionCode == newMsg.functionCode);
            QVERIFY(msg.timeStamp == newMsg.timeStamp);

            QVERIFY(msg.ip == newMsg.ip);
            QVERIFY(msg.port == newMsg.port);
            QVERIFY(msg.clientID == newMsg.clientID);
            QVERIFY(msg.fileCount == newMsg.fileCount);
            QVERIFY(msg.totalFileSize == newMsg.totalFileSize);
            QVERIFY(msg.firstTransferFileName == newMsg.firstTransferFileName);
            QVERIFY(msg.firstTransferFileSize == newMsg.firstTransferFileSize);
        }
    }

    void test_status_info_notify_msg()
    {
        {
            StatusInfoNotifyMsg msg;
            msg.statusCode = 5;

            QByteArray send_data = StatusInfoNotifyMsg::toByteArray(msg);
            qInfo() << send_data.toHex().toUpper().constData();

            StatusInfoNotifyMsg newMsg;
            StatusInfoNotifyMsg::fromByteArray(send_data, newMsg);
            QVERIFY(newMsg.headerInfo.type == PipeMessageType::Notify);
            QVERIFY(static_cast<uint32_t>(send_data.length()) == newMsg.getMessageLength());
            QVERIFY(msg.statusCode == newMsg.statusCode);
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

