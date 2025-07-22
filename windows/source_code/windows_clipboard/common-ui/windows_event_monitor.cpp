#include "windows_event_monitor.h"
#include "common_signals.h"
#include "common_utils.h"
#include <physicalmonitorenumerationapi.h>
#include <highlevelmonitorconfigurationapi.h>
#include <lowlevelmonitorconfigurationapi.h>
#include "ddcci/crossShareVcpCtrl.h"
#include <physicalmonitorenumerationapi.h>
#include <highlevelmonitorconfigurationapi.h>
#include <lowlevelmonitorconfigurationapi.h>
#include <QElapsedTimer>

namespace {

QPoint g_pressPos;
bool g_dragging = false;
std::atomic<bool> g_keyStatus { false };

void sendEsc()
{
    keybd_event(VK_ESCAPE, 0, 0, 0); // down
    keybd_event(VK_ESCAPE, 0, KEYEVENTF_KEYUP, 0); // up
}

// FIXME: Due to the probability of failure, a number of repeated attempts is set here
const int g_vcpRetryCnt = 5;

}

WindowsEventMonitor *WindowsEventMonitor::m_instance = nullptr;

WindowsEventMonitor::WindowsEventMonitor()
    : QObject(nullptr)
{
    {
        auto handle = SetWindowsHookEx(WH_MOUSE_LL, hookProc_mouse, NULL, 0);
        m_hookVec.push_back(handle);
    }

    // {
    //     auto handle = SetWindowsHookEx(WH_KEYBOARD_LL, hookProc_keyboard, NULL, 0);
    //     m_hookVec.push_back(handle);
    // }
}

WindowsEventMonitor::~WindowsEventMonitor()
{
    for (const auto &handle : m_hookVec) {
        UnhookWindowsHookEx(handle);
    }
}

WindowsEventMonitor *WindowsEventMonitor::getInstance()
{
    if (m_instance == nullptr) {
        m_instance = new WindowsEventMonitor;
    }
    return m_instance;
}

LRESULT CALLBACK WindowsEventMonitor::hookProc_mouse(int nCode, WPARAM wParam, LPARAM lParam) {
    if (nCode >= 0) {
        PMSLLHOOKSTRUCT pMouse = reinterpret_cast<PMSLLHOOKSTRUCT>(lParam);

        switch (wParam) {
        case WM_MBUTTONDOWN: {
            g_pressPos = QPoint(pMouse->pt.x, pMouse->pt.y);
            g_dragging = false;
            Q_EMIT WindowsEventMonitor::getInstance()->clickedPos(g_pressPos);
            break;
        }
        case WM_MOUSEMOVE: {
            if (GetAsyncKeyState(VK_LBUTTON) & 0x8000) { // Hold down the left button
                QPoint currentPos(pMouse->pt.x, pMouse->pt.y);
                if (!g_dragging && (currentPos - g_pressPos).manhattanLength() > 3) {
                    g_dragging = true;
                    if (g_keyStatus.load()) {
                        Q_EMIT CommonSignals::getInstance()->userSelectedFiles();
                    }
                }
            }
            break;
        }
        case WM_MBUTTONUP: {
            if (g_dragging) {
                g_dragging = false;
                if (g_keyStatus.load()) {
                    QTimer::singleShot(0, QThread::currentThread(), [] {
                        sendEsc();
                    });
                    return 1;
                }
            }
            break;
        }
        default: {
            break;
        }
        }
    }
    return CallNextHookEx(NULL, nCode, wParam, lParam);
}

LRESULT WindowsEventMonitor::hookProc_keyboard(int nCode, WPARAM wParam, LPARAM lParam)
{
    if (nCode >= 0) {
        KBDLLHOOKSTRUCT *pKey = reinterpret_cast<KBDLLHOOKSTRUCT*>(lParam);
        if (pKey->vkCode == VK_F9) {
            if (wParam == WM_KEYUP) {
                if (g_keyStatus.load() == true) {
                    g_keyStatus.store(false);
                }
            } else if (wParam == WM_KEYDOWN) {
                if (g_keyStatus.load() == false) {
                    g_keyStatus.store(true);
                }
            }
        }
    }
    return CallNextHookEx(NULL, nCode, wParam, lParam);
}

