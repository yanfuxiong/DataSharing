#pragma once
#include <QLocalServer>
#include <QLocalSocket>
#include <QByteArray>
#include <QObject>
#include <QPointer>

class NamedPipeServer : public QObject
{
    Q_OBJECT
public:
    NamedPipeServer(QObject *parent = nullptr);
    ~NamedPipeServer();

    void startServer(const QString &serverName);

Q_SIGNALS:
    void recvData(const QByteArray &data);

private Q_SLOTS:
    void onNewConnection();

    void onReadyRead();
    void onDisconnected();
    void onSendDataForTestServer(const QByteArray &data);
    void onAddTestClient();

private:
    QPointer<QLocalServer> m_server;
    QList<QPointer<QLocalSocket>> m_clientList;
};
