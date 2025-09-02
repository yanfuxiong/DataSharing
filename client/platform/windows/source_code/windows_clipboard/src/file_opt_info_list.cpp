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
        connect(CommonSignals::getInstance(), &CommonSignals::updateOptRecordStatus, this, &FileOptInfoList::onUpdateOptRecordStatus);
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

    const auto &fileOptRecordVec = g_getGlobalData()->cacheFileOptRecord.get<tag_db_timestamp>();
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
    backgoundWidget->setObjectName("FileRecordWidgetBox");
    auto flowLayout = new QVBoxLayout;
    flowLayout->setSpacing(8);
    flowLayout->setMargin(0);
    backgoundWidget->setLayout(flowLayout);
    area->setWidget(backgoundWidget); // 放入滚动区域
    ui->content->addWidget(area);
    Q_ASSERT(area->parent() != nullptr);
    Q_ASSERT(backgoundWidget->parent() != nullptr);

    for (auto itr = fileOptRecordVec.begin(); itr != fileOptRecordVec.end(); ++itr) {
        const auto &data = *itr;
        if (data.direction == 0) {
            // FIXME: According to the latest requirements,
            // the operation of actively sending files will not be recorded, so it will be skipped here and not displayed
#ifdef NDEBUG
            continue; // release 模式下执行
#endif
        }
        QGroupBox *box = new QGroupBox;
        box->setFixedHeight(70);
        box->setProperty(PR_ADJUST_WINDOW_Y_SIZE, true);
        box->setObjectName("FileRecordWidgetBox");
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
    auto &cacheFileOptRecord = g_getGlobalData()->cacheFileOptRecord.get<tag_db_timestamp>();

    for (auto itr = cacheFileOptRecord.begin(); itr != cacheFileOptRecord.end(); ++itr) {
        const auto &recordData = *itr;
        if (recordData.progressValue == -1
            || recordData.progressValue > currentVal
            || recordData.optStatus == FileOperationRecord::TransferFileCancelStatus) {
            continue;
        }
        if (recordData.toRecordData().getHashID() == hashID) {
            //recordData.progressValue = currentVal;
            cacheFileOptRecord.modify(itr, [currentVal] (FileOperationRecord &data) {
                data.progressValue = currentVal;
            });
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

void FileOptInfoList::updateCacheFileOptRecord(const QByteArray &hashID, UpdateProgressMsgPtr ptrMsg)
{
    auto &cacheFileOptRecord = g_getGlobalData()->cacheFileOptRecord.get<tag_db_timestamp>();

    for (auto itr = cacheFileOptRecord.begin(); itr != cacheFileOptRecord.end(); ++itr) {
        const auto &recordData = *itr;
        if (recordData.toRecordData().getHashID() == hashID) {
            cacheFileOptRecord.modify(itr, [ptrMsg] (FileOperationRecord &data) {
                data.fileName = ptrMsg->fileName.toStdString();
                data.fileSize = ptrMsg->fileSize;

                data.sentFileCount = ptrMsg->sentFilesCount;
                data.totalFileCount = ptrMsg->totalFilesCount;
                data.sentFileSize = ptrMsg->totalSentSize;
                data.totalFileSize = ptrMsg->totalFilesSize;
                data.currentTransferFileName = ptrMsg->currentFileName;
                data.currentTransferFileSize = ptrMsg->currentFileSize;

                if (ptrMsg->totalFilesCount <= 1) {
                    data.fileName = ptrMsg->currentFileName.toStdString();
                    data.fileSize = ptrMsg->currentFileSize;
                }
            });
            break;
        }
    }
}

void FileOptInfoList::updateCacheFileOptRecord(const QByteArray &hashID, int optStatus)
{
    auto &cacheFileOptRecord = g_getGlobalData()->cacheFileOptRecord.get<tag_db_timestamp>();

    for (auto itr = cacheFileOptRecord.begin(); itr != cacheFileOptRecord.end(); ++itr) {
        const auto &recordData = *itr;
        if (recordData.toRecordData().getHashID() == hashID) {
            cacheFileOptRecord.modify(itr, [optStatus] (FileOperationRecord &data) {
                if (data.optStatus == FileOperationRecord::TransferFileCancelStatus) {
                    return;
                }
                data.optStatus = optStatus;
            });
            break;
        }
    }
}

void FileOptInfoList::onUpdateOptRecordStatus(const QByteArray &hashID, int optStatus)
{
    updateCacheFileOptRecord(hashID, optStatus);
    for (const auto &widget : m_recordWidgetList) {
        if (widget->getHashID() == hashID) {
            widget->updateStatusInfo();
            break;
        }
    }
}
