#include "windows_event_monitor.h"
#include "crossShareVcpCtrl.h"
#include <physicalmonitorenumerationapi.h>
#include <highlevelmonitorconfigurationapi.h>
#include <lowlevelmonitorconfigurationapi.h>
#include <algorithm>
#include <chrono>
#include <thread>
#include <unordered_map>
#define TASK_QUEUE_MSG (WM_APP + 1)

namespace {

// FIXME: Due to the probability of failure, a number of repeated attempts is set here
const int g_vcpRetryCnt = 5;

HWND g_hWnd { 0 };
std::list<WindowsEventMonitor::TaskCallback> g_taskQeue;
std::mutex g_taskQueueMutex;
std::atomic<uint32_t> g_currentThreadId { 0 };

struct TimerData
{
    WindowsEventMonitor::TaskCallback callback;
    bool isSingleShot { false };
};
std::unordered_map<uint64_t, TimerData> g_timerDataMap;
std::mutex g_timerMutex;

uint64_t getCurrentTimeInMS()
{
    using namespace std::chrono;
    return duration_cast<milliseconds>(system_clock::now().time_since_epoch()).count();
}

std::string toLowerString(const std::string &str)
{
    std::string lowerStr = str;
    std::transform(lowerStr.begin(), lowerStr.end(), lowerStr.begin(), [](unsigned char c) {
        return std::tolower(c);
    });
    return lowerStr;
}

bool contains(const std::string &left, const std::string &right)
{
    return left.find(right) != std::string::npos;
}

}

WindowsEventMonitor *WindowsEventMonitor::m_instance { nullptr };

WindowsEventMonitor::WindowsEventMonitor()
{

}

WindowsEventMonitor::~WindowsEventMonitor()
{
    m_instance = nullptr;
}

WindowsEventMonitor *WindowsEventMonitor::getInstance()
{
    if (m_instance == nullptr) {
        m_instance = new WindowsEventMonitor;
    }
    return m_instance;
}

int WindowsEventMonitor::getMonitorCount() const
{
    int count = 0;
    EnumDisplayMonitors(nullptr, nullptr, [](HMONITOR, HDC, LPRECT, LPARAM lParam) {
        (*reinterpret_cast<int*>(lParam))++;
        return TRUE;
    }, reinterpret_cast<LPARAM>(&count));
    return count;
}

void WindowsEventMonitor::updateMonitorDataList()
{
    std::vector<MonitorData> newDataList;
    getCurrentAllMonitroData(newDataList);
    if (!newDataList.empty()) {
        m_cacheMonitorData = newDataList.front();
        m_monitorDataList = newDataList;
    }
}

void WindowsEventMonitor::debugDeviceInfo(DWORD dwDevType, const std::string &devPath)
{
    LOG_INFO << "[Device Event] Type:" << dwDevType
              << "\nPath:" << devPath
              << "\n----------------------------------";
}

BOOL WindowsEventMonitor::monitorEnumProc(HMONITOR hMonitor, HDC hdcMonitor, LPRECT lprcMonitor, LPARAM dwData)
{
    (void)hdcMonitor;
    (void)lprcMonitor;

    DWORD numPhysicalMonitors = 0;

    if (!::GetNumberOfPhysicalMonitorsFromHMONITOR(hMonitor, &numPhysicalMonitors)) {
        LOG_INFO << "GetNumberOfPhysicalMonitorsFromHMONITOR failed. Error: " << GetLastError();
        return TRUE;
    }

    if (numPhysicalMonitors == 0) {
        LOG_INFO << "No physical monitors found for this HMONITOR.";
        return TRUE;
    }

    PHYSICAL_MONITOR *physicalMonitors = new PHYSICAL_MONITOR[numPhysicalMonitors];

    if (!::GetPhysicalMonitorsFromHMONITOR(hMonitor, numPhysicalMonitors, physicalMonitors)) {
        LOG_INFO << "GetPhysicalMonitorsFromHMONITOR failed. Error: " << GetLastError();
        delete[] physicalMonitors;
        return TRUE;
    }

    assert(dwData != 0);
    std::vector<MonitorData> &monitorDataList = *reinterpret_cast<std::vector<MonitorData>*>(dwData);
    for (DWORD index = 0; index < numPhysicalMonitors; ++index) {
        MonitorData data;
        data.hPhysicalMonitor = physicalMonitors[index].hPhysicalMonitor;
        data.desc = sunkang::Utils::toUtf8(std::wstring(physicalMonitors[index].szPhysicalMonitorDescription));

        monitorDataList.push_back(data);
    }

    delete[] physicalMonitors;
    return TRUE;
}

