#ifndef __MISC_UTILS__
#define __MISC_UTILS__

#include <string>
#include <vector>
#include <array>
#include <map>
#include <windows.h>
#include <utility>
#include <chrono>
#include <functional>
#include "typeDefine.h"

struct ConfigData {
    std::array<unsigned char, 6> mac;
    unsigned char connectedSrc = 0;
    unsigned char connectedPort = 0;
    int testRounds = 0;
    double timeOutCriteria = 0.0;
};

void DumpVcpStrMap(const VcpCodeStringVectorMap& vcpMap);
void DumpVcpHexMap(VcpCodeUIntVectorMap& uintMap);
void DumpVcpKey(const unsigned int& key);
void DumpVcpValues(const std::vector<unsigned int>& values);
void DumpGetVcpCmdInfo(BYTE vcpCode, const std::pair<DWORD, DWORD>& result, bool isCrossShareVcpCode = 0, const std::string& csCmdInfo = "");
void DumpSetVcpCmdInfo(BYTE vcpCode, DWORD value, bool isCrossShareVcpCode = 0, const std::string& csCmdInfo = "");
void DumpMacAddr(const std::array<unsigned char, 6>& macAddr);
VcpCodeUIntVectorMap MapTypeConverter(const VcpCodeStringVectorMap& stringMap);
bool ExtractVcpCodeMap(const std::string& capabilities, VcpCodeUIntVectorMap& vcpOpCodeMap);
bool readIniFile(const std::string& filename, ConfigData& config);

class ExecutionTimer {
private:
    std::string mName;
    std::chrono::time_point<std::chrono::high_resolution_clock> mStart;

public:
    ExecutionTimer(const std::string& funcName);
    ~ExecutionTimer();
    double Stop();
};
#endif // __MISC_UTILS__