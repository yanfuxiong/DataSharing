#pragma once
#include "global_utils.h"
#include "callbacks.h"

namespace sunkang {

class EventLoop
{
public:
    typedef std::function<void()> Functor;

    EventLoop();
    EventLoop(const EventLoop &) = delete ;
    EventLoop &operator = (const EventLoop &) = delete ;
    ~EventLoop();
    void loop();
    void quit();
    void runInLoop(const Functor &cb);
    void queueInLoop(const Functor &cb);
    bool isInLoopThread() const;
    asio::io_context &ioContext() { return ioContext_; }

    void runInLoop(Functor &&cb);
    void queueInLoop(Functor &&cb);

private:
    std::thread::id threadId_;
    asio::io_context ioContext_;

    static std::atomic<int64_t> s_numCreated_;
};

} // namespace sunkang