uint64_t WindowsEventMonitor::runTimer(std::chrono::milliseconds interval, const TaskCallback &callback, bool isSingleShot)
{
    static std::atomic<uint64_t> s_timerID { 1 };
    uint64_t timerID = s_timerID.fetch_add(1);
    {
        std::lock_guard<std::mutex> locker(g_timerMutex);
        (void)locker;
        TimerData data { callback, isSingleShot };
        g_timerDataMap[timerID] = std::move(data);
    }
    ::SetTimer(g_hWnd, timerID, interval.count(), timerProc);
    return timerID;
}

void WindowsEventMonitor::timerProc(HWND hwnd, UINT msg, UINT_PTR idTimer, DWORD dwTime)
{
    assert(hwnd == g_hWnd);
    assert(msg == WM_TIMER);
    (void)dwTime;

    TimerData timerData;
    {
        std::lock_guard<std::mutex> locker(g_timerMutex);
        (void)locker;
        auto itr = g_timerDataMap.find(static_cast<uint64_t>(idTimer));
        if (itr != g_timerDataMap.end()) {
            timerData = itr->second;
        } else {
            ::KillTimer(hwnd, idTimer);
            return;
        }
    }

    if (timerData.callback) {
        timerData.callback();
    }

    if (timerData.isSingleShot) {
        ::KillTimer(hwnd, idTimer);
        std::lock_guard<std::mutex> locker(g_timerMutex);
        (void)locker;
        auto itr = g_timerDataMap.find(static_cast<uint32_t>(idTimer));
        assert(itr != g_timerDataMap.end());
        g_timerDataMap.erase(itr);
    }
}

