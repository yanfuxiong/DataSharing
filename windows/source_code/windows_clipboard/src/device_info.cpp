#include "device_info.h"
#include "ui_device_info.h"
#include "common_signals.h"
#include <QFileInfo>
#include <QMessageBox>
#include <QFileDialog>
#include <QPainter>
#include <QPixmap>

DeviceInfo::DeviceInfo(QWidget *parent) :
    QWidget(parent),
    ui(new Ui::DeviceInfo)
{
    ui->setupUi(this);
}

DeviceInfo::~DeviceInfo()
{
    delete ui;
}

void DeviceInfo::setClientName(const QString &name)
{
    m_clientName = name;
}

void DeviceInfo::paintEvent(QPaintEvent *event)
{
    Q_UNUSED(event)
    QPainter painter(this);
    painter.setRenderHint(QPainter::Antialiasing, true);
    const QPixmap pixmap(":/resource/icon/DeviceComputer.svg");
    const int marginDelta = 4;

    const QRect pixRect((width() - pixmap.width() - marginDelta) / 2, marginDelta, pixmap.width(), pixmap.height());

    if (m_hoverStatus || m_selectedStatus) {
        painter.save();
        painter.setPen(QPen(QColor(0, 122, 255), 3, Qt::SolidLine));
        painter.drawEllipse(pixRect.adjusted(-3, -3, 3, 3));
        painter.restore();
    } else {
        //painter.fillRect(event->rect(), Qt::lightGray);
    }

    {
        painter.drawPixmap(pixRect, pixmap);
    }

    // 绘制选中图标
    if (m_selectedStatus) {
        QPixmap checkPixmap(":/resource/icon/check_big.svg");
        double r_val = pixmap.width() / 2.0;
        double delta_x = r_val * cos(M_PI / 4.0);
        double delta_y = r_val * sin(M_PI / 4.0);
        QPointF centerPoint(pixRect.center().x() + delta_x, pixRect.center().y() + delta_y);
        QPoint newCenterPoint = centerPoint.toPoint();
        const int check_pix_width = 20;
        QRect checkPixRect(newCenterPoint.x() - check_pix_width / 2 + 4,
                           newCenterPoint.y() - check_pix_width / 2 + 4,
                           check_pix_width,
                           check_pix_width);
        painter.drawPixmap(checkPixRect, checkPixmap);
    }

    {
        painter.save();

        QFont font;
        font.setPixelSize(14);
        font.setBold(true);
        painter.setFont(font);
        QTextOption option;
        option.setWrapMode(QTextOption::WrapMode::WrapAtWordBoundaryOrAnywhere);
        option.setAlignment(Qt::AlignHCenter | Qt::AlignTop);
        int delta = 10 + marginDelta;
        painter.drawText(QRectF(0, pixmap.height() + delta, width(), height() - delta),
                         m_clientName,
                         option);

        painter.restore();
    }

}

void DeviceInfo::mousePressEvent(QMouseEvent *event)
{
    Q_UNUSED(event)
    setSelected(!m_selectedStatus);
}

void DeviceInfo::setSelected(bool status)
{
    if (m_selectedStatus == status) {
        return;
    }
    m_selectedStatus = status;

    if (m_selectedStatus == false) {
        for (auto itr = g_getGlobalData()->m_selectedClientVec.begin(); itr != g_getGlobalData()->m_selectedClientVec.end(); ++itr) {
            if ((*itr)->clientID == m_clientID) {
                g_getGlobalData()->m_selectedClientVec.erase(itr);
                break;
            }
        }
    } else {
        UpdateClientStatusMsgPtr ptr_client = getClientStatusPtr();
        g_getGlobalData()->m_selectedClientVec.push_back(ptr_client);
    }
    Q_EMIT CommonSignals::getInstance()->updateUserSelectedInfo();
    repaint();
}

void DeviceInfo::resetStatus()
{
    m_hoverStatus = false;
    setSelected(false);
}

void DeviceInfo::enterEvent(QEvent *event)
{
    Q_UNUSED(event)
    if (m_hoverStatus || m_selectedStatus) {
        return;
    }
    m_hoverStatus = true;
    repaint();
}

void DeviceInfo::leaveEvent(QEvent *event)
{
    Q_UNUSED(event)
    if (m_hoverStatus == false || m_selectedStatus) {
        return;
    }
    m_hoverStatus = false;
    repaint();
}

UpdateClientStatusMsgPtr DeviceInfo::getClientStatusPtr() const
{
    UpdateClientStatusMsgPtr ptr_client;
    for (const auto &data : g_getGlobalData()->m_clientVec) {
        if (data->clientID == m_clientID) {
            ptr_client = data;
            break;
        }
    }
    return ptr_client;
}

void DeviceInfo::sendData()
{
    UpdateClientStatusMsgPtr ptr_client = getClientStatusPtr();

    Q_ASSERT(ptr_client != nullptr);
    // FIXME: 有待完善
    {
        FileOperationRecord record;
        record.fileName = g_getGlobalData()->selectedFileName.toStdString();
        record.fileSize = QFileInfo(g_getGlobalData()->selectedFileName).size();
        record.timeStamp = QDateTime::currentDateTime().toMSecsSinceEpoch();
        record.progressValue = 0;
        record.clientName = ptr_client->clientName.toStdString();
        record.clientID = ptr_client->clientID.toStdString();
        record.ip = ptr_client->ip;
        record.direction = 0;

        g_getGlobalData()->cacheFileOptRecord.push_back(record);
        Q_EMIT CommonSignals::getInstance()->updateFileOptInfoList();
    }

    SendFileRequestMsg msg;
    msg.ip = ptr_client->ip;
    msg.port = ptr_client->port;
    msg.clientID = ptr_client->clientID;
    msg.fileSize = QFileInfo(g_getGlobalData()->selectedFileName).size();
    msg.timeStamp = QDateTime::currentDateTime().toUTC().toMSecsSinceEpoch();
    msg.fileName = g_getGlobalData()->selectedFileName;

    QByteArray send_data = SendFileRequestMsg::toByteArray(msg);
    Q_EMIT CommonSignals::getInstance()->sendDataToServer(send_data);
}
