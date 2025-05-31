#include "pipe_server.h"
#include "pipe_connection.h"

namespace sunkang {

using std::placeholders::_1;
using std::placeholders::_2;

PipeServer::PipeServer(EventLoop *loop, const std::string &nameArg)
    : loop_(loop)
    , acceptPipeThread_()
    , name_(nameArg)
    , threadPool_(new EventLoopThreadPool(loop))
    , connectionCallback_(defaultConnectionCallback)
    , messageCallback_(defaultMessageCallback)
    , started_(0)
    , nextConnId_(1)
    , connections_()
{

}

PipeServer::~PipeServer()
{
    assert(loop_->isInLoopThread());
    LOG_DEBUG << "PipeServer::~PipeServer [" << name_ << "] destructing";

    for (auto &item : connections_) {
        PipeConnectionPtr conn(item.second);
        item.second.reset();
        conn->getLoop()->runInLoop(std::bind(&PipeConnection::connectDestroyed, conn));
    }
}

void PipeServer::setThreadNum(int numThreads)
{
    assert(numThreads >= 0);
    threadPool_->setThreadNum(numThreads);
}

void PipeServer::start()
{
    if (started_.fetch_add(1) == 0) {
        threadPool_->start(threadInitCallback_);
        LOG_INFO << "[" << name_ << "] The NamedPipe server is listening......";
        assert(acceptPipeThread_ == nullptr);
        acceptPipeThread_.reset(new std::thread(std::bind(&PipeServer::doAccept, this)));
        // FIXME: Simplified processing, resources are reclaimed by the operating system
        acceptPipeThread_->detach();
    }
}

void PipeServer::doAccept()
{
    static std::string s_pipeName(R"(\\.\pipe\)" + name());

    assert(loop_->isInLoopThread() == false);
    HANDLE namedPipeHandle = CreateNamedPipeA(
                s_pipeName.c_str(),
                PIPE_ACCESS_DUPLEX | FILE_FLAG_OVERLAPPED,       // read/write access
                PIPE_TYPE_BYTE |          // byte type pipe
                PIPE_READMODE_BYTE |      // byte-read mode
                PIPE_WAIT,                // blocking mode
                PIPE_UNLIMITED_INSTANCES, // max. instances
                0,                  // output buffer size
                0,                  // input buffer size
                3000,               // client time-out
                nullptr);

    if (namedPipeHandle == INVALID_HANDLE_VALUE) {
        LOG_WARN << "CreateNamedPipe failed, GLE=" << GetLastError();
    }

    if (!ConnectNamedPipe(namedPipeHandle, nullptr)) {
        if (GetLastError() != ERROR_PIPE_CONNECTED) {
            LOG_WARN << "ConnectNamedPipe failed, error: " << GetLastError();
            CloseHandle(namedPipeHandle);
            return;
        }
    }

    loop_->runInLoop([this, namedPipeHandle] {
        EventLoop *ioLoop = threadPool_->getNextLoop();
        newConnection(namedPipeHandle, ioLoop);
    });

    doAccept();
}

void PipeServer::newConnection(const HANDLE &pipeHandle, EventLoop *ioLoop)
{
    assert(loop_->isInLoopThread());

    std::stringstream strStream;
    strStream << name_ << "-" << "NamedPipeConn" << "#" << nextConnId_;
    std::string connName(strStream.str());
    ++nextConnId_;

    LOG_INFO << "PipeServer::newConnection [" << name_
             << "] - new connection [" << connName;

    PipeConnectionPtr conn = std::make_shared<PipeConnection>(ioLoop, connName, pipeHandle);
    connections_[connName] = conn;
    assert(connectionCallback_);
    assert(messageCallback_);
    conn->setConnectionCallback(connectionCallback_);
    conn->setMessageCallback(messageCallback_);
    conn->setWriteCompleteCallback(writeCompleteCallback_);

    conn->setCloseCallback(std::bind(&PipeServer::removeConnection, this, _1));
    ioLoop->runInLoop(std::bind(&PipeConnection::connectEstablished, conn));
}

void PipeServer::removeConnection(const PipeConnectionPtr &conn)
{
    loop_->runInLoop(std::bind(&PipeServer::removeConnectionInLoop, this, conn));
}

void PipeServer::removeConnectionInLoop(const PipeConnectionPtr &conn)
{
    assert(loop_->isInLoopThread());
    LOG_INFO << "PipeServer::removeConnectionInLoop [" << name_ << "] - connection " << conn->name();

    auto itr = connections_.find(conn->name());
    if (itr != connections_.end()) {
        connections_.erase(itr);
    } else {
        assert(false);
    }

    EventLoop *ioLoop = conn->getLoop();
    ioLoop->queueInLoop(std::bind(&PipeConnection::connectDestroyed, conn));
}

void PipeServer::broadcastData(const std::string &data)
{
    loop_->runInLoop([this, data] {
        for (const auto &item : connections_) {
            item.second->send(data);
        }
    });
}

void PipeServer::closeAllConnection()
{
    for (const auto &item : connections_) {
        item.second->forceClose();
    }
}

}
