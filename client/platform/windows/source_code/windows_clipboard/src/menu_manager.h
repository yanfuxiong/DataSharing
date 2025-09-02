#pragma once
#include <QDialog>
#include <QWidgetAction>
#include <QMenu>
#include <QLabel>
#include <QPushButton>
#include <QHBoxLayout>
#include <QMouseEvent>

extern QPoint g_menuPos;

class SystemInfoMenuManager : public QObject
{
    Q_OBJECT
public:
    SystemInfoMenuManager(QObject *parent = nullptr);
    ~SystemInfoMenuManager();

    static QPixmap createPixmap(const QString &imgPath, int size);
    static void updateMenuPos(QMenu *menu);

    QMenu *createMainMenu();
    QMenu *createOpenSourceLicenseMenu();
    QMenu *createClientInfoListMenu();
    QMenu *createTextInfoMenu(const QString &textInfo);
};

class LicenseItem : public QWidget
{
    Q_OBJECT
public:
    LicenseItem(const QString &name, QWidget *parent = nullptr);
    ~LicenseItem();

    const QString &itemName() const { return m_name; }

Q_SIGNALS:
    void actionTriggered();

protected:
    void enterEvent(QEvent *event) override;
    void leaveEvent(QEvent *event) override;
    void mousePressEvent(QMouseEvent *event) override;

private:
    QString m_name;
};

class LicenseItemAction : public QWidgetAction
{
    Q_OBJECT
public:
    LicenseItemAction(const QString &name, QObject *parent = nullptr);
    ~LicenseItemAction();
    const QString &itemName() const { return m_name; }
    QWidget *createWidget(QWidget *parent) override;

private:
    QString m_name;
};


class LicenseTitleItem : public QWidget
{
    Q_OBJECT
public:
    LicenseTitleItem(QWidget *parent = nullptr);
    ~LicenseTitleItem();
};

class LicenseTitleItemAction : public QWidgetAction
{
    Q_OBJECT
public:
    LicenseTitleItemAction(QObject *parent = nullptr);
    ~LicenseTitleItemAction();
    QWidget *createWidget(QWidget *parent) override;
};

class HLineItem : public QWidget
{
    Q_OBJECT
public:
    HLineItem(QWidget *parent = nullptr);
    ~HLineItem();
};

class HLineItemAction : public QWidgetAction
{
    Q_OBJECT
public:
    HLineItemAction(QObject *parent = nullptr);
    ~HLineItemAction();
    QWidget *createWidget(QWidget *parent) override;
};

class DownloadPathItem : public QWidget
{
    Q_OBJECT
public:
    DownloadPathItem(QWidget *parent = nullptr);
    ~DownloadPathItem();

Q_SIGNALS:
    void actionTriggered();
};

class DownloadPathItemAction : public QWidgetAction
{
    Q_OBJECT
public:
    DownloadPathItemAction(QObject *parent = nullptr);
    ~DownloadPathItemAction();
    QWidget *createWidget(QWidget *parent) override;
};

class BugReportItem : public QWidget
{
    Q_OBJECT
public:
    BugReportItem(QWidget *parent = nullptr);
    ~BugReportItem();

Q_SIGNALS:
    void actionTriggered();
};


class BugReportItemAction : public QWidgetAction
{
    Q_OBJECT
public:
    BugReportItemAction(QObject *parent = nullptr);
    ~BugReportItemAction();
    QWidget *createWidget(QWidget *parent) override;
};

class ClientInfoListItem : public QWidget
{
    Q_OBJECT
public:
    ClientInfoListItem(const QString &iconPath, const QString &info, QWidget *parent = nullptr);
    ~ClientInfoListItem();

Q_SIGNALS:
    void actionTriggered();
};

class ClientInfoListItemAction : public QWidgetAction
{
    Q_OBJECT
public:
    ClientInfoListItemAction(QObject *parent = nullptr);
    ~ClientInfoListItemAction();
    void setIconPath(const QString &path) { m_iconPath = path; }
    void setInfo(const QString &info) { m_info = info; }

    QWidget *createWidget(QWidget *parent) override;

private:
    QString m_iconPath;
    QString m_info;
};

class TextInfoItem : public QWidget
{
    Q_OBJECT
public:
    TextInfoItem(const QString &info, QWidget *parent = nullptr);
    ~TextInfoItem();

Q_SIGNALS:
    void actionTriggered();
};

class TextInfoItemAction : public QWidgetAction
{
    Q_OBJECT
public:
    TextInfoItemAction(QObject *parent = nullptr);
    ~TextInfoItemAction();
    void setInfo(const QString &info) { m_info = info; }

    QWidget *createWidget(QWidget *parent) override;

private:
    QString m_info;
};
