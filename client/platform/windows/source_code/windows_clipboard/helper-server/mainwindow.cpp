#include "mainwindow.h"
#include "ui_mainwindow.h"
#include "event_filter_process.h"
#include "windows_event_monitor.h"
#include "load_plugin.h"
#include <QWindowStateChangeEvent>
#include <QProcess>
#include <QMenu>
#include <unordered_map>
#include <QScreen>
#include <QApplication>

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
        updateSystemTrayIcon();
        m_systemTrayIcon->setVisible(true);
        m_systemTrayIcon->setToolTip("CrossShare");
        connect(m_systemTrayIcon, &QSystemTrayIcon::activated, this, &MainWindow::onSystemTrayIconActivated);

        setWindowFlag(Qt::WindowType::WindowMinMaxButtonsHint, false);

        connect(QApplication::primaryScreen(), &QScreen::geometryChanged, [this] (const QRect &geometry) {
            qDebug() << "screen geometry:" << geometry.size();
            updateSystemTrayIcon();
        });

        connect(QApplication::primaryScreen(), &QScreen::logicalDotsPerInchChanged, [this] (qreal dpi) {
            qDebug() << "logicalDotsPerInchChanged:" << dpi;
            updateSystemTrayIcon();
        });
    }

    {
        connect(CommonSignals::getInstance(), &CommonSignals::dispatchMessage, this, &MainWindow::onDispatchMessage);
        connect(CommonSignals::getInstance(), &CommonSignals::logMessage, this, &MainWindow::onLogMessage);
        connect(LoadPlugin::getInstance(), &LoadPlugin::showWarningIcon, this, &MainWindow::onShowWarningIcon);
        connect(MonitorPlugEvent::getInstance(), &MonitorPlugEvent::statusChanged, this, &MainWindow::onDIASStatusChanged);
    }

    QTimer::singleShot(0, this, [] {
        Q_EMIT CommonSignals::getInstance()->systemConfigChanged();
    });

    QTimer::singleShot(50, Qt::TimerType::PreciseTimer, this, [] {
        MonitorPlugEvent::getInstance()->initData();
    });
    QTimer::singleShot(1000, Qt::TimerType::PreciseTimer, this, &MainWindow::processUITheme);
}

MainWindow::~MainWindow()
{
    delete ui;
}

void MainWindow::updateSystemTrayIcon()
{
    if (m_systemTrayIcon) {
        m_systemTrayIcon->setIcon(QIcon(":/resource/application.ico"));
    }
}

void MainWindow::initPlugin()
{
    QTimer::singleShot(0, this, [] {
        LoadPlugin::getInstance()->initPlugin();
    });
    updateSystemTrayIcon();
}

void MainWindow::processUITheme()
{
    qInfo() << "--------------------------processUITheme";

    bool isInited = false;
    const uint32_t oldCustomerID = g_getCustomerIDForUITheme(isInited);
    if (isInited == false) {
        return;
    }
    while (MonitorPlugEvent::getInstance()->getCacheMonitorData().macAddress.empty()) {
        MonitorPlugEvent::getInstance()->refreshCachedMonitorData();
        MonitorPlugEvent::delayInEventLoop(10);
    }
    uint32_t customerID = 0;
    MonitorPlugEvent::getCustomerThemeCode(MonitorPlugEvent::getInstance()->getCacheMonitorData().hPhysicalMonitor, customerID);
    if (customerID == oldCustomerID) {
        qInfo() << "The detection shows that the customerIDs are the same. customerID=" << customerID;
        initPlugin();
        return;
    }

    {
        qInfo() << "old customerID=" << oldCustomerID << "; new customerID=" << customerID;
        g_getGlobalData()->localConfig["UITheme"]["customerID"] = customerID;
        g_updateLocalConfig();
    }

    QString windowsClipboardExePath = qApp->applicationDirPath() + "/" + WINDOWS_CLIPBOARD_NAME;
#ifndef NDEBUG
    windowsClipboardExePath = qApp->applicationDirPath() + "/../src/" + WINDOWS_CLIPBOARD_NAME;
#endif
    Q_ASSERT(QFile::exists(windowsClipboardExePath) == true);
    if (CommonUtils::processIsRunning(WINDOWS_CLIPBOARD_NAME)) {
        qInfo() << "kill" << WINDOWS_CLIPBOARD_NAME << ";" << windowsClipboardExePath;
        CommonUtils::killWindowsClipboard();
        QTimer::singleShot(1000, this, [windowsClipboardExePath] {
            qInfo() << "start" << WINDOWS_CLIPBOARD_NAME << ";" << windowsClipboardExePath;
            CommonUtils::startDetachedWithoutInheritance(windowsClipboardExePath, {});
        });
    }
    initPlugin();
}

void MainWindow::onDIASStatusChanged(bool status)
{
    if (status) {
        QTimer::singleShot(100, Qt::TimerType::PreciseTimer, this, &MainWindow::processUITheme);
    }
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

    if (data.canConvert<UpdateDownloadPathMsgPtr>() == true) {
        UpdateDownloadPathMsgPtr ptr_msg = data.value<UpdateDownloadPathMsgPtr>();
        qInfo() << *ptr_msg;
        LoadPlugin::getInstance()->updateDownloadPath(ptr_msg->downloadPath);
        return;
    }

    if (data.canConvert<UpdateLocalConfigInfoMsgPtr>() == true) {
        UpdateLocalConfigInfoMsgPtr ptr_msg = data.value<UpdateLocalConfigInfoMsgPtr>();
        qInfo() << *ptr_msg;
        qInfo() << ptr_msg->appFilePath;
        try {
            g_getGlobalData()->localConfig = nlohmann::json::parse(ptr_msg->configData.constData());
            qInfo() << g_getGlobalData()->localConfig.dump(4).c_str();
            updateSystemTrayIcon();
        } catch (const std::exception &e) {
            qWarning() << e.what();
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
    Q_EMIT CommonSignals::getInstance()->quitAllEventLoop();
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
    if (reason == QSystemTrayIcon::ActivationReason::Trigger || reason == QSystemTrayIcon::ActivationReason::Context) {
        QString windowsClipboardExePath = qApp->applicationDirPath() + "/" + WINDOWS_CLIPBOARD_NAME;
#ifndef NDEBUG
        windowsClipboardExePath = qApp->applicationDirPath() + "/../src/" + WINDOWS_CLIPBOARD_NAME;
#endif
        Q_ASSERT(QFile::exists(windowsClipboardExePath));
        if (CommonUtils::processIsRunning(windowsClipboardExePath) == false) {
            CommonUtils::startDetachedWithoutInheritance(windowsClipboardExePath, {});
        } else {
            ShowWindowsClipboardMsg message;
            message.desc = "activateWindow";
            g_broadcastData(ShowWindowsClipboard_code, ShowWindowsClipboardMsg::toByteArray(message));
        }
    }
}

void MainWindow::onShowWarningIcon()
{
    m_systemTrayIcon->setIcon(QIcon(":/resource/warning.ico"));
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
