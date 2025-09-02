#include "tips_list_area.h"
#include "ui_tips_list_area.h"
#include "device_info_lite.h"
#include "common_signals.h"
#include "common_utils.h"
#include <QHBoxLayout>
#include <QVBoxLayout>

TipsListArea::TipsListArea(QWidget *parent) :
    QScrollArea(parent),
    ui(new Ui::TipsListArea)
{
    ui->setupUi(this);
    setObjectName("TipsListArea");

    {
        this->setSizePolicy(QSizePolicy::Expanding, QSizePolicy::Expanding);
        this->setHorizontalScrollBarPolicy(Qt::ScrollBarPolicy::ScrollBarAlwaysOff);
        this->setVerticalScrollBarPolicy(Qt::ScrollBarPolicy::ScrollBarAsNeeded);
        this->setWidgetResizable(true);
    }

    {
        // Remove the client list display
        //connect(CommonSignals::getInstance(), &CommonSignals::updateClientList, this, &TipsListArea::onUpdateClientList);
    }
}

TipsListArea::~TipsListArea()
{
    delete ui;
}

void TipsListArea::onUpdateClientList()
{
    if (auto pWidget = takeWidget()) {
        pWidget->deleteLater();
    }

    QWidget *pWidget = new QWidget;
    pWidget->setSizePolicy(QSizePolicy::Expanding, QSizePolicy::Expanding);
    QVBoxLayout *pLayout = new QVBoxLayout;
    pLayout->setContentsMargins(30, 0, 5, 0);
    pLayout->setSpacing(5);
    pWidget->setLayout(pLayout);

    for (const auto &data : g_getGlobalData()->m_clientVec) {
        DeviceInfoLite *ptr_info_widget = new DeviceInfoLite;
        QString info;
        info += data->clientName + "\n";
        info += data->ip + ":" + QString::number(data->port);
        ptr_info_widget->setDisplayInfo(info);

        pLayout->addWidget(ptr_info_widget);
    }

    pLayout->addStretch();

    setWidget(pWidget);
}

