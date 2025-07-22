#include "mainwindow.h"
#include "ui_mainwindow.h"
#include "drag_drop_widget.h"
#include "event_filter_process.h"
#include "tips_list_area.h"
#include "file_explorer.h"
#include "common_utils.h"
#include "device_list_dialog.h"
#include "common_signals.h"
#include "accept_file_dialog.h"
#include "file_opt_info_list.h"
#include <QWindowStateChangeEvent>
#include <QHBoxLayout>
#include <QVBoxLayout>

namespace {

int g_mainWindowTitleHeight = 0;
int g_topAreaHeight = 0;
int g_fileExplorerMinHeight = 460;

}

MainWindow::MainWindow(QWidget *parent)
    : QMainWindow(parent)
    , ui(new Ui::MainWindow)
    , m_testTimer(nullptr)
    , m_currentProgressVal(0)
{
    ui->setupUi(this);
    if (0)
    {
        m_systemTrayIcon = new QSystemTrayIcon(this);
        m_systemTrayIcon->setIcon(QIcon(":/resource/application.ico"));
        m_systemTrayIcon->setVisible(true);
        connect(m_systemTrayIcon, &QSystemTrayIcon::activated, this, &MainWindow::onSystemTrayIconActivated);
        //setWindowFlag(Qt::WindowStaysOnTopHint, true);
    }

    {
        ui->title_icon_1->clear();
        ui->title_icon_2->clear();
        ui->clear_all_icon->clear();
        ui->clear_all_icon->setCursor(Qt::PointingHandCursor);
        ui->empty_label->clear();
        ui->status_info_label->clear();
        //ui->title_label->clear();

        while (ui->drag_drop_area->count()) {
            ui->drag_drop_area->removeWidget(ui->drag_drop_area->widget(0));
        }
        ui->drag_drop_area->addWidget(new DragDropWidget);
    }

    {
        ui->middle_stacked_widget->setFixedHeight(g_fileExplorerMinHeight);
        ui->tag_box_1->setFixedHeight(g_fileExplorerMinHeight);
        ui->tag_box_2->setFixedHeight(g_fileExplorerMinHeight);

        QTimer::singleShot(0, this, [this] {
            g_mainWindowTitleHeight = frameGeometry().height() - ui->centralwidget->height();
            g_topAreaHeight = ui->top_box->height();
            qInfo() << "g_mainWindowTitleHeight:" << g_mainWindowTitleHeight << "; g_topAreaHeight:" << g_topAreaHeight;
        });
    }

    {
        while (ui->tips_stacked_widget->count()) {
            ui->tips_stacked_widget->removeWidget(ui->tips_stacked_widget->widget(0));
        }

        ui->tips_stacked_widget->addWidget(new TipsListArea);
        ui->tips_stacked_widget->setCurrentIndex(0);
    }

    {
        QGroupBox *pParentBox = new QGroupBox;
        pParentBox->setObjectName("middle_box");
        {
            QHBoxLayout *pHBoxLayout = new QHBoxLayout;
            pHBoxLayout->setMargin(0);
            pHBoxLayout->setSpacing(0);
            pParentBox->setLayout(pHBoxLayout);

            auto fileExplorer = new FileExplorer;
            pHBoxLayout->addWidget(fileExplorer->createNaviWindow());
            pHBoxLayout->addWidget(fileExplorer);
        }

        {
            QHBoxLayout *pHBoxLayout = new QHBoxLayout;
            pHBoxLayout->setMargin(0);
            pHBoxLayout->setSpacing(0);
            ui->file_explorer_page->setLayout(pHBoxLayout);
            pHBoxLayout->addWidget(pParentBox);
        }
        ui->middle_stacked_widget->setCurrentIndex(1);
    }

    {
        connect(CommonSignals::getInstance(), &CommonSignals::dispatchMessage, this, &MainWindow::onDispatchMessage);
        connect(CommonSignals::getInstance(), &CommonSignals::logMessage, this, &MainWindow::onLogMessage);
        connect(CommonSignals::getInstance(), &CommonSignals::systemConfigChanged, this, &MainWindow::onSystemConfigChanged);
        connect(CommonSignals::getInstance(), &CommonSignals::updateClientList, this, &MainWindow::onSystemConfigChanged);

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

        g_loadDataFromSqliteDB();
        Q_EMIT CommonSignals::getInstance()->updateFileOptInfoList();
    });

    {
        EventFilterProcess::getInstance()->registerFilterEvent({ ui->title_icon_1, std::bind(&MainWindow::processTopTitleLeftClicked, this) });
        EventFilterProcess::getInstance()->registerFilterEvent({ ui->clear_all_icon, std::bind(&MainWindow::clearAllUserOptRecord, this) });
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

    if (data.canConvert<UpdateSystemInfoMsgPtr>() == true) {
        UpdateSystemInfoMsgPtr ptr_msg = data.value<UpdateSystemInfoMsgPtr>();

        g_getGlobalData()->systemConfig.localIpAddress = ptr_msg->ip;
        g_getGlobalData()->systemConfig.port = ptr_msg->port;
        g_getGlobalData()->systemConfig.serverVersionStr = ptr_msg->serverVersion;

        Q_EMIT CommonSignals::getInstance()->systemConfigChanged();
        return;
    }

    if (data.canConvert<SendFileRequestMsgPtr>() == true) {
        SendFileRequestMsgPtr ptr_msg = data.value<SendFileRequestMsgPtr>();
        if (ptr_msg->flag == SendFileRequestMsg::DragFlag) {
            {
                UpdateClientStatusMsgPtr ptr_client = nullptr;
                for (const auto &data : g_getGlobalData()->m_clientVec) {
                    if (data->clientID == ptr_msg->clientID) {
                        ptr_client = data;
                        break;
                    }
                }

                if (ptr_client == nullptr) {
                    return;
                }
                g_getGlobalData()->m_selectedClientVec.clear();
                g_getGlobalData()->m_selectedClientVec.push_back(ptr_client);

                FileOperationRecord record;
                record.fileName = ptr_msg->fileName.toStdString();
                record.fileSize = ptr_msg->fileSize;
                record.timeStamp = ptr_msg->timeStamp;
                record.progressValue = 0;
                record.clientName = ptr_client->clientName.toStdString();
                record.clientID = ptr_client->clientID.toStdString();
                record.ip = ptr_client->ip;
                record.port = ptr_client->port;
                record.direction = FileOperationRecord::DragSingleFileType;

                record.uuid = CommonUtils::createUuid();

                g_getGlobalData()->cacheFileOptRecord.push_back(record);
                Q_EMIT CommonSignals::getInstance()->updateFileOptInfoList();
            }
            return;
        }
        AcceptFileDialog dialog;
        dialog.setFileInfo(ptr_msg);
        dialog.setWindowTitle("Do you want to receive files?");
        //Prevent the dialog from popping up
        if (/*dialog.exec() == QDialog::Accepted*/true) {
            QString newFileName = dialog.filePath();

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
                record.port = ptr_client->port;
                record.direction = 1;

                record.uuid = CommonUtils::createUuid();

                g_getGlobalData()->cacheFileOptRecord.push_back(record);
                Q_EMIT CommonSignals::getInstance()->updateFileOptInfoList();
            }

            SendFileResponseMsg responseMessage;
            responseMessage.statusCode = 1; // accept
            responseMessage.ip = ptr_msg->ip;
            responseMessage.port = ptr_msg->port;
            responseMessage.clientID = ptr_msg->clientID;
            responseMessage.fileSize = ptr_msg->fileSize;
            responseMessage.timeStamp = ptr_msg->timeStamp;
            responseMessage.fileName = newFileName;

            QByteArray data = SendFileResponseMsg::toByteArray(responseMessage);
            Q_EMIT CommonSignals::getInstance()->sendDataToServer(data);

            // Block user clicks
            Q_EMIT CommonSignals::getInstance()->updateControlStatus(false);
        } else {
            SendFileResponseMsg responseMessage;
            responseMessage.statusCode = 0; // reject
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

        if (ptr_msg->functionCode == UpdateProgressMsg::FuncCode::SingleFile) {
            uint64_t totalSize = ptr_msg->fileSize;
            uint64_t sentSize = ptr_msg->sentSize;
            int progressVal = static_cast<int>((sentSize / double(totalSize)) * 100);
            if (sentSize >= totalSize) {
                progressVal = 100;
            }

            RecordDataHash hashData;
            {
                hashData.clientID = ptr_msg->clientID.toStdString();
                hashData.ip = ptr_msg->ip.toStdString();
                hashData.timeStamp = ptr_msg->timeStamp;
            }
            Q_EMIT CommonSignals::getInstance()->updateProgressInfoWithID(progressVal, hashData.getHashID());
        } else if (ptr_msg->functionCode == UpdateProgressMsg::FuncCode::MultiFile) {
            uint64_t totalSize = ptr_msg->totalFilesSize;
            uint64_t sentSize = ptr_msg->totalSentSize;
            int progressVal = static_cast<int>((sentSize / double(totalSize)) * 100);
            if (sentSize >= totalSize) {
                progressVal = 100;
            }

            RecordDataHash hashData;
            {
                hashData.clientID = ptr_msg->clientID.toStdString();
                hashData.ip = ptr_msg->ip.toStdString();
                hashData.timeStamp = ptr_msg->timeStamp;
            }
            FileOptInfoList::updateCacheFileOptRecord(hashData.getHashID(), ptr_msg);
            Q_EMIT CommonSignals::getInstance()->updateProgressInfoWithID(progressVal, hashData.getHashID());
        }
        return;
    }

    if (data.canConvert<StatusInfoNotifyMsgPtr>() == true) {
        StatusInfoNotifyMsgPtr ptr_msg = data.value<StatusInfoNotifyMsgPtr>();
        qInfo() << "status_code:" << ptr_msg->statusCode;
        QString message;
        switch (ptr_msg->statusCode) {
        case 1: {
            message = "Wait for connecting to DIAS monitor...";
            break;
        }
        case 2: {
            message = "Searching DIAS service in LAN...";
            break;
        }
        case 3: {
            message = "Checking authorization...";
            break;
        }
        case 4: {
            message = "Wait for screen casting...";
            break;
        }
        case 5: {
            message = "Failed! Authorization not available";
            break;
        }
        case 6: {
            message = "Connected. Wait for other clients";
            break;
        }
        case 7: {
            message = "Connected";
            break;
        }
        default: {
            break;
        }
        }

        ui->status_info_label->setText(message);
    }

    if (data.canConvert<DragFilesMsgPtr>() == true) {
        DragFilesMsgPtr ptr_msg = data.value<DragFilesMsgPtr>();
        Q_ASSERT(ptr_msg->functionCode == DragFilesMsg::FuncCode::ReceiveFileInfo);
        qInfo() << "totalFileSize:" << CommonUtils::byteCountDisplay(ptr_msg->totalFileSize)
                << "; firstTransferFileName:" << ptr_msg->firstTransferFileName.toUtf8().constData()
                << "; firstTransferFileSize:" << CommonUtils::byteCountDisplay(ptr_msg->firstTransferFileSize);

        {
            UpdateClientStatusMsgPtr ptr_client = nullptr;
            for (const auto &data : g_getGlobalData()->m_clientVec) {
                if (data->clientID == ptr_msg->clientID) {
                    ptr_client = data;
                    break;
                }
            }

            if (ptr_client == nullptr) {
                return;
            }
            g_getGlobalData()->m_selectedClientVec.clear();
            g_getGlobalData()->m_selectedClientVec.push_back(ptr_client);

            FileOperationRecord record;
            record.clientName = ptr_client->clientName.toStdString();
            record.clientID = ptr_client->clientID.toStdString();
            record.ip = ptr_client->ip;
            record.port = ptr_client->port;
            record.timeStamp = ptr_msg->timeStamp;
            record.progressValue = 0;
            record.direction = FileOperationRecord::DragMultiFileType;
            record.uuid = CommonUtils::createUuid();

            record.totalFileCount = ptr_msg->fileCount;
            record.totalFileSize = ptr_msg->totalFileSize;
            record.currentTransferFileName = ptr_msg->firstTransferFileName;
            record.currentTransferFileSize = ptr_msg->firstTransferFileSize;

            if (record.totalFileCount <= 1) {
                record.direction = FileOperationRecord::DragSingleFileType;
                record.fileName = ptr_msg->firstTransferFileName.toStdString();
                record.fileSize = ptr_msg->firstTransferFileSize;
            }

            {
                record.cacheFileName = ptr_msg->firstTransferFileName;
                record.cacheFileSize = ptr_msg->firstTransferFileSize;
            }

            g_getGlobalData()->cacheFileOptRecord.push_back(record);
            Q_EMIT CommonSignals::getInstance()->updateFileOptInfoList();
        }
    }
}

