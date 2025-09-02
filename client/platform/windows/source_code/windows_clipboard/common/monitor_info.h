#pragma once
#include "common_utils.h"
#include <string>
#include <vector>
#include <windows.h>

// In the code of knife, there is a struct MonitorInfo. To avoid symbol conflicts, MonitorInfoEx is used here
struct MonitorInfoEx
{
    std::string manufacturer;
    std::string productName;
    std::string productSerialNumber;
    std::string edidVersion;
    std::optional<double> gamma; // [1.0, 3.54]
    int manufactureYear;
    std::optional<int> manufactureWeek;
    double size; // unit: inch
    int maxHSize; // unit: cm
    int maxVSize; // unit: cm
    //std::vector<std::string> videoMode;
};

class MonitorInfoUtils
{
public:
    MonitorInfoUtils();
    ~MonitorInfoUtils();

    std::vector<MonitorInfoEx> parseAllEdidInfo() const;

private:
    static bool getAllEdidData(std::vector<std::vector<uint8_t> > &edidData);
    static bool parseFullEDID(const std::vector<uint8_t> &rawData, MonitorInfoEx &info);

private:
    std::vector<std::vector<uint8_t>> m_edidInfoVec;
};

class MonitorController
{
public:
    MonitorController();
    ~MonitorController();

    bool control(uint8_t vcpCode, uint32_t *pNewVal = nullptr, uint32_t *pCurrentVal = nullptr, uint32_t *pMaxVal = nullptr);
};
