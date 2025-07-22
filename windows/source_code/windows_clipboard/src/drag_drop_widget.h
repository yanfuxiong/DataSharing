#ifndef DRAG_DROP_WIDGET_H
#define DRAG_DROP_WIDGET_H

#include <QWidget>
#include <QDragEnterEvent>
#include <QDragLeaveEvent>
#include <QMimeData>
#include <QPaintEvent>
#include <QPainter>

namespace Ui {
class DragDropWidget;
}

class DragDropWidget : public QWidget
{
    Q_OBJECT

public:
    explicit DragDropWidget(QWidget *parent = nullptr);
    ~DragDropWidget();

protected:
    void dragEnterEvent(QDragEnterEvent *event) override;
    void dragLeaveEvent(QDragLeaveEvent *event) override;
    void dropEvent(QDropEvent *event) override;

    void paintEvent(QPaintEvent *event) override;

private:
    Ui::DragDropWidget *ui;
    bool m_hoverStatus { false };
};

#endif // DRAG_DROP_WIDGET_H
