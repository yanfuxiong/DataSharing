#include "mainwindow.h"
#include "ui_mainwindow.h"
#include "event_filter_process.h"
#include <QWindowStateChangeEvent>
#include <QProcess>
#include <QMenu>

MainWindow::MainWindow(QWidget *parent)
    : QMainWindow(parent)
    , ui(new Ui::MainWindow)
    , m_testTimer(nullptr)
{
    ui->setupUi(this);

    {
        m_systemTrayIcon = new QSystemTrayIcon(this);
        m_systemTrayIcon->setIcon(QIcon(":/resource/application.ico"));
        m_systemTrayIcon->setVisible(true);
        connect(m_systemTrayIcon, &QSystemTrayIcon::activated, this, &MainWindow::onSystemTrayIconActivated);

        setWindowFlag(Qt::WindowType::WindowMinMaxButtonsHint, false);
    }

    {
        connect(CommonSignals::getInstance(), &CommonSignals::dispatchMessage, this, &MainWindow::onDispatchMessage);
        connect(CommonSignals::getInstance(), &CommonSignals::logMessage, this, &MainWindow::onLogMessage);
    }

    QTimer::singleShot(0, this, [] {
        Q_EMIT CommonSignals::getInstance()->systemConfigChanged();
    });
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
//        GetConnStatusResponseMsgPtr ptr_msg = data.value<GetConnStatusResponseMsgPtr>();
//        if (ptr_msg->statusCode == 1) {
//            Q_EMIT CommonSignals::getInstance()->showInfoMessageBox("connectionStatus", "connected to server.");
//        } else {
//            Q_EMIT CommonSignals::getInstance()->showWarningMessageBox("connectionStatus", "The server is in a disconnected state.");
//        }
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

    if (data.canConvert<UpdateImageProgressMsgPtr>() == true) {
        UpdateImageProgressMsgPtr ptr_msg = data.value<UpdateImageProgressMsgPtr>();
        if (1)
        {
            nlohmann::json infoJson;
            infoJson["ip"] = ptr_msg->ip.toStdString();
            infoJson["port"] = ptr_msg->port;
            infoJson["clientID"] = ptr_msg->clientID.toStdString();
            infoJson["fileSize"] = ptr_msg->fileSize;
            infoJson["sentSize"] = ptr_msg->sentSize;
            infoJson["timeStamp"] = QDateTime::fromMSecsSinceEpoch(ptr_msg->timeStamp).toString("yyyy-MM-dd hh:mm:ss.zzz").toStdString();
            Q_EMIT CommonSignals::getInstance()->logMessage(infoJson.dump(4).c_str());
        }

        Q_EMIT CommonSignals::getInstance()->updateProgressInfoWithMsg(data);

//        {
//            uint64_t totalSize = ptr_msg->fileSize;
//            uint64_t sentSize = ptr_msg->sentSize;
//            int progressVal = static_cast<int>((sentSize / double(totalSize)) * 100);
//            if (sentSize >= totalSize) {
//                progressVal = 100;
//            }

//            RecordDataHash hashData;
//            {
//                hashData.fileSize = ptr_msg->fileSize;
//                hashData.clientID = ptr_msg->clientID.toStdString();
//                hashData.ip = ptr_msg->ip.toStdString();
//            }
//            Q_EMIT CommonSignals::getInstance()->updateProgressInfoWithID(progressVal, hashData.getHashID());
//        }
        return;
    }

    if (data.canConvert<NotifyMessagePtr>() == true) {
        NotifyMessagePtr ptr_msg = data.value<NotifyMessagePtr>();
        qInfo() << ptr_msg->toString().dump(4).c_str();
        QProcess process;
        for (const auto &exePath : statusTipsExePathList()) {
            if (QFile::exists(exePath)) {
                process.startDetached(exePath, { NotifyMessage::toByteArray(*ptr_msg).toHex().toUpper() });
                break;
            }
        }
        return;
    }
}

void MainWindow::closeEvent(QCloseEvent *event)
{
    if (m_exitsStatus == false) {
        hide();
        event->ignore();
        return;
    }
    QMainWindow::closeEvent(event);
}

void MainWindow::changeEvent(QEvent *event)
{
    if (event->type() == QEvent::WindowStateChange) {
        //QWindowStateChangeEvent *stateEvent = static_cast<QWindowStateChangeEvent*>(event);
        if (windowState().testFlag(Qt::WindowState::WindowMinimized)) {
            if (m_systemTrayIcon) {
                hide();
                event->accept();
                return;
            }
        }
    }
    return QMainWindow::changeEvent(event);
}

void MainWindow::onSystemTrayIconActivated(QSystemTrayIcon::ActivationReason reason)
{
    if (reason == QSystemTrayIcon::ActivationReason::Trigger) {
#ifndef NDEBUG
        QTimer::singleShot(0, this, [this] {
            showNormal();
            activateWindow();
        });
#endif
    } else if (reason == QSystemTrayIcon::ActivationReason::Context) {
        QMenu menu;
        menu.addAction(QIcon(":/resource/exit.svg"), "   Exit   ", [this] {
            m_exitsStatus = true;
            close();
        });
        menu.exec(QCursor::pos());
    }
}

QStringList MainWindow::statusTipsExePathList() const
{
    QStringList pathList;
    pathList << qApp->applicationDirPath() + "/status-tips.exe";
    pathList << qApp->applicationDirPath() + "/../status-tips/status-tips.exe";
    return pathList;
}
