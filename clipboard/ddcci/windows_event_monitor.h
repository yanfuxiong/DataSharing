#pragma once
#include "global_utils.h"
#include <iostream>
#include <functional>
#include <map>
#include <cassert>
#include <atomic>
#include <Windows.h>
#include <dbt.h>
#include <devguid.h>
#include "../clipboard.h"

class WindowsEventMonitor
{
public:
    using TaskCallback = std::function<void()>;
    struct MonitorData
    {
        HANDLE hPhysicalMonitor { nullptr };
        std::string desc;
        std::atomic<bool> isDIAS { false };
        std::string macAddress;

        MonitorData() = default;
        ~MonitorData() = default;
        MonitorData(const MonitorData &other)
            : hPhysicalMonitor(other.hPhysicalMonitor)
            , desc(other.desc)
            , isDIAS(other.isDIAS.load())
            , macAddress(other.macAddress)
        {}
        MonitorData &operator = (const MonitorData &other)
        {
            if (this == &other) {
                return *this;
            }
            hPhysicalMonitor = other.hPhysicalMonitor;
            desc = other.desc;
            isDIAS = other.isDIAS.load();
            macAddress = other.macAddress;
            return *this;
        }
    };
    ~WindowsEventMonitor();
    static WindowsEventMonitor *getInstance();

    bool initialize();
    void start();

    void initData();
    void clearData();
    void runInLoop(const TaskCallback &callback);
    uint64_t runAfter(std::chrono::milliseconds interval, const TaskCallback &callback);
    uint64_t runEvery(std::chrono::milliseconds interval, const TaskCallback &callback);
    void killTimer(uint64_t timerID);

    const std::vector<MonitorData> &getMonitorData() const { return m_monitorDataList; }
    MonitorData &getCacheMonitorData() { return m_cacheMonitorData; }

    static bool getSmartMonitorMacAddr(HANDLE hPhysicalMonitor, std::string &macAddr);
    static bool querySmartMonitorAuthStatus(HANDLE hPhysicalMonitor, uint32_t index, uint8_t &authResult);
    static bool getConnectedPortInfo(HANDLE hPhysicalMonitor, uint8_t &source, uint8_t &port);
    static bool updateMousePos(HANDLE hPhysicalMonitor, unsigned short width, unsigned short hight, short posX, short posY);

    void setGetMacAddressCallback(GetMacAddressCallback callback);
    void setExtractDIASCallback(ExtractDIASCallback callback);
    void setAuthStatusCodeCallback(AuthStatusCodeCallback callback);
    void setDIASSourceAndPortCallback(DIASSourceAndPortCallback callback);
    void macAddressNotify();
    void extractDIASMonitorNotify();
    void authViaIndex(uint32_t indexValue);
    void requestSourcePort();

private:
    WindowsEventMonitor();
    static LRESULT windowProc(HWND hWnd, UINT msg, WPARAM wParam, LPARAM lParam);
    int getMonitorCount() const;
    void updateMonitorDataList();
    void debugDeviceInfo(DWORD dwDevType, const std::string &devPath);
    void getCurrentAllMonitroData(std::vector<MonitorData> &data);
    uint64_t runTimer(std::chrono::milliseconds interval, const TaskCallback &callback, bool isSingleShot);
    static BOOL monitorEnumProc(HMONITOR hMonitor, HDC hdcMonitor, LPRECT lprcMonitor, LPARAM dwData);
    static void timerProc(HWND hwnd, UINT msg, UINT_PTR idTimer, DWORD dwTime);

    int m_currentMonitorCount { 0 };
    HDEVNOTIFY m_hDevNotify { nullptr };
    std::vector<MonitorData> m_monitorDataList;
    MonitorData m_cacheMonitorData;

    GetMacAddressCallback m_getMacAddressCallback { nullptr };
    ExtractDIASCallback m_extractDIASCallback { nullptr };
    AuthStatusCodeCallback m_authStatusCodeCallback { nullptr };
    DIASSourceAndPortCallback m_DIASSourceAndPortCallback { nullptr };

    static WindowsEventMonitor *m_instance;
};
