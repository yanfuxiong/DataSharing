#include "mainwindow.h"
#include "ui_mainwindow.h"
#include <QWindowStateChangeEvent>
#include <QMenu>
#include <QCursor>

MainWindow::MainWindow(QWidget *parent)
    : QMainWindow(parent)
    , ui(new Ui::MainWindow)
    , m_testTimer(nullptr)
    , m_moveHandler(std::make_unique<WindowMoveHandler>(this))
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
}

MainWindow::~MainWindow()
{
    delete ui;
}

void MainWindow::onLogMessage(const QString &message)
{
    Q_UNUSED(message)
}

void MainWindow::onDispatchMessage(const QVariant &data)
{
    if (data.canConvert<GetConnStatusResponseMsgPtr>() == true) {
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
        if (ptr_msg->timeStamp != g_timeStamp) {
            return;
        }

        {
            uint64_t totalSize = ptr_msg->fileSize;
            uint64_t sentSize = ptr_msg->sentSize;
            int progressVal = static_cast<int>((sentSize / double(totalSize)) * 100);
            if (sentSize >= totalSize) {
                progressVal = 100;
            }
            Q_EMIT CommonSignals::getInstance()->updateProgressInfoWithID(progressVal, {});
        }
        return;
    }
}

void MainWindow::closeEvent(QCloseEvent *event)
{
    QMainWindow::closeEvent(event);
    qApp->quit();
}
