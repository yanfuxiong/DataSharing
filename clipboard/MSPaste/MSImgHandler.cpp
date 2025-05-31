#include "MSImgHandler.h"
#include "MSCommon.h"
#include "MSUtils.h"

const unsigned int kDataShift = sizeof(BITMAPINFO);

MSImgHandler::MSImgHandler(ClipboardPasteFileCallback& pasteCb, DWORD fileSize, IMAGE_HEADER& picHeader, IStreamObserver* pObserver)
 : mHDib(NULL),
   mLpDib(NULL),
   mFileSize(fileSize),
   mProgressSize(0),
   mPosition(0),
   mWaitDataFlag(false),
   mCancel(false),
   mPasteCb(pasteCb),
   mpObserver(pObserver),
   mPicHeader(picHeader)
{
}

MSImgHandler::~MSImgHandler()
{
    // DEBUG_LOG("[%s %d]", __func__, __LINE__);
    if (mHDib) {
        GlobalFree(mHDib);
    }
}

bool MSImgHandler::GetImg(HGLOBAL *phglob)
{
    if (mHDib != NULL) {
        *phglob = mHDib;
        return true;
    }

    unsigned long totalSize = sizeof(BITMAPINFO) + mFileSize;

    mHDib = GlobalAlloc(GMEM_MOVEABLE, totalSize);
    if (mHDib == NULL) {
        return false;
    }
    mLpDib = static_cast<BYTE*>(GlobalLock(mHDib));
    if (mLpDib == NULL) {
        GlobalFree(mHDib);
        return false;
    }
    BITMAPINFO bmpInfo = {0};
    bmpInfo.bmiHeader.biSize = sizeof(BITMAPINFOHEADER);
    bmpInfo.bmiHeader.biWidth = mPicHeader.width;
    bmpInfo.bmiHeader.biHeight = mPicHeader.height;
    bmpInfo.bmiHeader.biPlanes = mPicHeader.planes;
    bmpInfo.bmiHeader.biBitCount = mPicHeader.bitCount;
    bmpInfo.bmiHeader.biCompression = mPicHeader.compression;
    memcpy(mLpDib, &bmpInfo, sizeof(BITMAPINFO));

    StartDownload();
    while(1) {
        if (!Check()) {
            break;
        }
    }

    GlobalUnlock(mHDib);
    *phglob = mHDib;
    return true;
}

bool MSImgHandler::Check()
{
    std::lock_guard<std::mutex> guard(mMutexRead);

    if (mPosition >= mFileSize) {
        MSUtils::PrintEndDownload();
        return false;
    }

    while (!mWaitDataFlag) {}

    if (mCancel) {
        return false;
    } else {
        mWaitDataFlag = false;
        mPosition = mProgressSize;
        // DEBUG_LOG("[%s %d] mPosition=%d", __func__, __LINE__, mPosition);
        return true;
    }
}

void MSImgHandler::StartDownload()
{
    // TODO: discuss callback param
    MSUtils::PrintStartDownload();
    mPasteCb((char*)"");
}

void MSImgHandler::WriteFile(BYTE* data, unsigned int size)
{
    std::lock_guard<std::mutex> guard(mMutexWrite);
    // DEBUG_LOG("[%s %d] Receive data and write to file. size: %u", __func__, __LINE__, size);
    if (!mLpDib) {
        DEBUG_LOG("[%s %d] Invalid Buffer", __func__, __LINE__);
        return;
    }
    if (!data) {
        DEBUG_LOG("[%s %d] Invalid data", __func__, __LINE__);
        return;
    }
    if (size == 0) {
        DEBUG_LOG("[%s %d] Invalid file size", __func__, __LINE__);
        return;
    }
    if (mProgressSize + size > mFileSize) {
        DEBUG_LOG("[%s %d] The file size exceeds. %ld/%ld", __func__, __LINE__, mProgressSize, mFileSize);
        return;
    }

    memcpy(mLpDib+kDataShift+mProgressSize, data, size);
    mProgressSize += size;
    MSUtils::PrintProgress(mProgressSize, mFileSize);
    mWaitDataFlag = true;
}

void MSImgHandler::Cancel()
{
    mCancel = true;
    mWaitDataFlag = true;
}
