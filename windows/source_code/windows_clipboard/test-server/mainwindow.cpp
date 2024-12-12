#include "mainwindow.h"
#include "ui_mainwindow.h"
#include "common_signals.h"
#include "progress_object.h"
#include <QTimer>
#include <QEventLoop>

MainWindow::MainWindow(QWidget *parent) :
    QMainWindow(parent),
    ui(new Ui::MainWindow)
{
    ui->setupUi(this);

    {
        connect(&m_server, &NamedPipeServer::recvData, this, &MainWindow::onRecvClientData);
        connect(CommonSignals::getInstance(), &CommonSignals::logMessage, this, &MainWindow::onLogMessage);
    }

    QTimer::singleShot(0, this, [this] {
        m_server.startServer("CrossSharePipe");
        Q_EMIT CommonSignals::getInstance()->logMessage("----------------start server");
    });
}

MainWindow::~MainWindow()
{
    delete ui;
}

void MainWindow::on_add_client_clicked()
{
    Q_EMIT CommonSignals::getInstance()->addTestClient();
}

void MainWindow::onLogMessage(const QString &message)
{
    ui->log_browser->append(QString("[%1]: %2").arg(QDateTime::currentDateTime().toString("yyyy-MM-dd hh:mm:ss.zzz")).arg(message));
}

void MainWindow::onRecvClientData(const QByteArray &data)
{
    Q_EMIT CommonSignals::getInstance()->logMessage(data.toHex().toUpper().constData());
    m_buffer.append(data);

    uint8_t typeValue = 0;
    uint8_t code = 0;
    // 解析消息
    while (g_getCodeFromByteArray(QByteArray(m_buffer.peek(), m_buffer.readableBytes()), typeValue, code)) {
        switch (code) {
        case 1: {
            GetConnStatusRequestMsg message;
            if (GetConnStatusRequestMsg::fromByteArray(QByteArray(m_buffer.peek(), m_buffer.readableBytes()), message)) {
                m_buffer.retrieve(message.getMessageLength());
                {
                    nlohmann::json infoJson;
                    infoJson["message"] = "GetConnStatusRequestMsg";
                    infoJson["type"] = message.headerInfo.type;
                    infoJson["code"] = message.headerInfo.code;
                    infoJson["contentLength"] = message.headerInfo.contentLength;
                    Q_EMIT CommonSignals::getInstance()->logMessage(infoJson.dump(4).c_str());
                }

                GetConnStatusResponseMsg responseMessage;
                responseMessage.statusCode = 1;

                Q_EMIT CommonSignals::getInstance()->sendDataForTestServer(GetConnStatusResponseMsg::toByteArray(responseMessage));
            }
            break;
        }
        case 2: {
            Q_ASSERT(false);
            break;
        }
        case 3: {
            Q_ASSERT(false);
            break;
        }
        case 4: {
            if (typeValue == PipeMessageType::Response) {
                SendFileResponseMsg message;
                if (SendFileResponseMsg::fromByteArray(QByteArray(m_buffer.peek(), m_buffer.readableBytes()), message)) {
                    m_buffer.retrieve(message.getMessageLength());

                    {
                        nlohmann::json infoJson;
                        infoJson["message"] = "SendFileResponseMsg";
                        infoJson["fileName"] = message.fileName.toStdString();
                        infoJson["clientID"] = message.clientID.toStdString();
                        infoJson["ip"] = message.ip.toStdString();
                        infoJson["port"] = message.port;
                        infoJson["statusCode"] = message.statusCode;
                        infoJson["desc"] = (message.statusCode == 0 ? "拒绝" : "接受");
                        Q_EMIT CommonSignals::getInstance()->logMessage(infoJson.dump(4).c_str());
                    }

                    if (message.statusCode == 0) {
                        break; // 拒绝了就不需要处理进度信息了, 直接 break
                    }

                    {
                        auto pProgressMsg = new UpdateProgressMsg;
                        pProgressMsg->ip = message.ip;
                        pProgressMsg->port = message.port;
                        pProgressMsg->clientID = message.clientID;
                        pProgressMsg->fileSize = 0; // 先初始化为0
                        pProgressMsg->timeStamp = message.timeStamp;
                        pProgressMsg->fileName = message.fileName;

                        //m_cacheProgressMsgPtr.reset(pProgressMsg);
                        auto ptr_object = new ProgressObject(pProgressMsg);
                        Q_UNUSED(ptr_object)
                    }
                }
            } else {
                Q_ASSERT(typeValue == PipeMessageType::Request);
                SendFileRequestMsg message;
                if (SendFileRequestMsg::fromByteArray(QByteArray(m_buffer.peek(), m_buffer.readableBytes()), message)) {
                    m_buffer.retrieve(message.getMessageLength());
                    {
                        nlohmann::json infoJson;
                        infoJson["message"] = "SendFileRequestMsg";
                        infoJson["fileName"] = message.fileName.toStdString();
                        infoJson["clientID"] = message.clientID.toStdString();
                        infoJson["ip"] = message.ip.toStdString();
                        infoJson["port"] = message.port;
                        Q_EMIT CommonSignals::getInstance()->logMessage(infoJson.dump(4).c_str());
                    }

                    {
                        auto pProgressMsg = new UpdateProgressMsg;
                        pProgressMsg->ip = message.ip;
                        pProgressMsg->port = message.port;
                        pProgressMsg->clientID = message.clientID;
                        pProgressMsg->fileSize = 0; // 先初始化为0
                        pProgressMsg->timeStamp = message.timeStamp;
                        pProgressMsg->fileName = message.fileName;

                        //m_cacheProgressMsgPtr.reset(pProgressMsg);
                        auto ptr_object = new ProgressObject(pProgressMsg);
                        Q_UNUSED(ptr_object)
                    }
                }
            }
            break;
        }
        case 5: {
            Q_ASSERT(false);
            break;
        }
        default: {
            Q_ASSERT(false);
            break;
        }
        }
    }
}

