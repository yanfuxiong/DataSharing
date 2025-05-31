#ifndef __INCLUDED_MS_FILE_DROP_IMPL__
#define __INCLUDED_MS_FILE_DROP_IMPL__

#include "MSCommon.h"
#include <windef.h>
#include <vector>
#include <mutex>

class MSFileDrop
{
public:
    MSFileDrop(FileDropCmdCallback& cmdCb);
    ~MSFileDrop();
    bool SetupDropFilePath(char* ip, std::vector<FILE_INFO>& fileList);
    bool UpdateProgressBar(unsigned long progress);
    void DeinitProgressBar();

private:
    void SetupDialog(char* ip);
    void PopSelectPathDialog(std::wstring &userSelectedPath);

    std::vector<FILE_INFO> mFileList;
    FileDropCmdCallback& mCmdCb;
    unsigned int mCurProgress;
};

#endif //__INCLUDED_MS_FILE_DROP_IMPL__