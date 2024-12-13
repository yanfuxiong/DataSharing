#include "MSDataObject.h"
#include "MSStream.h"
#include "MSImgHandler.h"
#include "MSProgressBar.h"
#include <windows.h>
#include <shlobj.h>
#include <iostream>
#include <initguid.h>
#include <objidl.h>
#include <strsafe.h>

MSDataObject::MSDataObject(ClipboardPasteFileCallback& pasteCb, std::vector<FILE_INFO>& fileList)
: mPasteCb(pasteCb),
  mFileList(fileList)
{
    Init();
    FORMATETC format;
    SetFORMATETC(&format, RegisterClipboardFormat(CFSTR_FILEDESCRIPTOR));
    mFormats.push_back(format);

    for (int i = 0; i < fileList.size(); i++) {
        SetFORMATETC(&format, RegisterClipboardFormat(CFSTR_FILECONTENTS), TYMED_ISTREAM, i);
        mFormats.push_back(format);
    }
}

MSDataObject::MSDataObject(ClipboardPasteFileCallback& pasteCb, IMAGE_INFO& picInfo)
: mPasteCb(pasteCb),
  mPicInfo(picInfo)
{
    Init();
    FORMATETC format;
    SetFORMATETC(&format, CF_DIB);
    mFormats.push_back(format);
}

void MSDataObject::Init()
{
    m_cRef = 1;
    mCurStream = NULL;
    mImgHandler = NULL;
    mCurProgressBar = NULL;
}

MSDataObject::~MSDataObject()
{
    DEBUG_LOG("[%s %d]", __func__, __LINE__);
    ReleaseObj();
}

void MSDataObject::ReleaseObj()
{
    // DEBUG_LOG("[MSDataObject][%s %d]", __func__, __LINE__);
    if (mCurStream) {
        mCurStream->Release();
        mCurStream = NULL;
    }
    if (mImgHandler) {
        delete mImgHandler;
        mImgHandler = NULL;
    }
    if (mCurProgressBar) {
        delete mCurProgressBar;
        mCurProgressBar = NULL;
    }
}

STDMETHODIMP MSDataObject::QueryInterface(REFIID riid, void **ppv)
{
    IUnknown *punk = NULL;
    if (riid == IID_IUnknown) {
        punk = static_cast<IUnknown*>(this);
    } else if (riid == IID_IDataObject) {
        punk = static_cast<IDataObject*>(this);
    }

    *ppv = punk;
    if (punk) {
        punk->AddRef();
        return S_OK;
    } else {
        return E_NOINTERFACE;
    }
}

STDMETHODIMP_(ULONG) MSDataObject::AddRef()
{
    return ++m_cRef;
}

STDMETHODIMP_(ULONG) MSDataObject::Release()
{
    ULONG cRef = --m_cRef;
    if (cRef == 0) delete this;
    return cRef;
}

HRESULT MSDataObject::CreateHGlobalFromBlob(const void *pvData, SIZE_T cbData, UINT uFlags, HGLOBAL *phglob)
{
    HGLOBAL hglob = GlobalAlloc(uFlags, cbData);
    if (hglob) {
        void *pvAlloc = GlobalLock(hglob);
        if (pvAlloc) {
            CopyMemory(pvAlloc, pvData, cbData);
            GlobalUnlock(hglob);
        } else {
            GlobalFree(hglob);
            hglob = NULL;
        }
    }
    *phglob = hglob;
    return hglob ? S_OK : E_OUTOFMEMORY;
}

STDMETHODIMP MSDataObject::GetData(FORMATETC *pfe, STGMEDIUM *pmed)
{
    ZeroMemory(pmed, sizeof(*pmed));
    PASTE_TYPE type = PASTE_TYPE_UNKNOWN;
    int idx = -1;
    if (!GetDataIndex(pfe, type, idx)) {
        return DV_E_FORMATETC;
    }

    if (type == PASTE_TYPE_FILE)
    {
        return GetFileData(pfe, pmed, idx);
    }
    else if (type == PASTE_TYPE_DIB)
    {
        return GetDibData(pfe, pmed);
    }

    return DV_E_FORMATETC;
}

