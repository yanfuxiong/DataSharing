#include "file_opt_info_list.h"
#include "ui_file_opt_info_list.h"
#include <QGroupBox>

FileOptInfoList::FileOptInfoList(QWidget *parent) :
    QWidget(parent),
    ui(new Ui::FileOptInfoList)
{
    ui->setupUi(this);
    {
        connect(CommonSignals::getInstance(), &CommonSignals::updateFileOptInfoList, this, &FileOptInfoList::onUpdateFileOptInfoList);
        connect(CommonSignals::getInstance(), &CommonSignals::updateProgressInfoWithID, this, &FileOptInfoList::onUpdateProgressInfoWithID);
        // FIXME: 仅用于测试
        //connect(CommonSignals::getInstance(), &CommonSignals::updateClientList, this, &FileOptInfoList::onUpdateFileOptInfoList);
    }

    onUpdateFileOptInfoList();
}

FileOptInfoList::~FileOptInfoList()
{
    delete ui;
}


void FileOptInfoList::onUpdateFileOptInfoList()
{
    while (ui->content->count() > 0) {
        auto widget = ui->content->widget(0);
        ui->content->removeWidget(widget);
        widget->deleteLater();
    }

    m_recordWidgetList.clear();

    const auto &fileOptRecordVec = g_getGlobalData()->cacheFileOptRecord;
//    if (fileOptRecordVec.empty()) {
//        auto imageWidget = new QLabel(this);
//        imageWidget->setStyleSheet("border-image:url(:/resource/background.jpg);");
//        ui->content->addWidget(imageWidget);
//        ui->content->setCurrentIndex(0);
//        return;
//    }

    auto area = new QScrollArea;
    {
        area->setSizePolicy(QSizePolicy::Expanding, QSizePolicy::Expanding);
        area->setHorizontalScrollBarPolicy(Qt::ScrollBarPolicy::ScrollBarAlwaysOff);
        area->setVerticalScrollBarPolicy(Qt::ScrollBarPolicy::ScrollBarAsNeeded);
        area->setWidgetResizable(true);
    }

    auto backgoundWidget = new QWidget;
    auto flowLayout = new QVBoxLayout;
    flowLayout->setSpacing(8);
    flowLayout->setMargin(0);
    backgoundWidget->setLayout(flowLayout);
    area->setWidget(backgoundWidget); // 放入滚动区域
    ui->content->addWidget(area);
    Q_ASSERT(area->parent() != nullptr);
    Q_ASSERT(backgoundWidget->parent() != nullptr);

    for (auto itr = fileOptRecordVec.rbegin(); itr != fileOptRecordVec.rend(); ++itr) {
        const auto &data = *itr;
        QGroupBox *box = new QGroupBox; // 作为 FileRecordWidget 外层的父窗口
        box->setObjectName("FileRecordWidgetBox"); // 名字与css文件对应
        {
            QHBoxLayout *pLayout = new QHBoxLayout;
            pLayout->setSpacing(0);
            pLayout->setMargin(1);
            box->setLayout(pLayout);

            auto widget = new FileRecordWidget;
            m_recordWidgetList.push_back(widget);
            widget->setFileOptInfo(data);
            pLayout->addWidget(widget);
        }
        flowLayout->addWidget(box);
    }
    flowLayout->addStretch();
    ui->content->setCurrentIndex(0);
}

void FileOptInfoList::onUpdateProgressInfoWithID(int currentVal, const QByteArray &hashID)
{
    auto &cacheFileOptRecord = g_getGlobalData()->cacheFileOptRecord;

    // 倒序查找, 以最新的为准
    for (auto itr = cacheFileOptRecord.rbegin(); itr != cacheFileOptRecord.rend(); ++itr) {
        auto &recordData = *itr;
        if (recordData.progressValue == -1 || recordData.progressValue > currentVal) {
            continue;
        }
        if (recordData.toRecordData().getHashID() == hashID) {
            recordData.progressValue = currentVal;
            break;
        }
    }

    for (const auto &widget : m_recordWidgetList) {
        if (widget->getHashID() == hashID) {
            widget->updateStatusInfo();
            break;
        }
    }
}
