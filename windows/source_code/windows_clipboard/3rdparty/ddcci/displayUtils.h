#ifndef __DISPLAY_UTILS__
#define __DISPLAY_UTILS__

#include <string>
#include <windows.h>
#include <vector>
#include <utility>


struct MonitorInfo {
    HMONITOR hMonitor;
    HANDLE hPhysicalMonitor;
    std::string ddcciCapabilities;
};

std::vector<MonitorInfo> GetAllMonitors();
bool GetMonitorCapabilities(HANDLE hMonitor, std::string& capabilities);
void PrintMonitorInfo(std::vector<MonitorInfo>& monitors);
bool GetMonitorVCPFeature(HANDLE hMonitor, BYTE vcpCode, std::pair<DWORD, DWORD>& result);
bool SetMonitorVCPFeature(HANDLE hMonitor, BYTE vcpCode, DWORD value);

#endif // __DISPLAY_UTILS__