void WindowsEventMonitor::getCurrentAllMonitroData(std::vector<MonitorData> &data)
{
    data.clear();
    std::vector<MonitorData> newData;
    if (::EnumDisplayMonitors(nullptr, nullptr, monitorEnumProc, reinterpret_cast<LPARAM>(&newData)) == FALSE) {
        LOG_INFO << "EnumDisplayMonitors failed. Error: " << GetLastError();
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

LRESULT WindowsEventMonitor::windowProc(HWND hWnd, UINT msg, WPARAM wParam, LPARAM lParam)
{
    switch (msg) {
    case WM_CREATE: {
        LOG_INFO << "------------------------WndProc WM_CREATE";
        g_hWnd = hWnd;
        g_currentThreadId.store(::GetCurrentThreadId());
        // Delay for 100 milliseconds.
        // Wait until the initialize function finishes initializing, and then execute the initData function.
        WindowsEventMonitor::getInstance()->runAfter(std::chrono::milliseconds(100), [] {
            WindowsEventMonitor::getInstance()->initData();
        });
        break;
    }
    case WM_DEVICECHANGE: {
        if (wParam == DBT_DEVICEARRIVAL || wParam == DBT_DEVICEREMOVECOMPLETE) {
            DEV_BROADCAST_HDR *pHdr = reinterpret_cast<DEV_BROADCAST_HDR*>(lParam);
            if (pHdr->dbch_devicetype != DBT_DEVTYP_DEVICEINTERFACE) {
                return TRUE;
            }

            auto pDevInf = reinterpret_cast<DEV_BROADCAST_DEVICEINTERFACE_W*>(pHdr);
            std::string devPath { reinterpret_cast<const char*>(pDevInf->dbcc_name) };
            devPath = toLowerString(devPath);
            if (contains(devPath, "display") == false && contains(devPath, "monitor") == false) {
                return TRUE;
            }
            if (wParam == DBT_DEVICEARRIVAL) {
                static std::atomic<bool> s_processingStatus { false };
                if (s_processingStatus.load() == false
                        && WindowsEventMonitor::getInstance()->getCacheMonitorData().isDIAS.load() == false) {
                    s_processingStatus.store(true);
                    std::thread([] {
                        auto currentTimeMS = getCurrentTimeInMS();
                        while (getCurrentTimeInMS() - currentTimeMS <= 3000) { // Exit after timeout of 3000ms
                            std::vector<MonitorData> newDataList;
                            WindowsEventMonitor::getInstance()->getCurrentAllMonitroData(newDataList);
                            if (!newDataList.empty()) {
                                std::atomic<bool> doneStatus { false };
                                WindowsEventMonitor::getInstance()->runInLoop([&doneStatus] {
                                    WindowsEventMonitor::getInstance()->updateMonitorDataList();
                                    WindowsEventMonitor::getInstance()->macAddressNotify();
                                    doneStatus.store(true);
                                });
                                while (doneStatus.load() == false) {
                                    std::this_thread::sleep_for(std::chrono::milliseconds(10));
                                }
                                break;
                            }
                            std::this_thread::sleep_for(std::chrono::milliseconds(10)); // Avoid rapid repeated testing
                        }
                        s_processingStatus.store(false);
                    }).detach();
                }
            } else {
                static std::atomic<bool> s_processingStatus { false };
                if (s_processingStatus.load() == false
                        && WindowsEventMonitor::getInstance()->getCacheMonitorData().isDIAS.load() == true) {
                    s_processingStatus.store(true);
                    std::thread([] {
                        auto currentTimeMS = getCurrentTimeInMS();
                        while (getCurrentTimeInMS() - currentTimeMS <= 3000) { // Exit after timeout of 3000ms
                            std::vector<MonitorData> newDataList;
                            WindowsEventMonitor::getInstance()->getCurrentAllMonitroData(newDataList);
                            if (newDataList.empty()) {
                                std::atomic<bool> doneStatus { false };
                                WindowsEventMonitor::getInstance()->runInLoop([&doneStatus] {
                                    WindowsEventMonitor::getInstance()->updateMonitorDataList();
                                    WindowsEventMonitor::getInstance()->extractDIASMonitorNotify();
                                    doneStatus.store(true);
                                });
                                while (doneStatus.load() == false) {
                                    std::this_thread::sleep_for(std::chrono::milliseconds(10));
                                }
                                break;
                            }
                            std::this_thread::sleep_for(std::chrono::milliseconds(10)); // Avoid rapid repeated testing
                        }
                        s_processingStatus.store(false);
                    }).detach();
                }
            }
        }
        return TRUE;
    }
    case WM_DESTROY: {
        UnregisterDeviceNotification(WindowsEventMonitor::getInstance()->m_hDevNotify);
        PostQuitMessage(0);
        break;
    }
    case TASK_QUEUE_MSG: {
        std::list<TaskCallback> taskQeue;
        {
            std::lock_guard<std::mutex> locker(g_taskQueueMutex);
            (void)locker;
            taskQeue.swap(g_taskQeue);
        }
        while (taskQeue.empty() == false) {
            auto callback = taskQeue.front();
            taskQeue.pop_front();
            if (callback) {
                callback();
            }
        }
        break;
    }
    }
    return DefWindowProc(hWnd, msg, wParam, lParam);
}


bool WindowsEventMonitor::initialize()
{
    WNDCLASSEX wc;
    memset(&wc, 0, sizeof (wc));
    wc.cbSize = sizeof(WNDCLASSEX);
    wc.lpfnWndProc = windowProc;
    wc.hInstance = GetModuleHandle(nullptr);
    wc.lpszClassName = "WindowsEventMonitorClass";

    if (!RegisterClassEx(&wc)) {
        return false;
    }

    CreateWindowEx(0, wc.lpszClassName, "", 0, 0, 0, 0, 0, nullptr, nullptr, wc.hInstance, this);
    if (!g_hWnd) {
        return false;
    }

    DEV_BROADCAST_DEVICEINTERFACE dbdi{};
    dbdi.dbcc_size = sizeof(DEV_BROADCAST_DEVICEINTERFACE);
    dbdi.dbcc_devicetype = DBT_DEVTYP_DEVICEINTERFACE;
    dbdi.dbcc_classguid = GUID_DEVCLASS_MONITOR;

    m_hDevNotify = RegisterDeviceNotification(g_hWnd, &dbdi, DEVICE_NOTIFY_WINDOW_HANDLE | DEVICE_NOTIFY_ALL_INTERFACE_CLASSES);

    if (!m_hDevNotify) {
        DestroyWindow(g_hWnd);
        return false;
    }

    m_currentMonitorCount = getMonitorCount();
    LOG_INFO << "m_currentMonitorCount: " << m_currentMonitorCount;
    return true;
}

void WindowsEventMonitor::start()
{
    MSG msg;
    while (GetMessage(&msg, nullptr, 0, 0)) {
        TranslateMessage(&msg);
        DispatchMessage(&msg);
    }
}

void WindowsEventMonitor::initData()
{
    for (int i = 0; i < g_vcpRetryCnt; ++i) {
        m_monitorDataList.clear();
        getCurrentAllMonitroData(m_monitorDataList);
        if (m_monitorDataList.empty() == true) {
            break;
        }

        m_cacheMonitorData = m_monitorDataList.front();
        if (m_cacheMonitorData.macAddress.empty()) {
            continue;
        }

        for (const auto &itemData : m_monitorDataList) {
            LOG_INFO << "macAddress(hex):" << sunkang::Utils::toHex(itemData.macAddress)
                     << "; monitor desc:" << itemData.desc;
        }
        macAddressNotify();
        break;
    }
}

void WindowsEventMonitor::clearData()
{
    m_cacheMonitorData = MonitorData();
}

void WindowsEventMonitor::runInLoop(const TaskCallback &callback)
{
    assert(g_currentThreadId.load() != 0);
    if (::GetCurrentThreadId() == g_currentThreadId.load()) {
        if (callback) {
            callback();
        }
        return;
    }
    std::lock_guard<std::mutex> locker(g_taskQueueMutex);
    (void)locker;
    g_taskQeue.push_back(callback);
    ::PostMessage(g_hWnd, TASK_QUEUE_MSG, 0, 0);
}

uint64_t WindowsEventMonitor::runAfter(std::chrono::milliseconds interval, const TaskCallback &callback)
{
    return runTimer(interval, callback, true);
}

uint64_t WindowsEventMonitor::runEvery(std::chrono::milliseconds interval, const TaskCallback &callback)
{
    return runTimer(interval, callback, false);
}

void WindowsEventMonitor::killTimer(uint64_t timerID)
{
    std::lock_guard<std::mutex> locker(g_timerMutex);
    (void)locker;
    auto itr = g_timerDataMap.find(timerID);
    if (itr != g_timerDataMap.end()) {
        g_timerDataMap.erase(itr);
    }
    ::KillTimer(g_hWnd, timerID);
}

bool WindowsEventMonitor::getSmartMonitorMacAddr(HANDLE hPhysicalMonitor, std::string &macAddr)
{
    bool retVal = false;
    std::array<unsigned char, 6> tmpVal;
    for (int i = 0; i < g_vcpRetryCnt; ++i) {
        retVal = GetSmartMonitorMacAddr(hPhysicalMonitor, tmpVal);
        if (retVal == true) {
            macAddr = std::string(reinterpret_cast<const char*>(tmpVal.data()), tmpVal.size());
            break;
        }
    }
    return retVal;
}

bool WindowsEventMonitor::querySmartMonitorAuthStatus(HANDLE hPhysicalMonitor, uint32_t index, unsigned char &authResult)
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

bool WindowsEventMonitor::getConnectedPortInfo(HANDLE hPhysicalMonitor, uint8_t &source, uint8_t &port)
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

bool WindowsEventMonitor::updateMousePos(HANDLE hPhysicalMonitor, unsigned short width, unsigned short hight, short posX, short posY)
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

void WindowsEventMonitor::setGetMacAddressCallback(GetMacAddressCallback callback)
{
    assert(callback != nullptr);
    m_getMacAddressCallback = callback;
}

void WindowsEventMonitor::setExtractDIASCallback(ExtractDIASCallback callback)
{
    assert(callback != nullptr);
    m_extractDIASCallback = callback;
}

void WindowsEventMonitor::setAuthStatusCodeCallback(AuthStatusCodeCallback callback)
{
    assert(callback != nullptr);
    m_authStatusCodeCallback = callback;
}

void WindowsEventMonitor::setDIASSourceAndPortCallback(DIASSourceAndPortCallback callback)
{
    assert(callback != nullptr);
    m_DIASSourceAndPortCallback = callback;
}

void WindowsEventMonitor::macAddressNotify()
{
    runInLoop([this] {
        auto cacheData = getCacheMonitorData();
        if (cacheData.macAddress.empty() == false) {
            m_getMacAddressCallback(const_cast<char*>(cacheData.macAddress.data()), static_cast<int>(cacheData.macAddress.size()));
        }
    });
}

void WindowsEventMonitor::extractDIASMonitorNotify()
{
    runInLoop([this] {
        auto cacheData = getCacheMonitorData();
        if (cacheData.isDIAS) {
            m_extractDIASCallback();
        }
        clearData();
    });
}

void WindowsEventMonitor::authViaIndex(uint32_t indexValue)
{
    runInLoop([this, indexValue] {
        auto cacheData = getCacheMonitorData();
        uint8_t authResult = 0;
        if (querySmartMonitorAuthStatus(cacheData.hPhysicalMonitor, indexValue, authResult) == false) {
            LOG_WARN << "querySmartMonitorAuthStatus failed ......";
            return;
        }
        m_authStatusCodeCallback(static_cast<unsigned char>(authResult));
    });
}

void WindowsEventMonitor::requestSourcePort()
{
    runInLoop([this] {
        auto cacheData = getCacheMonitorData();
        uint8_t source = 0;
        uint8_t port = 0;
        if (getConnectedPortInfo(cacheData.hPhysicalMonitor, source, port) == false) {
            return;
        }
        m_DIASSourceAndPortCallback(static_cast<unsigned char>(source), static_cast<unsigned char>(port));
    });
}
