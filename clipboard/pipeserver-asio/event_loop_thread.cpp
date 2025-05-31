#include "event_loop_thread.h"
#include "event_loop.h"

namespace sunkang {

EventLoopThread::EventLoopThread(const ThreadInitCallback &cb)
  : loop_(nullptr),
    exiting_(false),
    thread_(),
    mutex_(),
    callback_(cb)
{

}

EventLoopThread::~EventLoopThread()
{
    exiting_ = true;
    if (loop_ != nullptr) {
        loop_->quit();
        assert(thread_ != nullptr);
        thread_->join();
    }
}

EventLoop *EventLoopThread::startLoop()
{
    thread_.reset(new std::thread(std::bind(&EventLoopThread::threadFunc, this)));
    EventLoop *loop = nullptr;

    {
        std::unique_lock<std::mutex> lock(mutex_);
        while (loop_ == nullptr) {
            cond_.wait(lock);
        }
        loop = loop_;
    }

    return loop;
}

void EventLoopThread::threadFunc()
{
    EventLoop loop;

    if (callback_) {
        callback_(&loop);
    }

    {
        std::lock_guard<std::mutex> lockerGuard(mutex_);
        (void)lockerGuard;
        loop_ = &loop;
        cond_.notify_all();
    }

    loop.loop();

    {
        std::unique_lock<std::mutex> lock(mutex_);
        (void)lock;
        loop_ = nullptr;
    }
}

}