QStringList WindowsEventMonitor::getSelectedPathList()
{
    QStringList selectedPaths;

    QAxObject shellApp("Shell.Application");
    if (shellApp.isNull()) {
        qWarning() << "Failed to create Shell.Application object";
        return selectedPaths;
    }

    QAxObject *shellWindows = shellApp.querySubObject("Windows()");
    Q_ASSERT(shellWindows->parent() == &shellApp);
    if (!shellWindows) {
        qWarning() << "Failed to get shell windows collection";
        return selectedPaths;
    }

    HWND foregroundWindow = GetForegroundWindow();
    int windowCount = shellWindows->property("Count").toInt();
    for (int i = 0; i < windowCount; ++i) {
        QAxObject *window = shellWindows->querySubObject("Item(QVariant)", QVariant(i));
        if (!window) continue;

        qlonglong hwnd = window->property("HWND").toLongLong();
        if (reinterpret_cast<HWND>(hwnd) != foregroundWindow) {
            delete window;
            continue;
        }

        QAxObject *document = window->querySubObject("Document");
        if (!document) {
            delete window;
            continue;
        }

        QAxObject *selection = document->querySubObject("SelectedItems()");
        if (!selection) {
            delete document;
            delete window;
            continue;
        }

        int itemCount = selection->property("Count").toInt();
        for (int j = 0; j < itemCount; ++j) {
            QAxObject *item = selection->querySubObject("Item(QVariant)", QVariant(j));
            if (item) {
                QString path = item->property("Path").toString();
                selectedPaths.append(path);
                delete item;
            }
        }

        delete selection;
        delete document;
        delete window;

        break;
    }

    delete shellWindows;
    return selectedPaths;
}

// ---------------------------------- MonitorPlugEvent

MonitorPlugEvent *MonitorPlugEvent::m_instance = nullptr;

MonitorPlugEvent::MonitorPlugEvent()
{
    setVisible(false);
    registerDeviceNotification();
}

MonitorPlugEvent::~MonitorPlugEvent()
{
    unregisterDeviceNotification();
    m_instance = nullptr;
}

MonitorPlugEvent *MonitorPlugEvent::getInstance()
{
    if (m_instance == nullptr) {
        m_instance = new MonitorPlugEvent;
    }
    return m_instance;
}

void MonitorPlugEvent::initDataImpl()
{
    const int vcpRetryCnt = 1;
    for (int i = 0; i < vcpRetryCnt; ++i) {
        m_monitorDataList.clear();
        getCurrentAllMonitroData(m_monitorDataList);
        if (m_monitorDataList.isEmpty() == true) {
            break;
        }

        m_cacheMonitorData = m_monitorDataList.front();
        if (m_cacheMonitorData.macAddress.empty()) {
            continue;
        }

        for (const auto &itemData : m_monitorDataList) {
            qInfo() << "macAddress(hex):" << QByteArray::fromStdString(itemData.macAddress).toHex().toUpper().constData()
            << "; monitor desc:" << itemData.desc.toUtf8().constData();
        }
        Q_EMIT statusChanged(true);
        break;
    }
}

void MonitorPlugEvent::stopProcessDDCCI()
{
    m_currentTimeout = 0;
}

void MonitorPlugEvent::restartProcessDDCCI()
{
    m_currentTimeout = DDCCI_TIMEOUT;
}

void MonitorPlugEvent::initData()
{
    static bool s_initStatus = false;
    if (s_initStatus == false) {
        s_initStatus = true;
        std::thread(std::bind(&MonitorPlugEvent::processDDCCI, this)).detach();
        return;
    }
    initDataImpl();
    if (getCacheMonitorData().macAddress.empty() == false) {
        stopProcessDDCCI();
        return;
    }
}

void MonitorPlugEvent::processDDCCI()
{
    m_currentTimeout = DDCCI_TIMEOUT;
    const int interval = 2000; // 2s
    while (true) {
        if (m_currentTimeout <= 0) {
            QThread::msleep(500); // 500ms
            continue;
        }
        QTimer::singleShot(0, this, [this] {
            initData();
        });
        m_currentTimeout -= interval;
        QThread::msleep(interval);
    }
}

void MonitorPlugEvent::clearData()
{
    m_cacheMonitorData = MonitorData();
}

bool MonitorPlugEvent::registerDeviceNotification()
{
    DEV_BROADCAST_DEVICEINTERFACE dbdi{};
    dbdi.dbcc_size = sizeof(DEV_BROADCAST_DEVICEINTERFACE);
    dbdi.dbcc_devicetype = DBT_DEVTYP_DEVICEINTERFACE;
    dbdi.dbcc_classguid = GUID_DEVCLASS_MONITOR;

    m_hDevNotify = RegisterDeviceNotification((HWND)winId(), &dbdi, DEVICE_NOTIFY_WINDOW_HANDLE | DEVICE_NOTIFY_ALL_INTERFACE_CLASSES);

    if (!m_hDevNotify) {
        qWarning() << "RegisterDeviceNotification failed. Error:" << GetLastError();
    }
    return m_hDevNotify != nullptr;
}

