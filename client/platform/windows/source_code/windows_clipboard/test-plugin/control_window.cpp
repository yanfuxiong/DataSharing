#include "control_window.h"
#include "ui_control_window.h"
#include "plugin.h"
#include <QMimeData>
#include <QApplication>
#include <QClipboard>
#include <QBuffer>
#include <QPainter>
#include <random>

ControlWindow::ControlWindow(QWidget *parent)
    : QMainWindow(parent)
    , ui(new Ui::ControlWindow)
{
    ui->setupUi(this);
    setWindowTitle("testPlugin");
    setAttribute(Qt::WA_DeleteOnClose, true);

    {
        connect(CommonSignals::getInstance(), &CommonSignals::logMessage, this, &ControlWindow::appendLogInfo);
    }
}

ControlWindow::~ControlWindow()
{
    delete ui;
    g_testWindow = nullptr;
}

void ControlWindow::appendLogInfo(const QString &logString)
{
    ui->logInfoEdit->append(logString);
}

void ControlWindow::on_UpdateClientStatus_btn_clicked()
{
    m_workerThread.runInThread([] {
        static const wchar_t *s_nameArry[] = { L"HDMI1-\nTEST", L"HDMI2", L"Miracast", L"USBC1", L"DP2" };
        Q_ASSERT(sizeof(s_nameArry) / sizeof (s_nameArry[0]) == 5);
        static const char *s_deviceType[] = { "HDMI1", "HDMI2", "Miracast", "USBC1", "DP2" };
        Q_ASSERT(sizeof(s_deviceType) / sizeof (s_deviceType[0]) == 5);
        std::string clientID { "QmQ7obXFx1XMFr6hCYXtovn9zREFqSXEtH5hdtpBDLjrAz" };
        static int s_index = 0;
        ++s_index;
        for (int index = 0; index < 5; ++index) {
            std::string newClientID = clientID;
            newClientID.back() = std::to_string(index + 1).back();
            std::string ip { "192.168.0.1" };
            uint8_t port = 80 + index;
            std::string ipPortString = ip + ":" + std::to_string(port);
            int arryIndex = QRandomGenerator::global()->generate() % 5;
            if (s_index % 2 == 0) {
                g_UpdateClientStatus(1,
                                     ipPortString.c_str(),
                                     newClientID.data(),
                                     s_nameArry[arryIndex],
                                     s_deviceType[arryIndex]);
            } else {
                g_UpdateClientStatus((index % 2 == 0) ? 0 : 1,
                                     ipPortString.c_str(),
                                     newClientID.data(),
                                     s_nameArry[arryIndex],
                                     s_deviceType[arryIndex]);
            }
        }
    });
}

void ControlWindow::on_UpdateSystemInfo_btn_clicked()
{
    m_workerThread.runInThread([] {
        uint32_t index = QRandomGenerator::global()->generate() % 3;
        if (index == 0) {
            g_UpdateSystemInfo("192.168.6.123:66", L"server_v1.0.1");
        } else if (index == 1) {
            g_UpdateSystemInfo("192.168.6.124:77", L"server_v1.0.2");
        } else {
            g_UpdateSystemInfo("192.168.6.125:88", L"server_v1.0.3");
        }
    });
}

void ControlWindow::on_GetDownloadPath_btn_clicked()
{
    m_workerThread.runInThread([] {
        Q_EMIT CommonSignals::getInstance()->logMessage("GetDownloadPath: " + g_downloadPath);
    });
}

void ControlWindow::on_GetDeviceName_btn_clicked()
{
    m_workerThread.runInThread([] {
        Q_EMIT CommonSignals::getInstance()->logMessage("GetDeviceName: " + g_deviceName);
    });
}

void ControlWindow::on_StartClipboardMonitor_btn_clicked()
{
    m_workerThread.runInThread([] {
        g_StartClipboardMonitor();
    });
}

void ControlWindow::on_StopClipboardMonitor_btn_clicked()
{
    m_workerThread.runInThread([] {
        g_StopClipboardMonitor();
    });
}

void ControlWindow::on_AuthViaIndex_btn_clicked()
{
    m_workerThread.runInThread([] {
        g_AuthViaIndex(static_cast<uint32_t>(0x12345678));
    });
}

void ControlWindow::on_RequestSourceAndPort_btn_clicked()
{
    m_workerThread.runInThread([] {
        g_RequestSourceAndPort();
    });
}

