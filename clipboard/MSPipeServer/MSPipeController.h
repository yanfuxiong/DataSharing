#ifndef __INCLUDED_MS_PIPE_CONTROLLER__
#define __INCLUDED_MS_PIPE_CONTROLLER__

#include "IMSPipeConnection.h"
#include "MSPipeCommon.h"

class MSPipeController : public IMSPipeConnection
{
public:
    MSPipeController(const std::atomic<bool>& running_pipe,
                        FileDropRequestCallback& fileDropReqCb,
                        FileDropResponseCallback& fileDropRespCb,
                        PipeConnectedCallback& connCb);
    ~MSPipeController();
    void onConnected() override;
    void onDisconnected() override;
    void onReadData(unsigned char* data, unsigned int length) override;
    // Sender
    void UpdateClientStatus(unsigned int status, char* ip, char* id, wchar_t* name);
    void UpdateProgress(char* ip, char* id, uint64_t fileSize, uint64_t sentSize, uint64_t timestamp, wchar_t* fileName);
    void UpdateSystemInfo(char* ip, wchar_t* serviceVer);
    void SendFileReq(char* ip, char* id, uint64_t fileSize, uint64_t timestamp, wchar_t* fileName);

private:
    // Reader
    bool HandlePipeDataHeader(unsigned char* buffer, RTK_PIPE_TYPE &type, RTK_PIPE_CODE &code, uint32_t &lenContent);
    void HandleSendFileReq(unsigned char* content, uint32_t length);
    void HandleSendFileResp(unsigned char* content, uint32_t length);

    const std::atomic<bool>& g_running_pipe;
    FileDropRequestCallback mFileDropReqCb;
    FileDropResponseCallback mFileDropRespCb;
    PipeConnectedCallback mConnCb;
};

#endif //__INCLUDED_MS_PIPE_CONTROLLER__