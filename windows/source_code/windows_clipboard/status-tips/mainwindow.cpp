#include "mainwindow.h"
#include "ui_mainwindow.h"
#include "event_filter_process.h"
#include <QWindowStateChangeEvent>
#include <QMenu>
#include <QScreen>
#include <Windows.h>

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

    updateStatusInfo();

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

void MainWindow::setProcessIndex(int indexVal)
{
    m_processIndex = indexVal;
}

void MainWindow::updateWindowPos()
{
    int x_pos = qApp->primaryScreen()->availableSize().width() - width();
    int y_pos = qApp->primaryScreen()->availableSize().height();
    y_pos -= (height() + 4) * m_processIndex;
    move(x_pos, y_pos);
}

void MainWindow::onLogMessage(const QString &message)
{
    Q_UNUSED(message)
    //ui->log_browser->append(QString("[%1]: %2").arg(QDateTime::currentDateTime().toString("yyyy-MM-dd hh:mm:ss.zzz")).arg(message));
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
}

void MainWindow::closeEvent(QCloseEvent *event)
{
    Q_UNUSED(event)
    qApp->quit();
}

bool MainWindow::nativeEvent(const QByteArray &eventType, void *message, long *result)
{
    Q_UNUSED(eventType)
    Q_UNUSED(result)
    MSG *msg = reinterpret_cast<MSG*>(message);
    if (msg->message == WM_COPYDATA) {
        Q_ASSERT(msg->hwnd == reinterpret_cast<HWND>(winId()));
        COPYDATASTRUCT *cds = reinterpret_cast<COPYDATASTRUCT*>(msg->lParam);
        switch (cds->dwData) {
        case UPDATE_WINDOW_POS_TAG: {
            if (m_processIndex >= 2) {
                --m_processIndex;
                QTimer::singleShot(0, this, [this] {
                    updateWindowPos();
                });
            }
            break;
        }
        case UPDATE_STATUS_TIPS_MSG_TAG: {
            QByteArray data(reinterpret_cast<const char*>(cds->lpData), cds->cbData);
            NotifyMessage notifyMessage;
            NotifyMessage::fromByteArray(QByteArray::fromHex(data), notifyMessage);
            if (notifyMessage.timeStamp != g_notifyMessage.timeStamp) {
                break;
            }
            g_notifyMessage = std::move(notifyMessage);
            QTimer::singleShot(0, this, [this] {
                updateStatusInfo();
            });
            break;
        }
        default: {
            break;
        }
        }
        return true;
    }
    return false;
}

void MainWindow::updateStatusInfo()
{
    try {
        nlohmann::json infoJson = g_notifyMessage.toString();
        ui->status_title_label->setText(infoJson.at("title").get<std::string>().c_str());
        ui->status_content_label->setText(infoJson.at("content").get<std::string>().c_str());
        repaint();
    } catch (const std::exception &e) {
        qWarning() << e.what();
    }

}
