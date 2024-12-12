#include "progress_bar_dialog.h"
#include "ui_progress_bar_dialog.h"
#include "common_signals.h"

ProgressBarDialog::ProgressBarDialog(QWidget *parent) :
    QDialog(parent),
    ui(new Ui::ProgressBarDialog)
{
    setWindowFlag(Qt::WindowType::WindowCloseButtonHint, false);
    ui->setupUi(this);
    ui->progressBar->setRange(0, 100);
    ui->progressBar->setValue(0);
    connect(CommonSignals::getInstance(), &CommonSignals::updateProgressInfo, this, &ProgressBarDialog::onUpdateProgressInfo);
}

ProgressBarDialog::~ProgressBarDialog()
{
    delete ui;
}

void ProgressBarDialog::onUpdateProgressInfo(int currentVal)
{
    ui->progressBar->setValue(currentVal);
    if (currentVal >= 100) {
        accept();
        deleteLater();
    }
}