HRESULT MSDataObject::GetFileData(FORMATETC *pfe, STGMEDIUM *pmed, int idx)
{
    if (idx == 0) {
        DEBUG_LOG("FILEDESCRIPTOR");

        int fgd_size = sizeof(FILEGROUPDESCRIPTOR) + (mFileList.size() - 1) * sizeof(FILEDESCRIPTOR);
        LPFILEGROUPDESCRIPTOR fgd = (LPFILEGROUPDESCRIPTOR)new BYTE[fgd_size];
        ZeroMemory(fgd, fgd_size);
        fgd->cItems = mFileList.size();
        for (int i = 0; i < mFileList.size(); i++) {
            fgd->fgd[i].dwFlags = FD_FILESIZE;
            fgd->fgd[i].nFileSizeHigh = mFileList[i].fileSizeHigh;
            fgd->fgd[i].nFileSizeLow = mFileList[i].fileSizeLow;

            char cvtChar[MAX_PATH];
            std::wcstombs(cvtChar, mFileList[i].fileName.c_str(), MAX_PATH);
            StringCchCopy(fgd->fgd[i].cFileName, ARRAYSIZE(fgd->fgd[i].cFileName), cvtChar);
        }

        pmed->tymed = TYMED_HGLOBAL;

        HRESULT hr = CreateHGlobalFromBlob(fgd, fgd_size, GMEM_MOVEABLE, &pmed->hGlobal);
        delete[] fgd;
        return hr;
    }
    else if (idx > 0 && idx <= mFileList.size()) {
        DEBUG_LOG("FILECONTENTS");
        DEBUG_LOG("DO file transfer by MSStream");

        pmed->tymed = TYMED_ISTREAM;

        ReleaseObj();
        mCurProgressBar = new MSProgressBar(mFileList[pfe->lindex].fileSizeHigh, mFileList[pfe->lindex].fileSizeLow);
        mFileList[pfe->lindex].pStream = new MSStream(mPasteCb, mFileList[pfe->lindex].fileSizeHigh, mFileList[pfe->lindex].fileSizeLow, this);
        mCurStream = (MSStream*)mFileList[pfe->lindex].pStream;

        MSProgressBar::WaitSetupReady();
        mCurStream->StartDownload();

        pmed->pstm = mFileList[pfe->lindex].pStream;

        return S_OK;
    }

    return DV_E_FORMATETC;
}

HRESULT MSDataObject::GetDibData(FORMATETC *pfe, STGMEDIUM *pmed)
{
    DEBUG_LOG("DIB Data");
    ReleaseObj();
    pmed->tymed = TYMED_HGLOBAL;

    // Note: picture copy should not be allowed over 3.9GB
    if (!mCurProgressBar) {
        mCurProgressBar = new MSProgressBar(0, mPicInfo.dataSize);
    }

    if (!mImgHandler) {
        mImgHandler = new MSImgHandler(mPasteCb, mPicInfo.dataSize, mPicInfo.picHeader, this);
    }

    MSProgressBar::WaitSetupReady();
    if (!mImgHandler->GetImg(&pmed->hGlobal)) {
        DeinitProgressBar();
        return DV_E_FORMATETC;
    }

    DeinitProgressBar();
    return S_OK;
}

STDMETHODIMP MSDataObject::QueryGetData(FORMATETC *pfe)
{
    PASTE_TYPE type = PASTE_TYPE_UNKNOWN;
    int idx = -1;
    return GetDataIndex(pfe, type, idx) ? S_OK : S_FALSE;
}

STDMETHODIMP MSDataObject::GetCanonicalFormatEtc(FORMATETC *pfeIn, FORMATETC *pfeOut)
{
    *pfeOut = *pfeIn;
    pfeOut->ptd = NULL;
    return DATA_S_SAMEFORMATETC;
}