void MainWindow::onSystemConfigChanged()
{
    QString info;
    if (g_getGlobalData()->namedPipeConnected) {
        info += g_getGlobalData()->systemConfig.localIpAddress;
        info += "\n";
        info += g_getGlobalData()->systemConfig.serverVersionStr;
        info += "\n";
        info += g_getGlobalData()->systemConfig.clientVersionStr;
    } else {
        info += g_getGlobalData()->systemConfig.clientVersionStr;
    }
    ui->system_info_label->setText(info);
}

void MainWindow::closeEvent(QCloseEvent *event)
{
#if STABLE_VERSION_CONTROL > 0
    CommonUtils::killServer();
#endif
    g_updateCacheFileOptRecord();
    g_saveDataToSqliteDB();
    QMainWindow::closeEvent(event);
}

void MainWindow::changeEvent(QEvent *event)
{
    if (event->type() == QEvent::WindowStateChange) {
        //QWindowStateChangeEvent *stateEvent = static_cast<QWindowStateChangeEvent*>(event);
        if (windowState().testFlag(Qt::WindowState::WindowMinimized)) {
            if (m_systemTrayIcon) {
                hide();
                return;
            }
        }
    }
    return QMainWindow::changeEvent(event);
}

void MainWindow::on_select_file_clicked()
{
    {
        g_getGlobalData()->selectedFileName.clear();

        QString fileName = QFileDialog::getOpenFileName(this, tr("Select File"),
                                                        CommonUtils::desktopDirectoryPath(),
                                                        ("Files (*.*)"));
        if (fileName.isEmpty()) {
            return;
        }
        qInfo() << "[FILE]:" << fileName;
        Q_EMIT CommonSignals::getInstance()->logMessage(QString("[SELECT]: %1").arg(fileName));
        g_getGlobalData()->selectedFileName = fileName; // Save the selected file name
    }

    {
        DeviceListDialog dialog;
        dialog.setWindowTitle("Select device");
        dialog.exec();
    }
}

void MainWindow::processTopTitleLeftClicked()
{
    qInfo() << "----------------------clicked title icon......";
    //CommonUtils::setAutoRun(false);
}

void MainWindow::clearAllUserOptRecord()
{
    g_getGlobalData()->cacheFileOptRecord.clear();
    Q_EMIT CommonSignals::getInstance()->updateFileOptInfoList();
}

void MainWindow::onUpdateClientList()
{
    QString infoText = QString("Online devices: %1").arg(g_getGlobalData()->m_clientVec.size());
    ui->online_devices_label->setText(infoText);
}

void MainWindow::onSystemTrayIconActivated(QSystemTrayIcon::ActivationReason reason)
{
    if (reason == QSystemTrayIcon::ActivationReason::Trigger) {
        QTimer::singleShot(0, this, [this] {
            showNormal();
            activateWindow();
        });
    }
}
