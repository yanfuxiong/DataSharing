#pragma once
#include "common_signals.h"
#include "common_utils.h"


class ProgressObject : public QObject
{
    Q_OBJECT
public:
    ProgressObject(UpdateProgressMsg *ptr_msg, QObject *parent = nullptr);
    ~ProgressObject();

private Q_SLOTS:
    void sendProgressData();

private:
    std::unique_ptr<UpdateProgressMsg> m_cacheProgressMsgPtr;
    QPointer<QTimer> m_timer;
    int m_cacheProgressVal; // 进度值, 只用于模拟
};
