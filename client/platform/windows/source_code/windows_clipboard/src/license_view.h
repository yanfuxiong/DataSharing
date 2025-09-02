#ifndef LICENSE_VIEW_H
#define LICENSE_VIEW_H

#include <QWidget>

namespace Ui {
class LicenseView;
}

class LicenseView : public QWidget
{
    Q_OBJECT

public:
    explicit LicenseView(QWidget *parent = nullptr);
    ~LicenseView();

    void setTitle(const QString &title);
    void setDisplayInfo(const QString &info);

private:
    void clickedLeftArrow();

private:
    Ui::LicenseView *ui;
};

#endif // LICENSE_VIEW_H
