#include "accept_file_dialog.h"
#include "ui_accept_file_dialog.h"
#include "common_signals.h"
#include <QFileDialog>

AcceptFileDialog::AcceptFileDialog(QWidget *parent) :
    QDialog(parent),
    ui(new Ui::AcceptFileDialog)
{
    ui->setupUi(this);
    ui->icon_left->clear();
    ui->file_info_label->clear();
    ui->path_label->clear();
    ui->device_icon_label->clear();
}

AcceptFileDialog::~AcceptFileDialog()
{
    delete ui;
}

void AcceptFileDialog::setFileInfo(SendFileRequestMsgPtr ptr_msg)
{
    m_cacheMsgPtr = ptr_msg;

    {
        QString infoText;
        for (const auto &client :g_getGlobalData()->m_clientVec) {
            if (client->clientID == m_cacheMsgPtr->clientID) {
                infoText += client->clientName;
                infoText += "\n";
                break;
            }
        }

        infoText += m_cacheMsgPtr->ip;
        ui->device_info_label->setText(infoText);
    }

    {
        QString infoText = QString("%1 (%2)")
                .arg(CommonUtils::getFileNameByPath(ptr_msg->fileName))
                .arg(CommonUtils::byteCountDisplay(ptr_msg->fileSize));
        ui->file_info_label->setText(infoText);

        QString newFilePath = CommonUtils::desktopDirectoryPath() + "/" + CommonUtils::getFileNameByPath(ptr_msg->fileName);
        ui->path_label->setText(newFilePath);
    }
}

QString AcceptFileDialog::filePath() const
{
    return ui->path_label->text();
}

void AcceptFileDialog::on_accept_btn_clicked()
{
    Q_EMIT CommonSignals::getInstance()->userAcceptFile(true);
    accept();
}

void AcceptFileDialog::on_reject_btn_clicked()
{
    Q_EMIT CommonSignals::getInstance()->userAcceptFile(false);
    reject();
}

void AcceptFileDialog::on_modify_button_clicked()
{
    QString dir = QFileDialog::getExistingDirectory(this, "Open Directory",
                                                    CommonUtils::desktopDirectoryPath(),
                                                    QFileDialog::ShowDirsOnly | QFileDialog::DontResolveSymlinks);
    if (dir.isEmpty()) {
        return;
    }

    QString newFilePath = dir + "/" + CommonUtils::getFileNameByPath(m_cacheMsgPtr->fileName);
    ui->path_label->setText(newFilePath);
}
