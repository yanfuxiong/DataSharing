#include "pipe_connection.h"

namespace sunkang {

using std::placeholders::_1;
using std::placeholders::_2;

void defaultConnectionCallback(const PipeConnectionPtr &conn)
{
    LOG_INFO << "defaultConnectionCallback: " << conn->name();
}

void defaultMessageCallback(const PipeConnectionPtr &conn, Buffer *buffer)
{
    (void)conn;
    buffer->retrieveAll();
}

PipeConnection::PipeConnection(EventLoop *loop, const std::string &nameArg, HANDLE pipeHandle)
    : loop_(loop)
    , name_(nameArg)
    , handle_(loop->ioContext(), pipeHandle)
    , state_(kConnecting)
    , isWritting_(false)
{
    assert(handle_.is_open() == true);
}

PipeConnection::~PipeConnection()
{
    LOG_DEBUG << "PipeConnection::dtor[" <<  name_ << "] at " << this
              << " fd=" << handle_.native_handle()
              << " state=" << stateToString();
    assert(state_ == kDisconnected);
}

void PipeConnection::send(const std::string &message)
{
    if (state_ == kConnected) {
        loop_->runInLoop(std::bind(&PipeConnection::sendInLoop, shared_from_this(), message));
    }
}

void PipeConnection::send(std::string &&message)
{
    if (state_ == kConnected) {
        loop_->runInLoop(std::bind(&PipeConnection::sendInLoop, shared_from_this(), std::move(message)));
    }
}

void PipeConnection::send(Buffer *buffer)
{
    if (state_ == kConnected) {
        loop_->runInLoop(std::bind(&PipeConnection::sendInLoop, shared_from_this(), buffer->retrieveAllAsString()));
    }
}

void PipeConnection::send(const void *message, size_t len)
{
    if (state_ == kConnected) {
        send(std::string(reinterpret_cast<const char*>(message), len));
    }
}

void PipeConnection::shutdown()
{
    if (state_ == kConnected) {
        state_.store(kDisconnecting);
        loop_->runInLoop(std::bind(&PipeConnection::shutdownInLoop, shared_from_this()));
    }
}

void PipeConnection::forceClose()
{
    if (state_ == kConnected || state_ == kDisconnecting) {
        state_.store(kDisconnecting);
        loop_->queueInLoop(std::bind(&PipeConnection::forceCloseInLoop, shared_from_this()));
    }
}

void PipeConnection::forceCloseInLoop()
{
    if (state_ == kConnected || state_ == kDisconnecting) {
       asio::error_code err_code;
        handle_.close(err_code);
        if (err_code) {
            LOG_WARN << err_code.message();
        }
        handleClose();
    }
}

void PipeConnection::shutdownInLoop()
{
    assert(state_ == kDisconnecting);
    if (outputBuffer_.empty()) {
       asio::error_code err_code;
       handle_.close(err_code);
       if (err_code) {
           LOG_WARN << err_code.message();
       }
    }
}

void PipeConnection::sendInLoop(const std::string &message)
{
    assert(loop_->isInLoopThread());

    if (state_ == kDisconnected) {
        LOG_WARN << "disconnected, give up writing";
        return;
    }

    Buffer newBuffer(message.data(), message.length());
    outputBuffer_.push_back(std::move(newBuffer));

    if (!isWritting_) {
        isWritting_ = true;
        handle_.async_write_some(asio::const_buffer(outputBuffer_.front().peek(), outputBuffer_.front().readableBytes()),
                                 std::bind(&PipeConnection::writeDataInLoop, shared_from_this(), _1, _2));
    }
}

void PipeConnection::writeDataInLoop(const asio::error_code &err_code, size_t len)
{
    assert(loop_->isInLoopThread());
    assert(isWritting_);
    assert(!outputBuffer_.empty());
    if (err_code || len == 0) {
        //LOG_WARN << "[ERROR] writeDataInLoop : " << err_code.message();
        handleClose();
        return;
    }
    assert(state_ != kDisconnected);

    outputBuffer_.front().retrieve(len);
    if (outputBuffer_.front().readableBytes() == 0) {
        outputBuffer_.pop_front();
        if (writeCompleteCallback_) {
            writeCompleteCallback_(shared_from_this());
        }
    }

    if (outputBuffer_.empty()) {
        isWritting_ = false;
        if (state_ == kDisconnecting) {
            shutdownInLoop();
        }
    } else {
        assert(isWritting_);
        handle_.async_write_some(asio::const_buffer(outputBuffer_.front().peek(), outputBuffer_.front().readableBytes()),
                                 std::bind(&PipeConnection::writeDataInLoop, shared_from_this(), _1, _2));
    }
}

void PipeConnection::asyncReadData()
{
    if (disconnected()) {
        return;
    }
    // Reserved space for single data reception
    inputBuffer_.ensureWritableBytes(1024); // 1kB
    handle_.async_read_some(asio::mutable_buffer(inputBuffer_.beginWrite(), inputBuffer_.writableBytes()),
                            std::bind(&PipeConnection::asyncReadDataInLoop, shared_from_this(), _1, _2));
}

void PipeConnection::asyncReadDataInLoop(const asio::error_code &err_code, size_t len)
{
    assert(loop_->isInLoopThread());
    if (err_code) {
        LOG_WARN << "asyncReadDataInLoop : " << err_code.message();
        handleClose();
        return;
    }
    inputBuffer_.hasWritten(len);

    messageCallback_(shared_from_this(), &inputBuffer_);
    asyncReadData();
}

void PipeConnection::connectEstablished()
{
    assert(loop_->isInLoopThread());
    assert(state_ == kConnecting);
    state_.store(kConnected);
    connectionCallback_(shared_from_this());
    asyncReadData();
}

void PipeConnection::connectDestroyed()
{
    assert(loop_->isInLoopThread());
    if (state_ == kConnected) {
        state_.store(kDisconnected);
        connectionCallback_(shared_from_this());
    }
}

void PipeConnection::handleClose()
{
    LOG_INFO << "fd = " << handle_.native_handle() << " state = " << stateToString();
    if (state_ == kDisconnected) {
        return;
    }

    assert(state_ == kConnected || state_ == kDisconnecting);
    state_.store(kDisconnected);

    PipeConnectionPtr guardThis(shared_from_this());
    connectionCallback_(guardThis);
    closeCallback_(guardThis);
}

const char *PipeConnection::stateToString() const
{
    switch (state_.load()) {
    case kDisconnected:
        return "kDisconnected";
    case kConnecting:
        return "kConnecting";
    case kConnected:
        return "kConnected";
    case kDisconnecting:
        return "kDisconnecting";
    default:
        return "unknown state";
  }
}

}