void ControlWindow::on_DIASStatus_clicked()
{
    m_workerThread.runInThread([] {
        static uint32_t s_statusCode = 0;
        ++s_statusCode;
        if (s_statusCode > 7) {
            s_statusCode = 1;
        }

        g_DIASStatus(s_statusCode);
    });
}

void ControlWindow::on_DIAS_Connected_clicked()
{
    g_DIASStatus(7);
}

void ControlWindow::on_CleanClipboard_btn_clicked()
{
    m_workerThread.runInThread([] {
        g_CleanClipboard();
    });
}

void ControlWindow::on_NotiMessage_btn_clicked()
{
    m_workerThread.runInThread([] {
        static int s_index = 0;
        QString info = QString::number(++s_index);
        std::wstring info_wstring = info.toStdWString();
        const wchar_t *message_arry[] = { L"HDMI_TEST", info_wstring.c_str() };
        uint64_t timestamp = QDateTime::currentDateTime().toMSecsSinceEpoch();
        g_NotiMessage(timestamp, 2, message_arry, sizeof (message_arry) / sizeof (message_arry[0]));
    });
}

static uint64_t s_timeStamp_for_list { 0 };
static std::atomic<bool> s_runningStatus_for_list { false };
void ControlWindow::on_DragFileListNotify_btn_clicked()
{
    s_timeStamp_for_list = QDateTime::currentDateTime().toMSecsSinceEpoch();
    const char *ipPort = "192.168.0.1:80";
    const char *clientID = "QmQ7obXFx1XMFr6hCYXtovn9zREFqSXEtH5hdtpBDLjrA1";
    QString filePath(R"(C:\Users\TS\Desktop\test_data.mp4)");
    uint32_t cFileCount = 10;
    uint64_t totalSize = 1024 * 1024 * 100;
    uint64_t timeStamp = s_timeStamp_for_list;
    QString currentFileName = R"(C:\Users\TS\Desktop\test_1.log)";
    std::wstring currentFileName_wstring = currentFileName.toStdWString();
    uint64_t firstFileSize = 10 * 1024 * 1024;

    g_DragFileListNotify(ipPort,
                         clientID,
                         cFileCount,
                         totalSize,
                         timeStamp,
                         currentFileName_wstring.c_str(),
                         firstFileSize);
}

void ControlWindow::on_UpdateMultipleProgress_btn_clicked()
{
    s_runningStatus_for_list.store(true);
    std::thread([] {
        if (s_runningStatus_for_list.load() == false) {
            return;
        }

        std::vector<QString> fileNameVec = {
            R"(C:\Users\TS\Desktop\test_1.log)",
            R"(C:\Users\TS\Desktop\test_2.log)",
            R"(C:\Users\TS\Desktop\test_3.log)",
            R"(C:\Users\TS\Desktop\test_4.log)",
            R"(C:\Users\TS\Desktop\test_5.log)",

            R"(C:\Users\TS\Desktop\test_6.log)",
            R"(C:\Users\TS\Desktop\test_7.log)",
            R"(C:\Users\TS\Desktop\test_8.log)",
            R"(C:\Users\TS\Desktop\test_9.log)",
            R"(C:\Users\TS\Desktop\test_10.log)"
        };

        const char *ipPort = "192.168.0.1:80";
        const char *clientID = "QmQ7obXFx1XMFr6hCYXtovn9zREFqSXEtH5hdtpBDLjrA1";
        uint64_t sentSize = 0;

        for (int index = 0; index < static_cast<int>(fileNameVec.size()); ++index) {
            QString filePath = fileNameVec[index];
            std::wstring currentFileName = filePath.toStdWString();
            uint32_t sentFilesCnt = index + 1;
            uint32_t totalFilesCnt = static_cast<uint32_t>(fileNameVec.size());
            uint64_t currentFileSize = 10 * 1024 * 1024;
            uint64_t totalSize = 100 * 1024 * 1024;
            uint64_t timeStamp = s_timeStamp_for_list;
            for (int i = 0; i < 10; ++i) {
                if (s_runningStatus_for_list.load() == false) {
                    break;
                }
                sentSize += 1024 * 1024;
                g_UpdateMultipleProgressBar(ipPort,
                                            clientID,
                                            currentFileName.c_str(),
                                            sentFilesCnt,
                                            totalFilesCnt,
                                            currentFileSize,
                                            totalSize,
                                            sentSize,
                                            timeStamp);

                QThread::msleep(100);
            }
        }
    }).detach();
}

