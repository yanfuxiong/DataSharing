#include "mainwindow.h"
#include "ui_mainwindow.h"
#include "event_filter_process.h"
#include <QWindowStateChangeEvent>
#include <QMenu>

MainWindow::MainWindow(QWidget *parent)
    : QMainWindow(parent)
    , ui(new Ui::MainWindow)
    , m_testTimer(nullptr)
{
    ui->setupUi(this);

    {
        ui->title_icon_2->clear();
        ui->close_label->clear();

        ui->status_title_label->clear();
        ui->status_content_label->clear();
    }

    try {
        nlohmann::json infoJson = g_notifyMessage.toString();
        ui->status_title_label->setText(infoJson.at("title").get<std::string>().c_str());
        ui->status_content_label->setText(infoJson.at("content").get<std::string>().c_str());
    } catch (const std::exception &e) {
        qWarning() << e.what();
    }

    {
        connect(CommonSignals::getInstance(), &CommonSignals::dispatchMessage, this, &MainWindow::onDispatchMessage);
        connect(CommonSignals::getInstance(), &CommonSignals::logMessage, this, &MainWindow::onLogMessage);
    }

    EventFilterProcess::getInstance()->registerFilterEvent({ ui->close_label, std::bind(&MainWindow::close, this) });
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
}

void MainWindow::closeEvent(QCloseEvent *event)
{
    qInfo() << "------------------close";
    QMainWindow::closeEvent(event);
    qApp->quit();
}
