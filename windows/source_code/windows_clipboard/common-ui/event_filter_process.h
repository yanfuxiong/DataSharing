#pragma once
#include <QObject>
#include "common_signals.h"
#include "common_utils.h"

struct EventFilterData
{
    EventFilterData(QPointer<QObject> pObject, const EventCallback &cb, QEvent::Type type = QEvent::Type::MouseButtonPress)
        : monitoredObject(pObject)
        , callback(cb)
        , eventType(type)
    {}

    QPointer<QObject> monitoredObject;
    EventCallback callback;
    QEvent::Type eventType;
};

class EventFilterProcess : public QObject
{
    Q_OBJECT
public:
    ~EventFilterProcess();
    static EventFilterProcess *getInstance();

    void registerFilterEvent(const EventFilterData &filterData);

protected:
    bool eventFilter(QObject *obj, QEvent *event) override;

private:
    EventFilterProcess(QObject *parent = nullptr);

    static EventFilterProcess *m_instance;
    QList<EventFilterData> m_filterDataList;
};
