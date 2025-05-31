#pragma once
#include "global.h"

namespace sunkang {

class Utils
{
public:
    static uint32_t getCurrentThreadId();
    static std::string getCurrentTimeString();
    static int gettimeofday(struct timeval *tp, void *tzp);

    static std::wstring toUtf16LE(const std::string &utf8_str, bool *ok = nullptr);
    static std::string toUtf8(const std::wstring &wstr, bool *ok = nullptr);
    static std::string createUuid();
    static std::string toHex(const std::string &data, bool uppercase = true);
    static std::string fromHex(const std::string &data);

    static std::optional<std::string> getFileContent(const std::string &fileName);
    static std::optional<std::string> getFileContent(const std::wstring &fileName);
    static uint32_t toIPv4Address(const std::string &ipAddress);
    static std::string fromIPv4Address(uint32_t address);
    static std::string getFileNameByPath(const std::string &filePath);
    static std::wstring stringToWString(const std::string &data);
};

class Logging
{
public:
    Logging(const std::string &fileName, uint32_t fileLine, const std::string &levelStr);
    ~Logging();

    Logging(const Logging &) = delete ;
    Logging(Logging &&) = delete ;
    Logging &operator = (const Logging &) = delete;
    Logging &operator = (Logging &&) = delete;

    template<class T>
    Logging &operator << (const T &val)
    {
        strStream_ << val;
        return *this;
    }
private:
    std::string fileName_;
    uint32_t fileLine_;
    std::stringstream strStream_;
    std::string levelStr_;
};

#define LOG_DEBUG sunkang::Logging(__FILE__, __LINE__, "DEBUG")
#define LOG_INFO sunkang::Logging(__FILE__, __LINE__, "INFO")
#define LOG_WARN sunkang::Logging(__FILE__, __LINE__, "WARN")

}
