#include "mainwindow.h"
#include "ui_mainwindow.h"
#include "event_filter_process.h"
#include "windows_event_monitor.h"
#include <QWindowStateChangeEvent>
#include <QProcess>
#include <QMenu>
#include <unordered_map>

namespace {

struct StatusTipsWindowInfo
{
    bool windowIsVisible { false };
};

std::unordered_map<uint64_t, StatusTipsWindowInfo> g_stausInfoMap;

std::list<qint64> g_processIdList;

static BOOL CALLBACK EnumWindowsProc(HWND hwnd, LPARAM lParam)
{
    std::list<HWND> &windowHandleList = *reinterpret_cast<std::list<HWND>*>(lParam);
    DWORD processId = 0;
    ::GetWindowThreadProcessId(hwnd, &processId);

    for (const auto &idData : g_processIdList) {
        if (idData == processId) {
            windowHandleList.push_back(hwnd);
            break;
        }
    }

    return TRUE;
}

std::list<HWND> getWindowsByProcessId()
{
    if (g_processIdList.empty()) {
        return {};
    }
    std::list<HWND> windowHandleList;
    ::EnumWindows(EnumWindowsProc, reinterpret_cast<LPARAM>(&windowHandleList));
    return windowHandleList;
}

}

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
}

void MainWindow::onDispatchMessage(const QVariant &data)
{
    if (data.canConvert<GetConnStatusResponseMsgPtr>() == true) {
        return;
    }

    if (data.canConvert<NotifyMessagePtr>() == true) {
        NotifyMessagePtr ptr_msg = data.value<NotifyMessagePtr>();
        uint64_t timeStamp = ptr_msg->timeStamp;
        qInfo() << "recv NotifyMessage:" << "timeStamp=" << timeStamp << "; notiCode=" << ptr_msg->notiCode;
        auto itr = g_stausInfoMap.find(timeStamp);
        if (itr == g_stausInfoMap.end() || itr->second.windowIsVisible == false) {
            g_stausInfoMap[ptr_msg->timeStamp] = { true };
            processNotifyMessage(ptr_msg);
        } else {
            m_workerThread.runInThread([ptr_msg] {
                for (const auto &hwnd : getWindowsByProcessId()) {
                    QByteArray sendData = NotifyMessage::toByteArray(*ptr_msg).toHex().toUpper();
                    COPYDATASTRUCT cds;
                    cds.dwData = UPDATE_STATUS_TIPS_MSG_TAG;
                    cds.cbData = sendData.length();
                    cds.lpData = sendData.data();
                    ::SendMessage(hwnd, WM_COPYDATA, 0, reinterpret_cast<LPARAM>(&cds));
                }
            });
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
    qApp->exit();
}

void MainWindow::changeEvent(QEvent *event)
{
    if (event->type() == QEvent::WindowStateChange) {
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

int MainWindow::getNotifyMessageDuration() const
{
    try {
        return g_getGlobalData()->localConfig.at("crossShareServer").at("notifyMessageDuration").get<int>();
    } catch (const std::exception &e) {
        qWarning() << e.what();
        return 3000; // 3000ms
    }
}

UpdateClientStatusMsgPtr MainWindow::getClientStatusMsgByClientID(const QByteArray &clientID) const
{
    UpdateClientStatusMsgPtr ptr_client = nullptr;
    for (const auto &data : g_getGlobalData()->m_clientVec) {
        if (data->clientID == clientID) {
            ptr_client = data;
            break;
        }
    }
    return ptr_client;
}

void MainWindow::processNotifyMessage(NotifyMessagePtr ptrMsg)
{
    qInfo() << ptrMsg->toString().dump(4).c_str();
    QProcess process;
    for (const auto &exePath : statusTipsExePathList()) {
        if (QFile::exists(exePath)) {
            ++m_processIndex;
            ++m_processCount;
            qint64 pid = 0;
            process.startDetached(exePath,
                                  { NotifyMessage::toByteArray(*ptrMsg).toHex().toUpper(), QString::number(m_processIndex) },
                                  QString(),
                                  &pid);
            Q_ASSERT(pid != 0);
            g_processIdList.push_back(pid);
            uint64_t timeStamp = ptrMsg->timeStamp;
            QTimer::singleShot(getNotifyMessageDuration(), Qt::TimerType::PreciseTimer, this, [this, pid, timeStamp] {
                --m_processCount;
                if (m_processCount <= 0) {
                    m_processIndex = 0;
                }
                QProcess process;
                process.setProcessChannelMode(QProcess::ProcessChannelMode::MergedChannels);
                process.startDetached(QString("taskkill /PID %1").arg(pid));
                Q_ASSERT(g_processIdList.empty() == false && g_processIdList.front() == pid);
                g_processIdList.pop_front();

                {
                    auto itr = g_stausInfoMap.find(timeStamp);
                    if (itr != g_stausInfoMap.end()) {
                        itr->second.windowIsVisible = false;
                    }
                }

                m_workerThread.runInThread([] {
                    for (const auto &hwnd : getWindowsByProcessId()) {
                        COPYDATASTRUCT cds;
                        cds.dwData = UPDATE_WINDOW_POS_TAG;
                        cds.cbData = 0;
                        cds.lpData = NULL;
                        ::SendMessage(hwnd, WM_COPYDATA, 0, reinterpret_cast<LPARAM>(&cds));
                    }
                });
            });
            break;
        }
    }
}
