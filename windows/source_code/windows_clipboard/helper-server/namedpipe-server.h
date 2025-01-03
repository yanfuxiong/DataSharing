#pragma once
#include <QLocalServer>
#include <QLocalSocket>
#include <QByteArray>
#include <QObject>
#include <QPointer>
#include "pipe-connection.h"

class HelperServer : public QObject
{
    Q_OBJECT
public:
    HelperServer(QObject *parent = nullptr);
    ~HelperServer();

    void startServer(const QString &serverName);

private Q_SLOTS:
    void onNewConnection();
    void onUpdateProgressInfoWithMsg(const QVariant &msgData);

private:
    void updateConnList();
    QStringList pasteProgressExePathList() const;
    void processMessageData();

private:
    QPointer<QLocalServer> m_server;
    QList<QPointer<PipeConnection> > m_connList;
    QList<QByteArray> m_cacheHashIdList;
    QList<QVariant> m_cacheMessageList;
};
