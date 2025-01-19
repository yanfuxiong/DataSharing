#include "worker_thread.h"
#include <QCoreApplication>

WorkerThread::WorkerThread(QPointer<QThread> newThread)
    : QObject(nullptr)
    , m_thread(nullptr)
{
    static int s_retVal = qRegisterMetaType<TaskCallback>("TaskCallback");
    Q_UNUSED(s_retVal)

    if (newThread) {
        moveToThread(newThread);
        newThread->start();
        m_thread = newThread;
    }
    connect(this, &WorkerThread::runTaskCallback, this, &WorkerThread::onRunTaskCallback);
}

WorkerThread::~WorkerThread()
{
    if (m_thread) {
        QPointer<QThread> tmpThread(m_thread);
        connect(m_thread, &QThread::finished, tmpThread, [tmpThread] {
            if (tmpThread) {
                delete tmpThread.data();
                qDebug() << "线程对象析构......";
            }
        });
        m_thread->quit();
    }
}

void WorkerThread::runInThread(const TaskCallback &cb)
{
    Q_EMIT runTaskCallback(cb, QPrivateSignal());
}

void WorkerThread::onRunTaskCallback(const TaskCallback &cb)
{
    Q_ASSERT(cb != nullptr);
    cb();
}


