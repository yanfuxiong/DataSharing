#include <iostream>
#include "crossShareVcpCtrl.h"
#include "crossShareDefine.h"
#include <atomic>
#include <QDebug>
#include <QString>

static QString g_getLastErrorAsString(DWORD errorCode)
{
    if (errorCode == 0) {
        return {};
    }

    LPSTR messageBuffer = nullptr;
    size_t size = FormatMessageA(
        FORMAT_MESSAGE_ALLOCATE_BUFFER | FORMAT_MESSAGE_FROM_SYSTEM | FORMAT_MESSAGE_IGNORE_INSERTS,
        nullptr,
        errorCode,
        MAKELANGID(LANG_NEUTRAL, SUBLANG_DEFAULT),
        (LPSTR)&messageBuffer,
        0,
        nullptr
    );

    QString message = QString::fromLocal8Bit(messageBuffer, size);
    LocalFree(messageBuffer);
    return message.trimmed();
}


#define LOG_ONCE \
if (s_status.load()) \
    qWarning()

bool GetSmartMonitorMacAddr(HANDLE hPhysicalMonitor, std::array<unsigned char, 6>& macAddr)
{
    static std::atomic<bool> s_status{ true };

    macAddr.fill(0);
    std::pair<DWORD, DWORD> result1, result2;

    if (!GetMonitorVCPFeature(hPhysicalMonitor, DDCCI_CrossShare_CMD_GET_MAC_1, result1)) {
        auto errCode = GetLastError();
        QString errMessage = QString::asprintf("errorCode: 0x%lX; errorMessage: %s", errCode, g_getLastErrorAsString(errCode).toStdString().c_str());
        LOG_ONCE << "Can't get DDCCI_CrossShare_CMD_GET_MAC_1:" << errMessage;
        s_status.store(false);
        return false;
    }
    if (!GetMonitorVCPFeature(hPhysicalMonitor, DDCCI_CrossShare_CMD_GET_MAC_2, result2)) {
        auto errCode = GetLastError();
        QString errMessage = QString::asprintf("errorCode: 0x%lX; errorMessage: %s", errCode, g_getLastErrorAsString(errCode).toStdString().c_str());
        LOG_ONCE << "Can't get DDCCI_CrossShare_CMD_GET_MAC_2:" << errMessage;
        s_status.store(false);
        return false;
    }

    macAddr[0] = (result1.first & 0xff00) >> 8;
    macAddr[1] = (result1.first & 0xff);
    macAddr[2] = (result1.second & 0xff00) >> 8;
    macAddr[3] = (result1.second & 0xff);
    macAddr[4] = (result2.first & 0xff00) >> 8;
    macAddr[5] = (result2.first & 0xff);

    s_status.store(true);
    return true;
}

bool QuerySmartMonitorAuthStatus(HANDLE hPhysicalMonitor, unsigned int index, unsigned char& authResult)
{
    static std::atomic<bool> s_status{ true };

    index = ((index & 0xFF) << 8);
    if (!SetMonitorVCPFeature(hPhysicalMonitor, DDCCI_CrossShare_CMD_AUTH_DEVICE, index)) {
        auto errCode = GetLastError();
        QString errMessage = QString::asprintf("errorCode: 0x%lX; errorMessage: %s", errCode, g_getLastErrorAsString(errCode).toStdString().c_str());
        LOG_ONCE << "Can't set DDCCI_CrossShare_CMD_AUTH_DEVICE:" << errMessage;
        s_status.store(false);
        return false;
    }

    std::pair<DWORD, DWORD> result;
    if (!GetMonitorVCPFeature(hPhysicalMonitor, DDCCI_CrossShare_CMD_AUTH_DEVICE, result)) {
        auto errCode = GetLastError();
        QString errMessage = QString::asprintf("errorCode: 0x%lX; errorMessage: %s", errCode, g_getLastErrorAsString(errCode).toStdString().c_str());
        LOG_ONCE << "Can't get DDCCI_CrossShare_CMD_AUTH_DEVICE:" << errMessage;
        s_status.store(false);
        return false;
    }
    authResult = static_cast<unsigned char>((result.first & 0xff00) >> 8);

    s_status.store(true);
    return true;
}

