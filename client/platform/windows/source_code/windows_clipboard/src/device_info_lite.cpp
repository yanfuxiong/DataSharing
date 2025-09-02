#include "device_info_lite.h"
#include "ui_device_info_lite.h"

DeviceInfoLite::DeviceInfoLite(QWidget *parent) :
    QGroupBox(parent),
    ui(new Ui::DeviceInfoLite)
{
    ui->setupUi(this);
}

DeviceInfoLite::~DeviceInfoLite()
{
    delete ui;
}

void DeviceInfoLite::setDisplayInfo(const QString &info)
{
    ui->info_label->setText(info);
}
