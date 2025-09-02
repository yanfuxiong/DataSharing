#ifndef WARNING_DIALOG_H
#define WARNING_DIALOG_H

#include <QDialog>

namespace Ui {
class WarningDialog;
}

class WarningDialog : public QDialog
{
    Q_OBJECT

public:
    explicit WarningDialog(QWidget *parent = nullptr);
    ~WarningDialog();

    void updateWarningInfo(const QString &versionInfo);

private slots:
    void on_warning_ok_btn_clicked();

private:
    Ui::WarningDialog *ui;
};

#endif // WARNING_DIALOG_H
