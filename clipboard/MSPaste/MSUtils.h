#ifndef __INCLUDED_MS_UTILS__
#define __INCLUDED_MS_UTILS__

namespace MSUtils
{
    void PrintStartDownload();
    void PrintEndDownload();
    void PrintProgress(unsigned long progress, unsigned long size);
};

#endif //__INCLUDED_MS_UTILS__