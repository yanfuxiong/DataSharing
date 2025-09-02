#include "progress_bar_widget.h"
#include "ui_progress_bar_widget.h"
#include "common_signals.h"
#include "common_utils.h"
#include "event_filter_process.h"

ProgressBarWidget::ProgressBarWidget(QWidget *parent) :
    QWidget(parent),
    ui(new Ui::ProgressBarWidget)
{
    ui->setupUi(this);
    {
        ui->title_icon_2->clear();
        ui->close_label->clear();
    }
    ui->progressBar->setRange(0, 100);

    ui->progressBar->setValue(0);
    connect(CommonSignals::getInstance(), &CommonSignals::updateProgressInfoWithID, this, &ProgressBarWidget::onUpdateProgressInfoWithID);

    EventFilterProcess::getInstance()->registerFilterEvent({ ui->close_label, [] { qApp->quit(); } });
}

ProgressBarWidget::~ProgressBarWidget()
{
    delete ui;
}

int ProgressBarWidget::currentValue() const
{
    return ui->progressBar->value();
}

void ProgressBarWidget::onUpdateProgressInfoWithID(int currentVal, const QByteArray &hashID)
{
    Q_UNUSED(hashID)
    ui->progressBar->setValue(currentVal);
    if (currentVal >= 100) {
        qApp->quit();
    }
}
