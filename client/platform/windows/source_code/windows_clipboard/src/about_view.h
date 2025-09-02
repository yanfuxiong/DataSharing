#ifndef ABOUT_VIEW_H
#define ABOUT_VIEW_H

#include <QWidget>

namespace Ui {
class AboutView;
}

class AboutView : public QWidget
{
    Q_OBJECT

public:
    explicit AboutView(QWidget *parent = nullptr);
    ~AboutView();

    QString getVersionInfoKey() const;
    QString getVersionInfo() const;
    QString getReadmeInfo() const;

private:
    Ui::AboutView *ui;
};

#endif // ABOUT_VIEW_H
