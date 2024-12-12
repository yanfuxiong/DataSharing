#pragma once
#include "common_signals.h"
#include "common_utils.h"

class CommonUiProcess : public QObject
{
    Q_OBJECT
public:
    ~CommonUiProcess();
    static CommonUiProcess *getInstance();

private:
    void initConnect();

private Q_SLOTS:
    void onShowInfoMessageBox(const QString &title, const QString &message);
    void onShowWarningMessageBox(const QString &title, const QString &message);

private:
    CommonUiProcess();
    static CommonUiProcess *m_instance;
};
