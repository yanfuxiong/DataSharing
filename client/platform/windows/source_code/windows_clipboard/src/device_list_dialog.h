#ifndef DEVICE_LIST_DIALOG_H
#define DEVICE_LIST_DIALOG_H

#include <QDialog>
#include "flowlayout.h"
#include "device_info.h"

namespace Ui {
class DeviceListDialog;
}

class DeviceListDialog : public QDialog
{
    Q_OBJECT

public:
    explicit DeviceListDialog(QWidget *parent = nullptr);
    ~DeviceListDialog();

private Q_SLOTS:
    void onSendDataToServer(const QByteArray &data);
    void onUpdateClientList();
    void onUpdateUserSelectedInfo();

    void onCliked_clear_selected_label();

    void on_confirm_btn_clicked();

private:
    Ui::DeviceListDialog *ui;
    QList<QPointer<DeviceInfo> > m_deviceList;
};

#endif // DEVICE_LIST_DIALOG_H
