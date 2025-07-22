#include <iostream>
#include <vector>
#include <sstream>
#include <fstream>
#include <stdexcept>
#include <iomanip>
#include "utils.h"
#include "strUtils.h"

using namespace std;

void DumpVcpStrMap(const VcpCodeStringVectorMap& vcpMap)
{
    std::cout << "Dump VCP Map:\n";
    for (const auto& [key, values] : vcpMap) {
        std::cout << key << ": ";
        if (values.empty()) {
            std::cout << "Continuous";
        } else {
            std::cout << "{ ";
            for (const auto& val : values) {
                std::cout << val << " ";
            }
            std::cout << "}";
        }
        std::cout << "\n";
    }
}

void DumpVcpKey(const unsigned int& key)
{
    std::cout << "0x" << std::hex << std::uppercase << std::setw(2) << std::setfill('0') << key << std::dec;
}

void DumpVcpValues(const std::vector<unsigned int>& values)
{
    if (values.empty()) {
        std::cout << ": Empty\n";
    } else {
        std::cout << ": { ";
        for (unsigned int v : values) {
            std::cout << "0x" << std::hex << std::uppercase << std::setw(2) << std::setfill('0') << v << " ";
        }
        std::cout << "}\n" << std::dec;
    }
}

void DumpVcpHexMap(VcpCodeUIntVectorMap& uintMap)
{
    std::cout << "Key   Values\n";
    for (const auto& [key, values] : uintMap) {
        DumpVcpKey(key);
        DumpVcpValues(values);
    }
}

void DumpGetVcpCmdInfo(BYTE vcpCode, const std::pair<DWORD, DWORD>& result, bool isCrossShareVcpCode, const std::string& csCmdInfo)
{
    if (isCrossShareVcpCode) {
        // cross share vcp code
        unsigned char data[4] = {};
        data[0] = (result.first & 0xff00) >> 8;
        data[1] = (result.first & 0xff);
        data[2] = (result.second & 0xff00) >> 8;
        data[3] = (result.second & 0xff);

        std::cout << std::hex << std::uppercase\
              << "CrossShare" << csCmdInfo << "(" << "0x"\
              << std::setw(2) << std::setfill('0') << static_cast<int>(vcpCode) << "): 0x"\
              << std::setw(2) << std::setfill('0') << static_cast<int>(data[0]) << ", 0x"\
              << std::setw(2) << std::setfill('0') << static_cast<int>(data[1]) << ", 0x"\
              << std::setw(2) << std::setfill('0') << static_cast<int>(data[2]) << ", 0x"\
              << std::setw(2) << std::setfill('0') << static_cast<int>(data[3]) << std::endl;
    } else {
        // normal vcp code
        std::cout << "VCP Code " << "0x" << std::hex << (int)vcpCode << " Current Value: 0x" << result.second
                  << " (Max: 0x" << result.first << ")" << std::dec << std::endl;
    }
}

void DumpSetVcpCmdInfo(BYTE vcpCode, DWORD value, bool isCrossShareVcpCode, const std::string& csCmdInfo)
{
    std::string info = isCrossShareVcpCode? (string("set CrossShare ") + csCmdInfo + string(" 0x")) : "Successfully set VCP Code 0x";
    std::cout << info\
                << std::hex << std::uppercase\
                << std::setw(2) << std::setfill('0') << static_cast<int>(vcpCode)\
                << " to " << "0x"\
                << std::setw(2) << std::setfill('0') << static_cast<int>(value)\
                << std::dec << "(" << value << ")" << std::endl;
}

void DumpMacAddr(const std::array<unsigned char, 6>& macAddr)
{
    std::ostringstream oss;
    for (size_t i = 0; i < macAddr.size(); ++i) {
        if (i > 0) oss << ":";
        oss << std::hex << std::setw(2) << std::setfill('0') << static_cast<int>(macAddr[i]);
    }
    std::cout << "mac addr: " << oss.str() << std::dec << std::endl;
}

bool IsValidHex(const std::string& str) {
    if (str.empty()) return false;
    for (char c : str) {
        if (!std::isxdigit(c)) return false;  // check if {0-9, A-F, a-f}
    }
    return true;
}

unsigned int HexStringToUInt(const std::string& hexStr) {
    if (!IsValidHex(hexStr)) {
        throw std::invalid_argument("invalid HEX value: " + hexStr);
    }

    unsigned int value;
    std::stringstream ss;
    ss << std::hex << hexStr;
    ss >> value;
    return value;
}

VcpCodeUIntVectorMap MapTypeConverter(
    const VcpCodeStringVectorMap& stringMap)
{
    VcpCodeUIntVectorMap uintMap;

    for (const auto& [key, values] : stringMap) {
        unsigned int uintKey = HexStringToUInt(key);

        std::vector<unsigned int> uintValues;
        for (const std::string& value : values) {
            uintValues.push_back(HexStringToUInt(value));
        }

        uintMap[uintKey] = uintValues;
    }

    return uintMap;
}

bool ExtractVcpCodeMap(const std::string& capabilities, VcpCodeUIntVectorMap& vcpOpCodeMap)
{
    VcpCodeStringVectorMap vcpStringMap = GetVcpCodeStrMap(capabilities);

    try {
        vcpOpCodeMap = MapTypeConverter(vcpStringMap);
    } catch (const std::exception& e) {
        std::cerr << "error: " << e.what() << std::endl;
        return false;
    }
    return true;
}

bool readIniFile(const std::string& filename, ConfigData& config) {
    std::ifstream file(filename);
    if (!file.is_open()) {
        std::cerr << "can't open file: " << filename << std::endl;
        return false;
    }

    std::string line;
    while (std::getline(file, line)) {
        std::istringstream iss(line);
        std::string key, value;

        if (std::getline(iss, key, '=') && std::getline(iss, value)) {
            if (key == "mac") {
                std::istringstream macStream(value);
                std::string byte;
                std::vector<unsigned char> macBytes;
                while (std::getline(macStream, byte, ':')) {
                    macBytes.push_back(static_cast<unsigned char>(std::stoul(byte, nullptr, 16)));
                }
                if (macBytes.size() == 6) {
                    std::copy(macBytes.begin(), macBytes.end(), config.mac.begin());
                } else {
                    std::cerr << "invalid mac addr: " << value << std::endl;
                    return false;
                }
            } else if (key == "source") {
                config.connectedSrc = static_cast<unsigned char>(std::stoi(value));
            } else if (key == "port") {
                config.connectedPort = static_cast<unsigned char>(std::stoi(value));
            } else if (key == "testRounds") {
                config.testRounds = std::stoi(value);
            } else if (key == "timeOutCriteria") {
                config.timeOutCriteria = std::stod(value);
            }
        }
    }

    file.close();
    return true;
}

ExecutionTimer::ExecutionTimer(const std::string& funcName) : mName(funcName)
{
    mStart = std::chrono::high_resolution_clock::now();
}

ExecutionTimer::~ExecutionTimer()
{
}

double ExecutionTimer::Stop()
{
    auto end = std::chrono::high_resolution_clock::now();
    double durationMs = std::chrono::duration<double, std::milli>(end - mStart).count();
    std::cout << mName << " excution time: " << durationMs << " ms" << std::endl;
    return durationMs;
}
