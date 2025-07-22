#ifndef __CS_VCP_CTRL__
#define __CS_VCP_CTRL__

#include <string>
#include <array>
#include <map>
#include <windows.h>
#include "displayUtils.h"

bool GetSmartMonitorMacAddr(HANDLE hPhysicalMonitor, std::array<unsigned char, 6>& macAddr);
bool QuerySmartMonitorAuthStatus(HANDLE hPhysicalMonitor, unsigned int index, unsigned char& authResult);
bool GetConnectedPortInfo(HANDLE hPhysicalMonitor, unsigned char& source, unsigned char& port);
bool UpdateMousePos(HANDLE hPhysicalMonitor, unsigned short width, unsigned short hight, short posX, short posY);

#endif // __CS_VCP_CTRL__