#include "boost_global_def.h"
#include <cstring>

// https://learn.microsoft.com/zh-cn/cpp/c-runtime-library/reference/localtime-s-localtime32-s-localtime64-s?view=msvc-170
std::tm *localtime_r(const std::time_t *t, std::tm *result)
{
    auto errCode = ::localtime_s(result, t);
    if (errCode != 0) {
        char buffer[1024] = { 0 };
        strerror_s(buffer, sizeof (buffer), errCode);
        fprintf(stderr, "%s\n", buffer);
        return nullptr;
    }
    return result;
}

// https://learn.microsoft.com/zh-cn/previous-versions/3stkd9be(v%3dvs.110)
std::tm *gmtime_r(const std::time_t *t, std::tm *result)
{
    auto errCode = ::gmtime_s(result, t);
    if (errCode != 0) {
        char buffer[1024] = { 0 };
        strerror_s(buffer, sizeof (buffer), errCode);
        fprintf(stderr, "%s\n", buffer);
        return nullptr;
    }
    return result;
}
