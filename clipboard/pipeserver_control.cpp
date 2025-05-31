#include "pipeserver_control.h"
#include "common_message.h"
#include <filesystem>

using namespace sunkang;

enum FieldInfo
{
    BufferField = 0
};

PipeServerControl::PipeServerControl(const std::string &name)
    : m_eventLoop(nullptr)
    , m_pipeServer(nullptr)
    , m_name(name)
    , m_runningStatus(false)
{

}

PipeServerControl::~PipeServerControl()
{

}

void PipeServerControl::setPipeConnectedCallback(PipeConnectedCallback callback)
{
    assert(callback != nullptr);
    m_pipeConnectedCb = callback;
}

void PipeServerControl::setFileDropRequestCallback(FileDropRequestCallback callback)
{
    assert(callback != nullptr);
    m_fileDropRequestCb = callback;
}

void PipeServerControl::setMultiFilesDropRequestCallback(MultiFilesDropRequestCallback callback)
{
    assert(callback != nullptr);
    g_multiFilesReqCallback = callback;
}

void PipeServerControl::setFileDropResponseCallback(FileDropResponseCallback callback)
{
    assert(callback != nullptr);
    m_fileDropResponseCb = callback;
}

void PipeServerControl::setDragFileCallback(DragFileCallback callback)
{
    assert(callback != nullptr);
    m_dragFileCallback = callback;
}

void PipeServerControl::setDragFileListCallback(DragFileListRequestCallback callback)
{
    assert(callback != nullptr);
    m_dragFileListCallback = callback;
}

void PipeServerControl::setCancelFileTransferCallback(CancelFileTransferCallback callback)
{
    assert(callback != nullptr);
    m_cancelFileTransferCallback = callback;
}

void PipeServerControl::startServer()
{
    std::thread(std::bind(&PipeServerControl::processStartServer, this)).detach();
}

void PipeServerControl::closeAllConnection()
{
    //Just disconnect all connections here, keep the server running
    if (m_runningStatus) {
        m_pipeServer->closeAllConnection();
    }
}

void PipeServerControl::sendData(const std::string &data, const std::string &connName)
{
    (void)connName;
    if (m_runningStatus) {
        m_pipeServer->broadcastData(data);
    }
}

void PipeServerControl::processStartServer()
{
    assert(m_eventLoop == nullptr);
    assert(m_pipeServer == nullptr);
    m_eventLoop = new sunkang::EventLoop;
    m_pipeServer = new sunkang::PipeServer(m_eventLoop, m_name);
    m_pipeServer->setConnectionCallback(std::bind(&PipeServerControl::connCallBack, this, std::placeholders::_1));
    m_pipeServer->setMessageCallback(std::bind(&PipeServerControl::recvMessageCallBack, this, std::placeholders::_1, std::placeholders::_2));
    m_pipeServer->start();
    m_runningStatus.store(true);
    m_eventLoop->loop();
}

void PipeServerControl::connCallBack(const sunkang::PipeConnectionPtr &conn)
{
    assert(m_pipeConnectedCb != nullptr);
    if (conn->connected()) {
        LOG_INFO << "--------------------------new connection: " << conn->name();
        m_pipeConnectedCb();
    } else {
        LOG_INFO << "--------------------------conn disconnected: " << conn->name();
    }
}

std::pair<std::string, uint16_t> PipeServerControl::getIpPort(const std::string &ipPortString) const
{
    auto pos = ipPortString.find_first_of(':');
    assert(pos != std::string::npos);
    std::string ip = ipPortString.substr(0, pos);
    uint16_t port = static_cast<uint16_t>(std::atoi(ipPortString.substr(pos + 1).c_str()));
    return { ip, port };
}

