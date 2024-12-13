#ifndef __INCLUDED_MS_INTERFACE__
#define __INCLUDED_MS_INTERFACE__

#include <objidl.h>

class MSITransData
{
public:
    virtual ~MSITransData() {};
    virtual void StartDownload() = 0;
    virtual void WriteFile(BYTE* data, unsigned int size) = 0;
    virtual void Cancel() = 0;
};

#endif //__INCLUDED_MS_INTERFACE__