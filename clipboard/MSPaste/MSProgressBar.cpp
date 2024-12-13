#include "MSProgressBar.h"
#include "MSCommon.h"
#include "MSUtils.h"

#define WM_USER_UPDATE_PROGRESS (WM_USER + 1)
#define WM_USER_TRANS_ERR       (WM_USER + 2)
#define WM_USER_TRANS_TERM      (WM_USER + 3)
static HWND g_win = NULL;
static HWND g_progressBar = NULL;
static HWND g_msgLabel = NULL;
static HWND g_btn = NULL;
static std::mutex g_Mutex;

std::mutex MSProgressBar::m_Mutex;
std::condition_variable MSProgressBar::m_Cv;
bool MSProgressBar::m_SetupReady = false;
LRESULT CALLBACK WndProgressProc(HWND hwnd, UINT message, WPARAM wParam, LPARAM lParam) {
    switch (message) {
    case WM_CREATE:
        DEBUG_LOG("++ [%s %d][WM_CREATE] ++", __func__, __LINE__);
        g_progressBar = CreateWindowEx(0, PROGRESS_CLASS, NULL, WS_CHILD | WS_VISIBLE | PBS_SMOOTH,
            10, 30, 260, 20, hwnd, NULL,
            ((LPCREATESTRUCT)lParam)->hInstance, NULL);
        if (g_progressBar == NULL) {
            DEBUG_LOG("-- [%s %d][WM_CREATE]1 --", __func__, __LINE__);
            break;
        }

        SendMessage(g_progressBar, PBM_SETRANGE, 0, MAKELPARAM(0, 100));
        SendMessage(g_progressBar, PBM_SETPOS, 0, 0);

        g_msgLabel = CreateWindowEx(0, "Warning", "", WS_CHILD | SS_CENTER,
            10, 30, 260, 20, hwnd, NULL,
            ((LPCREATESTRUCT)lParam)->hInstance, NULL);
        if (g_msgLabel == NULL) {
            DEBUG_LOG("-- [%s %d][WM_CREATE]2 --", __func__, __LINE__);
            break;
        }
        DEBUG_LOG("-- [%s %d][WM_CREATE] --", __func__, __LINE__);
        break;
    case WM_USER_UPDATE_PROGRESS:
    {
        std::lock_guard<std::mutex> lock(g_Mutex);
        DEBUG_LOG("++ [%s %d][WM_USER_UPDATE_PROGRESS] ++", __func__, __LINE__);
        DEBUG_LOG("[%s %d] WM_USER_UPDATE_PROGRESS, progress=%llu", __func__, __LINE__, wParam);
        if (g_progressBar && IsWindow(g_progressBar)) {
            SendMessage(g_progressBar, PBM_SETPOS, wParam, 0);
        }
        DEBUG_LOG("-- [%s %d][WM_USER_UPDATE_PROGRESS] --", __func__, __LINE__);
    }
        break;
    case WM_USER_TRANS_ERR:
    {
        DEBUG_LOG("++ [%s %d][WM_USER_TRANS_ERR] ++", __func__, __LINE__);
        ShowWindow(g_progressBar, SW_HIDE);
        ShowWindow(g_win, SW_HIDE);
        int result = MessageBox(hwnd, "File cannot be copied.\nPlease check if the file is currently in use.", "Warning",
            MB_OK | MB_ICONERROR);
        if (result == IDOK) {
            SendMessage(hwnd, WM_CLOSE, 0, 0);
        }
        DEBUG_LOG("-- [%s %d][WM_USER_TRANS_ERR] --", __func__, __LINE__);
    }
        break;
    case WM_USER_TRANS_TERM:
    {
        DEBUG_LOG("++ [%s %d][WM_USER_TRANS_TERM] ++", __func__, __LINE__);
        // TODO: create cancel button and re-send the action to cgo
        // g_btn = CreateWindowEx(0, "BUTTON", "Cancel", WS_CHILD | WS_VISIBLE,
        //     10, 60, 80, 30, hwnd, (HMENU)1,
        //     ((LPCREATESTRUCT)lParam)->hInstance, NULL);
        ShowWindow(g_progressBar, SW_HIDE);
        ShowWindow(g_win, SW_HIDE);
        int result = MessageBox(hwnd, "Transmission failed: Network error\nPlease check your network connection and retry.", "Warning",
            MB_OK | MB_ICONERROR);
        if (result == IDOK) {
            SendMessage(hwnd, WM_CLOSE, 0, 0);
        }
        DEBUG_LOG("-- [%s %d][WM_USER_TRANS_TERM] --", __func__, __LINE__);
    }
        break;
    case WM_DESTROY:
        DEBUG_LOG("++ [%s %d][WM_DESTROY] ++", __func__, __LINE__);
        PostQuitMessage(0);
        g_progressBar = NULL;
        g_msgLabel = NULL;
        DEBUG_LOG("-- [%s %d][WM_DESTROY] --", __func__, __LINE__);
        break;
    case WM_CLOSE:
        DEBUG_LOG("++ [%s %d][WM_CLOSE] ++", __func__, __LINE__);
        DestroyWindow(hwnd);
        DEBUG_LOG("-- [%s %d][WM_CLOSE] --", __func__, __LINE__);
        break;
    default:
        return DefWindowProc(hwnd, message, wParam, lParam);
    }
    return 0;
}

