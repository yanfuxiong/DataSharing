#include "IMSPipeConnection.h"
#include "MSPipeCommon.h"
#include <thread>

using namespace std;

IMSPipeConnection::IMSPipeConnection(const std::atomic<bool>& running_pipe
                                        ) : g_running_pipe(running_pipe),
                                            m_event_pipe(NULL),
                                            m_thread_pipe(NULL),
                                            m_hPipe(NULL)
{
}

IMSPipeConnection::~IMSPipeConnection()
{
}

void IMSPipeConnection::Connect()
{
    m_event_pipe = CreateEvent(NULL, TRUE, FALSE, NULL);
    m_thread_pipe = CreateThread(NULL, 0, CreatePipeServerThreadStatic, this, 0, NULL);

    WaitForSingleObject(m_event_pipe, INFINITE);
    CloseHandle(m_event_pipe);
}

void IMSPipeConnection::Disconnect()
{
    WaitForSingleObject(m_thread_pipe, INFINITE);
    CloseHandle(m_thread_pipe);
    m_thread_pipe = NULL;
}

DWORD WINAPI IMSPipeConnection::CreatePipeServerThreadStatic(LPVOID lpParam)
{
    IMSPipeConnection* pConnection = static_cast<IMSPipeConnection*>(lpParam);
    return pConnection->CreatePipeServerThread(lpParam);
}

DWORD WINAPI IMSPipeConnection::CreatePipeServerThread(LPVOID lpParam)
{
    while(g_running_pipe.load()) {
        std::unique_lock<std::mutex> lock(m_pipeMutex);

        if (m_hPipe != NULL) {
            DEBUG_LOG("[%s %d] Create redundantly", __func__, __LINE__);
            SetEvent(m_event_pipe);
            return 1;
        }

        m_hPipe = CreateNamedPipe(RTK_PIPE_NAME,
                                    PIPE_ACCESS_DUPLEX | FILE_FLAG_OVERLAPPED,
                                    PIPE_TYPE_BYTE | PIPE_READMODE_BYTE | PIPE_WAIT,
                                    PIPE_UNLIMITED_INSTANCES,
                                    RTK_PIPE_BUFF_SIZE, RTK_PIPE_BUFF_SIZE, 0, NULL);
        if (m_hPipe == INVALID_HANDLE_VALUE) {
            DEBUG_LOG("[%s %d] Err: creating pipe: %lu", __func__, __LINE__, GetLastError());
            SetEvent(m_event_pipe);
            return 1;
        }

        DEBUG_LOG("[PIPE] Create successfully");
        DEBUG_LOG("[PIPE] Waiting for client to connect...");
        SetEvent(m_event_pipe);
        lock.unlock();

        if (ConnectNamedPipe(m_hPipe, NULL) != FALSE) {
            onConnected();
            std::thread reader(&IMSPipeConnection::ReadDataThread, this);
            reader.join();
        } else {
            DWORD error = GetLastError();
            if (error == ERROR_PIPE_CONNECTED) {
                DEBUG_LOG("[%s %d] Client already connected pipe", __func__, __LINE__);
                onConnected();
                std::thread reader(&IMSPipeConnection::ReadDataThread, this);
                reader.join();
            } else {
                DEBUG_LOG("[%s %d] Err: Connection failed", __func__, __LINE__);
            }
        }

        lock.lock();
        onDisconnected();
        DisconnectNamedPipe(m_hPipe);
        CloseHandle(m_hPipe);
        m_hPipe = NULL;
        lock.unlock();
    }

    return 0;
}

void DumpError(int errCode)
{
    switch (errCode)
    {
        case ERROR_BROKEN_PIPE:
            DEBUG_LOG("[%s %d] Err:%d Pipe disconnected", __func__, __LINE__, errCode);
            break;
        case ERROR_NO_DATA:
            DEBUG_LOG("[%s %d] Err:%d No data available in non-blocking mode", __func__, __LINE__, errCode);
            break;
        case ERROR_MORE_DATA:
            DEBUG_LOG("[%s %d] Err:%d Data oversize", __func__, __LINE__, errCode);
            break;
        default:
            DEBUG_LOG("[%s %d] Err:%d Other failed", __func__, __LINE__, errCode);
            break;
    }
}