void MainWindow::on_send_file_clicked()
{
    SendFileRequestMsg msg;
    msg.ip = "192.168.30.1";
    msg.port = 12345;
    msg.clientID = "QmQ7obXFx1XMFr6hCYXtovn9zREFqSXEtH5hdtpBDLjrAz";
    //msg.fileSize = static_cast<uint64_t>(QFileInfo(__FILE__).size());
    //msg.fileSize = 60727169;
    msg.timeStamp = QDateTime::currentDateTime().toUTC().toMSecsSinceEpoch();
    msg.fileName = R"(C:\Users\TS\Desktop\test_data.mp4)";
    msg.fileSize = QFileInfo(msg.fileName).size();

    {
        nlohmann::json infoJson;
        infoJson["message"] = "SendFileRequestMsg";
        infoJson["ip"] = msg.ip.toStdString();
        infoJson["port"] = msg.port;
        infoJson["clientID"] = msg.clientID.toStdString();
        infoJson["fileSize"] = msg.fileSize;
        infoJson["timeStamp"] = msg.timeStamp;
        infoJson["fileName"] = msg.fileName.toUtf8().constData();
        Q_EMIT CommonSignals::getInstance()->logMessage(infoJson.dump(4).c_str());
    }

    Q_EMIT CommonSignals::getInstance()->sendDataForTestServer(SendFileRequestMsg::toByteArray(msg));
}

void MainWindow::on_disconnect_client_clicked()
{
    QByteArray clientStatusMsgData;
    {
        UpdateClientStatusMsg msg;
        msg.status = 0;
        msg.ip = "192.168.30.1";
        msg.port = 12345;
        msg.clientID = "QmQ7obXFx1XMFr6hCYXtovn9zREFqSXEtH5hdtpBDLjrAz";
        msg.clientName = QString("测试电脑_%1").arg(1);

        clientStatusMsgData = UpdateClientStatusMsg::toByteArray(msg);
    }

    Q_EMIT CommonSignals::getInstance()->sendDataForTestServer(clientStatusMsgData);
}