MSProgressBar::MSProgressBar(DWORD filesizeHigh, DWORD filesizeLow)
{
    DEBUG_LOG("++ [%s %d] ++", __func__, __LINE__);
    m_Thread = std::thread(&MSProgressBar::Init, this);
    MSUtils::PrintStartDownload();
    DEBUG_LOG("-- [%s %d] --", __func__, __LINE__);
}

MSProgressBar::~MSProgressBar()
{
    DEBUG_LOG("++ [%s %d] ++", __func__, __LINE__);
    std::lock_guard<std::mutex> lock(g_Mutex);
    MSUtils::PrintEndDownload();
    if (g_win) {
        {
            std::lock_guard<std::mutex> lock(m_Mutex);
            m_SetupReady = false;
        }
        SendMessage(g_win, WM_CLOSE, 0, 0);
        g_win = NULL;
        DEBUG_LOG("[%s %d] Destroy windows Win", __func__, __LINE__);
    }
    if (m_Thread.joinable()) {
        DEBUG_LOG("++ [%s %d] Thread join ++", __func__, __LINE__);
        m_Thread.join();
        DEBUG_LOG("-- [%s %d] Thread join --", __func__, __LINE__);
    } else {
        DEBUG_LOG("[%s %d] Err: Thread cannot join", __func__, __LINE__);
    }
    DEBUG_LOG("-- [%s %d] --", __func__, __LINE__);
}

bool MSProgressBar::Init()
{
    DEBUG_LOG("++ [%s %d] ++", __func__, __LINE__);
    {
        std::lock_guard<std::mutex> lock(m_Mutex);
        m_SetupReady = true;
        HINSTANCE hInstance = GetModuleHandle(NULL);
        WNDCLASSEX wc = { sizeof(WNDCLASSEX), CS_HREDRAW | CS_VREDRAW, WndProgressProc, 0L, 0L,
                        hInstance, LoadIcon(NULL, IDI_APPLICATION), LoadCursor(NULL, IDC_ARROW),
                        (HBRUSH)(COLOR_WINDOW + 1), NULL, TEXT("ProgressWindow"), NULL };
        RegisterClassEx(&wc);

        int windowWidth = 350;
        int windowHeight = 100;

        int screenWidth = GetSystemMetrics(SM_CXSCREEN);
        int screenHeight = GetSystemMetrics(SM_CYSCREEN);

        int xPos = (screenWidth - windowWidth) / 2;
        int yPos = (screenHeight - windowHeight) / 2;

        g_win = CreateWindow(TEXT("ProgressWindow"), TEXT("Copy..."),
            WS_OVERLAPPED | WS_CAPTION, xPos, yPos,
            windowWidth, windowHeight, NULL, NULL, hInstance, NULL);

        if (g_win == NULL) {
            DEBUG_LOG("-- [%s %d] --", __func__, __LINE__);
            return false;
        }

        ShowWindow(g_win, SW_SHOW);
        SetForegroundWindow(g_win);
        SetWindowPos(g_win, HWND_TOPMOST, 0, 0, 0, 0, SWP_NOMOVE | SWP_NOSIZE);
        UpdateWindow(g_win);

        m_Cv.notify_one();
    }

    MSG msg = { 0 };
    while (GetMessage(&msg, NULL, 0, 0)) {
        TranslateMessage(&msg);
        DispatchMessage(&msg);
    }

    DEBUG_LOG("-- [%s %d] --", __func__, __LINE__);
    return true;
}

void MSProgressBar::WaitSetupReady()
{
    DEBUG_LOG("++ [%s %d] ++", __func__, __LINE__);
    std::unique_lock<std::mutex> lock(m_Mutex);
    m_Cv.wait(lock, [] { return m_SetupReady; });
    DEBUG_LOG("-- [%s %d] --", __func__, __LINE__);
}

void MSProgressBar::SetProgress(int progress)
{
    DEBUG_LOG("++ [%s %d] ++", __func__, __LINE__);
    WaitSetupReady();
    if (g_win) {
        DEBUG_LOG("[%s %d] progress:%d", __func__, __LINE__, progress);
        PostMessage(g_win, WM_USER_UPDATE_PROGRESS, progress, 0);
    } else {
        DWORD err = GetLastError();
        DEBUG_LOG("[%s %d] gwin not exist: %lu", __func__, __LINE__, err);
    }
    //SendMessage(hwndEdit, WM_PAINT, 0, (LPARAM)RGB(0, 255, 0));
    DEBUG_LOG("-- [%s %d] --", __func__, __LINE__);
}

void MSProgressBar::SetErrorMsg()
{
    DEBUG_LOG("++ [%s %d] ++", __func__, __LINE__);
    if (g_win) {
        PostMessage(g_win, WM_USER_TRANS_ERR, 0, 0);
    }
    DEBUG_LOG("-- [%s %d] --", __func__, __LINE__);
}

void MSProgressBar::SetTransTerm()
{
    DEBUG_LOG("++ [%s %d] ++", __func__, __LINE__);
    if (g_win) {
        PostMessage(g_win, WM_USER_TRANS_TERM, 0, 0);
    }
    DEBUG_LOG("-- [%s %d] --", __func__, __LINE__);
}