void PipeServerControl::recvMessageCallBack(const sunkang::PipeConnectionPtr &conn, sunkang::Buffer *buffer)
{
    (void)conn;

    uint8_t typeValue = 0;
    uint8_t code = 0;
    // parse message
    while (g_getCodeFromByteArray(std::string(buffer->peek(), buffer->readableBytes()), typeValue, code)) {
        switch (code) {
        case SendFile_code: {
            if (typeValue == PipeMessageType::Request) {
                SendFileRequestMsg message;
                if (SendFileRequestMsg::fromByteArray(std::string(buffer->peek(), buffer->readableBytes()), message)) {
                    buffer->retrieve(message.getMessageLength());
                    //void MainWindow::onFileDropRequestCallback(char *ipString, char *idString, unsigned long long fileSize, unsigned long long timestamp, wchar_t *filePath)
                    std::string ipString = message.ip + ":" + std::to_string(message.port);
                    std::string clientID = message.clientID;
                    uint64_t fileSize = message.fileSize;
                    uint64_t timestamp = message.timeStamp;
                    if (message.filePathVec.empty()) {
                        std::wstring filePath = Utils::toUtf16LE(message.fileName);
                        m_fileDropRequestCb(const_cast<char*>(ipString.c_str()),
                                           const_cast<char*>(clientID.c_str()),
                                           fileSize,
                                           timestamp,
                                           const_cast<wchar_t*>(filePath.c_str()));
                    } else {
                        std::vector<std::wstring> filePathVec;
                        for (const auto &filePathUtf8 : message.filePathVec) {
                            filePathVec.push_back(Utils::toUtf16LE(filePathUtf8));
                        }
                        wchar_t **filePathArry = new wchar_t*[filePathVec.size()];
                        for (int index = 0; index < static_cast<int>(filePathVec.size()); ++index) {
                            filePathArry[index] = const_cast<wchar_t*>(filePathVec[index].c_str());
                        }
                        g_multiFilesReqCallback(const_cast<char*>(ipString.c_str()),
                                                const_cast<char*>(clientID.c_str()),
                                                timestamp,
                                                filePathArry,
                                                message.filePathVec.size());
                        delete []filePathArry;
                    }
                }
            } else if (typeValue == PipeMessageType::Response) {
                SendFileResponseMsg message;
                if (SendFileResponseMsg::fromByteArray(std::string(buffer->peek(), buffer->readableBytes()), message)) {
                    buffer->retrieve(message.getMessageLength());
                    // void MainWindow::onFileDropResponseCallback(int status, char *ipString, char *clientID, unsigned long long fileSize, unsigned long long timestamp, wchar_t *filePath)
                    int statusCode = message.statusCode;
                    std::string ipString = message.ip + ":" + std::to_string(message.port);
                    std::string clientID = message.clientID;
                    uint64_t fileSize = message.fileSize;
                    uint64_t timestamp = message.timeStamp;
                    std::wstring filePath = Utils::toUtf16LE(message.fileName);
                    m_fileDropResponseCb(statusCode,
                                        const_cast<char*>(ipString.c_str()),
                                        const_cast<char*>(clientID.c_str()),
                                        fileSize,
                                        timestamp,
                                        const_cast<wchar_t*>(filePath.c_str()));
                }
            } else {
                assert(false);
            }
            break;
        }
        case DDCCIMsg_code: {
            assert(typeValue == PipeMessageType::Notify);
            DDCCINotifyMsg message;
            if (DDCCINotifyMsg::fromByteArray(std::string(buffer->peek(), buffer->readableBytes()), message)) {
                buffer->retrieve(message.getMessageLength());
            }
            break;
        }
        case DragFilesMsg_code: {
            assert(typeValue == PipeMessageType::Notify);
            DragFilesMsg message;
            if (DragFilesMsg::fromByteArray(std::string(buffer->peek(), buffer->readableBytes()), message)) {
                buffer->retrieve(message.getMessageLength());
                if (message.functionCode == DragFilesMsg::FuncCode::CancelFileTransfer) {
                    std::string ipString = message.ip + ":" + std::to_string(message.port);
                    std::string clientID = message.clientID;
                    m_cancelFileTransferCallback(const_cast<char*>(ipString.c_str()),
                                                 const_cast<char*>(clientID.c_str()),
                                                 message.timeStamp);
                    break;
                }
                assert(message.filePathVec.empty() == false);
                int fileCount = 0;
                int folderCount = 0;
                for (const auto &filePath : message.filePathVec) {
                    if (std::filesystem::is_regular_file(filePath)) {
                        ++fileCount;
                    } else if (std::filesystem::is_directory(filePath)) {
                        ++folderCount;
                    }
                }

                assert(static_cast<std::size_t>(fileCount + folderCount) == message.filePathVec.size());
                if (fileCount == 1 && folderCount == 0) { // Single file situation
                    std::wstring filePath = Utils::toUtf16LE(message.filePathVec.front());
                    m_dragFileCallback(message.timeStamp, const_cast<wchar_t*>(filePath.c_str()));
                } else {
                    std::vector<std::wstring> filePathVec;
                    for (const auto &filePath : message.filePathVec) {
                        filePathVec.push_back(Utils::toUtf16LE(filePath));
                    }
                    wchar_t **filePathArry = new wchar_t*[fileCount + folderCount];
                    for (int index = 0; index < fileCount + folderCount; ++index) {
                        filePathArry[index] = const_cast<wchar_t*>(filePathVec[index].c_str());
                    }
                    m_dragFileListCallback(filePathArry, fileCount + folderCount, message.timeStamp);
                    delete []filePathArry;
                }
            }
            break;
        }
        default: {
            assert(false);
            break;
        }
        }
    }
}

