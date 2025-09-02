#include "monitor_info.h"
#include <windows.h>
#include <SetupAPI.h>
#include <iostream>
#include <iomanip>
#include <cfgmgr32.h>
#include <devguid.h> // GUID_DEVCLASS_MONITOR
#include <physicalmonitorenumerationapi.h>
#include <highlevelmonitorconfigurationapi.h>
#include <lowlevelmonitorconfigurationapi.h>

#define EDID_SIZE 128

namespace {

// https://glenwing.github.io/docs/VESA-EEDID-A2.pdf
// 3. Extended Display Identification Data (EDID) Version 1 Revision 4

// EDID 1.4
struct EDID_DATA
{
    uint8_t header[8]; // 00 FF FF FF FF FF FF 00
    // Vendor & Product Identification:
    uint16_t manufacturerID;
    uint16_t productCode;
    uint32_t serialNumber;
    uint8_t manufactureWeek;
    uint8_t manufactureYear;

    //EDID Structure Version & Revision:
    uint8_t edidVersion;
    uint8_t edidRevision;

    // Basic Display Parameters & Features:
    uint8_t videoInputParams;
    uint8_t maxHSize;
    uint8_t maxVSize;
    uint8_t gamma;
    uint8_t features;

    // Color Characteristics:
    uint8_t redGreenLow;
    uint8_t blueWhiteLow;
    uint8_t redXHigh;
    uint8_t redYHigh;
    uint8_t greenXHigh;
    uint8_t greenYHigh;
    uint8_t blueXHigh;
    uint8_t blueYHigh;
    uint8_t whiteXHigh;
    uint8_t whiteYHigh;

    // Established Timings
    uint8_t establishedTimings[3];

    // Standard Timings: Identification 1 → 8
    uint8_t standardTimings[8][2];

    // 18 Byte Data Blocks
    uint8_t detailedTimingBlocks[4][18];

    // Extension Block Count N
    uint8_t extensionFlag;

    uint8_t checksum;
};

static_assert (sizeof (EDID_DATA) == EDID_SIZE, "");
}

MonitorInfoUtils::MonitorInfoUtils()
{
    if (getAllEdidData(m_edidInfoVec) == false) {
        qWarning() << "----------------get EDID data failed !!!";
    }
}

MonitorInfoUtils::~MonitorInfoUtils()
{

}

bool MonitorInfoUtils::getAllEdidData(std::vector<std::vector<uint8_t> > &edidData)
{
    edidData.clear();
    HDEVINFO devInfo = SetupDiGetClassDevs(&GUID_DEVCLASS_MONITOR, 0, 0, DIGCF_PRESENT);
    if (devInfo == INVALID_HANDLE_VALUE) {
        return false;
    }

    SP_DEVINFO_DATA devInfoData;
    memset(&devInfoData, 0, sizeof (devInfoData));
    devInfoData.cbSize = sizeof(SP_DEVINFO_DATA);

    for (DWORD index = 0; SetupDiEnumDeviceInfo(devInfo, index, &devInfoData); ++index) {
        HKEY hKey = SetupDiOpenDevRegKey(devInfo, &devInfoData, DICS_FLAG_GLOBAL, 0, DIREG_DEV, KEY_READ);
        if (hKey == INVALID_HANDLE_VALUE) {
            continue;
        }

        DWORD edidSize = 0;
        DWORD type = 0;
        LSTATUS status = RegQueryValueEx(hKey, "EDID", 0, &type, nullptr, &edidSize);
        if (status == ERROR_SUCCESS && type == REG_BINARY) {
            std::vector<uint8_t> tmpEdidData;
            tmpEdidData.resize(edidSize);
            RegQueryValueEx(hKey, "EDID", 0, nullptr, tmpEdidData.data(), &edidSize);
            edidData.push_back(tmpEdidData);
        }
        RegCloseKey(hKey);
    }

    SetupDiDestroyDeviceInfoList(devInfo);
    return true;
}