void IMSPipeConnection::ReadDataThread()
{
    OVERLAPPED overlapped = {0};
    overlapped.hEvent = CreateEvent(NULL, TRUE, FALSE, NULL);

    unsigned char buffer[RTK_PIPE_BUFF_SIZE] = {0};
    DWORD bytesRead = 0;

    while(g_running_pipe.load()) {
        std::unique_lock<std::mutex> lock(m_pipeMutex);

        if (m_hPipe == NULL) {
            lock.unlock();
            break;
        }

        bool success = ReadFile(m_hPipe, buffer, RTK_PIPE_BUFF_SIZE, &bytesRead, &overlapped);
        lock.unlock();

        if (!success || bytesRead == 0) {
            DWORD error = GetLastError();
            if (error == ERROR_IO_PENDING) {
                DWORD waitResult = WaitForSingleObject(overlapped.hEvent, INFINITE);
                if (waitResult == WAIT_OBJECT_0) {
                    GetOverlappedResult(m_hPipe, &overlapped, &bytesRead, FALSE);
                } else {
                    DEBUG_LOG("[%s %d] Err:%lu WaitForSingleObject failed", __func__, __LINE__, GetLastError());
                    break;
                }
            } else {
                DEBUG_LOG("[%s %d] Err: Read data", __func__, __LINE__);
                DumpError(error);
                char buffer[512];
                FormatMessage(FORMAT_MESSAGE_FROM_SYSTEM | FORMAT_MESSAGE_IGNORE_INSERTS,
                NULL, error, 0, buffer, sizeof(buffer), NULL);
                DEBUG_LOG("[%s %d] Error description: %s", __func__, __LINE__, buffer);
                break;
            }
        }

        DEBUG_LOG("[PIPE] Receive data length: %lu", bytesRead);

#if 1
        printf("[PIPE] Receive data length: %lu\n", bytesRead);
        for(int i=0; i<bytesRead; i++) {
            printf("0x%02X ", buffer[i]);
        }
        printf("\n\n");
#endif

        onReadData(buffer, bytesRead);
    }
    CloseHandle(overlapped.hEvent);
}

void IMSPipeConnection::SendData(unsigned char* data, unsigned int length)
{
    std::thread writer([&]() {
        std::unique_lock<std::mutex> lock(m_pipeMutex);
        if (m_hPipe == NULL) {
            lock.unlock();
            return;
        }

        if (m_hPipe == INVALID_HANDLE_VALUE) {
            DEBUG_LOG("[%s %d] Invalid pipe handle", __func__, __LINE__);
            lock.unlock();
            return;
        }

        if (!data) {
            DEBUG_LOG("[%s %d] Err: Send null msg", __func__, __LINE__);
            lock.unlock();
            return;
        }

        OVERLAPPED overlapped = {0};
        overlapped.hEvent = CreateEvent(NULL, TRUE, FALSE, NULL);

        DWORD bytesWrite = 0;
        bool result = WriteFile(m_hPipe, data, length, &bytesWrite, &overlapped);
        lock.unlock();

        if (!result) {
            DWORD error = GetLastError();
            if (error == ERROR_IO_PENDING) {
                DWORD waitResult = WaitForSingleObject(overlapped.hEvent, INFINITE);
                if (waitResult == WAIT_OBJECT_0) {
                    GetOverlappedResult(m_hPipe, &overlapped, &bytesWrite, FALSE);
                } else {
                    DEBUG_LOG("[%s %d] Err: WaitForSingleObject failed with error: %lu", __func__, __LINE__, GetLastError());
                }
            } else {
                DEBUG_LOG("[%s %d] Err: Send data", __func__, __LINE__);
                DumpError(error);
                char buffer[512];
                FormatMessage(FORMAT_MESSAGE_FROM_SYSTEM | FORMAT_MESSAGE_IGNORE_INSERTS,
                NULL, error, 0, buffer, sizeof(buffer), NULL);
                DEBUG_LOG("[%s %d] Error description: %s", __func__, __LINE__, buffer);
            }
        }
        DEBUG_LOG("[PIPE] Send data length: %d", length);

#if 1
        printf("[PIPE] Send data length: %d\n", length);
        for(int i=0; i<length; i++) {
            printf("0x%02X ", data[i]);
        }
        printf("\n\n");
#endif

        ResetEvent(overlapped.hEvent);
        CloseHandle(overlapped.hEvent);
    });

    writer.join();
}