void ControlWindow::on_MultiFilesDropNotify_btn_clicked()
{
    s_timeStamp_for_list = QDateTime::currentDateTime().toMSecsSinceEpoch();
    const char *ipPort = "192.168.0.1:80";
    const char *clientID = "QmQ7obXFx1XMFr6hCYXtovn9zREFqSXEtH5hdtpBDLjrA1";
    QString filePath(R"(C:\Users\TS\Desktop\test_data.mp4)");
    uint32_t cFileCount = 10;
    uint64_t totalSize = 1024 * 1024 * 100;
    uint64_t timeStamp = s_timeStamp_for_list;
    QString currentFileName = R"(C:\Users\TS\Desktop\test_1.log)";
    std::wstring currentFileName_wstring = currentFileName.toStdWString();
    uint64_t firstFileSize = 10 * 1024 * 1024;

    g_MultiFilesDropNotify(ipPort,
                         clientID,
                         cFileCount,
                         totalSize,
                         timeStamp,
                         currentFileName_wstring.c_str(),
                         firstFileSize);
}

void ControlWindow::on_SetCancelFileTransfer(const char* ipPort, const char* clientID, uint64_t timeStamp)
{
    qInfo() << "[SetCancelFileTransfer]\n"
            << "  IP: " << (ipPort ? ipPort : "NULL") << "\n"
            << "  ClientID: " << (clientID ? clientID : "NULL") << "\n"
            << "  Timestamp: " << timeStamp;

    s_runningStatus_for_list.store(false);

    uint32_t errorCode = GoErrorCode::ERR_BIZ_FD_DST_COPY_FILE_TIMEOUT;
    QByteArray ipPortString = ipPort;
    QByteArray timeStamp_string = QByteArray::number(timeStamp);
    s_runningStatus_for_list.store(false);
    g_NotifyErrEventCallback(clientID, errorCode, ipPortString.constData(), timeStamp_string.constData(), "", "");
}

void ControlWindow::on_UpdateClientVersion_btn_clicked()
{
    std::thread([] {
        g_RequestUpdateClientVersionCallback("1.2.3");
    }).detach();
}

void ControlWindow::on_NotifyErrorEvent_btn_clicked()
{
    QByteArray clientID = "QmQ7obXFx1XMFr6hCYXtovn9zREFqSXEtH5hdtpBDLjrA1";
    uint32_t errorCode = GoErrorCode::ERR_BIZ_FD_DST_COPY_FILE_TIMEOUT;
    QByteArray ipPortString = "192.168.0.1:80";
    QByteArray timeStamp = QByteArray::number(s_timeStamp_for_list);
    s_runningStatus_for_list.store(false);
    g_NotifyErrEventCallback(clientID.constData(), errorCode, ipPortString.constData(), timeStamp.constData(), "", "");
}

void ControlWindow::on_SendXClipData_btn_clicked()
{
    QByteArray textData("hello world 123456 !!!");
    QByteArray imageData = CommonUtils::getFileContent(":/resource/test.jpg").toBase64();
    QByteArray htmlData = CommonUtils::getFileContent(CommonUtils::downloadDirectoryPath() + "/test.html");
    g_SetupDstPasteXClipDataCallback(textData.constData(), imageData.constData(), htmlData.constData());
}

void ControlWindow::on_LoadClipboardData_btn_clicked()
{
    auto writeFileFunc = [] (const QString &fileName, const QByteArray &data) {
        QString filePath = CommonUtils::downloadDirectoryPath() + "/" + fileName;
        QFile file(filePath);
        if (file.open(QFile::WriteOnly)) {
            file.write(data);
            Q_EMIT CommonSignals::getInstance()->logMessage("write: " + filePath);
        }
    };
    const QMimeData *mimeData = qApp->clipboard()->mimeData();
    if (mimeData) {
        if (mimeData->hasText()) {
            writeFileFunc("test.txt", mimeData->text().toUtf8());
            qInfo() << "---------------------write test.txt";
        }

        if (mimeData->hasImage()) {
            QByteArray jpegData;
            {
                QImage image = qApp->clipboard()->image();
                if (image.hasAlphaChannel()) {
                    QImage opaque(image.size(), QImage::Format_RGB32);
                    opaque.fill(Qt::white);
                    QPainter painter(&opaque);
                    painter.drawImage(0, 0, image);
                    image = opaque;
                }

                image = image.convertToFormat(QImage::Format_RGB888);
                QBuffer buffer(&jpegData);
                buffer.open(QIODevice::WriteOnly);
                image.save(&buffer, "JPEG");
            }
            writeFileFunc("test.jpg", jpegData);
            qInfo() << "---------------------write test.jpg";
        }

        if (mimeData->hasHtml()) {
            writeFileFunc("test.html", mimeData->html().toUtf8());
            qInfo() << "---------------------write test.html";
        }
    }
}

