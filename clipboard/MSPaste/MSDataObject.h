#ifndef __INCLUDED_MS_DATAOBJECT__
#define __INCLUDED_MS_DATAOBJECT__

#include <ole2.h>
#include <vector>
#include <string>
#include "MSCommon.h"

class MSStream;
class MSImgHandler;
class MSProgressBar;
class MSITransData;

class MSDataObject : public IDataObject, public IStreamObserver
{
public:
    explicit MSDataObject(ClipboardPasteFileCallback& pasteCb, std::vector<FILE_INFO>& fileList);
    explicit MSDataObject(ClipboardPasteFileCallback& pasteCb, IMAGE_INFO& picInfo);
    ~MSDataObject();
    // IUnknown
    STDMETHODIMP QueryInterface(REFIID riid, void **ppv) override;
    STDMETHODIMP_(ULONG) AddRef() override;
    STDMETHODIMP_(ULONG) Release() override;
    // IDataObject
    STDMETHODIMP GetData(FORMATETC *pfe, STGMEDIUM *pmed) override;
    STDMETHODIMP GetDataHere(FORMATETC *pfe, STGMEDIUM *pmed) override;
    STDMETHODIMP QueryGetData(FORMATETC *pfe) override;
    STDMETHODIMP GetCanonicalFormatEtc(FORMATETC *pfeIn, FORMATETC *pfeOut) override;
    STDMETHODIMP SetData(FORMATETC *pfe, STGMEDIUM *pmed, BOOL fRelease) override;
    STDMETHODIMP EnumFormatEtc(DWORD dwDirection, LPENUMFORMATETC *ppefe) override;
    STDMETHODIMP DAdvise(FORMATETC *pfe, DWORD grfAdv, IAdviseSink *pAdvSink, DWORD *pdwConnection) override;
    STDMETHODIMP DUnadvise(DWORD dwConnection) override;
    STDMETHODIMP EnumDAdvise(LPENUMSTATDATA *ppefe) override;
    void WriteFile(BYTE* data, unsigned int size);
    void EventHandle(EVENT_TYPE event);
    void OnStreamEOF() override;

private:
    void Init();
    void ReleaseObj();
    HRESULT CreateHGlobalFromBlob(const void *pvData, SIZE_T cbData, UINT uFlags, HGLOBAL *phglob);
    void SetFORMATETC(FORMATETC *pfe, UINT cf,
                        TYMED tymed = TYMED_HGLOBAL, LONG lindex = -1,
                        DWORD dwAspect = DVASPECT_CONTENT,
                        DVTARGETDEVICE *ptd = NULL);
    bool GetDataIndex(const FORMATETC *pfe, PASTE_TYPE &type, int &idx);
    HRESULT GetFileData(FORMATETC *pfe, STGMEDIUM *pmed, int idx);
    HRESULT GetDibData(FORMATETC *pfe, STGMEDIUM *pmed);
    MSITransData* GetInterface();
    void EventOpenFileErr();
    void EventRecvTimeout();
    void DeinitProgressBar();

    ULONG m_cRef;
    std::vector<FILE_INFO> mFileList;
    IMAGE_INFO mPicInfo;
    std::vector<FORMATETC> mFormats;
    MSStream* mCurStream;
    MSImgHandler* mImgHandler;
    MSProgressBar* mCurProgressBar;
    ClipboardPasteFileCallback& mPasteCb;
};

#endif //__INCLUDED_MS_DATAOBJECT__