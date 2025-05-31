#pragma once
#include "global.h"
#include "pipe_server.h"
#include "pipe_connection.h"
#include "clipboard.h"

class PipeServerControl
{
public:
    PipeServerControl(const std::string &name);
    ~PipeServerControl();

    void setPipeConnectedCallback(PipeConnectedCallback callback);
    void setFileDropRequestCallback(FileDropRequestCallback callback);
    void setMultiFilesDropRequestCallback(MultiFilesDropRequestCallback callback);
    void setFileDropResponseCallback(FileDropResponseCallback callback);

    void setDragFileCallback(DragFileCallback callback);
    void setDragFileListCallback(DragFileListRequestCallback callback);
    void setCancelFileTransferCallback(CancelFileTransferCallback callback);

    void startServer();
    void closeAllConnection();
    void sendData(const std::string &data, const std::string &connName = std::string());

    void updateProgress(const std::string &ipPortString, const std::string &id, uint64_t fileSize, uint64_t sentSize, uint64_t timestamp, const std::wstring &fileName);
    void updateImageProgress(const std::string &ipPortString, const std::string &id, uint64_t fileSize, uint64_t sentSize, uint64_t timestamp);
    void updateClientStatus(unsigned int status, const std::string &ip, const std::string &id, const std::wstring &clientName, const std::string &deviceType);
    void updateSystemInfo(const std::string &ipPortString, const std::wstring &serviceVer);
    void notiMessage(uint64_t timestamp, unsigned int notiCode, const std::vector<std::wstring> &notiParamVec);
    void sendFileRequest(const std::string &ipPortString, const std::string &id, uint64_t fileSize, uint64_t timestamp, const std::wstring &fileName);
    void dragFileNotify(const std::string &ipPortString,
                        const std::string &id,
                        uint64_t fileSize,
                        uint64_t timestamp,
                        const std::wstring &fileName);
    void dragFileListNotify(const std::string &ipPortString,
                            const std::string &id,
                            uint32_t cFileCount,
                            uint64_t totalSize,
                            uint64_t timestamp,
                            const std::wstring &firstFileName,
                            uint64_t firstFileSize);
    void updateProgressForMultiFileTransfers(const std::string &ipPortString,
                                             const std::string &id,
                                             const std::wstring &currentFileName,
                                             uint32_t sentFilesCnt,
                                             uint32_t totalFilesCnt,
                                             uint64_t currentFileSize,
                                             uint64_t totalSize,
                                             uint64_t sentSize,
                                             uint64_t timestamp);
    void statusInfoNotify(uint32_t statusCode);

private:
    void processStartServer();
    void connCallBack(const sunkang::PipeConnectionPtr &conn);
    void recvMessageCallBack(const sunkang::PipeConnectionPtr &conn, sunkang::Buffer *buffer);
    std::pair<std::string, uint16_t> getIpPort(const std::string &ipPortString) const;

private:
    sunkang::EventLoop *m_eventLoop;
    sunkang::PipeServer *m_pipeServer;
    const std::string m_name;
    std::atomic<bool> m_runningStatus;

    PipeConnectedCallback m_pipeConnectedCb { nullptr };
    FileDropRequestCallback m_fileDropRequestCb { nullptr };
    MultiFilesDropRequestCallback g_multiFilesReqCallback { nullptr };
    FileDropResponseCallback m_fileDropResponseCb { nullptr };

    DragFileCallback m_dragFileCallback { nullptr };
    DragFileListRequestCallback m_dragFileListCallback { nullptr };
    CancelFileTransferCallback m_cancelFileTransferCallback { nullptr };
};