bool MonitorInfoUtils::parseFullEDID(const std::vector<uint8_t> &rawData, MonitorInfoEx &info)
{
    if (rawData.size() < EDID_SIZE) {
        qWarning() << "Incomplete EDID data";
        return false;
    }

    EDID_DATA edidData;
    memcpy(&edidData, rawData.data(), sizeof (EDID_DATA));

    {
        QByteArray headerData = QByteArray::fromHex("00FFFFFFFFFFFF00");
        if (memcmp(headerData.data(), &edidData.header[0], headerData.size()) != 0) {
            return false;
        }
    }

    {
        info.edidVersion = std::to_string(edidData.edidVersion) + "." + std::to_string(edidData.edidRevision);
    }

    {
        union {
            uint8_t code[2];
            uint16_t value;
        } bitData;
        bitData.value = edidData.manufacturerID;
        char chr_1 = static_cast<char>((bitData.code[0] >> 2) & 0B11111);
        char chr_2 = static_cast<char>(((bitData.code[0] & 0B11) << 3) | (bitData.code[1] >> 5));
        char chr_3 = static_cast<char>(bitData.code[1] & 0B11111);
        info.manufacturer.push_back(static_cast<char>(chr_1 + 'A' - 1));
        info.manufacturer.push_back(static_cast<char>(chr_2 + 'A' - 1));
        info.manufacturer.push_back(static_cast<char>(chr_3 + 'A' - 1));
    }

    {
        if (edidData.manufactureWeek >= 0x01 && edidData.manufactureWeek <= 0x36) {
            info.manufactureWeek = edidData.manufactureWeek;
        }

        info.manufactureYear = edidData.manufactureYear + 1990;
    }

    {
        if (edidData.gamma != 0xFF) {
            info.gamma = static_cast<double>((edidData.gamma + 100.0) / 100.0);
        }
    }

    {
        // 1 inch = 2.54cm
        auto val = std::sqrt(std::pow(edidData.maxHSize, 2.0) + std::pow(edidData.maxVSize, 2.0)) / 2.54;
        info.size = val;

        info.maxHSize = edidData.maxHSize;
        info.maxVSize = edidData.maxVSize;
    }

    {
        for (int index = 0; index < 4; ++index) {
            const auto &blockData = edidData.detailedTimingBlocks[index];
            if (blockData[0] == 0x00 && blockData[1] == 0x00 && blockData[2] == 0x00 && blockData[3] == 0xFC) {
                for (int i = 5; i <= 17; ++i) {
                    info.productName.push_back(static_cast<char>(blockData[i]));
                }

                info.productName = QString::fromStdString(info.productName).trimmed().toStdString();
                break;
            }
        }

        {
            union {
                uint8_t code[2];
                uint16_t value;
            } bitData;
            bitData.value = edidData.productCode;
            std::swap(bitData.code[0], bitData.code[1]);
            QString productCodeHex = QByteArray(reinterpret_cast<const char*>(&bitData), sizeof (uint16_t)).toHex().toUpper();
            qInfo() << "productCode hex:" << productCodeHex.toUtf8().constData();
            if (info.productName.empty()) {
                info.productName = info.manufacturer + productCodeHex.toStdString();
            }
        }
    }

    {
        for (int index = 0; index < 4; ++index) {
            const auto &blockData = edidData.detailedTimingBlocks[index];
            if (blockData[0] == 0x00 && blockData[1] == 0x00 && blockData[2] == 0x00 && blockData[3] == 0xFF) {
                for (int i = 5; i <= 17; ++i) {
                    info.productSerialNumber.push_back(static_cast<char>(blockData[i]));
                }

                info.productSerialNumber = QString::fromStdString(info.productSerialNumber).trimmed().toStdString();
                break;
            }
        }
    }

    return true;
}


std::vector<MonitorInfoEx> MonitorInfoUtils::parseAllEdidInfo() const
{
    std::vector<MonitorInfoEx> infoVec;
    for (const auto &edidData: m_edidInfoVec) {
        MonitorInfoEx tmpInfo;
        if (parseFullEDID(edidData, tmpInfo)) {
            infoVec.push_back(tmpInfo);
        }
    }
    return infoVec;
}


////////////////////
// https://learn.microsoft.com/zh-cn/windows/win32/api/_monitor/

MonitorController::MonitorController()
{

}

MonitorController::~MonitorController()
{

}

bool MonitorController::control(uint8_t vcpCode, uint32_t *pNewVal, uint32_t *pCurrentVal, uint32_t *pMaxVal)
{
    HMONITOR hMonitor = MonitorFromWindow(nullptr, MONITOR_DEFAULTTOPRIMARY);
    if (!hMonitor) {
        qWarning() << "------------------ERROR CODE: " << GetLastError();
        return false;
    }

    DWORD numPhysicalMonitors = 0;
    if (!GetNumberOfPhysicalMonitorsFromHMONITOR(hMonitor, &numPhysicalMonitors)) {
        qWarning() << "Failed to get number of physical monitors.";
        return false;
    }

    if (numPhysicalMonitors == 0) {
        qWarning() << "No physical monitors found.";
        return false;
    }

    LPPHYSICAL_MONITOR ptrPhysicalMonitors = new PHYSICAL_MONITOR[numPhysicalMonitors];
    if (GetPhysicalMonitorsFromHMONITOR(hMonitor, numPhysicalMonitors, ptrPhysicalMonitors) == FALSE) {
        qWarning() << "Failed to get physical monitors.";
        delete [] ptrPhysicalMonitors;
        return false;
    }

    DWORD currentValue = 0;
    DWORD maxValue = 0;

    if (!GetVCPFeatureAndVCPFeatureReply(ptrPhysicalMonitors->hPhysicalMonitor, vcpCode, nullptr, &currentValue, &maxValue)) {
        qWarning() << "Failed to get VCP feature and reply. Error: " << GetLastError();
        DestroyPhysicalMonitors(numPhysicalMonitors, ptrPhysicalMonitors);
        delete[] ptrPhysicalMonitors;
        return false;
    }

    qInfo() << "VCP Code: " << static_cast<int>(vcpCode);
    qInfo() << "Current Value: " << currentValue;
    qInfo() << "Maximum Value: " << maxValue;

    if (pCurrentVal) {
        *pCurrentVal = currentValue;
    }

    if (pMaxVal) {
        *pMaxVal = maxValue;
    }

    bool retVal = true;
    if (pNewVal) {
        DWORD newValue = *pNewVal;
        if (!SetVCPFeature(ptrPhysicalMonitors->hPhysicalMonitor, vcpCode, newValue)) {
            qWarning() << "Failed to set VCP feature. Error: " << GetLastError();
            retVal = false;
        }
    }

    DestroyPhysicalMonitors(numPhysicalMonitors, ptrPhysicalMonitors);
    delete[] ptrPhysicalMonitors;
    return retVal;
}
