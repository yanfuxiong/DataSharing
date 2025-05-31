#include "event_loop.h"

using std::placeholders::_1;
using std::placeholders::_2;

namespace sunkang {

std::atomic<int64_t> EventLoop::s_numCreated_(1);

EventLoop::EventLoop()
{
    threadId_ = std::this_thread::get_id();
}

EventLoop::~EventLoop()
{

}

void EventLoop::loop()
{
    assert(isInLoopThread());
    auto guard = asio::make_work_guard(ioContext_);
    (void)guard;
    ioContext_.run();
}

void EventLoop::quit()
{
    try {
        ioContext_.stop();
    } catch (const std::exception &e) {
        LOG_DEBUG << e.what();
    }
}

void EventLoop::runInLoop(const Functor &cb)
{
    asio::dispatch(ioContext_, cb);
}

void EventLoop::runInLoop(Functor &&cb)
{
    asio::dispatch(ioContext_, std::move(cb));
}

void EventLoop::queueInLoop(const Functor &cb)
{
    asio::post(ioContext_, cb);
}

void EventLoop::queueInLoop(Functor &&cb)
{
    asio::post(ioContext_, std::move(cb));
}

bool EventLoop::isInLoopThread() const
{
    return threadId_ == std::this_thread::get_id();
}

}
