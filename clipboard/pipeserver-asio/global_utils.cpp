#include "global_utils.h"
#include <objbase.h>
#include <iomanip>
#include <filesystem>

namespace sunkang {

uint32_t Utils::getCurrentThreadId()
{
    static thread_local uint32_t g_currentThreadId = ::GetCurrentThreadId();
    return g_currentThreadId;
}

std::string Utils::getCurrentTimeString()
{
    struct timeval time_val;
    gettimeofday(&time_val, nullptr);
    struct tm tm_val;

    time_t seconds_val = static_cast<time_t>(time_val.tv_sec);
    localtime_s(&tm_val, &seconds_val);

    char time_str[64] = { 0 };
    snprintf(time_str, sizeof(time_str), "%4d%02d%02d %02d:%02d:%02d.%06ld",
            tm_val.tm_year + 1900,
            tm_val.tm_mon + 1,
            tm_val.tm_mday,
            tm_val.tm_hour,
            tm_val.tm_min,
            tm_val.tm_sec,
            time_val.tv_usec);
    return time_str;
}

int Utils::gettimeofday(struct timeval *tp, void *tzp)
{
    (void)tzp;

    uint64_t  intervals;
    FILETIME  ft;

    GetSystemTimeAsFileTime(&ft);

    intervals = ((uint64_t)ft.dwHighDateTime << 32) | ft.dwLowDateTime;
    intervals -= 116444736000000000;

    tp->tv_sec = (long)(intervals / 10000000);
    tp->tv_usec = (long)((intervals % 10000000) / 10);
    return (0);
}

std::wstring Utils::toUtf16LE(const std::string &utf8_str, bool *ok)
{
    int wchar_count = ::MultiByteToWideChar(CP_UTF8, 0, utf8_str.c_str(), -1, nullptr, 0);
    if (wchar_count == 0) {
        LOG_WARN << "toUtf16LE:" << ::GetLastError();
        if (ok) {
            *ok = false;
        }
        return {};
    }

    std::wstring wstr(wchar_count - 1, 0);
    ::MultiByteToWideChar(CP_UTF8, 0, utf8_str.c_str(), -1, wstr.data(), wchar_count);
    if (ok) {
        *ok = true;
    }
    return wstr;
}

std::string Utils::toUtf8(const std::wstring &wstr, bool *ok)
{
    int utf8_length = ::WideCharToMultiByte(CP_UTF8, 0, wstr.c_str(), -1, nullptr, 0, nullptr, nullptr);
    if (utf8_length == 0) {
        LOG_WARN << "toUtf8:" << ::GetLastError();
        if (ok) {
            *ok = false;
        }
        return {};
    }

    std::string utf8_str(utf8_length - 1, 0);
    ::WideCharToMultiByte(CP_UTF8, 0, wstr.c_str(), -1, utf8_str.data(), utf8_length, nullptr, nullptr);
    if (ok) {
        *ok = true;
    }
    return utf8_str;
}

std::string Utils::createUuid()
{
    GUID guid_data;
    auto retVal = ::CoCreateGuid(&guid_data);
    if (retVal != S_OK) {
        LOG_WARN << "createUuid:" << ::GetLastError();
        return {};
    }

    std::stringstream str_stream;
    str_stream << std::hex << std::uppercase;
    str_stream << std::setw(8) << std::setfill('0') << guid_data.Data1;
    str_stream << "-";
    str_stream << std::setw(4) << std::setfill('0') << guid_data.Data2;
    str_stream << "-";
    str_stream << std::setw(4) << std::setfill('0') << guid_data.Data3;
    str_stream << "-";
    str_stream << std::setw(2) << std::setfill('0') << (int)guid_data.Data4[0];
    str_stream << std::setw(2) << std::setfill('0') << (int)guid_data.Data4[1];
    str_stream << "-";
    for (int i = 2; i < 8; ++i) {
        str_stream << std::setw(2) << std::setfill('0') << (int)guid_data.Data4[i];
    }
    return str_stream.str();
}

std::string Utils::toHex(const std::string &data, bool uppercase)
{
    static const char *s_upper_hex = "0123456789ABCDEF";
    static const char *s_lower_hex = "0123456789abcdef";

    std::string new_data;
    new_data.reserve(data.size() * 2);
    for (const auto &chr : data) {
        uint8_t left_code = ((chr & 0xF0) >> 4);
        uint8_t right_code = (chr & 0x0F);
        if (uppercase) {
            new_data.push_back(s_upper_hex[left_code]);
            new_data.push_back(s_upper_hex[right_code]);
        } else {
            new_data.push_back(s_lower_hex[left_code]);
            new_data.push_back(s_lower_hex[right_code]);
        }
    }
    return new_data;
}

std::string Utils::fromHex(const std::string &data)
{
    if (data.empty() || data.size() % 2 != 0) {
        return {};
    }
    auto chrToCode = [] (char chr) {
        if (chr >= '0' && chr <= '9') {
            return static_cast<int>(chr - '0');
        }

        if (chr >= 'a' && chr <= 'f') {
            return static_cast<int>(chr - 'a' + 10);
        }

        if (chr >= 'A' && chr <= 'F') {
            return static_cast<int>(chr - 'A' + 10);
        }
        assert(false);
        return 0;
    };
    std::string new_data;
    new_data.reserve(data.size() / 2);

    for (std::size_t index = 0; index < data.size(); index += 2) {
        int left_code = chrToCode(data[index]);
        int right_code = chrToCode(data[index + 1]);
        char value = static_cast<char>(((left_code & 0x0F) << 4) | (right_code & 0x0F));
        new_data.push_back(value);
    }
    return new_data;
}

std::optional<std::string> Utils::getFileContent(const std::string &fileName)
{
    std::ifstream read_file(fileName, std::ios_base::binary);
    if (read_file.is_open() == false) {
        return {};
    }
    std::stringstream data_stream;
    data_stream << read_file.rdbuf();
    return data_stream.str();
}

std::optional<std::string> Utils::getFileContent(const std::wstring &fileName)
{
    return getFileContent(toUtf8(fileName));
}

uint32_t Utils::toIPv4Address(const std::string &ipAddress)
{
    asio::error_code errCode;
    auto address_v4 = asio::ip::make_address_v4(ipAddress, errCode);
    if (errCode) {
        LOG_WARN << errCode.message();
        return 0;
    }
    return address_v4.to_uint();
}

std::string Utils::fromIPv4Address(uint32_t address)
{
    return asio::ip::address_v4(address).to_string();
}

std::string Utils::getFileNameByPath(const std::string &filePath)
{
    do {
        auto pos = filePath.find_last_of('/');
        if (pos == std::string::npos) {
            break;
        }
        return filePath.substr(pos + 1).c_str();
    } while (false);

    do {
        auto pos = filePath.find_last_of('\\');
        if (pos == std::string::npos) {
            break;
        }
        return filePath.substr(pos + 1).c_str();
    } while (false);

    return filePath;
}

std::wstring Utils::stringToWString(const std::string &data)
{
    assert(data.size() % sizeof (wchar_t) == 0);
    return std::wstring(reinterpret_cast<const wchar_t*>(data.data()), data.size() / sizeof (wchar_t));
}


//------------------------------------------------------------------------
Logging::Logging(const std::string &fileName, uint32_t fileLine, const std::string &levelStr)
    : fileName_(fileName)
    , fileLine_(fileLine)
    , levelStr_(levelStr)
{
    strStream_ << Utils::getCurrentTimeString() << " " << Utils::getCurrentThreadId() << " " << levelStr_ << " ";
}

Logging::~Logging()
{
    fileName_ = std::filesystem::path(fileName_).filename().string();
    strStream_ << " - " << fileName_ << ":" << fileLine_;
    fprintf(stderr, "%s\n", strStream_.str().c_str());
    fflush(stderr);
}

}
