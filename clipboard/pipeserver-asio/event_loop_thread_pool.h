#pragma once
#include "global.h"
#include "event_loop.h"
#include "event_loop_thread.h"

namespace sunkang {


class EventLoopThreadPool
{
public:
    typedef std::function<void(EventLoop*)> ThreadInitCallback;

    EventLoopThreadPool(EventLoop *baseLoop);
    EventLoopThreadPool(const EventLoopThreadPool &) = delete ;
    EventLoopThreadPool &operator = (const EventLoopThreadPool &) = delete ;
    ~EventLoopThreadPool();
    void setThreadNum(int numThreads);
    void start(const ThreadInitCallback &cb = ThreadInitCallback());

    // valid after calling start()
    // round-robin
    EventLoop *getNextLoop();

    // with the same hash code, it will always return the same EventLoop
    EventLoop *getLoopForHash(size_t hashCode);
    std::vector<EventLoop*> getAllLoops() const;

private:
    EventLoop *baseLoop_;
    bool started_;
    int numThreads_;
    uint32_t next_;
    std::vector<std::unique_ptr<EventLoopThread> > threads_;
    std::vector<EventLoop*> loops_;
};

}