void ControlWindow::writeClipboardData(const QByteArray &textData, const QByteArray &imageData, const QByteArray &htmlData)
{
    QTimer::singleShot(0, qApp, [textData, imageData, htmlData] {
        auto writeFileFunc = [] (const QString &fileName, const QByteArray &data) {
            QString filePath = CommonUtils::downloadDirectoryPath() + "/" + fileName;
            QFile file(filePath);
            if (file.open(QFile::WriteOnly)) {
                file.write(data);
                Q_EMIT CommonSignals::getInstance()->logMessage("write: " + filePath);
                qInfo() << "write:" << filePath.toUtf8().constData();
            }
        };

        if (!textData.isEmpty()) {
            writeFileFunc("recv_test.txt", textData);
        }

        if (!imageData.isEmpty()) {
            writeFileFunc("recv_test.jpg", QByteArray::fromBase64(imageData));
        }

        if (!htmlData.isEmpty()) {
            writeFileFunc("recv_test.html", htmlData);
        }
    });
}

// https://vendorjira.realtek.com/browse/TSTAS-384
void ControlWindow::on_UpdateClientStatusEx_btn_clicked()
{
    m_workerThread.runInThread([] {
        static const wchar_t *s_nameArry[] = { L"HDMI1-\nTEST", L"HDMI2", L"Miracast", L"USBC1", L"DP2" };
        Q_ASSERT(sizeof(s_nameArry) / sizeof (s_nameArry[0]) == 5);
        static const char *s_deviceType[] = { "HDMI1", "HDMI2", "Miracast", "USBC1", "DP2" };
        Q_ASSERT(sizeof(s_deviceType) / sizeof (s_deviceType[0]) == 5);
        std::string clientID { "QmQ7obXFx1XMFr6hCYXtovn9zREFqSXEtH5hdtpBDLjrAz" };
        static int s_index = 0;
        ++s_index;
        std::random_device currentRandomDevice;
        for (int index = 0; index < 5; ++index) {
            std::string newClientID = clientID;
            newClientID.back() = std::to_string(index + 1).back();
            std::string ip { "192.168.0.1" };
            uint8_t port = 80 + index;
            std::string ipPortString = ip + ":" + std::to_string(port);
            int arryIndex = std::uniform_int_distribution<int>(0, 4)(currentRandomDevice);
            if (s_index % 2 == 0) {
                nlohmann::json jsonInfo;
                jsonInfo["TimeStamp"] = QDateTime::currentDateTime().toMSecsSinceEpoch();
                jsonInfo["Status"] = 1;
                jsonInfo["ID"] = newClientID.c_str();
                jsonInfo["IpAddr"] = ipPortString.c_str();
                jsonInfo["Platform"] = "Windows";
                jsonInfo["DeviceName"] = s_nameArry[arryIndex];
                jsonInfo["SourcePortType"] = s_deviceType[arryIndex];
                jsonInfo["Version"] = "server_v1.2.3";
                g_UpdateClientStatusExCallback(jsonInfo.dump().c_str());
            } else {
                nlohmann::json jsonInfo;
                jsonInfo["TimeStamp"] = QDateTime::currentDateTime().toMSecsSinceEpoch();
                jsonInfo["Status"] = (index % 2 == 0) ? 0 : 1;
                jsonInfo["ID"] = newClientID.c_str();
                jsonInfo["IpAddr"] = ipPortString.c_str();
                jsonInfo["Platform"] = "Windows";
                jsonInfo["DeviceName"] = s_nameArry[arryIndex];
                jsonInfo["SourcePortType"] = s_deviceType[arryIndex];
                jsonInfo["Version"] = "server_v1.2.3";
                g_UpdateClientStatusExCallback(jsonInfo.dump().c_str());
            }
        }
    });
}

