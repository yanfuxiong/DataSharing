#pragma once
#include <QObject>
#include "common_signals.h"
#include "common_utils.h"

struct EventFilterData
{
    EventFilterData(QPointer<QObject> pObject, const EventCallback &cb, QEvent::Type type = QEvent::Type::MouseButtonPress)
        : monitoredObject(pObject)
        , callback(QVariant::fromValue(cb))
        , eventType(type)
    {}

    EventFilterData(QPointer<QObject> pObject, const EventCallbackWithEvent &cb, QEvent::Type type)
        : monitoredObject(pObject)
        , callback(QVariant::fromValue(cb))
        , eventType(type)
    {}

    QObject *objectPointer() const { return monitoredObject.data(); }

    QPointer<QObject> monitoredObject;
    QVariant callback;
    QEvent::Type eventType;
};

struct tag_obj_pointer{};
struct tag_obj_event{};
struct tag_obj_composite{};

using EventFilterDataContainer = multi_index_container<
    EventFilterData,
    indexed_by<
        sequenced<>,
        ordered_non_unique<tag<tag_obj_pointer>, key<&EventFilterData::objectPointer> >,
        ordered_non_unique<tag<tag_obj_event>, key<&EventFilterData::eventType> >,
        ordered_unique<
            tag<tag_obj_composite>,
            composite_key<
                EventFilterData,
                key<&EventFilterData::objectPointer>,
                key<&EventFilterData::eventType>
            >
        >
    >
>;

class EventFilterProcess : public QObject
{
    Q_OBJECT
public:
    ~EventFilterProcess();
    static EventFilterProcess *getInstance();

    void registerFilterEvent(const EventFilterData &filterData);
    void removeFilterEvent(QObject *object, const QList<QEvent::Type> &eventList);

protected:
    bool eventFilter(QObject *obj, QEvent *event) override;

private:
    EventFilterProcess(QObject *parent = nullptr);

    static EventFilterProcess *m_instance;
    EventFilterDataContainer m_filterDataList;
};
