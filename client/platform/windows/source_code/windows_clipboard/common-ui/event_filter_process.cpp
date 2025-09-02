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
    const auto &objectIndex = m_filterDataList.get<tag_obj_pointer>();
    auto exists = (objectIndex.find(filterData.objectPointer()) != objectIndex.end());
    auto retVal = m_filterDataList.push_back(filterData);
    if (retVal.second == false) {
        Q_ASSERT(false);
        qWarning() << "---------------------registerFilterEvent failed !!!";
        return;
    }

    if (exists) {
        return;
    }
    filterData.monitoredObject->installEventFilter(this);

    // Remove data when destroying objects
    connect(filterData.monitoredObject, &QObject::destroyed, this, [this] (QObject *object) {
        auto &objectIndex = m_filterDataList.get<tag_obj_pointer>();
        const auto &itr_left = objectIndex.lower_bound(object);
        const auto &itr_right = objectIndex.upper_bound(object);
        objectIndex.erase(itr_left, itr_right);
    });
}

void EventFilterProcess::removeFilterEvent(QObject *object, const QList<QEvent::Type> &eventList)
{
    auto &eventIndex = m_filterDataList.get<tag_obj_composite>();
    for (const auto &eventType : eventList) {
        auto itr = eventIndex.find(std::make_tuple(object, eventType));
        if (itr != eventIndex.end()) {
            eventIndex.erase(itr);
        }
    }

    const auto &objectIndex = m_filterDataList.get<tag_obj_pointer>();
    if (objectIndex.find(object) == objectIndex.end()) {
        object->removeEventFilter(this);
        disconnect(object, &QObject::destroyed, this, nullptr);
    }
}

bool EventFilterProcess::eventFilter(QObject *obj, QEvent *event)
{
    auto processEvent = [event, obj, this] {
        auto &eventIndex = m_filterDataList.get<tag_obj_composite>();
        auto itr = eventIndex.find(std::make_tuple(obj, event->type()));
        if (itr != eventIndex.end()) {
            if (itr->callback.canConvert<EventCallback>()) {
                const auto &callback = itr->callback.value<EventCallback>();
                callback();
            } else {
                Q_ASSERT(itr->callback.canConvert<EventCallbackWithEvent>() == true);
                const auto &callback = itr->callback.value<EventCallbackWithEvent>();
                callback(event);
            }
            return true;
        }
        return false;
    };

    switch (static_cast<int>(event->type())) {
    case QEvent::MouseButtonPress: {
        QMouseEvent *mouseEvent = static_cast<QMouseEvent*>(event);
        if (mouseEvent->button() == Qt::MouseButton::LeftButton) {
            processEvent();
        }
        break;
    }
    case QEvent::MouseButtonRelease: {
        QMouseEvent *mouseEvent = static_cast<QMouseEvent*>(event);
        if (mouseEvent->button() == Qt::MouseButton::LeftButton) {
            processEvent();
        }
        break;
    }
    case QEvent::MouseMove: {
        QMouseEvent *mouseEvent = static_cast<QMouseEvent*>(event);
        Q_ASSERT(mouseEvent->button() == Qt::MouseButton::NoButton);
        processEvent();
        break;
    }
    default: {
        break;
    }
    }
    return QObject::eventFilter(obj, event);
}


