#include "drag_drop_widget.h"
#include "ui_drag_drop_widget.h"
#include "device_list_dialog.h"
#include "common_utils.h"

DragDropWidget::DragDropWidget(QWidget *parent) :
    QWidget(parent),
    ui(new Ui::DragDropWidget)
{
    ui->setupUi(this);
    setAcceptDrops(true);

    {
        ui->drag_drop_image_label->clear();
    }
}

DragDropWidget::~DragDropWidget()
{
    delete ui;
}

void DragDropWidget::paintEvent(QPaintEvent *event)
{
    QWidget::paintEvent(event);

    if (m_hoverStatus == false) {
        return;
    }

    QPainter painter(this);
    painter.setRenderHint(QPainter::Antialiasing, true);
    painter.fillRect(event->rect(), QColor(255, 0, 0, 150));
}

void DragDropWidget::dragEnterEvent(QDragEnterEvent *event)
{
    if (event->mimeData()->hasUrls()) {
        QString filePath = event->mimeData()->urls().front().toLocalFile();
        if (QFileInfo(filePath).isFile()) {
            {
                m_hoverStatus = true;
                repaint();
            }
            event->acceptProposedAction();
        }
    }
}

void DragDropWidget::dragLeaveEvent(QDragLeaveEvent *event)
{
    qInfo() << "-------------- Drag and drop to abandon......";
    event->accept();

    {
        m_hoverStatus = false;
        repaint();
    }
}

void DragDropWidget::dropEvent(QDropEvent *event)
{
    Q_ASSERT(event->mimeData()->hasUrls() && event->mimeData()->urls().isEmpty() == false);

    {
        m_hoverStatus = false;
        repaint();
    }

    g_getGlobalData()->selectedFileName = event->mimeData()->urls().front().toLocalFile();
    qInfo() << "Decided to drag and drop:" << g_getGlobalData()->selectedFileName.toUtf8().constData();
    event->acceptProposedAction();
    {
        DeviceListDialog dialog;
        dialog.setWindowTitle("Select device");
        dialog.exec();
    }
}

