#include "file_record_widget.h"
#include "ui_file_record_widget.h"
#include "event_filter_process.h"
#include "file_opt_info_list.h"
#include <QDesktopServices>
#include <QUrl>
#include <QProcess>
#include <QDialog>
#include <QMessageBox>
#include <QPushButton>
#include <QRegularExpression>

FileRecordWidget::FileRecordWidget(QWidget *parent) :
    QWidget(parent),
    ui(new Ui::FileRecordWidget),
    m_openFileLabel(nullptr)
{
    ui->setupUi(this);
    QFont font;
    font.setPointSize(9);
    {
        ui->icon_left->clear();
        ui->icon_right->clear();
        ui->file_info_label->setFont(font);
        ui->file_info_label->clear();
        ui->progressBar->setRange(0, 100);
        ui->progressBar->setValue(0);

        ui->file_opt_desc_label->clear();
        ui->failed_icon_label->clear();
        ui->stackedWidget->setCurrentIndex(0);

        if (cancelTransferFunctionIsEnabled() == false) {
            ui->icon_right->setStyleSheet("QLabel#icon_right{border-image:none;}");
        } else {
            ui->icon_right->setCursor(Qt::PointingHandCursor);
        }

        {
            m_openFileLabel = new QLabel;
            m_openFileLabel->setObjectName("openFileLabel");
            m_openFileLabel->setAlignment(Qt::AlignCenter);
            m_openFileLabel->setFixedSize(ui->icon_right->size());
            ui->horizontalLayout->insertWidget(ui->horizontalLayout->indexOf(ui->icon_right), m_openFileLabel);
            m_openFileLabel->hide();
            m_openFileLabel->setCursor(Qt::PointingHandCursor);
        }
    }

    {
        EventFilterProcess::getInstance()->registerFilterEvent({ m_openFileLabel, std::bind(&FileRecordWidget::processClickedOpenFileIcon, this) });
        EventFilterProcess::getInstance()->registerFilterEvent({ ui->icon_right, std::bind(&FileRecordWidget::processClickedRightIcon, this) });
    }
}

FileRecordWidget::~FileRecordWidget()
{
    delete ui;
}

void FileRecordWidget::processClickedOpenFileIcon()
{
    if (m_fileOptRecord.direction == FileOperationRecord::DirectionType::ReceiveType ||
        m_fileOptRecord.direction == FileOperationRecord::DirectionType::DragSingleFileType) {
        QString filePath = m_fileOptRecord.fileName.c_str();
        if (QFile::exists(filePath) == false && isFullPath(filePath) == false) {
            filePath = CommonUtils::downloadDirectoryPath() + "/" + filePath;
        }
        filePath = QFileInfo(filePath).absoluteFilePath();
        qInfo() << "open file: " << filePath;
        if (QFileInfo::exists(filePath) == false) {
            Q_EMIT CommonSignals::getInstance()->showWarningMessageBox("warning", QString("The file does not exist. path: %1").arg(filePath));
            return;
        }
        QDesktopServices::openUrl(QUrl::fromLocalFile(filePath));
    } else if (m_fileOptRecord.direction == FileOperationRecord::DirectionType::DragMultiFileType) {
        QString folderPath = CommonUtils::downloadDirectoryPath();
        qInfo() << "open file: " << folderPath;
        QDesktopServices::openUrl(QUrl::fromLocalFile(folderPath));
    }
}

void FileRecordWidget::processClickedRightIcon()
{
    auto &uuidIndex = g_getGlobalData()->cacheFileOptRecord.get<tag_db_uuid>();
    auto itr = uuidIndex.find(m_fileOptRecord.uuid);
    if (itr != uuidIndex.end()) {
        if (itr->progressValue >= 100 ||
            itr->optStatus == FileOperationRecord::TransferFileCancelStatus ||
            itr->optStatus == FileOperationRecord::TransferFileErrorStatus) {
            uuidIndex.erase(itr);
            Q_EMIT CommonSignals::getInstance()->updateFileOptInfoList();
        } else {
            if (cancelTransferFunctionIsEnabled() == false) {
                return;
            }

            if (itr->direction == FileOperationRecord::DragSingleFileType || itr->direction == FileOperationRecord::ReceiveType) {
                if (showTerminateSingleFileTransfer(itr->fileName.c_str()) == false) {
                    return;
                }
            } else if (itr->direction == FileOperationRecord::DragMultiFileType) {
                if (showCancelAllTransferDialog() == false) {
                    return;
                }
            } else {
                return;
            }

            FileOptInfoList::updateCacheFileOptRecord(m_fileOptRecord.toRecordData().getHashID(), FileOperationRecord::OptStatusType::TransferFileCancelStatus);

            DragFilesMsg message;
            message.functionCode = DragFilesMsg::FuncCode::CancelFileTransfer;
            message.timeStamp = itr->timeStamp;
            message.ip = itr->ip;
            message.port = itr->port;
            message.clientID = itr->clientID.c_str();

            auto data = DragFilesMsg::toByteArray(message);
            Q_EMIT CommonSignals::getInstance()->sendDataToServer(data);

            updateUI(*itr);
        }
    }
}

