#ifndef __INCLUDED_MS_PIPE_CONNECTION__
#define __INCLUDED_MS_PIPE_CONNECTION__

#include <windows.h>
#include <atomic>
#include <mutex>

class IMSPipeConnection
{
public:
    IMSPipeConnection(const std::atomic<bool>& running_pipe);
    virtual ~IMSPipeConnection();
    virtual void onConnected() = 0;
    virtual void onDisconnected() = 0;
    virtual void onReadData(unsigned char* data, unsigned int length) = 0;

protected:
    void Connect();
    void Disconnect();
    void ReadDataThread();
    void SendData(unsigned char* data, unsigned int length);

private:
    static DWORD WINAPI CreatePipeServerThreadStatic(LPVOID lpParam);
    DWORD WINAPI CreatePipeServerThread(LPVOID lpParam);

    const std::atomic<bool>& g_running_pipe;
    HANDLE m_event_pipe;
    HANDLE m_thread_pipe;
    HANDLE m_hPipe;
    std::mutex m_pipeMutex;
};

#endif //__INCLUDED_MS_PIPE_CONNECTION__