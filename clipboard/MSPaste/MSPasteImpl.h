#ifndef __INCLUDED_MS_PASTE_IMPL__
#define __INCLUDED_MS_PASTE_IMPL__

#include "MSCommon.h"
#include <windef.h>
#include <vector>
#include <mutex>

class MSDataObject;

class MSPasteImpl
{
public:
    MSPasteImpl(ClipboardPasteFileCallback& pasteCb);
    ~MSPasteImpl();
    bool SetupPasteFile(std::vector<FILE_INFO>& fileList, std::mutex& clipboardMutex, bool& isOleClipboardOperation);
    bool SetupPasteImage(IMAGE_INFO& picInfo, std::mutex& clipboardMutex, bool& isOleClipboardOperation);
    void WriteFile(BYTE* data, unsigned int size);
    void EventHandle(EVENT_TYPE event);
    void CleanClipboard();

private:
    bool SetOleClipboard();
    bool InitOle(std::mutex& clipboardMutex, bool& isOleClipboardOperation);
    void ReleaseOle();
    void ReleaseObj();

    PASTE_TYPE mPasteType;
    bool mIsInited;
    std::vector<FILE_INFO> mFileList;
    IMAGE_INFO mPicInfo;
    MSDataObject* mCurDataObject;
    ClipboardPasteFileCallback& mPasteCb;
};

#endif //__INCLUDED_MS_PASTE_IMPL__