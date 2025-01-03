#ifndef DEVICE_INFO_LITE_H
#define DEVICE_INFO_LITE_H

#include <QGroupBox>

namespace Ui {
class DeviceInfoLite;
}

class DeviceInfoLite : public QGroupBox
{
    Q_OBJECT

public:
    explicit DeviceInfoLite(QWidget *parent = nullptr);
    ~DeviceInfoLite();

    void setDisplayInfo(const QString &info);

private:
    Ui::DeviceInfoLite *ui;
};

#endif // DEVICE_INFO_LITE_H
