#include "event_loop_thread_pool.h"

namespace sunkang {

EventLoopThreadPool::EventLoopThreadPool(EventLoop *eventLoop)
    : baseLoop_(eventLoop)
    , started_(false)
    , numThreads_(0)
    , next_(0)
{

}

EventLoopThreadPool::~EventLoopThreadPool()
{

}

void EventLoopThreadPool::setThreadNum(int numThreads)
{
    assert(numThreads_ == 0);
    numThreads_ = numThreads;
}

void EventLoopThreadPool::start(const ThreadInitCallback &cb)
{
    assert(baseLoop_->isInLoopThread());

    started_ = true;

    for (int index = 0; index < numThreads_; ++index) {
        EventLoopThread *pLoopThread = new EventLoopThread(cb);
        threads_.push_back(std::unique_ptr<EventLoopThread>(pLoopThread));
        loops_.push_back(pLoopThread->startLoop());
    }

    if (numThreads_ == 0 && cb) {
        cb(baseLoop_);
    }
}

EventLoop *EventLoopThreadPool::getNextLoop()
{
    assert(baseLoop_->isInLoopThread());
    EventLoop *pLoopThread = baseLoop_;

    if (!loops_.empty()) {
        pLoopThread = loops_[next_];
        next_++;
        if (next_ >= loops_.size()) {
            next_ = 0;
        }
    }

    return pLoopThread;
}

EventLoop *EventLoopThreadPool::getLoopForHash(size_t hashCode)
{
    assert(baseLoop_->isInLoopThread());
    EventLoop *loop = baseLoop_;

    if (!loops_.empty()) {
        loop = loops_[hashCode % loops_.size()];
    }

    return loop;
}

std::vector<EventLoop*> EventLoopThreadPool::getAllLoops() const
{
    assert(baseLoop_->isInLoopThread());
    assert(started_);

    if (loops_.empty()) {
        return { baseLoop_ };
    } else {
        return loops_;
    }
}

}