STDMETHODIMP MSDataObject::EnumFormatEtc(DWORD dwDirection, LPENUMFORMATETC *ppefe)
{
    if (dwDirection == DATADIR_GET) {
        return SHCreateStdEnumFmtEtc(mFormats.size(), &mFormats[0], ppefe);
    }
    *ppefe = NULL;
    return E_NOTIMPL;
}

STDMETHODIMP MSDataObject::GetDataHere(FORMATETC *pfe, STGMEDIUM *pmed) { return E_NOTIMPL;}
STDMETHODIMP MSDataObject::SetData(FORMATETC *pfe, STGMEDIUM *pmed, BOOL fRelease) { return E_NOTIMPL;}
STDMETHODIMP MSDataObject::DAdvise(FORMATETC *pfe, DWORD grfAdv, IAdviseSink *pAdvSink, DWORD *pdwConnection) { return OLE_E_ADVISENOTSUPPORTED;}
STDMETHODIMP MSDataObject::DUnadvise(DWORD dwConnection) { return OLE_E_ADVISENOTSUPPORTED;}
STDMETHODIMP MSDataObject::EnumDAdvise(LPENUMSTATDATA *ppefe) {return OLE_E_ADVISENOTSUPPORTED;}

void MSDataObject::SetFORMATETC(FORMATETC *pfe, UINT cf, TYMED tymed, LONG lindex, DWORD dwAspect, DVTARGETDEVICE *ptd)
{
    pfe->cfFormat = (CLIPFORMAT)cf;
    pfe->tymed    = tymed;
    pfe->lindex   = lindex;
    pfe->dwAspect = dwAspect;
    pfe->ptd      = ptd;
}

PASTE_TYPE GetDataType(const FORMATETC *pfe)
{
    if (pfe->cfFormat == CF_DIB) {
        return PASTE_TYPE_DIB;
    } else {
        return PASTE_TYPE_FILE;
    }
}

bool MSDataObject::GetDataIndex(const FORMATETC *pfe, PASTE_TYPE &type, int &idx)
{
    for (int i = 0; i < mFormats.size(); i++) {
        if (pfe->cfFormat == mFormats[i].cfFormat &&
            (pfe->tymed    &  mFormats[i].tymed)   &&
            pfe->dwAspect == mFormats[i].dwAspect &&
            pfe->lindex   == mFormats[i].lindex) {

            type = GetDataType(pfe);
            idx = i;
            return true;
        }
    }
    return false;
}

MSITransData* MSDataObject::GetInterface()
{
    if (mCurStream) {
        return dynamic_cast<MSITransData*>(mCurStream);
    } else if (mImgHandler) {
        return dynamic_cast<MSITransData*>(mImgHandler);
    } else {
        DEBUG_LOG("[%s %d] Unknown class", __func__, __LINE__);
        return NULL;
    }
}

void MSDataObject::WriteFile(BYTE* data, unsigned int size)
{
    MSITransData* pInterface = GetInterface();
    if (!pInterface)
        return;

    pInterface->WriteFile(data, size);
}

void MSDataObject::EventHandle(EVENT_TYPE event)
{
    switch (event)
    {
        case EVENT_TYPE_OPEN_FILE_ERR:
            EventOpenFileErr();
            break;
        case EVENT_TYPE_RECV_TIMEOUT:
            EventRecvTimeout();
            break;
    }
}

void MSDataObject::EventOpenFileErr()
{
    DEBUG_LOG("[%s %d] Open file err", __func__, __LINE__);
    MSITransData* pInterface = GetInterface();
    if (!pInterface)
        return;

    pInterface->Cancel();

    MSProgressBar::SetErrorMsg();
}

void MSDataObject::EventRecvTimeout()
{
    DEBUG_LOG("[%s %d] Receive data timeout", __func__, __LINE__);
    MSITransData* pInterface = GetInterface();
    if (!pInterface)
        return;

    pInterface->Cancel();

    MSProgressBar::SetTransTerm();
}

void MSDataObject::OnStreamEOF()
{
    ReleaseObj();
}

void MSDataObject::DeinitProgressBar()
{
    if (mCurProgressBar) {
        delete mCurProgressBar;
        mCurProgressBar = NULL;
    }
}
