#include "device_list_dialog.h"
#include "ui_device_list_dialog.h"
#include "common_signals.h"
#include "event_filter_process.h"
#include <QTimer>
#include <QVBoxLayout>
#include <QEventLoop>

DeviceListDialog::DeviceListDialog(QWidget *parent) :
    QDialog(parent),
    ui(new Ui::DeviceListDialog)
{
    ui->setupUi(this);
    {
        ui->icon_left->clear();
        ui->display_selection->clear();
        ui->file_info_label->clear();

        QString fileName = g_getGlobalData()->selectedFileName;
        ui->file_info_label->setText(QString("%1 (%2)")
                                     .arg(CommonUtils::getFileNameByPath(fileName))
                                     .arg(CommonUtils::byteCountDisplay(QFileInfo(fileName).size())));
    }

    {
        //connect(CommonSignals::getInstance(), &CommonSignals::sendDataToServer, this, &DeviceListDialog::onSendDataToServer);
        connect(CommonSignals::getInstance(), &CommonSignals::updateClientList, this, &DeviceListDialog::onUpdateClientList);
        connect(CommonSignals::getInstance(), &CommonSignals::updateUserSelectedInfo, this, &DeviceListDialog::onUpdateUserSelectedInfo);

        EventFilterProcess::getInstance()->registerFilterEvent({ ui->clear_selected_label, std::bind(&DeviceListDialog::onCliked_clear_selected_label, this) });
    }

    QTimer::singleShot(0, this, [this] {
        onUpdateClientList();
    });

    g_getGlobalData()->m_selectedClientVec.clear(); // The handling here is necessary
}

DeviceListDialog::~DeviceListDialog()
{
    delete ui;
}

void DeviceListDialog::onSendDataToServer(const QByteArray &data)
{
    Q_UNUSED(data)
}

void DeviceListDialog::onUpdateClientList()
{
    {
        g_getGlobalData()->m_selectedClientVec.clear();
        ui->display_selection->clear();
    }

    while (ui->content->count() > 0) {
        auto widget = ui->content->widget(0);
        ui->content->removeWidget(widget);
        widget->deleteLater();
    }

    m_deviceList.clear();

    auto backgoundWidget = new QWidget;
    auto flowLayout = new FlowLayout(5, 8, 8);
    //auto flowLayout = new QVBoxLayout;
    flowLayout->setSpacing(8);
    flowLayout->setMargin(5);
    backgoundWidget->setLayout(flowLayout);
    ui->content->addWidget(backgoundWidget);

    const auto &clientVec = g_getGlobalData()->m_clientVec;
    for (const auto &data : clientVec) {
        auto widget = new DeviceInfo;
        m_deviceList.push_back(widget);
        {
            nlohmann::json deviceInfoJson;
            deviceInfoJson["clientName"] = data->clientName.toStdString();
            deviceInfoJson["clientID"] = data->clientID.toStdString();
        }
        widget->setClientName(data->clientName);
        widget->setClientID(data->clientID);
        flowLayout->addWidget(widget);
    }
    //flowLayout->addStretch();
    ui->content->setCurrentIndex(0);
}

void DeviceListDialog::onUpdateUserSelectedInfo()
{
    qInfo() << "--------------------selected:" << g_getGlobalData()->m_selectedClientVec.size();
    ui->display_selection->setText(QString("Number of selected devices: %1").arg(g_getGlobalData()->m_selectedClientVec.size()));
}

void DeviceListDialog::onCliked_clear_selected_label()
{
    for (const auto &device : m_deviceList) {
        device->resetStatus();
    }
}

void DeviceListDialog::on_confirm_btn_clicked()
{
    bool exists = false;
    for (const auto &device : m_deviceList) {
        if (device->isSelected()) {
            device->sendData(); // Here, data transmission and related operation records are processed
            exists = true;

            {
                QEventLoop eventLoop;
                QTimer::singleShot(10, Qt::TimerType::PreciseTimer, &eventLoop, [&eventLoop] {
                    eventLoop.quit();
                });
                eventLoop.exec();
            }
        }
    }

    if (exists == false) {
        CommonSignals::getInstance()->showWarningMessageBox("warning", "No device selected !!!");
        return;
    }
    accept();
    QTimer::singleShot(0, qApp, [] {
        Q_EMIT CommonSignals::getInstance()->updateFileOptInfoList();
    });
}
