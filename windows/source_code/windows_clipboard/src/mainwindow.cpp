#include "mainwindow.h"
#include "ui_mainwindow.h"
#include "drag_drop_widget.h"
#include "event_filter_process.h"

MainWindow::MainWindow(QWidget *parent)
    : QMainWindow(parent)
    , ui(new Ui::MainWindow)
    , m_testTimer(nullptr)
    , m_currentProgressVal(0)
{
    ui->setupUi(this);
    {
        ui->title_icon_1->clear();
        ui->title_icon_2->clear();
        ui->empty_label->clear();

        while (ui->drag_drop_area->count()) {
            ui->drag_drop_area->removeWidget(ui->drag_drop_area->widget(0));
        }
        ui->drag_drop_area->addWidget(new DragDropWidget);
//        if (STABLE_VERSION_CONTROL == 0) {
//            ui->record_title_label->hide();
//            ui->bottom_box->hide();
//            ui->verticalLayout->addStretch();
//        }
    }

    {
        connect(CommonSignals::getInstance(), &CommonSignals::dispatchMessage, this, &MainWindow::onDispatchMessage);
        connect(CommonSignals::getInstance(), &CommonSignals::logMessage, this, &MainWindow::onLogMessage);
        connect(CommonSignals::getInstance(), &CommonSignals::systemConfigChanged, this, &MainWindow::onSystemConfigChanged);

        connect(CommonSignals::getInstance(), &CommonSignals::pipeDisconnected, this, &MainWindow::onUpdateClientList);
        connect(CommonSignals::getInstance(), &CommonSignals::updateClientList, this, &MainWindow::onUpdateClientList);
        //connect(CommonSignals::getInstance(), &CommonSignals::userAcceptFile, this, &MainWindow::onUserAcceptFile);
    }

    {
        auto deviceList = new FileOptInfoList(this);
        ui->fileOptListWidget->addWidget(deviceList);
        ui->fileOptListWidget->setCurrentWidget(deviceList);
    }

    QTimer::singleShot(0, this, [] {
        Q_EMIT CommonSignals::getInstance()->systemConfigChanged();
    });

    // 处理部分事件过滤
    {
        EventFilterProcess::getInstance()->registerFilterEvent({ ui->title_icon_1, std::bind(&MainWindow::processTopTitleLeftClicked, this) });
    }
}

MainWindow::~MainWindow()
{
    delete ui;
}

void MainWindow::onLogMessage(const QString &message)
{
    Q_UNUSED(message)
    //ui->log_browser->append(QString("[%1]: %2").arg(QDateTime::currentDateTime().toString("yyyy-MM-dd hh:mm:ss.zzz")).arg(message));
}

// 只用测试
void MainWindow::startTestTimer()
{
    if (m_testTimer) {
        if (m_testTimer->isActive()) {
            m_testTimer->stop();
            m_testTimer->deleteLater();
        }
    }

    m_testTimer = new QTimer(this);
    m_currentProgressVal = 0;
    m_testTimer->setInterval(500);
    connect(m_testTimer, &QTimer::timeout, this, [this] {
        m_currentProgressVal += 5;
        if (m_currentProgressVal > 100) {
            m_currentProgressVal = 100;
        }
        Q_EMIT CommonSignals::getInstance()->updateProgressInfo(m_currentProgressVal);
        if (m_currentProgressVal >= 100) {
            m_testTimer->stop();
        }
    });
    m_testTimer->start();
}

