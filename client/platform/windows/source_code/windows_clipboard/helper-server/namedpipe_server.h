#pragma once
#include <QLocalServer>
#include <QLocalSocket>
#include <QByteArray>
#include <QObject>
#include <QPointer>
#include "pipe_connection.h"

class HelperServer : public QObject
{
    Q_OBJECT
public:
    HelperServer(QObject *parent = nullptr);
    ~HelperServer();

    void startServer(const QString &serverName);

private Q_SLOTS:
    void onNewConnection();
    void onBroadcastData(const QByteArray &data);

private:
    void removeConnection(const QString &connName);
    QStringList pasteProgressExePathList() const;

private:
    QPointer<QLocalServer> m_server;
    QList<QPointer<PipeConnection> > m_connList;
};
