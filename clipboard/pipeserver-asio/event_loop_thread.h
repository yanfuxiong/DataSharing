#pragma once
#include "global.h"

namespace sunkang {

class EventLoop;

class EventLoopThread
{
public:
    typedef std::function<void(EventLoop*)> ThreadInitCallback;

    EventLoopThread(const ThreadInitCallback &cb = ThreadInitCallback());
    EventLoopThread(const EventLoopThread &) = delete ;
    EventLoopThread &operator = (const EventLoopThread &) = delete ;
    ~EventLoopThread();
    EventLoop *startLoop();

private:
    void threadFunc();

    EventLoop *loop_;
    std::atomic<bool> exiting_;
    std::unique_ptr<std::thread> thread_;
    std::mutex mutex_;
    std::condition_variable cond_;
    ThreadInitCallback callback_;
};

}
