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

    // --------------------- server部分的消息, 用于测试
    void connectdForTestServer();
    void sendDataForTestServer(const QByteArray &data);
    void addTestClient();

    //---------------------------------------------

    void sendDataToServer(const QByteArray &data);
    void updateProgressInfo(int currentVal);
    void updateProgressInfoWithID(int currentVal, const QByteArray &hashID);

    void recvServerData(const QByteArray &data);
    void pipeDisconnected();

    void dispatchMessage(const QVariant &data);
    // 刷新客户端列表展示
    void updateClientList();
    // true: 接受, false: 拒绝
    void userAcceptFile(bool status);

    void systemConfigChanged();
    void updateControlStatus(bool status);
    void updateFileOptInfoList();
    void updateUserSelectedInfo();

private:
    CommonSignals();
    static CommonSignals *m_instance;
};
