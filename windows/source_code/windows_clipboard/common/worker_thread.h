#pragma once
#include "common_utils.h"
#include "common_signals.h"

class WorkerThread : public QObject
{
    Q_OBJECT
public:
    typedef std::function<void()> TaskCallback;

    explicit WorkerThread(QPointer<QThread> newThread = new QThread);
    ~WorkerThread();

    void runInThread(const TaskCallback &cb);

Q_SIGNALS:
    void runTaskCallback(const TaskCallback &cb, QPrivateSignal);

private Q_SLOTS:
    void onRunTaskCallback(const TaskCallback &cb);

private:
    QPointer<QThread> m_thread;
};
