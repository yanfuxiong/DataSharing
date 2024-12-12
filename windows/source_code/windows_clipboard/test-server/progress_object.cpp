#include "progress_object.h"
ProgressObject::ProgressObject(UpdateProgressMsg *ptr_msg, QObject *parent)
    : QObject(parent)
    , m_cacheProgressMsgPtr(ptr_msg)
    , m_timer(new QTimer)
    , m_cacheProgressVal(0)
{
    m_timer->setTimerType(Qt::TimerType::PreciseTimer);
    m_timer->setInterval(50 + qrand() % 20);
    m_cacheProgressVal = 0;
    connect(m_timer, &QTimer::timeout, this, &ProgressObject::sendProgressData);
    m_timer->start();
}

ProgressObject::~ProgressObject()
{

}

void ProgressObject::sendProgressData()
{
    if (m_cacheProgressMsgPtr == nullptr) {
        return;
    }

    if (m_cacheProgressVal >= 100) {
        m_timer->stop();
        m_timer->deleteLater();
        m_timer = nullptr;
        m_cacheProgressMsgPtr.reset(nullptr);
        deleteLater(); // 销毁自己
        return;
    }

    UpdateProgressMsg message;
    message.ip = m_cacheProgressMsgPtr->ip;
    message.port = m_cacheProgressMsgPtr->port;
    message.clientID = m_cacheProgressMsgPtr->clientID;
    message.fileSize = QFileInfo(m_cacheProgressMsgPtr->fileName).size();
    auto delta = message.fileSize / 100.0;
    if (delta < 1.0) {
        delta = 1.0;
    }
    uint64_t currentSentFileSize = static_cast<uint64_t>((++m_cacheProgressVal) * delta);
    if (m_cacheProgressVal >= 100) {
        currentSentFileSize = message.fileSize;
    }
    message.sentSize = currentSentFileSize;
    message.timeStamp = QDateTime::currentDateTime().toUTC().toMSecsSinceEpoch();
    message.fileName = m_cacheProgressMsgPtr->fileName;

    Q_EMIT CommonSignals::getInstance()->sendDataForTestServer(UpdateProgressMsg::toByteArray(message));
}
