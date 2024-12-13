#include "MSUtils.h"
#include "MSCommon.h"
#include "MSProgressBar.h"
#include <chrono>

namespace MSUtils
{
    std::chrono::steady_clock::time_point mTimeCntBegin;
    std::chrono::steady_clock::time_point mTmeCntEnd;
    const int kProgressLevel = 5;
    int curProgressLevel = 0;

    void PrintStartDownload()
    {
        curProgressLevel = 0;
        mTimeCntBegin = std::chrono::steady_clock::now();
        DEBUG_LOG("[%s] Start download", __func__);
    }

    void PrintEndDownload()
    {
        mTmeCntEnd = std::chrono::steady_clock::now();
        std::chrono::milliseconds dur = std::chrono::duration_cast<std::chrono::milliseconds> (mTmeCntEnd - mTimeCntBegin);
        DEBUG_LOG("[%s] Download completed. It took %.3fs", __func__, (float)dur.count()/1000);
    }

    void PrintProgress(unsigned long progress, unsigned long size)
    {
        float ret = (float)progress/size*100;
        int level = ret/kProgressLevel;
        if (level >= curProgressLevel) {
            curProgressLevel++;
            DEBUG_LOG("[%s] Download progress: %.1f%% (%lu/%ld)", __func__, (float)ret, progress, size);
            MSProgressBar::SetProgress(ret);
        }
    }
};