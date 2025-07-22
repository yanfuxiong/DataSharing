#ifndef TIPS_LIST_AREA_H
#define TIPS_LIST_AREA_H

#include <QScrollArea>

namespace Ui {
class TipsListArea;
}

class TipsListArea : public QScrollArea
{
    Q_OBJECT

public:
    explicit TipsListArea(QWidget *parent = nullptr);
    ~TipsListArea();

private Q_SLOTS:
    void onUpdateClientList();

private:
    Ui::TipsListArea *ui;
};

#endif // TIPS_LIST_AREA_H
