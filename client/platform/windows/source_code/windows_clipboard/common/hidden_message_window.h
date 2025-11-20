#pragma once
#include "common_utils.h"
#include <windows.h>

class HiddenMessageWindow
{
public:
    HiddenMessageWindow();
    ~HiddenMessageWindow();

    bool create();
    HWND getHandle() const { return m_hwnd; }
    virtual bool handleMessage(UINT msg, WPARAM wParam, LPARAM lParam);

private:
    static LRESULT CALLBACK WindowProc(HWND hwnd, UINT uMsg, WPARAM wParam, LPARAM lParam);
    HWND m_hwnd;
};
