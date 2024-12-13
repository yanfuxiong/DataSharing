#include "MSStream.h"
#include "MSCommon.h"
#include "MSUtils.h"
#include <iostream>

MSStream::MSStream(ClipboardPasteFileCallback& pasteCb, DWORD fileSizeHigh, DWORD fileSizeLow, IStreamObserver* pObserver)
 : m_cRef(1),
   mData(NULL),
   m_pBuffer(NULL),
   mFileSize(0),
   mProgressSize(0),
   mPosition(0),
   mWaitDataFlag(false),
   mCancel(false),
   mPasteCb(pasteCb),
   mpObserver(pObserver)
{
    if (InitFile(fileSizeHigh, fileSizeLow) != S_OK) {
        DEBUG_LOG("[%s %d] Err: Init File failed", __func__, __LINE__);
    }
}

MSStream::~MSStream()
{
    // DEBUG_LOG("[%s %d]", __func__, __LINE__);
    if (m_pBuffer) {
        UnmapViewOfFile(m_pBuffer);
    }
    if (mData) {
        CloseHandle(mData);
    }
}

HRESULT MSStream::InitFile(DWORD fileSizeHigh, DWORD fileSizeLow)
{
    mData = ::CreateFileMappingW(INVALID_HANDLE_VALUE, NULL, PAGE_READWRITE, fileSizeHigh, fileSizeLow, NULL);
    if (mData == NULL) {
        return HRESULT_FROM_WIN32(GetLastError());
    }

    m_pBuffer = static_cast<BYTE*>(MapViewOfFile(mData, FILE_MAP_WRITE, 0, 0, 0));
    if (m_pBuffer == NULL) {
        CloseHandle(mData);
        return HRESULT_FROM_WIN32(GetLastError());
    }

    // TODO: consider large file with fileSizeHigh
    mFileSize = fileSizeLow;

    return S_OK;
}

STDMETHODIMP MSStream::QueryInterface(REFIID riid, void **ppv)
{
    IUnknown *punk = NULL;
    if (riid == IID_IUnknown) {
        punk = static_cast<IUnknown*>(this);
    } else if (riid == IID_IStream) {
        punk = static_cast<IStream*>(this);
    }

    *ppv = punk;
    if (punk) {
        punk->AddRef();
        return S_OK;
    } else {
        return E_NOINTERFACE;
    }
}

STDMETHODIMP_(ULONG) MSStream::AddRef()
{
    return ++m_cRef;
}

STDMETHODIMP_(ULONG) MSStream::Release()
{
    ULONG cRef = --m_cRef;
    if (cRef == 0) delete this;
    return cRef;
}

HRESULT STDMETHODCALLTYPE MSStream::Read(void* pv, ULONG cb, ULONG* pcbRead)
{
    std::lock_guard<std::mutex> guard(mMutexRead);
    if (!pcbRead) {
        DEBUG_LOG("[%s %d] Invalid pcbRead", __func__, __LINE__);
        return S_FALSE;
    }

    if (mPosition >= mFileSize) {
        *pcbRead = 0;
        MSUtils::PrintEndDownload();
        if (mpObserver) {
            mpObserver->OnStreamEOF();
        }
        return S_FALSE;
    }

    while (!mWaitDataFlag) {}

    mWaitDataFlag = false;

    if (mCancel) {
        if (mpObserver) {
            mpObserver->OnStreamEOF();
        }
        return S_FALSE;
    } else {
        memcpy(pv, (m_pBuffer + mPosition), (mProgressSize - mPosition));
        *pcbRead = mProgressSize - mPosition;
        mPosition = mProgressSize;
        // DEBUG_LOG("[%s %d] mPosition=%lu, cb=%lu, pcbRead=%lu", __func__, __LINE__, mPosition, cb, *pcbRead);
        return S_OK;
    }
}

HRESULT STDMETHODCALLTYPE MSStream::Seek(LARGE_INTEGER dlibMove, DWORD dwOrigin, ULARGE_INTEGER* plibNewPosition) { return E_NOTIMPL;}
HRESULT STDMETHODCALLTYPE MSStream::Write(const void* pv, ULONG cb, ULONG* pcbWritten) { return E_NOTIMPL;}
HRESULT STDMETHODCALLTYPE MSStream::SetSize(ULARGE_INTEGER libNewSize) { return E_NOTIMPL;}
HRESULT STDMETHODCALLTYPE MSStream::CopyTo(IStream* pstm, ULARGE_INTEGER cb, ULARGE_INTEGER* pcbRead, ULARGE_INTEGER* pcbWritten) { return E_NOTIMPL;}
HRESULT STDMETHODCALLTYPE MSStream::Commit(DWORD grfCommitFlags) { return E_NOTIMPL;}
HRESULT STDMETHODCALLTYPE MSStream::Revert() { return E_NOTIMPL;}
HRESULT STDMETHODCALLTYPE MSStream::LockRegion(ULARGE_INTEGER libOffset, ULARGE_INTEGER cb, DWORD dwLockType) { return E_NOTIMPL;}
HRESULT STDMETHODCALLTYPE MSStream::UnlockRegion(ULARGE_INTEGER libOffset, ULARGE_INTEGER cb, DWORD dwLockType) { return E_NOTIMPL;}
HRESULT STDMETHODCALLTYPE MSStream::Stat(STATSTG* pstatstg, DWORD grfStatFlag) { return E_NOTIMPL;}
HRESULT STDMETHODCALLTYPE MSStream::Clone(IStream** ppstm) { return E_NOTIMPL;}

void MSStream::StartDownload()
{
    // TODO: discuss callback param
    MSUtils::PrintStartDownload();
    mPasteCb((char*)"");
}

void MSStream::WriteFile(BYTE* data, unsigned int size)
{
    std::lock_guard<std::mutex> guard(mMutexWrite);
    DEBUG_LOG("[%s %d] Receive data and write to file. size: %u", __func__, __LINE__, size);
    if (!m_pBuffer) {
        DEBUG_LOG("[%s %d] Invalid Buffer", __func__, __LINE__);
        return;
    }
    if (!data) {
        DEBUG_LOG("[%s %d] Invalid data", __func__, __LINE__);
        return;
    }
    if (size == 0) {
        DEBUG_LOG("[%s %d] Invalid file size=%d", __func__, __LINE__, size);
        return;
    }
    if (mProgressSize + size > mFileSize) {
        DEBUG_LOG("[%s %d] The file size exceeds. %ld/%ld", __func__, __LINE__, mProgressSize, mFileSize);
        return;
    }

    memcpy(m_pBuffer+mProgressSize, data, size);
    mProgressSize += size;
    mWaitDataFlag = true;
    MSUtils::PrintProgress(mProgressSize, mFileSize);
}

void MSStream::Cancel()
{
    mCancel = true;
    mWaitDataFlag = true;
}
