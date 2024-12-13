#ifndef __INCLUDED_MS_PIPE_UTILS__
#define __INCLUDED_MS_PIPE_UTILS__
#include <string>

namespace MSUtils
{
    template <typename T>
    void intToBigEndianBytes(T value, unsigned char* buffer);
    template <typename T>
    void bigEndianBytesToInt(unsigned char* buffer, T* value);
    std::string ConvertIp2Str(unsigned char* content);
    void ConvertIp2Bytes(char* ip, unsigned char* ipBytes);
    std::string ConvertId2Str(unsigned char* content);
};

#endif //__INCLUDED_MS_PIPE_UTILS__