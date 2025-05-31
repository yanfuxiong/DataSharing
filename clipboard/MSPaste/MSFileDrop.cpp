#include "MSFileDrop.h"
#include "MSUtils.h"
#include <iostream>
#include <shobjidl.h>
#include <windows.h>

MSFileDrop::MSFileDrop(FileDropCmdCallback& cmdCb):
    mFileList({}),
    mCmdCb(cmdCb),
    mCurProgress(0)
{

}

MSFileDrop::~MSFileDrop()
{

}

bool MSFileDrop::SetupDropFilePath(char* ip, std::vector<FILE_INFO>& fileList)
{
    mFileList = fileList;
    SetupDialog(ip);
    return true;
}

bool MSFileDrop::UpdateProgressBar(unsigned long progress)
{
    // FIXME
    if (mFileList.size() == 0) {
        DEBUG_LOG("[%s %d] Err: File list is empty", __func__, __LINE__);
        return false;
    }
    int64_t fileSize = (int64_t)mFileList[0].fileSizeHigh << 32 | (int64_t)mFileList[0].fileSizeLow;
    mCurProgress += progress;
    return true;
}

void MSFileDrop::DeinitProgressBar()
{
    mCurProgress = 0;
}

void MSFileDrop::SetupDialog(char* ip)
{
    HWND hwnd = GetForegroundWindow();
    if (!hwnd) {
        hwnd = GetConsoleWindow();
        SetForegroundWindow(hwnd);
    }

    int response = MessageBox(
        hwnd,
        TEXT("Do you accept a remote file drop?"),
        TEXT(""),
        MB_OKCANCEL | MB_SETFOREGROUND
    );

    if (!mCmdCb) {
        DEBUG_LOG("[%s %d] Err: callback is null", __func__, __LINE__);
        return;
    }

    if (response == IDOK) {
        std::wstring userSelectedPath;
        PopSelectPathDialog(userSelectedPath);
        if (userSelectedPath.empty()) {
            DEBUG_LOG("[%s %d] Err: userSelectedPath is empty", __func__, __LINE__);
            mCmdCb(ip, (unsigned long)FILE_DROP_REJECT, NULL);
            return;
        }
        if (mFileList.size() == 0) {
            DEBUG_LOG("[%s %d] Err: mFileList is empty", __func__, __LINE__);
            mCmdCb(ip, (unsigned long)FILE_DROP_REJECT, NULL);
            return;
        }
        int length = userSelectedPath.size() + mFileList[0].fileName.size();
        wchar_t* wideCPath = new wchar_t[length]();
        std::wcscat(wideCPath, userSelectedPath.c_str());
        const wchar_t* back_slash = L"\\";
        std::wcscat(wideCPath, back_slash);
        std::wcscat(wideCPath, mFileList[0].fileName.c_str());
        mCmdCb(ip, (unsigned long)FILE_DROP_ACCEPT, wideCPath);
        delete[] wideCPath;
    } else if (response == IDCANCEL) {
        mCmdCb(ip, (unsigned long)FILE_DROP_REJECT, NULL);
    }
}

void MSFileDrop::PopSelectPathDialog(std::wstring &userSelectedPath)
{
    HRESULT hr = CoInitialize(NULL);
    if (SUCCEEDED(hr)) {
        IFileDialog *pFileDialog;
        hr = CoCreateInstance(CLSID_FileOpenDialog, NULL, CLSCTX_INPROC_SERVER, IID_PPV_ARGS(&pFileDialog));

        if (SUCCEEDED(hr)) {
            DWORD options;
            hr = pFileDialog->GetOptions(&options);
            if (SUCCEEDED(hr)) {
                pFileDialog->SetOptions(options | FOS_PICKFOLDERS);

                hr = pFileDialog->Show(NULL);
                if (hr == HRESULT_FROM_WIN32(ERROR_CANCELLED)) {
                    DEBUG_LOG("[%s %d] Cancel file path selection", __func__, __LINE__);
                    pFileDialog->Release();
                    CoUninitialize();
                    return;
                } else if (SUCCEEDED(hr)) {
                    IShellItem *pItem;
                    hr = pFileDialog->GetResult(&pItem);
                    if (SUCCEEDED(hr)) {
                        PWSTR pszFilePath;
                        hr = pItem->GetDisplayName(SIGDN_FILESYSPATH, &pszFilePath);
                        if (SUCCEEDED(hr)) {
                            userSelectedPath = pszFilePath;
                            // std::wcout << L"Selected folder: " << pszFilePath << std::endl;
                            CoTaskMemFree(pszFilePath);
                        }
                        pItem->Release();
                    }
                } else {
                    DEBUG_LOG("[%s %d] Unknown file path selection", __func__, __LINE__);
                    pFileDialog->Release();
                    CoUninitialize();
                    return;
                }
            }
            pFileDialog->Release();
        }
        CoUninitialize();
    }
}