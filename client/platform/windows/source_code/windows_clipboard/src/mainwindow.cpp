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
#include "menu_manager.h"
#include "worker_thread.h"
#include "navibar_widget.h"
#include "warning_dialog.h"
#include "common_proxy_style.h"
#include <windows.h>
#include <QWindowStateChangeEvent>
#include <QHBoxLayout>
#include <QVBoxLayout>
#include <QRegularExpression>

namespace {

int g_mainWindowTitleHeight = 0;
int g_topAreaHeight = 0;
int g_fileExplorerMinHeight = 460;

}

QPoint g_menuPos{0, 0};

MainWindow::MainWindow(QWidget *parent)
    : QMainWindow(parent)
    , ui(new Ui::MainWindow)
    , m_testTimer(nullptr)
    , m_currentProgressVal(0)
{
    ui->setupUi(this);
    g_mainWindow = this;
    m_lastNormalState = windowState();

    {
        {
            QHBoxLayout *pHBoxLayout = new QHBoxLayout;
            pHBoxLayout->setMargin(0);
            pHBoxLayout->setSpacing(0);
            ui->top_box->setLayout(pHBoxLayout);
            pHBoxLayout->addWidget(new NaviBarWidget);
        }
        ui->clear_all_icon->clear();
        ui->clear_all_icon->setCursor(Qt::PointingHandCursor);
        ui->status_info_label->clear();
        ui->system_info_label->clear();
        //ui->title_label->clear();

        {
            ui->record_title_label->setProperty(PR_ADJUST_WINDOW_X_SIZE, true);
            constexpr int boxWidth = 96;
            ui->tag_box_1->setFixedWidth(boxWidth);
            ui->tag_box_2->setFixedWidth(boxWidth);
            ui->tag_box_3->setFixedWidth(boxWidth);
            ui->tag_box_4->setFixedWidth(boxWidth);
            ui->left_middle_h_spacer->changeSize(boxWidth, -1);
        }

        while (ui->drag_drop_area->count()) {
            ui->drag_drop_area->removeWidget(ui->drag_drop_area->widget(0));
        }
        ui->drag_drop_area->addWidget(new DragDropWidget);
    }

    {
        ui->middle_stacked_widget->setMaximumHeight(g_fileExplorerMinHeight);
        ui->middle_stacked_widget->setSizePolicy(QSizePolicy::Expanding, QSizePolicy::Expanding);

        ui->tag_box_1->setMaximumHeight(g_fileExplorerMinHeight);
        ui->tag_box_1->setSizePolicy(QSizePolicy::Expanding, QSizePolicy::Expanding);

        ui->tag_box_2->setMaximumHeight(g_fileExplorerMinHeight);
        ui->tag_box_2->setSizePolicy(QSizePolicy::Expanding, QSizePolicy::Expanding);

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
            {
                QFrame *naviParent = new QFrame;
                naviParent->setObjectName("NaviWindowParent");
                naviParent->setMaximumWidth(220);
                naviParent->setSizePolicy(QSizePolicy::Expanding, QSizePolicy::Expanding);
                QVBoxLayout *pVBoxLayout = new QVBoxLayout;
                pVBoxLayout->setMargin(0);
                pVBoxLayout->setSpacing(0);
                naviParent->setLayout(pVBoxLayout);
                pVBoxLayout->addSpacing(10);
                pVBoxLayout->addWidget(fileExplorer->createNaviWindow());

                pHBoxLayout->addWidget(naviParent);
            }

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
        //connect(CommonSignals::getInstance(), &CommonSignals::systemConfigChanged, this, &MainWindow::onSystemConfigChanged);
        //connect(CommonSignals::getInstance(), &CommonSignals::updateClientList, this, &MainWindow::onSystemConfigChanged);

        connect(CommonSignals::getInstance(), &CommonSignals::updateCurrentMenuPosData, this, &MainWindow::onUpdateCurrentMenPosData);
        connect(CommonSignals::getInstance(), &CommonSignals::showOpenSourceLicenseMenu, this, &MainWindow::onShowOpenSourceLicenseMenu);

        connect(CommonSignals::getInstance(), &CommonSignals::modifyDefaulutDownloadPath, this, &MainWindow::onModifyDefaulutDownloadPath);
        connect(CommonSignals::getInstance(), &CommonSignals::bugReport, this, &MainWindow::onBugReport);
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
        EventFilterProcess::getInstance()->registerFilterEvent({ ui->clear_all_icon, std::bind(&MainWindow::clearAllUserOptRecord, this) });
    }
}

