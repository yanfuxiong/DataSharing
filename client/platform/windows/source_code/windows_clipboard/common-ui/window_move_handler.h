#pragma once
#include <QWidget>
#include <QPoint>
#include <QMouseEvent>
#include <QEvent>

#include "common_signals.h"
#include "common_utils.h"

class WindowMoveHandler
{
public:
    WindowMoveHandler(QWidget *window);
    ~WindowMoveHandler();

private:
    void mousePressEvent(QEvent *event);
    void mouseReleaseEvent(QEvent *event);
    void mouseMoveEvent(QEvent *event);

private:
    QPointer<QWidget> m_window;
    QPoint m_clickedPos;
    bool m_clickedStatus { false };
};
