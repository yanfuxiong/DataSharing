#include <iostream>
#include "crossShareVcpCtrl.h"
#include "crossShareDefine.h"

bool GetSmartMonitorMacAddr(HANDLE hPhysicalMonitor, std::array<unsigned char, 6>& macAddr)
{
    macAddr.fill(0);
    std::pair<DWORD, DWORD> result1, result2;

    if (!GetMonitorVCPFeature(hPhysicalMonitor, DDCCI_CrossShare_CMD_GET_MAC_1, result1)) {
        std::cerr << "Can't get DDCCI_CrossShare_CMD_GET_MAC_1:" << std::endl;
        return false;
    }
    if (!GetMonitorVCPFeature(hPhysicalMonitor, DDCCI_CrossShare_CMD_GET_MAC_2, result2)) {
        std::cerr << "Can't get DDCCI_CrossShare_CMD_GET_MAC_2:" << std::endl;
        return false;
    }

    macAddr[0] = (result1.first & 0xff00) >> 8;
    macAddr[1] = (result1.first & 0xff);
    macAddr[2] = (result1.second & 0xff00) >> 8;
    macAddr[3] = (result1.second & 0xff);
    macAddr[4] = (result2.first & 0xff00) >> 8;
    macAddr[5] = (result2.first & 0xff);
    return true;
}

bool QuerySmartMonitorAuthStatus(HANDLE hPhysicalMonitor, unsigned int index, unsigned char& authResult)
{
    index = ((index & 0xFF) << 8);
    if (!SetMonitorVCPFeature(hPhysicalMonitor, DDCCI_CrossShare_CMD_AUTH_DEVICE, index)) {
        std::cerr << "Can't set DDCCI_CrossShare_CMD_AUTH_DEVICE:" << std::endl;
        return false;
    }

    std::pair<DWORD, DWORD> result;
    if (!GetMonitorVCPFeature(hPhysicalMonitor, DDCCI_CrossShare_CMD_AUTH_DEVICE, result)) {
        std::cerr << "Can't get DDCCI_CrossShare_CMD_AUTH_DEVICE:" << std::endl;
        return false;
    }
    authResult = static_cast<unsigned char>((result.first & 0xff00) >> 8);
    return true;
}

bool GetConnectedPortInfo(HANDLE hPhysicalMonitor, unsigned char& source, unsigned char& port)
{
    std::pair<DWORD, DWORD> result;
    if (!GetMonitorVCPFeature(hPhysicalMonitor, DDCCI_CrossShare_CMD_GET_TV_SRC, result)) {
        std::cerr << "Can't set DDCCI_CrossShare_CMD_GET_TV_SRC:" << std::endl;
        return false;
    }
    source = (result.first & 0xff00) >> 8;
    port = (result.first & 0xff);
    return true;
}

bool UpdateMousePos(HANDLE hPhysicalMonitor, unsigned short width, unsigned short hight, short posX, short posY)
{
    if (!SetMonitorVCPFeature(hPhysicalMonitor, DDCCI_CrossShare_CMD_SET_DESKTOP_RESOLUTION_W, width)) {
        std::cerr << "Can't set DDCCI_CrossShare_CMD_SET_DESKTOP_RESOLUTION_W:" << std::endl;
        return false;
    }

    if (!SetMonitorVCPFeature(hPhysicalMonitor, DDCCI_CrossShare_CMD_SET_DESKTOP_RESOLUTION_H, hight)) {
        std::cerr << "Can't set DDCCI_CrossShare_CMD_SET_DESKTOP_RESOLUTION_H:" << std::endl;
        return false;
    }

    if (!SetMonitorVCPFeature(hPhysicalMonitor, DDCCI_CrossShare_CMD_SET_CURSOR_POSITION_X, posX)) {
        std::cerr << "Can't set DDCCI_CrossShare_CMD_SET_CURSOR_POSITION_X:" << std::endl;
        return false;
    }

    if (!SetMonitorVCPFeature(hPhysicalMonitor, DDCCI_CrossShare_CMD_SET_CURSOR_POSITION_Y, posY)) {
        std::cerr << "Can't set DDCCI_CrossShare_CMD_SET_CURSOR_POSITION_Y:" << std::endl;
        return false;
    }
    return true;
}
