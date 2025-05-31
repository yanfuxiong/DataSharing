#include <highlevelmonitorconfigurationapi.h>
#include <lowlevelmonitorconfigurationapi.h>
#include <physicalmonitorenumerationapi.h>
#include <iostream>
#include <iomanip>
#include "displayUtils.h"
#include "crossShareDefine.h"
#include "utils.h"


using namespace std;

BOOL CALLBACK MonitorEnumProc(HMONITOR hMonitor, HDC hdcMonitor, LPRECT lprcMonitor, LPARAM pData) {
    std::vector<MonitorInfo>* monitors = reinterpret_cast<std::vector<MonitorInfo>*>(pData);

    DWORD monitorCount = 0;
    if (!GetNumberOfPhysicalMonitorsFromHMONITOR(hMonitor, &monitorCount) || monitorCount < 1) {
        std::cerr << "Physical monitor not found" << std::endl;
        return TRUE;
    }

    std::vector<PHYSICAL_MONITOR> physicalMonitors(monitorCount);
    if (!GetPhysicalMonitorsFromHMONITOR(hMonitor, monitorCount, physicalMonitors.data())) {
        std::cerr << "Unable to get physical monitor" << std::endl;
        return TRUE;
    }

    for (const auto& pm : physicalMonitors) {
        monitors->push_back({hMonitor, pm.hPhysicalMonitor, ""});
    }

    return TRUE;
}

std::vector<MonitorInfo> GetAllMonitors() {
    std::vector<MonitorInfo> monitors;
    EnumDisplayMonitors(NULL, NULL, MonitorEnumProc, reinterpret_cast<LPARAM>(&monitors));
    return monitors;
}

bool GetMonitorCapabilities(HANDLE hMonitor, std::string& capabilities) {
    char capabilitiesString[1024] = {0};

    if (!CapabilitiesRequestAndCapabilitiesReply(hMonitor, capabilitiesString, sizeof(capabilitiesString))) {
        capabilities = "";
        return false;
    }

    capabilities = capabilitiesString;
    return true;
}

void PrintMonitorInfo(std::vector<MonitorInfo>& monitors)
{
    for (size_t i = 0; i < monitors.size(); ++i) {
        std::cout << "\nMonitor " << i << ":" << std::endl;
        std::string capabilities = "";
        if (!GetMonitorCapabilities(monitors[i].hPhysicalMonitor, capabilities)) {
            std::cerr << "Get Monitor Capabilities Fail" << std::endl;
        } else {
            std::cout << "Monitor Capabilities:" << std::endl;
            std::cout << capabilities << std::endl;
            monitors[i].ddcciCapabilities = capabilities;
        }
    }
}

bool GetMonitorVCPFeature(HANDLE hMonitor, BYTE vcpCode, std::pair<DWORD, DWORD>& result) {
    DWORD currentValue, maxValue;

    if (!GetVCPFeatureAndVCPFeatureReply(hMonitor, vcpCode, NULL, &currentValue, &maxValue)) {
        std::cerr << "Can't get VCP Code: 0x" << std::hex << (int)vcpCode << std::endl;
        return false;
    }
    result = {maxValue, currentValue};

    auto it = kCrossShareVcpCodeMap.find(vcpCode);
    bool isCrossShareVcpCode = (it != kCrossShareVcpCodeMap.end());
    std::string csCmdInfo = kCrossShareVcpCodeMap.at(vcpCode).first;
    // DumpGetVcpCmdInfo(vcpCode, result, isCrossShareVcpCode, csCmdInfo);

    return true;
}

bool SetMonitorVCPFeature(HANDLE hMonitor, BYTE vcpCode, DWORD value) {
    auto it = kCrossShareVcpCodeMap.find(vcpCode);
    bool isCrossShareVcpCode = (it != kCrossShareVcpCodeMap.end());

    if (!SetVCPFeature(hMonitor, vcpCode, value)) {
        std::cerr << "Can't set VCP Code: 0x" << std::hex << (int)vcpCode << std::endl;
        return false;
    }
    std::string cmdInfo = kCrossShareVcpCodeMap.at(DDCCI_CrossShare_CMD_AUTH_DEVICE).second;
    // DumpSetVcpCmdInfo(vcpCode, value, isCrossShareVcpCode, cmdInfo);

    return true;
}