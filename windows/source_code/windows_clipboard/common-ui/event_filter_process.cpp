#include "event_filter_process.h"
#include <QMouseEvent>

EventFilterProcess *EventFilterProcess::m_instance = nullptr;

EventFilterProcess::EventFilterProcess(QObject *parent)
    : QObject(parent)
{

}

EventFilterProcess::~EventFilterProcess()
{
    if (m_instance) {
        m_instance->deleteLater();
        m_instance = nullptr;
    }
}

EventFilterProcess *EventFilterProcess::getInstance()
{
    if (m_instance == nullptr) {
        m_instance = new EventFilterProcess;
    }
    return m_instance;
}

void EventFilterProcess::registerFilterEvent(const EventFilterData &filterData)
{
    Q_ASSERT(filterData.eventType == QEvent::Type::MouseButtonPress);
    m_filterDataList.push_back(filterData);
    filterData.monitoredObject->installEventFilter(this);

    // 对象销毁的时候移除数据
    connect(filterData.monitoredObject, &QObject::destroyed, this, [this] (QObject *object) {
        for (auto itr = m_filterDataList.begin(); itr != m_filterDataList.end(); ++itr) {
            if (itr->monitoredObject == object) {
                m_filterDataList.erase(itr);
                qInfo() << "--------remove:" << object->objectName().toUtf8().constData();
                break;
            }
        }
    });
}


bool EventFilterProcess::eventFilter(QObject *obj, QEvent *event)
{
    // FIXME: 目前简化处理
    if (event->type() == QEvent::MouseButtonPress) {
        QMouseEvent *mouseEvent = static_cast<QMouseEvent*>(event);
        if (mouseEvent->button() == Qt::MouseButton::LeftButton) {
            for (const auto &data : m_filterDataList) {
                if (data.monitoredObject == obj) {
                    data.callback();
                    break;
                }
            }
            return true;
        } else {
            return QObject::eventFilter(obj, event);
        }
    } else {
        return QObject::eventFilter(obj, event);
    }
}


