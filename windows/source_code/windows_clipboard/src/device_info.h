#ifndef DEVICE_INFO_H
#define DEVICE_INFO_H

#include <QWidget>
#include <QLabel>
#include <QPushButton>
#include <QPointer>
#include <QHBoxLayout>
#include <QPaintEvent>
#include <QMouseEvent>
#include "common_utils.h"

namespace Ui {
class DeviceInfo;
}

class DeviceInfo : public QWidget
{
    Q_OBJECT

public:
    explicit DeviceInfo(QWidget *parent = nullptr);
    ~DeviceInfo();

    void setClientName(const QString &name);
    void setClientID(const QByteArray &id) { m_clientID = id; }
    bool isSelected() const { return m_selectedStatus; }
    void setSelected(bool status);
    void resetStatus();

    void sendData();

protected:
    void paintEvent(QPaintEvent *event) override;
    void mousePressEvent(QMouseEvent *event) override;
    void enterEvent(QEvent *event) override;
    void leaveEvent(QEvent *event) override;

    UpdateClientStatusMsgPtr getClientStatusPtr() const;
    QString getIconPath() const;

private:
    Ui::DeviceInfo *ui;
    QByteArray m_clientID;
    bool m_selectedStatus { false };
    bool m_hoverStatus { false };
    QString m_clientName;
    std::map<QByteArray, QByteArray> m_deviceIconMap;
};

#endif // DEVICE_INFO_H
