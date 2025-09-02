#ifndef PROGRESS_BAR_WIDGET_H
#define PROGRESS_BAR_WIDGET_H

#include <QWidget>

namespace Ui {
class ProgressBarWidget;
}

class ProgressBarWidget : public QWidget
{
    Q_OBJECT

public:
    explicit ProgressBarWidget(QWidget *parent = nullptr);
    ~ProgressBarWidget();

    int currentValue() const;

private Q_SLOTS:
    void onUpdateProgressInfoWithID(int currentVal, const QByteArray &hashID);

private:
    Ui::ProgressBarWidget *ui;
};

#endif // PROGRESS_BAR_WIDGET_H
