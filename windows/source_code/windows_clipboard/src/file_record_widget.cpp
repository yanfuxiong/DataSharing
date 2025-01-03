#include "file_record_widget.h"
#include "ui_file_record_widget.h"
#include "event_filter_process.h"

FileRecordWidget::FileRecordWidget(QWidget *parent) :
    QWidget(parent),
    ui(new Ui::FileRecordWidget)
{
    ui->setupUi(this);

    {
        ui->icon_left->clear();
        ui->icon_right->clear();
        ui->file_info_label->clear();
        ui->progressBar->setRange(0, 100);
        ui->progressBar->setValue(0);

        ui->file_opt_desc_label->clear();
        ui->failed_icon_label->clear();
        ui->stackedWidget->setCurrentIndex(0);
    }
}

FileRecordWidget::~FileRecordWidget()
{
    delete ui;
}

void FileRecordWidget::setFileOptInfo(const FileOperationRecord &record)
{
    m_fileOptRecord = record;

    // Synchronize refreshing UI
    {
        QString descInfo;
        descInfo += QString("%1 (%2)").arg(CommonUtils::getFileNameByPath(record.fileName.c_str())).arg(CommonUtils::byteCountDisplay(record.fileSize));
        descInfo += "\n";
        descInfo += QDateTime::fromMSecsSinceEpoch(record.timeStamp).toString("yyyy-MM-dd hh:mm:ss");
        descInfo += " " + QString("%1: %2").arg(record.direction == 0 ? "to" : "from").arg(record.clientName.c_str());
        ui->file_info_label->setText(descInfo);

        if (record.progressValue >= 100) {
            ui->stackedWidget->setCurrentIndex(1);
            ui->file_opt_desc_label->setText("Transmission completed  ");
            ui->file_opt_desc_label->setStyleSheet("background:transparent;color:green;");
        }
    }
}

void FileRecordWidget::updateStatusInfo()
{
    const auto &cacheFileOptRecord = g_getGlobalData()->cacheFileOptRecord.get<tag_db_timestamp>();
    for (auto itr = cacheFileOptRecord.begin(); itr != cacheFileOptRecord.end(); ++itr) {
        const auto &fileOptRecord = *itr;
        if (fileOptRecord.clientID == getClientID()) {
            ui->progressBar->setValue(fileOptRecord.progressValue);

            if (fileOptRecord.progressValue >= 100) {
                QTimer::singleShot(50, this, [this] {
                    ui->stackedWidget->setCurrentIndex(1);
                    ui->file_opt_desc_label->setText("Transmission completed  ");
                    ui->file_opt_desc_label->setStyleSheet("background:transparent;color:green;");
                });
            }
            break;
        }
    }
}

std::string FileRecordWidget::getClientID() const
{
    return m_fileOptRecord.clientID;
}

QByteArray FileRecordWidget::getHashID() const
{
    return m_fileOptRecord.toRecordData().getHashID();
}
