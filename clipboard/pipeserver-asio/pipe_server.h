#pragma once
#include "global_utils.h"
#include "event_loop.h"
#include "event_loop_thread_pool.h"

namespace sunkang {

class PipeServer
{
public:
    typedef std::function<void(EventLoop*)> ThreadInitCallback;

    PipeServer(EventLoop *loop, const std::string &nameArg);
    PipeServer(const PipeServer &) = delete ;
    PipeServer &operator = (const PipeServer &) = delete ;
    ~PipeServer();

    const std::string &name() const { return name_; }
    EventLoop *getLoop() const { return loop_; }

    void setThreadNum(int numThreads);
    void setThreadInitCallback(const ThreadInitCallback &cb) { threadInitCallback_ = cb; }

    std::shared_ptr<EventLoopThreadPool> threadPool() const { return threadPool_; }
    void start();

    // Send data to all connections
    void broadcastData(const std::string &data);
    void closeAllConnection();

    void setConnectionCallback(const ConnectionCallback &cb) { connectionCallback_ = cb; }
    void setMessageCallback(const MessageCallback &cb) { messageCallback_ = cb; }
    void setWriteCompleteCallback(const WriteCompleteCallback &cb) { writeCompleteCallback_ = cb; }

private:
    void newConnection(const HANDLE &pipeHandle, EventLoop *ioLoop);
    void doAccept();
    void removeConnection(const PipeConnectionPtr &conn);
    void removeConnectionInLoop(const PipeConnectionPtr &conn);

    typedef std::unordered_map<std::string, PipeConnectionPtr> ConnectionMap;

    EventLoop *loop_;
    std::unique_ptr<std::thread> acceptPipeThread_;
    const std::string name_;
    std::shared_ptr<EventLoopThreadPool> threadPool_;
    ConnectionCallback connectionCallback_;
    MessageCallback messageCallback_;
    WriteCompleteCallback writeCompleteCallback_;
    ThreadInitCallback threadInitCallback_;
    std::atomic<int> started_;

    int64_t nextConnId_;
    ConnectionMap connections_;
};

}