bool GetConnectedPortInfo(HANDLE hPhysicalMonitor, unsigned char& source, unsigned char& port)
{
    static std::atomic<bool> s_status{ true };

    std::pair<DWORD, DWORD> result;
    if (!GetMonitorVCPFeature(hPhysicalMonitor, DDCCI_CrossShare_CMD_GET_TV_SRC, result)) {
        auto errCode = GetLastError();
        QString errMessage = QString::asprintf("errorCode: 0x%lX; errorMessage: %s", errCode, g_getLastErrorAsString(errCode).toStdString().c_str());
        LOG_ONCE << "Can't set DDCCI_CrossShare_CMD_GET_TV_SRC:" << errMessage;
        s_status.store(false);
        return false;
    }
    source = (result.first & 0xff00) >> 8;
    port = (result.first & 0xff);

    s_status.store(true);
    return true;
}

bool UpdateMousePos(HANDLE hPhysicalMonitor, unsigned short width, unsigned short hight, short posX, short posY)
{
    static std::atomic<bool> s_status{ true };

    if (!SetMonitorVCPFeature(hPhysicalMonitor, DDCCI_CrossShare_CMD_SET_DESKTOP_RESOLUTION_W, width)) {
        auto errCode = GetLastError();
        QString errMessage = QString::asprintf("errorCode: 0x%lX; errorMessage: %s", errCode, g_getLastErrorAsString(errCode).toStdString().c_str());
        LOG_ONCE << "Can't set DDCCI_CrossShare_CMD_SET_DESKTOP_RESOLUTION_W:" << errMessage;
        s_status.store(false);
        return false;
    }

    if (!SetMonitorVCPFeature(hPhysicalMonitor, DDCCI_CrossShare_CMD_SET_DESKTOP_RESOLUTION_H, hight)) {
        auto errCode = GetLastError();
        QString errMessage = QString::asprintf("errorCode: 0x%lX; errorMessage: %s", errCode, g_getLastErrorAsString(errCode).toStdString().c_str());
        LOG_ONCE << "Can't set DDCCI_CrossShare_CMD_SET_DESKTOP_RESOLUTION_H:" << errMessage;
        s_status.store(false);
        return false;
    }

    if (!SetMonitorVCPFeature(hPhysicalMonitor, DDCCI_CrossShare_CMD_SET_CURSOR_POSITION_X, posX)) {
        auto errCode = GetLastError();
        QString errMessage = QString::asprintf("errorCode: 0x%lX; errorMessage: %s", errCode, g_getLastErrorAsString(errCode).toStdString().c_str());
        LOG_ONCE << "Can't set DDCCI_CrossShare_CMD_SET_CURSOR_POSITION_X:" << errMessage;
        s_status.store(false);
        return false;
    }

    if (!SetMonitorVCPFeature(hPhysicalMonitor, DDCCI_CrossShare_CMD_SET_CURSOR_POSITION_Y, posY)) {
        auto errCode = GetLastError();
        QString errMessage = QString::asprintf("errorCode: 0x%lX; errorMessage: %s", errCode, g_getLastErrorAsString(errCode).toStdString().c_str());
        LOG_ONCE << "Can't set DDCCI_CrossShare_CMD_SET_CURSOR_POSITION_Y:" << errMessage;
        s_status.store(false);
        return false;
    }

    s_status.store(true);
    return true;
}

bool GetCustomerThemeCode(HANDLE hPhysicalMonitor, CrossShareThemeCode& theme)
{
    static std::atomic<bool> s_status{ true };

    std::pair<DWORD, DWORD> result;
    if (!GetMonitorVCPFeature(hPhysicalMonitor, DDCCI_CrossShare_CMD_GET_CUSTOMIZED_THEME, result)) {
        auto errCode = GetLastError();
        QString errMessage = QString::asprintf("errorCode: 0x%lX; errorMessage: %s", errCode, g_getLastErrorAsString(errCode).toStdString().c_str());
        LOG_ONCE << "Can't get DDCCI_CrossShare_CMD_GET_CUSTOMIZED_THEME:" << errMessage;
        s_status.store(false);
        return false;
    }
    theme.byte[0] = (result.first & 0xff00) >> 8;
    theme.byte[1] = (result.first & 0xff);
    theme.byte[2] = (result.second & 0xff00) >> 8;
    theme.byte[3] = (result.second & 0xff);

    s_status.store(true);
    return true;
}