void MonitorPlugEvent::unregisterDeviceNotification()
{
    if (m_hDevNotify) {
        UnregisterDeviceNotification(m_hDevNotify);
        m_hDevNotify = nullptr;
    }
}

bool MonitorPlugEvent::nativeEvent(const QByteArray &, void *msg, long *)
{
    MSG *message = reinterpret_cast<MSG*>(msg);
    if (message->message == WM_DEVICECHANGE) {
        //qInfo() << "Received WM_DEVICECHANGE, wParam:" << message->wParam;
        DEV_BROADCAST_HDR *pHdr = reinterpret_cast<DEV_BROADCAST_HDR*>(message->lParam);
        if (!pHdr) {
            return false;
        }

        //qInfo() << "Device change type:" << pHdr->dbch_devicetype;

        if (message->wParam != DBT_DEVICEARRIVAL && message->wParam != DBT_DEVICEREMOVECOMPLETE) {
            return false;
        }

        if (pHdr->dbch_devicetype == DBT_DEVTYP_DEVICEINTERFACE) {
            auto pDevInf = reinterpret_cast<DEV_BROADCAST_DEVICEINTERFACE_W*>(pHdr);
            QString devPath = QString::fromWCharArray(pDevInf->dbcc_name).toLower();
            //debugDeviceInfo(pHdr->dbch_devicetype, devPath);

            if (devPath.contains("display") || devPath.contains("monitor")) {
                if (message->wParam == DBT_DEVICEARRIVAL) {
                    //qInfo() << "monitor Connected:" << devPath;
                    static std::atomic<bool> s_processingStatus { false };
                    if (s_processingStatus.load() == false && getCacheMonitorData().isDIAS.load() == false) {
                        s_processingStatus.store(true);
                        std::thread([this] {
                            QElapsedTimer timer;
                            timer.start();
                            while (timer.elapsed() <= 3000) { // Exit after timeout of 3000ms
                                QList<MonitorData> newDataList;
                                if (m_currentTimeout <= 0) {
                                    getCurrentAllMonitroData(newDataList);
                                }
                                if (!newDataList.empty()) {
                                    std::atomic<bool> doneStatus { false };
                                    QTimer::singleShot(0, this, [this, &doneStatus] {
                                        updateMonitorDataList();
                                        Q_EMIT statusChanged(true);
                                        doneStatus.store(true);
                                        stopProcessDDCCI();
                                    });
                                    while (doneStatus.load() == false) {
                                        QThread::msleep(10);
                                    }
                                    break;
                                }
                                if (m_currentTimeout <= 0) {
                                    restartProcessDDCCI();
                                }
                                QThread::msleep(10); // Avoid rapid repeated testing
                            }
                            s_processingStatus.store(false);
                        }).detach();
                    }
                } else {
                    //qInfo() << "monitor Disconnected:" << devPath;
                    static std::atomic<bool> s_processingStatus { false };
                    if (s_processingStatus.load() == false && getCacheMonitorData().isDIAS.load() == true) {
                        s_processingStatus.store(true);
                        std::thread([this] {
                            QElapsedTimer timer;
                            timer.start();
                            while (timer.elapsed() <= 3000) { // Exit after timeout of 3000ms
                                QList<MonitorData> newDataList;
                                getCurrentAllMonitroData(newDataList);
                                if (newDataList.empty()) {
                                    std::atomic<bool> doneStatus { false };
                                    QTimer::singleShot(0, this, [this, &doneStatus] {
                                        updateMonitorDataList();
                                        Q_EMIT statusChanged(false);
                                        doneStatus.store(true);
                                        stopProcessDDCCI();
                                    });
                                    while (doneStatus.load() == false) {
                                        QThread::msleep(10);
                                    }
                                    break;
                                }
                                QThread::msleep(10); // Avoid rapid repeated testing
                            }
                            s_processingStatus.store(false);
                        }).detach();
                    } else if (getCacheMonitorData().isDIAS.load() == false) {
                        stopProcessDDCCI();
                    }
                }
            }
        }
    }
    return false;
}

void MonitorPlugEvent::updateMonitorDataList()
{
    QList<MonitorData> newDataList;
    getCurrentAllMonitroData(newDataList);
    if (!newDataList.empty()) {
        m_cacheMonitorData = newDataList.front();
        m_monitorDataList = newDataList;
    }
}

void MonitorPlugEvent::debugDeviceInfo(DWORD dwDevType, const QString &devPath)
{
    qInfo() << "[Device Event] Type:" << dwDevType
            << "\nPath:" << devPath
            << "\n----------------------------------";
}

