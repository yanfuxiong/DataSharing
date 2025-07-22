#ifndef ACCEPT_FILE_DIALOG_H
#define ACCEPT_FILE_DIALOG_H

#include <QDialog>
#include "common_utils.h"
#include "common_signals.h"

namespace Ui {
class AcceptFileDialog;
}

class AcceptFileDialog : public QDialog
{
    Q_OBJECT

public:
    explicit AcceptFileDialog(QWidget *parent = nullptr);
    ~AcceptFileDialog();

    void setFileInfo(SendFileRequestMsgPtr ptr_msg);
    QString filePath() const;

private slots:
    void on_accept_btn_clicked();
    void on_reject_btn_clicked();

    void on_modify_button_clicked();

private:
    Ui::AcceptFileDialog *ui;
    SendFileRequestMsgPtr m_cacheMsgPtr;
};

#endif // ACCEPT_FILE_DIALOG_H
