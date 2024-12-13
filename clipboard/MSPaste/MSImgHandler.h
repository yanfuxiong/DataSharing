#ifndef __INCLUDED_MS_IMGHANDLER__
#define __INCLUDED_MS_IMGHANDLER__

#include "MSCommon.h"
#include "MSITransData.h"
#include <mutex>

class MSImgHandler : public MSITransData
{
public:
    MSImgHandler(ClipboardPasteFileCallback& pasteCb, DWORD fileSize, IMAGE_HEADER& picHeader, IStreamObserver* pObserver);
    ~MSImgHandler();

    void StartDownload() override;
    void WriteFile(BYTE* data, unsigned int size) override;
    void Cancel() override;
    bool GetImg(HGLOBAL *phglob);

private:
    bool Check();

    std::mutex mMutexRead;
    std::mutex mMutexWrite;
    HGLOBAL mHDib;
    BYTE* mLpDib;
    DWORD mFileSize;
    DWORD mProgressSize;
    ULONG mPosition;
    bool mWaitDataFlag;
    bool mCancel;
    IMAGE_HEADER mPicHeader;

    ClipboardPasteFileCallback& mPasteCb;
    IStreamObserver* mpObserver;
};

#endif //__INCLUDED_MS_IMGHANDLER__