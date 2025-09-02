#include "warning_dialog.h"
#include "ui_warning_dialog.h"
#include "common_utils.h"

WarningDialog::WarningDialog(QWidget *parent)
    : QDialog(parent)
    , ui(new Ui::WarningDialog)
{
    ui->setupUi(this);
    setWindowFlag(Qt::WindowType::WindowCloseButtonHint, false);
    setWindowFlag(Qt::WindowType::WindowMinMaxButtonsHint, false);

    {
        setProperty(PR_ADJUST_WINDOW_X_SIZE, true);
        setProperty(PR_ADJUST_WINDOW_Y_SIZE, true);

        ui->warning_bottom_box->setProperty(PR_ADJUST_WINDOW_Y_SIZE, true);

        ui->warning_ok_btn->setProperty(PR_ADJUST_WINDOW_X_SIZE, true);
        ui->warning_ok_btn->setProperty(PR_ADJUST_WINDOW_Y_SIZE, true);

        ui->warning_icon_box->setProperty(PR_ADJUST_WINDOW_X_SIZE, true);

        ui->warning_icon_label->setProperty(PR_ADJUST_WINDOW_X_SIZE, true);
        ui->warning_icon_label->setProperty(PR_ADJUST_WINDOW_Y_SIZE, true);
    }

    {
        ui->warning_icon_label->clear();
        setWindowTitle("Version Mismatch Detected");
        QFont font;
        font.setPointSizeF(9);
        ui->warning_ok_btn->setFont(font);
    }
}

WarningDialog::~WarningDialog()
{
    delete ui;
}

void WarningDialog::updateWarningInfo(const QString &versionInfo)
{
    QString formatStr = "Your CrossShare version %1 is lower than another connected client."
                        "Please update to continue.";
    ui->warning_content_label->setText(QString(formatStr).arg(versionInfo));
}

void WarningDialog::on_warning_ok_btn_clicked()
{
    accept();
}

