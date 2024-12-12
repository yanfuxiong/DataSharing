#pragma once
#include "common_signals.h"
#include "common_utils.h"
#include <QLocalSocket>
#include <QPointer>

class NamedPipeClient : public QObject
{
    Q_OBJECT
public:
    NamedPipeClient(QObject *parent = nullptr);
    ~NamedPipeClient();

    void connectToServer();
    void sendData(const QByteArray &data);
    bool connectdStatus() const;

private Q_SLOTS:
    void onConnected();
    void onDisconnected();
    void onReadyRead();

    void onError(QLocalSocket::LocalSocketError socketError);
    void onStateChanged(QLocalSocket::LocalSocketState socketState);

private:
    QPointer<QLocalSocket> m_client;
};