void FileRecordWidget::setFileOptInfo(const FileOperationRecord &record)
{
    m_fileOptRecord = record;

    // Synchronize refreshing UI
    if (record.direction == FileOperationRecord::SendType
        || record.direction == FileOperationRecord::ReceiveType
        || record.direction == FileOperationRecord::DragSingleFileType) {
        QString descInfo;
        descInfo += QString("%1 (%2)").arg(CommonUtils::getFileNameByPath(record.fileName.c_str())).arg(CommonUtils::byteCountDisplay(record.fileSize));
        descInfo += "\n";
        descInfo += QDateTime::fromMSecsSinceEpoch(record.timeStamp).toString("yyyy-MM-dd hh:mm:ss");
        descInfo += " " + QString("%1: %2").arg(record.direction == 0 ? "to" : "from").arg(record.clientName.c_str());
        ui->file_info_label->setText(descInfo);
        updateUI(record);
    } else {
        //qInfo() << "----------------------------------------- DragMultiFileType";
        QString descInfo;
        descInfo += QString("Total: %1/%2 files (%3 / %4)")
                            .arg(record.sentFileCount)
                            .arg(record.totalFileCount)
                            .arg(CommonUtils::byteCountDisplay(record.sentFileSize))
                            .arg(CommonUtils::byteCountDisplay(record.totalFileSize));
        descInfo += "\n";
        if (record.progressValue < 100) {
            descInfo += QString("%1 (%2)")
                        .arg(CommonUtils::getFileNameByPath(record.currentTransferFileName))
                        .arg(CommonUtils::byteCountDisplay(record.currentTransferFileSize));
        } else {
            descInfo += QString("All done");
        }
        descInfo += "\n";
        descInfo += QDateTime::fromMSecsSinceEpoch(record.timeStamp).toString("yyyy-MM-dd hh:mm:ss");
        descInfo += " " + QString("%1: %2").arg(record.direction == 0 ? "to" : "from").arg(record.clientName.c_str());
        ui->file_info_label->setText(descInfo);

        updateUI(record);
    }
}

void FileRecordWidget::updateStatusInfo()
{
    const auto &cacheFileOptRecord = g_getGlobalData()->cacheFileOptRecord.get<tag_db_timestamp>();
    for (auto itr = cacheFileOptRecord.begin(); itr != cacheFileOptRecord.end(); ++itr) {
        const auto &record = *itr;
        if (record.toRecordData().getHashID() == m_fileOptRecord.toRecordData().getHashID()) {
            if (record.direction == FileOperationRecord::DragMultiFileType) {
                QString descInfo;
                descInfo += QString("Total: %1/%2 files (%3 / %4)")
                                .arg(record.sentFileCount)
                                .arg(record.totalFileCount)
                                .arg(CommonUtils::byteCountDisplay(record.sentFileSize))
                                .arg(CommonUtils::byteCountDisplay(record.totalFileSize));
                descInfo += "\n";
                if (record.progressValue < 100) {
                    descInfo += QString("%1 (%2)")
                        .arg(CommonUtils::getFileNameByPath(record.currentTransferFileName))
                        .arg(CommonUtils::byteCountDisplay(record.currentTransferFileSize));
                } else {
                    descInfo += QString("All done");
                }
                descInfo += "\n";
                descInfo += QDateTime::fromMSecsSinceEpoch(record.timeStamp).toString("yyyy-MM-dd hh:mm:ss");
                descInfo += " " + QString("%1: %2").arg(record.direction == 0 ? "to" : "from").arg(record.clientName.c_str());
                ui->file_info_label->setText(descInfo);
            }

            QTimer::singleShot(0, this, [this, record] {
                updateUI(record);
            });
            break;
        }
    }
}

