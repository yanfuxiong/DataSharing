#include "hidden_message_window.h"

HiddenMessageWindow::HiddenMessageWindow()
    : m_hwnd(nullptr)
{

}

HiddenMessageWindow::~HiddenMessageWindow()
{
    if (m_hwnd) {
        DestroyWindow(m_hwnd);
        UnregisterClass("HiddenMessageWindowClass", GetModuleHandle(nullptr));
    }
}

bool HiddenMessageWindow::create()
{
    WNDCLASSEX wc = {};
    wc.cbSize = sizeof(WNDCLASSEX);
    wc.lpfnWndProc = WindowProc;
    wc.hInstance = GetModuleHandle(nullptr);
    wc.lpszClassName = "HiddenMessageWindowClass";

    if (!RegisterClassEx(&wc)) {
        qCritical() << "Failed to register window class:" << GetLastError();
        return false;
    }

    m_hwnd = CreateWindowEx(
        0,
        "HiddenMessageWindowClass",
        "HiddenMessageWindow",
        0,
        CW_USEDEFAULT, CW_USEDEFAULT,
        CW_USEDEFAULT, CW_USEDEFAULT,
        nullptr,
        nullptr,
        GetModuleHandle(nullptr),
        this
        );

    if (!m_hwnd) {
        qCritical() << "Failed to create window:" << GetLastError();
        return false;
    }

    SetWindowLongPtr(m_hwnd, GWLP_USERDATA, reinterpret_cast<LONG_PTR>(this));
    qInfo() << "Hidden message window created successfully";
    return true;
}

bool HiddenMessageWindow::handleMessage(UINT msg, WPARAM wParam, LPARAM lParam)
{
    Q_UNUSED(wParam)
    Q_UNUSED(lParam)
    switch (msg) {
    case WM_QUIT: {
        qInfo() << "Received WM_QUIT message, exit the service";
        _Exit(1);
        return true;
    }
    case WM_ENDSESSION: {
        qInfo() << "Received WM_ENDSESSION message, exit the service";
        _Exit(1);
        return true;
    }
    default:
        return false;
    }
}

LRESULT HiddenMessageWindow::WindowProc(HWND hwnd, UINT uMsg, WPARAM wParam, LPARAM lParam)
{
    HiddenMessageWindow* pThis = reinterpret_cast<HiddenMessageWindow*>(GetWindowLongPtr(hwnd, GWLP_USERDATA));
    if (uMsg == WM_NCCREATE) {
        CREATESTRUCT* pCreate = reinterpret_cast<CREATESTRUCT*>(lParam);
        pThis = reinterpret_cast<HiddenMessageWindow*>(pCreate->lpCreateParams);
        SetWindowLongPtr(hwnd, GWLP_USERDATA, reinterpret_cast<LONG_PTR>(pThis));
    }

    if (pThis && pThis->handleMessage(uMsg, wParam, lParam)) {
        return 0;
    }
    return DefWindowProc(hwnd, uMsg, wParam, lParam);
}
