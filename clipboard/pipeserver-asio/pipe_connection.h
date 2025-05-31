#pragma once
#include "callbacks.h"
#include "global_utils.h"
#include "event_loop.h"
#include "event_loop_thread.h"
#include "buffer_x.h"

namespace sunkang {

class PipeConnection : public std::enable_shared_from_this<PipeConnection>
{
public:
    enum ConnState {
        kDisconnected,
        kConnecting,
        kConnected,
        kDisconnecting
    };

public:
    PipeConnection(EventLoop *loop, const std::string &nameArg, HANDLE pipeHandle);
    PipeConnection(const PipeConnection &) = delete ;
    PipeConnection &operator = (const PipeConnection &) = delete ;
    ~PipeConnection();

    EventLoop *getLoop() const { return loop_; }
    const std::string &name() const { return name_; }
    bool connected() const { return state_ == kConnected; }
    bool disconnected() const { return state_ == kDisconnected; }

    void send(const std::string &message);
    void send(std::string &&message);
    void send(Buffer *buffer);
    void send(const void *message, size_t len);

    void shutdown();
    void forceClose();

    void setContext(const std::any &context, int index = 0)
    { assert(index < kContextCount && index >= 0); context_[index] = context; }

    const std::any &getContext(int index = 0) const
    { assert(index < kContextCount && index >= 0); return context_[index]; }

    std::any *getMutableContext(int index = 0)
    { assert(index < kContextCount && index >= 0); return &context_[index]; }

    void setConnectionCallback(const ConnectionCallback &cb) { connectionCallback_ = cb; }
    void setMessageCallback(const MessageCallback &cb) { messageCallback_ = cb; }
    void setWriteCompleteCallback(const WriteCompleteCallback &cb) { writeCompleteCallback_ = cb; }

    void setCloseCallback(const CloseCallback &cb) { closeCallback_ = cb; }
    void connectEstablished();
    void connectDestroyed();

private:
    void asyncReadData();
    void asyncReadDataInLoop(const asio::error_code &err_code, size_t len);
    void sendInLoop(const std::string &message);
    void writeDataInLoop(const asio::error_code &err_code, size_t len);
    void shutdownInLoop();
    void forceCloseInLoop();
    void handleClose();
    const char *stateToString() const;

private:
    enum { kContextCount = 16 };

    EventLoop *loop_;
    const std::string name_;
    asio::windows::stream_handle handle_;
    std::atomic<ConnState> state_;
    ConnectionCallback connectionCallback_;
    MessageCallback messageCallback_;
    WriteCompleteCallback writeCompleteCallback_;
    CloseCallback closeCallback_;
    Buffer inputBuffer_;
    std::list<Buffer> outputBuffer_;
    bool isWritting_; //Operate this variable only in one thread
    std::any context_[kContextCount];
};


typedef std::shared_ptr<PipeConnection> PipeConnectionPtr;

}