void FileRecordWidget::updateUI(const FileOperationRecord &record)
{
    static QString s_iconRightQss = "QLabel#icon_right{border-image: url(:/resource/icon/garbage.svg);}"
                                    "QLabel#icon_right:hover{border-image: url(:/resource/icon/garbage_h.svg);}";

    if (record.progressValue < 100) {
        ui->progressBar->setValue(record.progressValue);
        if (record.optStatus == FileOperationRecord::TransferFileCancelStatus) {
            ui->stackedWidget->setCurrentIndex(1);
            ui->file_opt_desc_label->setText("Transmission cancel  ");
            ui->file_opt_desc_label->setStyleSheet("background:transparent;color:red;");
            ui->icon_right->setStyleSheet(s_iconRightQss);
            ui->failed_icon_label->setStyleSheet("border-image:url(:/resource/icon/transfer_cancel.svg);");

            if (record.direction == FileOperationRecord::DragMultiFileType) {
                if (record.sentFileSize >= record.cacheFileSize) {
                    m_openFileLabel->show();
                }
            }
        } else if (record.optStatus == FileOperationRecord::TransferFileErrorStatus) {
            ui->stackedWidget->setCurrentIndex(1);
            ui->file_opt_desc_label->setText("Transmission failed  ");
            ui->file_opt_desc_label->setStyleSheet("background:transparent;color:red;");
            ui->icon_right->setStyleSheet(s_iconRightQss);
            ui->failed_icon_label->setStyleSheet("border-image:url(:/resource/icon/transfer_cancel.svg);");

            if (record.direction == FileOperationRecord::DragMultiFileType) {
                if (record.sentFileSize >= record.cacheFileSize) {
                    m_openFileLabel->show();
                }
            }
        }
        return;
    }

    if (record.direction == FileOperationRecord::DragMultiFileType) {
        ui->progressBar->setValue(100);
        ui->stackedWidget->setCurrentIndex(1);
        ui->file_opt_desc_label->setText("Transmission completed  ");
        ui->file_opt_desc_label->setStyleSheet("background:transparent;color:green;");
        ui->icon_right->setStyleSheet(s_iconRightQss);
        m_openFileLabel->show();
        return;
    }

    {
        ui->stackedWidget->setCurrentIndex(1);
        ui->file_opt_desc_label->setText("Transmission completed  ");
        ui->file_opt_desc_label->setStyleSheet("background:transparent;color:green;");
        ui->icon_right->setStyleSheet(s_iconRightQss);
        m_openFileLabel->show();
    }
}

bool FileRecordWidget::showCancelAllTransferDialog() const
{
    QMessageBox box;
    box.setFixedSize(250, 143);
    int retVal = 0;
    {
        QPushButton *button = new QPushButton;
        button->setFixedSize(85, 30);
        button->setText("YES");
        box.addButton(button, QMessageBox::ButtonRole::AcceptRole);
        connect(button, &QPushButton::clicked, &box, [&retVal] {
            retVal = 1;
        });
    }

    {
        QPushButton *button = new QPushButton;
        button->setFixedSize(85, 30);
        button->setText("NO");
        box.addButton(button, QMessageBox::ButtonRole::RejectRole);
    }

    box.setWindowTitle("Cancel all transfers in progress");
    box.setText("Are you sure you want to cancel all transfers?");
    box.setIcon(QMessageBox::Icon::Warning);

    QFont font;
    font.setPixelSize(14);
    box.setFont(font);
    box.exec();
    return retVal;
}

bool FileRecordWidget::showTerminateSingleFileTransfer(const QString &fileName) const
{
    QMessageBox box;
    box.setFixedSize(250, 143);
    int retVal = 0;
    {
        QPushButton *button = new QPushButton;
        button->setFixedSize(85, 30);
        button->setText("YES");
        box.addButton(button, QMessageBox::ButtonRole::AcceptRole);
        connect(button, &QPushButton::clicked, &box, [&retVal] {
            retVal = 1;
        });
    }

    {
        QPushButton *button = new QPushButton;
        button->setFixedSize(85, 30);
        button->setText("NO");
        box.addButton(button, QMessageBox::ButtonRole::RejectRole);
    }

    box.setWindowTitle("Cancel this transfers in progress");
    box.setText(QString("Are you sure you want to cancel the transfer of %1?").arg(CommonUtils::getFileNameByPath(fileName)));
    box.setIcon(QMessageBox::Icon::Warning);

    QFont font;
    font.setPixelSize(14);
    box.setFont(font);
    box.exec();
    return retVal;
}

std::string FileRecordWidget::getClientID() const
{
    return m_fileOptRecord.clientID;
}

QByteArray FileRecordWidget::getHashID() const
{
    return m_fileOptRecord.toRecordData().getHashID();
}

bool FileRecordWidget::cancelTransferFunctionIsEnabled() const
{
    try {
        return g_getGlobalData()->localConfig.at("filesRecords").at("enableCancelTransfer").get<bool>();
    } catch (const std::exception &e) {
        qWarning() << e.what();
        return false;
    }
}

bool FileRecordWidget::isFullPath(const QString &path)
{
    static QRegularExpression s_regExp(R"(^[a-zA-Z]:)");
    return s_regExp.match(path).hasMatch();
}