void MainWindow::onDispatchMessage(const QVariant &data)
{
    if (data.canConvert<GetConnStatusResponseMsgPtr>() == true) {
        GetConnStatusResponseMsgPtr ptr_msg = data.value<GetConnStatusResponseMsgPtr>();
        if (ptr_msg->statusCode == 1) {
            Q_EMIT CommonSignals::getInstance()->showInfoMessageBox("connectionStatus", "connected to server.");
        } else {
            Q_EMIT CommonSignals::getInstance()->showWarningMessageBox("connectionStatus", "The server is in a disconnected state.");
        }
        return;
    }

    if (data.canConvert<UpdateClientStatusMsgPtr>() == true) {
        UpdateClientStatusMsgPtr ptr_msg = data.value<UpdateClientStatusMsgPtr>();

        auto &clientVec = g_getGlobalData()->m_clientVec;
        bool exists = false;
        for (auto itr = clientVec.begin(); itr != clientVec.end(); ++itr) {
            if ((*itr)->clientID == ptr_msg->clientID) {
                exists = true;
                if (ptr_msg->status == 0) {
                    clientVec.erase(itr);
                } else {
                    *itr = ptr_msg;
                }
                break;
            }
        }
        if (exists == false && ptr_msg->status == 1) {
            clientVec.push_back(ptr_msg);
        }
        Q_EMIT CommonSignals::getInstance()->updateClientList();
        return;
    }

    if (data.canConvert<SendFileRequestMsgPtr>() == true) {
        SendFileRequestMsgPtr ptr_msg = data.value<SendFileRequestMsgPtr>();
        AcceptFileDialog dialog;
        dialog.setFileInfo(ptr_msg);
        {
//            nlohmann::json infoJson;
//            infoJson["ip"] = ptr_msg->ip.toStdString();
//            infoJson["port"] = ptr_msg->port;
//            infoJson["fileSize"] = ptr_msg->fileSize;
//            infoJson["timeStamp"] = ptr_msg->timeStamp;
//            infoJson["fileName"] = ptr_msg->fileName.toStdString();
//            infoJson["clientID"] = ptr_msg->clientID.toStdString();
        }
        dialog.setWindowTitle("Do you want to receive files?");
        if (dialog.exec() == QDialog::Accepted) {
            QString newFileName = dialog.filePath();

            // 用户操作记录相关
            {
                UpdateClientStatusMsgPtr ptr_client;
                for (const auto &data : g_getGlobalData()->m_clientVec) {
                    if (data->clientID == ptr_msg->clientID) {
                        ptr_client = data;
                        break;
                    }
                }

                g_getGlobalData()->m_selectedClientVec.clear();
                g_getGlobalData()->m_selectedClientVec.push_back(ptr_client);

                FileOperationRecord record;
                record.fileName = newFileName.toStdString();
                record.fileSize = ptr_msg->fileSize;
                record.timeStamp = ptr_msg->timeStamp;
                record.progressValue = 0;
                record.clientName = ptr_client->clientName.toStdString();
                record.clientID = ptr_client->clientID.toStdString();
                record.ip = ptr_client->ip;
                record.direction = 1;

                g_getGlobalData()->cacheFileOptRecord.push_back(record);
                Q_EMIT CommonSignals::getInstance()->updateFileOptInfoList();
            }

            SendFileResponseMsg responseMessage;
            responseMessage.statusCode = 1; // 接受
            responseMessage.ip = ptr_msg->ip;
            responseMessage.port = ptr_msg->port;
            responseMessage.clientID = ptr_msg->clientID;
            responseMessage.fileSize = ptr_msg->fileSize;
            responseMessage.timeStamp = ptr_msg->timeStamp;
            responseMessage.fileName = newFileName;

            QByteArray data = SendFileResponseMsg::toByteArray(responseMessage);
            Q_EMIT CommonSignals::getInstance()->sendDataToServer(data);

            // 屏蔽用户的点击操作
            Q_EMIT CommonSignals::getInstance()->updateControlStatus(false);
        } else {
            SendFileResponseMsg responseMessage;
            responseMessage.statusCode = 0; // 拒绝
            responseMessage.ip = ptr_msg->ip;
            responseMessage.port = ptr_msg->port;
            responseMessage.clientID = ptr_msg->clientID;
            responseMessage.fileSize = ptr_msg->fileSize;
            responseMessage.timeStamp = ptr_msg->timeStamp;
            responseMessage.fileName = ptr_msg->fileName;

            QByteArray data = SendFileResponseMsg::toByteArray(responseMessage);
            Q_EMIT CommonSignals::getInstance()->sendDataToServer(data);
        }
        return;
    }

    if (data.canConvert<UpdateProgressMsgPtr>() == true) {
        UpdateProgressMsgPtr ptr_msg = data.value<UpdateProgressMsgPtr>();
        if (1)
        {
            nlohmann::json infoJson;
            infoJson["ip"] = ptr_msg->ip.toStdString();
            infoJson["port"] = ptr_msg->port;
            infoJson["clientID"] = ptr_msg->clientID.toStdString();
            infoJson["fileSize"] = ptr_msg->fileSize;
            infoJson["sentSize"] = ptr_msg->sentSize;
            infoJson["timeStamp"] = QDateTime::fromMSecsSinceEpoch(ptr_msg->timeStamp).toString("yyyy-MM-dd hh:mm:ss.zzz").toStdString();
            infoJson["fileName"] = ptr_msg->fileName.toStdString();
            Q_EMIT CommonSignals::getInstance()->logMessage(infoJson.dump(4).c_str());
        }

//        if (m_progressDialog == nullptr) {
//            m_progressDialog = new ProgressBarDialog;
//            m_progressDialog->setWindowTitle(QString("send file [%1]").arg(ptr_msg->fileName));
//            m_progressDialog->setModal(true);
//            m_progressDialog->show();
//        } else {
//            uint64_t totalSize = ptr_msg->fileSize;
//            uint64_t sentSize = ptr_msg->sentSize;
//            int progressVal = static_cast<int>((sentSize / double(totalSize)) * 100);
//            if (sentSize >= totalSize) {
//                progressVal = 100;
//            }
//            Q_EMIT CommonSignals::getInstance()->updateProgressInfo(progressVal);
//        }
        {
            uint64_t totalSize = ptr_msg->fileSize;
            uint64_t sentSize = ptr_msg->sentSize;
            int progressVal = static_cast<int>((sentSize / double(totalSize)) * 100);
            if (sentSize >= totalSize) {
                progressVal = 100;
            }

            RecordDataHash hashData;
            {
                hashData.fileName = ptr_msg->fileName.toStdString();
                hashData.fileSize = ptr_msg->fileSize;
                hashData.clientID = ptr_msg->clientID.toStdString();
                hashData.ip = ptr_msg->ip.toStdString();
            }
            Q_EMIT CommonSignals::getInstance()->updateProgressInfoWithID(progressVal, hashData.getHashID());
        }
        return;
    }
}

