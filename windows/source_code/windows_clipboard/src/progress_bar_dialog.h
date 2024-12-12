#ifndef PROGRESS_BAR_DIALOG_H
#define PROGRESS_BAR_DIALOG_H

#include <QDialog>

namespace Ui {
class ProgressBarDialog;
}

class ProgressBarDialog : public QDialog
{
    Q_OBJECT

public:
    explicit ProgressBarDialog(QWidget *parent = nullptr);
    ~ProgressBarDialog();

private Q_SLOTS:
    void onUpdateProgressInfo(int currentVal);

private:
    Ui::ProgressBarDialog *ui;
};

#endif // PROGRESS_BAR_DIALOG_H
