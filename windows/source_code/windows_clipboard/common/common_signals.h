#pragma once
#include <QObject>
#include <QString>
#include <QByteArray>
#include "global_def.h"

class CommonSignals : public QObject
{
    Q_OBJECT
    Q_DISABLE_COPY(CommonSignals)
public:
    ~CommonSignals();
    static CommonSignals *getInstance();

Q_SIGNALS:
    void showInfoMessageBox(const QString &title, const QString &message);
    void showWarningMessageBox(const QString &title, const QString &message);
    void showStatusMessage(const QString &message);
    void logMessage(const QString &message);

    // --------------------- server
    void connectdForTestServer();
    void sendDataForTestServer(const QByteArray &data);
    void addTestClient();

    //---------------------------------------------

    void sendDataToServer(const QByteArray &data);
    void updateProgressInfo(int currentVal);
    void updateProgressInfoWithID(int currentVal, const QByteArray &hashID);
    void updateProgressInfoWithMsg(const QVariant &msgData);

    void recvServerData(const QByteArray &data);
    void pipeConnected();
    void pipeDisconnected();

    void dispatchMessage(const QVariant &data);
    void updateClientList();
    // true: accept, false: reject
    void userAcceptFile(bool status);

    void systemConfigChanged();
    void updateControlStatus(bool status);
    void updateFileOptInfoList();
    void updateUserSelectedInfo();

private:
    CommonSignals();
    static CommonSignals *m_instance;
};