MainWindow::~MainWindow()
{
    delete ui;
    g_mainWindow = nullptr;
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
        qDebug() << g_getGlobalData()->systemConfig;
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

    if (data.canConvert<UpdateClientVersionMsgPtr>() == true) {
        UpdateClientVersionMsgPtr ptr_msg = data.value<UpdateClientVersionMsgPtr>();
        qInfo() << "client version:" << ptr_msg->clientVersion;
        static QRegularExpression s_regExp(R"((\d+\.\d+\.\d+))");
        WarningDialog dialog(this);
        QString clientVersion = ptr_msg->clientVersion;
        QString goServerVersion = g_getGlobalData()->systemConfig.serverVersionStr;
        auto matchRet = s_regExp.match(goServerVersion);
        if (matchRet.hasMatch()) {
            clientVersion = matchRet.captured(1);
        }
        dialog.updateWarningInfo(clientVersion);
        dialog.exec();
        CommonUtils::killServer();
        QTimer::singleShot(0, this, [] {
            qApp->exit();
        });
        return;
    }

    if (data.canConvert<ShowWindowsClipboardMsgPtr>() == true) {
        ShowWindowsClipboardMsgPtr ptr_msg = data.value<ShowWindowsClipboardMsgPtr>();
        qInfo() << "ShowWindowsClipboardMsg:" << ptr_msg->desc;
        auto windowStateValue = m_lastNormalState;
        if (windowStateValue.testFlag(Qt::WindowState::WindowMaximized)) {
            showMaximized();
        } else {
            showNormal();
        }
        activateWindow();
        raise();

        // Send a simulated event to ensure that the window can be activated.
        {
            INPUT input;
            std::memset(&input, 0, sizeof (input));
            input.type = INPUT_MOUSE;
            input.mi.dx = 0;
            input.mi.dy = 0;
            input.mi.dwFlags = MOUSEEVENTF_MOVE;
            SendInput(1, &input, sizeof(INPUT));
        }
        return;
    }

    if (data.canConvert<NotifyErrorEventMsgPtr>() == true) {
        NotifyErrorEventMsgPtr ptr_msg = data.value<NotifyErrorEventMsgPtr>();
        qInfo() << "NotifyErrorEventMsg:" << ptr_msg->ipPortString << ";" << ptr_msg->timeStamp;
        qInfo() << *ptr_msg;
        qInfo() << "Error:" << g_goErrorCodeToString(ptr_msg->errorCode);

        RecordDataHash hashData;
        {
            hashData.clientID = ptr_msg->clientID.toStdString();
            int pos = ptr_msg->ipPortString.indexOf(':');
            hashData.ip = ptr_msg->ipPortString.left(pos).toStdString();
            hashData.timeStamp = ptr_msg->timeStamp.toULongLong();
        }
        Q_EMIT CommonSignals::getInstance()->updateOptRecordStatus(hashData.getHashID(), FileOperationRecord::TransferFileErrorStatus);
        return;
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
    Q_EMIT CommonSignals::getInstance()->quitAllEventLoop();
    QMainWindow::closeEvent(event);
}

void MainWindow::changeEvent(QEvent *event)
{
    if (event->type() == QEvent::WindowStateChange) {
        QWindowStateChangeEvent *stateEvent = static_cast<QWindowStateChangeEvent*>(event);
        if (windowState().testFlag(Qt::WindowState::WindowMinimized)) {
            m_lastNormalState = stateEvent->oldState();
        } else {
            m_lastNormalState = windowState();
        }
    }
    QMainWindow::changeEvent(event);
}

void MainWindow::moveEvent(QMoveEvent *event)
{
    QMainWindow::moveEvent(event);
}

void MainWindow::resizeEvent(QResizeEvent *event)
{
    QMainWindow::resizeEvent(event);
}

void MainWindow::showEvent(QShowEvent *event)
{
    QMainWindow::showEvent(event);
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

void MainWindow::onShowOpenSourceLicenseMenu()
{
    QTimer::singleShot(0, qApp, [] {
        SystemInfoMenuManager manager;
        QMenu *menu = manager.createOpenSourceLicenseMenu();
        SystemInfoMenuManager::updateMenuPos(menu);
        menu->exec(g_menuPos);
        menu->deleteLater();
    });
}

void MainWindow::onModifyDefaulutDownloadPath()
{
    QString oldDownloadPath;
    try {
        oldDownloadPath = g_getGlobalData()->localConfig.at("crossShareServer").at("downloadPath").get<std::string>().c_str();
    } catch (const std::exception &e) {
        qWarning() << e.what();
        return;
    }
    Q_ASSERT(QFile::exists(oldDownloadPath) == true);
    QString newDownloadPath = QFileDialog::getExistingDirectory(this, "download path",
                                                    oldDownloadPath,
                                                    QFileDialog::ShowDirsOnly | QFileDialog::DontResolveSymlinks);
    if (newDownloadPath.isEmpty()) {
        return;
    }
    qInfo() << "oldDownloadPath:" << oldDownloadPath << "; newDownloadPath:" << newDownloadPath;
    g_getGlobalData()->localConfig["crossShareServer"]["downloadPath"] = newDownloadPath.toStdString();

    {
        UpdateDownloadPathMsg message;
        message.downloadPath = newDownloadPath;
        g_sendDataToServer(UpdateDownloadPath_code, UpdateDownloadPathMsg::toByteArray(message));
    }
}

void MainWindow::onBugReport()
{
    static WorkerThread s_workThread;

    s_workThread.runInThread([] {
        static int s_index = 0;
        QString downloadPath;
        try {
            downloadPath = g_getGlobalData()->localConfig.at("crossShareServer").at("downloadPath").get<std::string>().c_str();
        } catch (const std::exception &e) {
            qWarning() << e.what();
        }
        if (QFile::exists(downloadPath) == false) {
            downloadPath = CommonUtils::downloadDirectoryPath();
        }

        QString fileName = QString::asprintf("CrossShareLog_%s_%d_%02d.tar.gz",
                                             QDateTime::currentDateTime().toString("yyyyMMdd_hhmmss").toStdString().c_str(),
                                             static_cast<int>(::GetCurrentProcessId()),
                                             ++s_index);
        fileName = downloadPath + "/" + fileName;
        CommonUtils::compressFolderToTarGz(CommonUtils::windowsLogFolderPath(), fileName);
        Q_EMIT CommonSignals::getInstance()->showInfoMessageBox("Bug Report", "Export report to download path.");
    });
}

void MainWindow::clearAllUserOptRecord()
{
    g_getGlobalData()->cacheFileOptRecord.clear();
    Q_EMIT CommonSignals::getInstance()->updateFileOptInfoList();
}

void MainWindow::onUpdateCurrentMenPosData()
{
    g_menuPos = ui->centralwidget->mapToGlobal(QPoint(ui->centralwidget->width(), ui->top_box->height()));
}