void PipeServerControl::updateProgress(const std::string &ipPortString, const std::string &id, uint64_t fileSize, uint64_t sentSize, uint64_t timestamp, const std::wstring &fileName)
{
    UpdateProgressMsg message;
    const auto &ipPort = getIpPort(ipPortString);
    message.ip = ipPort.first;
    message.port = ipPort.second;
    message.clientID = id;
    message.fileSize = fileSize;
    message.sentSize = sentSize;
    message.timeStamp = timestamp;
    message.fileName = Utils::toUtf8(fileName);

    //LOG_INFO << "-----------------------------updateProgress";
    auto data = UpdateProgressMsg::toByteArray(message);
    //LOG_INFO << Utils::toHex(data);
    sendData(data);
}

void PipeServerControl::updateImageProgress(const std::string &ipPortString, const std::string &id, uint64_t fileSize, uint64_t sentSize, uint64_t timestamp)
{
    UpdateImageProgressMsg message;
    const auto &ipPort = getIpPort(ipPortString);
    message.ip = ipPort.first;
    message.port = ipPort.second;
    message.clientID = id;
    message.fileSize = fileSize;
    message.sentSize = sentSize;
    message.timeStamp = timestamp;

    auto data = UpdateImageProgressMsg::toByteArray(message);
    //LOG_INFO << Utils::toHex(data);
    sendData(data);
}

void PipeServerControl::updateClientStatus(unsigned int status, const std::string &ipPortString, const std::string &id, const std::wstring &clientName, const std::string &deviceType)
{
    UpdateClientStatusMsg message;
    message.status = static_cast<int>(status);
    const auto &ipPort = getIpPort(ipPortString);
    message.ip = ipPort.first;
    message.port = ipPort.second;
    message.clientID = id;
    message.clientName = Utils::toUtf8(clientName);
    message.deviceType = deviceType;

    auto data = UpdateClientStatusMsg::toByteArray(message);
    //LOG_INFO << Utils::toHex(data);
    sendData(data);
}

void PipeServerControl::updateSystemInfo(const std::string &ipPortString, const std::wstring &serviceVer)
{
    UpdateSystemInfoMsg message;
    const auto &ipPort = getIpPort(ipPortString);
    message.ip = ipPort.first;
    message.port = ipPort.second;
    message.serverVersion = Utils::toUtf8(serviceVer);

    auto data = UpdateSystemInfoMsg::toByteArray(message);
    //LOG_INFO << Utils::toHex(data);
    sendData(data);
}

