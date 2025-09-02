#ifndef NAVIBAR_WIDGET_H
#define NAVIBAR_WIDGET_H

#include <QWidget>
#include <QMenu>
#include <QPointer>

namespace Ui {
class NaviBarWidget;
}

class OnlineStatusWidget : public QWidget
{
    Q_OBJECT
public:
    OnlineStatusWidget(QWidget *parent = nullptr);
    ~OnlineStatusWidget();

private Q_SLOTS:
    void onDispatchMessage(const QVariant &data);
    void onUpdateClientList();
    void onPipeDisconnected();

private:
    void paintEvent(QPaintEvent *event) override;
    void enterEvent(QEvent *event) override;
    void leaveEvent(QEvent *event) override;

    QString m_imagePath;
    int m_onlineDevicesCount{ 0 };
    int m_statusCode { 1 };
    bool m_enterStatus { false };
    QPointer<QMenu> m_currentMenu { nullptr };
    int64_t m_cacheIndex { 0 };
    QString m_statusMessage;
};

class NaviBarWidget : public QWidget
{
    Q_OBJECT

public:
    explicit NaviBarWidget(QWidget *parent = nullptr);
    ~NaviBarWidget();

private Q_SLOTS:
    void processMoreIconClicked();

private:
    Ui::NaviBarWidget *ui;
    OnlineStatusWidget *m_onlineStatusWidget;
};

#endif // NAVIBAR_WIDGET_H
