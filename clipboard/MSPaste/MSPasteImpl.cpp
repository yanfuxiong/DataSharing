#include "MSPasteImpl.h"
#include "MSDataObject.h"
#include <thread>

MSPasteImpl::MSPasteImpl(ClipboardPasteFileCallback& pasteCb)
 : mPasteType(PASTE_TYPE_UNKNOWN),
   mIsInited(false),
   mFileList({}),
   mPicInfo({}),
   mCurDataObject(NULL),
   mPasteCb(pasteCb)
{
}

MSPasteImpl::~MSPasteImpl()
{
    mFileList.clear();
    ReleaseOle();
    ReleaseObj();
}

bool MSPasteImpl::SetupPasteFile(std::vector<FILE_INFO>& fileList, std::mutex& clipboardMutex, bool& isOleClipboardOperation)
{
    mPasteType = PASTE_TYPE_FILE;
    mFileList = fileList;

    if (mFileList.size() <= 0) {
        DEBUG_LOG("[%s %d] Empty data", __func__, __LINE__);
        return false;
    }

    return InitOle(clipboardMutex, isOleClipboardOperation);
}

bool MSPasteImpl::SetupPasteImage(IMAGE_INFO& picInfo, std::mutex& clipboardMutex, bool& isOleClipboardOperation)
{
    mPasteType = PASTE_TYPE_DIB;
    mPicInfo = picInfo;

    if (mPicInfo.dataSize == 0) {
        DEBUG_LOG("[%s %d] Empty data", __func__, __LINE__);
        return false;
    }

    return InitOle(clipboardMutex, isOleClipboardOperation);
}

bool MSPasteImpl::InitOle(std::mutex& clipboardMutex, bool& isOleClipboardOperation)
{
    if (SUCCEEDED(OleInitialize(NULL))) {
        mIsInited = true;

        {
            std::lock_guard<std::mutex> lock(clipboardMutex);
            isOleClipboardOperation = true;
            DEBUG_LOG("[Paste] Lock ON");
        }
        SetOleClipboard(clipboardMutex, isOleClipboardOperation);
        std::thread worker([&]() {
            std::this_thread::sleep_for(std::chrono::milliseconds(500));
            {
                std::lock_guard<std::mutex> lock(clipboardMutex);
                isOleClipboardOperation = false;
                DEBUG_LOG("[Paste] Lock OFF");
            }
        });

        if (worker.joinable()) {
            worker.detach();
        }

        MSG msg;
        while(mIsInited) {
            BOOL res = GetMessage(&msg, NULL, 0, 0);
            if (res > 0) {
                TranslateMessage(&msg);
                DispatchMessage(&msg);
            } else if (res == 0) {
                break;
            } else {
                break;
            }
        }
        ReleaseOle();
    } else {
        DEBUG_LOG("[%s %d] OLE initialize failed", __func__, __LINE__);
        return false;
    }

    return true;
}

void MSPasteImpl::ReleaseOle()
{
    if (mIsInited) {
        DEBUG_LOG("[%s %d]", __func__, __LINE__);
        OleSetClipboard(NULL);
        OleUninitialize();
        mIsInited = false;
    }
}

void MSPasteImpl::ReleaseObj()
{
    // DEBUG_LOG("[MSPasteImpl][%s %d]", __func__, __LINE__);
    if (mCurDataObject) {
        mCurDataObject->Release();
        mCurDataObject = NULL;
    }
}

bool MSPasteImpl::SetOleClipboard(std::mutex& clipboardMutex, bool& isOleClipboardOperation)
{
    ReleaseObj();
    switch (mPasteType)
    {
        case PASTE_TYPE_FILE:
        {
            // if (mFileList.size() <= 0) {
            //     DEBUG_LOG("[%s %d] Invalid file list", __func__, __LINE__);
            //     return false;
            // }

            // mCurDataObject = new MSDataObject(mPasteCb, mFileList);
        }
            break;
        case PASTE_TYPE_DIB:
        {
            if (mPicInfo.dataSize == 0) {
                DEBUG_LOG("[%s %d] Invalid image info", __func__, __LINE__);
                return false;
            }

            mCurDataObject = new MSDataObject(mPasteCb, mPicInfo, clipboardMutex, isOleClipboardOperation);
        }
            break;
        default:
            DEBUG_LOG("[%s %d] Unknown paste type", __func__, __LINE__);
            break;
    }

    if (mCurDataObject) {
        if (FAILED(OleSetClipboard(mCurDataObject))) {
            DEBUG_LOG("[%s %d] Set paste object failed", __func__, __LINE__);
            return false;
        }

        DEBUG_LOG("[%s %d] Set paste DataObject successfully", __func__, __LINE__);
    }

    return true;
}

void MSPasteImpl::WriteFile(BYTE* data, unsigned int size)
{
    if (!mCurDataObject) {
        DEBUG_LOG("[%s %d] Current DataObject is empty", __func__, __LINE__);
        return;
    }
    mCurDataObject->WriteFile(data, size);
}

void MSPasteImpl::EventHandle(EVENT_TYPE event)
{
    mCurDataObject->EventHandle(event);
}

void MSPasteImpl::CleanClipboard()
{
    ReleaseOle();
}