void PipeServerControl::notiMessage(uint64_t timestamp, unsigned int notiCode, const std::vector<std::wstring> &notiParamVec)
{
    NotifyMessage message;
    message.timeStamp = timestamp;
    message.notiCode = static_cast<uint8_t>(notiCode);
    for (const auto &paramVal : notiParamVec) {
        message.paramInfoVec.push_back(NotifyMessage::ParamInfo{ Utils::toUtf8(paramVal) });
    }

    auto data = NotifyMessage::toByteArray(message);
    //LOG_INFO << Utils::toHex(data);
    sendData(data);
}

void PipeServerControl::sendFileRequest(const std::string &ipPortString, const std::string &id, uint64_t fileSize, uint64_t timestamp, const std::wstring &fileName)
{
    SendFileRequestMsg message;
    const auto &ipPort = getIpPort(ipPortString);
    message.ip = ipPort.first;
    message.port = ipPort.second;
    message.clientID = id;
    message.fileSize = fileSize;
    message.timeStamp = timestamp;
    message.fileName = Utils::toUtf8(fileName);

    auto data = SendFileRequestMsg::toByteArray(message);
    //LOG_INFO << Utils::toHex(data);
    sendData(data);
}

void PipeServerControl::dragFileNotify(const std::string &ipPortString, const std::string &id, uint64_t fileSize, uint64_t timestamp, const std::wstring &fileName)
{
    SendFileRequestMsg message;
    message.flag = SendFileRequestMsg::FlagType::DragFlag;
    const auto &ipPort = getIpPort(ipPortString);
    message.ip = ipPort.first;
    message.port = ipPort.second;
    message.clientID = id;
    message.fileSize = fileSize;
    message.timeStamp = timestamp;
    message.fileName = Utils::toUtf8(fileName);

    auto data = SendFileRequestMsg::toByteArray(message);
    //LOG_INFO << Utils::toHex(data);
    sendData(data);
}

void PipeServerControl::dragFileListNotify(const std::string &ipPortString,
                        const std::string &id,
                        uint32_t cFileCount,
                        uint64_t totalSize,
                        uint64_t timestamp,
                        const std::wstring &firstFileName,
                        uint64_t firstFileSize)
{
    DragFilesMsg message;
    message.functionCode = DragFilesMsg::FuncCode::ReceiveFileInfo;
    message.timeStamp = timestamp;
    const auto &ipPort = getIpPort(ipPortString);
    message.ip = ipPort.first;
    message.port = ipPort.second;
    message.clientID = id;
    message.fileCount = cFileCount;
    message.totalFileSize = totalSize;
    message.firstTransferFileName = Utils::toUtf8(firstFileName);
    message.firstTransferFileSize = firstFileSize;

    auto data = DragFilesMsg::toByteArray(message);
    sendData(data);
}

void PipeServerControl::updateProgressForMultiFileTransfers(const std::string &ipPortString,
                                         const std::string &id,
                                         const std::wstring &currentFileName,
                                         uint32_t sentFilesCnt,
                                         uint32_t totalFilesCnt,
                                         uint64_t currentFileSize,
                                         uint64_t totalSize,
                                         uint64_t sentSize,
                                         uint64_t timestamp)
{
    UpdateProgressMsg message;
    message.functionCode = UpdateProgressMsg::FuncCode::MultiFile;
    const auto &ipPort = getIpPort(ipPortString);
    message.ip = ipPort.first;
    message.port = ipPort.second;
    message.clientID = id;
    message.timeStamp = timestamp;

    message.currentFileName = Utils::toUtf8(currentFileName);
    message.sentFilesCount = sentFilesCnt;
    message.totalFilesCount = totalFilesCnt;
    message.currentFileSize = currentFileSize;
    message.totalFilesSize = totalSize;
    message.totalSentSize = sentSize;

    auto data = UpdateProgressMsg::toByteArray(message);
    sendData(data);
}

void PipeServerControl::statusInfoNotify(uint32_t statusCode)
{
    StatusInfoNotifyMsg message;
    message.statusCode = statusCode;

    auto data = StatusInfoNotifyMsg::toByteArray(message);
    sendData(data);
}