BOOL MonitorPlugEvent::monitorEnumProc(HMONITOR hMonitor, HDC hdcMonitor, LPRECT lprcMonitor, LPARAM dwData)
{
    Q_UNUSED(hdcMonitor)
    Q_UNUSED(lprcMonitor)

    DWORD numPhysicalMonitors = 0;

    if (!::GetNumberOfPhysicalMonitorsFromHMONITOR(hMonitor, &numPhysicalMonitors)) {
        qWarning() << "GetNumberOfPhysicalMonitorsFromHMONITOR failed. Error: " << GetLastError();
        return TRUE;
    }

    if (numPhysicalMonitors == 0) {
        qWarning() << "No physical monitors found for this HMONITOR.";
        return TRUE;
    }

    PHYSICAL_MONITOR* physicalMonitors = new PHYSICAL_MONITOR[numPhysicalMonitors];

    if (!::GetPhysicalMonitorsFromHMONITOR(hMonitor, numPhysicalMonitors, physicalMonitors)) {
        qWarning() << "GetPhysicalMonitorsFromHMONITOR failed. Error: " << GetLastError();
        delete[] physicalMonitors;
        return TRUE;
    }

    Q_ASSERT(dwData != 0);
    QList<MonitorData> &monitorDataList = *reinterpret_cast<QList<MonitorData>*>(dwData);
    for (DWORD index = 0; index < numPhysicalMonitors; ++index) {
        MonitorData data;
        data.hPhysicalMonitor = physicalMonitors[index].hPhysicalMonitor;
        data.desc = QString::fromWCharArray(physicalMonitors[index].szPhysicalMonitorDescription);

        monitorDataList.push_back(data);
    }

    delete[] physicalMonitors;
    return TRUE;
}

void MonitorPlugEvent::getCurrentAllMonitroData(QList<MonitorData> &data)
{
    data.clear();
    QList<MonitorData> newData;
    if (::EnumDisplayMonitors(nullptr, nullptr, monitorEnumProc, reinterpret_cast<LPARAM>(&newData)) == FALSE) {
        qWarning() << "EnumDisplayMonitors failed. Error: " << GetLastError();
        return;
    }

    for (auto &itemData : newData) {
        std::string macAddress;
        if (getSmartMonitorMacAddr(itemData.hPhysicalMonitor, macAddress)) {
            itemData.macAddress = macAddress;
            itemData.isDIAS.store(true);
            data.push_back(itemData);
        }
    }
}

bool MonitorPlugEvent::getSmartMonitorMacAddr(HANDLE hPhysicalMonitor, std::string &macAddr)
{
    bool retVal = false;
    std::array<unsigned char, 6> tmpVal;
    const int vcpRetryCnt = 1;
    for (int i = 0; i < vcpRetryCnt; ++i) {
        retVal = GetSmartMonitorMacAddr(hPhysicalMonitor, tmpVal);
        if (retVal == true) {
            macAddr = std::string(reinterpret_cast<const char*>(tmpVal.data()), tmpVal.size());
            break;
        }
    }
    return retVal;
}

bool MonitorPlugEvent::querySmartMonitorAuthStatus(HANDLE hPhysicalMonitor, uint32_t index, unsigned char &authResult)
{
    bool retVal = false;
    for (int i = 0; i < g_vcpRetryCnt; ++i) {
        retVal = QuerySmartMonitorAuthStatus(hPhysicalMonitor, index, authResult);
        if (retVal == true) {
            break;
        }
    }
    return retVal;
}

bool MonitorPlugEvent::getConnectedPortInfo(HANDLE hPhysicalMonitor, uint8_t &source, uint8_t &port)
{
    bool retVal = false;
    for (int i = 0; i < g_vcpRetryCnt; ++i) {
        retVal = GetConnectedPortInfo(hPhysicalMonitor, source, port);
        if (retVal == true) {
            break;
        }
    }
    return retVal;
}

bool MonitorPlugEvent::updateMousePos(HANDLE hPhysicalMonitor, unsigned short width, unsigned short hight, short posX, short posY)
{
    bool retVal = false;
    for (int i = 0; i < g_vcpRetryCnt; ++i) {
        retVal = UpdateMousePos(hPhysicalMonitor, width, hight, posX, posY);
        if (retVal == true) {
            break;
        }
    }
    return retVal;
}

bool MonitorPlugEvent::isDIASMonitor(uint8_t authResult)
{
    Q_UNUSED(authResult)
    // FIXME: The current plan is undecided, all return true
    return true;
    //return authResult == 0x01;
}
