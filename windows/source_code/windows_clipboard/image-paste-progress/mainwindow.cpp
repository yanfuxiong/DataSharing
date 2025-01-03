#include "mainwindow.h"
#include "ui_mainwindow.h"
#include "event_filter_process.h"
#include <QWindowStateChangeEvent>
#include <QMenu>
#include <QCursor>

MainWindow::MainWindow(QWidget *parent)
    : QMainWindow(parent)
    , ui(new Ui::MainWindow)
    , m_testTimer(nullptr)
{
    ui->setupUi(this);

    {
        m_progressBarWidget = new ProgressBarWidget;
        ui->stackedWidget->addWidget(m_progressBarWidget);
        Q_ASSERT(ui->stackedWidget->indexOf(m_progressBarWidget) == 1);
        ui->stackedWidget->setCurrentIndex(1);
    }

    {
        connect(CommonSignals::getInstance(), &CommonSignals::dispatchMessage, this, &MainWindow::onDispatchMessage);
        connect(CommonSignals::getInstance(), &CommonSignals::logMessage, this, &MainWindow::onLogMessage);
    }

    connect(CommonSignals::getInstance(), &CommonSignals::pipeConnected, this, [] {
        nlohmann::json dataJson;
        dataJson["hash_id"] = g_hashIdValue.toHex().toUpper();
        Buffer buffer;
        buffer.append(QByteArray::fromStdString(dataJson.dump()));
        buffer.prependUInt32(buffer.readableBytes());

        Q_EMIT CommonSignals::getInstance()->sendDataToServer(buffer.retrieveAllAsByteArray());
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

        {
            uint64_t totalSize = ptr_msg->fileSize;
            uint64_t sentSize = ptr_msg->sentSize;
            int progressVal = static_cast<int>((sentSize / double(totalSize)) * 100);
            if (sentSize >= totalSize) {
                progressVal = 100;
            }

            RecordDataHash hashData;
            {
                hashData.fileSize = ptr_msg->fileSize;
                hashData.clientID = ptr_msg->clientID.toStdString();
                hashData.ip = ptr_msg->ip.toStdString();
            }
            Q_EMIT CommonSignals::getInstance()->updateProgressInfoWithID(progressVal, hashData.getHashID());
        }
        return;
    }
}

void MainWindow::closeEvent(QCloseEvent *event)
{
    QMainWindow::closeEvent(event);
    qApp->quit();
}


void MainWindow::mousePressEvent(QMouseEvent *event)
{
    m_clickedStatus = true;
    m_clickedPos = event->globalPos() - frameGeometry().topLeft();
    event->accept();
}

void MainWindow::mouseReleaseEvent(QMouseEvent *event)
{
    m_clickedStatus = false;
    event->accept();
}

void MainWindow::mouseMoveEvent(QMouseEvent *event)
{
    if (m_clickedStatus == false) {
        event->accept();
        return;
    }
    move(this->mapToGlobal(event->pos() - m_clickedPos));
}