void MainWindow::onSystemConfigChanged()
{
    //const auto &config = g_getGlobalData()->systemConfig;
//    if (config.displayLogSwitch) {
//        ui->bottom_box->show();
//    } else {
//        ui->bottom_box->hide();
//    }
}

void MainWindow::on_settings_btn_clicked()
{
    auto &config = g_getGlobalData()->systemConfig;
    config.displayLogSwitch = !config.displayLogSwitch;
    Q_EMIT CommonSignals::getInstance()->systemConfigChanged();
}

void MainWindow::closeEvent(QCloseEvent *event)
{
    // FIXME: 暂时屏蔽
#if STABLE_VERSION_CONTROL > 0
    CommonUtils::killServer();
#endif
    QMainWindow::closeEvent(event);
}

//void MainWindow::on_conn_status_clicked()
//{
//    if (g_getGlobalData()->namedPipeConnected == false) {
//        Q_EMIT CommonSignals::getInstance()->showWarningMessageBox("connectionStatus", "The server is in a disconnected state.");
//        return;
//    }
//    GetConnStatusRequestMsg message;
//    QByteArray data = GetConnStatusRequestMsg::toByteArray(message);
//    Q_EMIT CommonSignals::getInstance()->sendDataToServer(data);
//}

void MainWindow::on_select_file_clicked()
{
//    UpdateClientStatusMsgPtr ptrClient = nullptr;
//    for (const auto &data : g_getGlobalData()->m_clientVec) {
//        if (data->clientID == getClientID()) {
//            ptrClient = data;
//            break;
//        }
//    }

//    if (ptrClient == nullptr) {
//        return;
//    }

    {
        g_getGlobalData()->selectedFileName.clear(); // 先清空

        QString fileName = QFileDialog::getOpenFileName(this, tr("Select File"),
                                                        CommonUtils::desktopDirectoryPath(),
                                                        ("Files (*.*)"));
        if (fileName.isEmpty()) {
            return;
        }
        qInfo() << "[FILE]:" << fileName;
        Q_EMIT CommonSignals::getInstance()->logMessage(QString("[SELECT]: %1").arg(fileName));
        g_getGlobalData()->selectedFileName = fileName; // 保存选中的文件名
    }

    {
        DeviceListDialog dialog;
        dialog.setWindowTitle("Select device");
        dialog.exec();
    }

//    {
//        SendFileRequestMsg message;
//        message.ip = ptrClient->ip;
//        message.port = ptrClient->port;
//        message.clientID = ptrClient->clientID;
//        message.fileSize = static_cast<uint64_t>(QFileInfo(g_getGlobalData()->selectedFileName).size());
//        message.timeStamp = QDateTime::currentDateTime().toUTC().toMSecsSinceEpoch();
//        message.fileName = g_getGlobalData()->selectedFileName;

//        QByteArray data = SendFileRequestMsg::toByteArray(message);
//        Q_EMIT CommonSignals::getInstance()->logMessage(QString("开始发送: %1").arg(message.fileName));
//        Q_EMIT CommonSignals::getInstance()->sendDataToServer(data);
//        Q_EMIT CommonSignals::getInstance()->updateControlStatus(false);
//    }
}

void MainWindow::processTopTitleLeftClicked()
{
    qInfo() << "----------------------标题栏icon点击......";
    //CommonUtils::setAutoRun(false);
}

void MainWindow::onUpdateClientList()
{
    QString infoText = QString("Online devices: %1").arg(g_getGlobalData()->m_clientVec.size());
    ui->online_devices_label->setText(infoText);
}
