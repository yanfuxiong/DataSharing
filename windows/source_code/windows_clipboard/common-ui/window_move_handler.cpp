#include "window_move_handler.h"
#include "event_filter_process.h"

WindowMoveHandler::WindowMoveHandler(QWidget *window)
    : m_window(window)
{
    EventFilterProcess::getInstance()->registerFilterEvent({
        m_window.data(), std::bind(&WindowMoveHandler::mousePressEvent, this, std::placeholders::_1), QEvent::MouseButtonPress
    });

    EventFilterProcess::getInstance()->registerFilterEvent({
        m_window.data(), std::bind(&WindowMoveHandler::mouseReleaseEvent, this, std::placeholders::_1), QEvent::MouseButtonRelease
    });

    EventFilterProcess::getInstance()->registerFilterEvent({
        m_window.data(), std::bind(&WindowMoveHandler::mouseMoveEvent, this, std::placeholders::_1), QEvent::MouseMove
    });
}

WindowMoveHandler::~WindowMoveHandler()
{
    if (m_window == nullptr) {
        return;
    }
    EventFilterProcess::getInstance()->removeFilterEvent(m_window,
                                        { QEvent::MouseButtonPress, QEvent::MouseButtonRelease, QEvent::MouseMove });
    qDebug() << "-----------------WindowMoveHandler::~WindowMoveHandler()";
}

void WindowMoveHandler::mousePressEvent(QEvent *ev)
{
    Q_ASSERT(dynamic_cast<QMouseEvent*>(ev) != nullptr);
    QMouseEvent *event = static_cast<QMouseEvent*>(ev);
    m_clickedStatus = true;
    m_clickedPos = event->globalPos() - m_window->frameGeometry().topLeft();
    event->accept();
}

void WindowMoveHandler::mouseReleaseEvent(QEvent *ev)
{
    Q_ASSERT(dynamic_cast<QMouseEvent*>(ev) != nullptr);
    QMouseEvent *event = static_cast<QMouseEvent*>(ev);
    m_clickedStatus = false;
    event->accept();
}

void WindowMoveHandler::mouseMoveEvent(QEvent *ev)
{
    Q_ASSERT(dynamic_cast<QMouseEvent*>(ev) != nullptr);
    QMouseEvent *event = static_cast<QMouseEvent*>(ev);
    if (m_clickedStatus == false) {
        event->accept();
        return;
    }
    m_window->move(m_window->mapToGlobal(event->pos() - m_clickedPos));
}

