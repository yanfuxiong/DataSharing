#include "namedpipe_server.h"
#include <QProcess>

HelperServer::HelperServer(QObject *parent)
    : QObject(parent)
{
    m_server = new QLocalServer;
    connect(m_server, &QLocalServer::newConnection, this, &HelperServer::onNewConnection);
    connect(CommonSignals::getInstance(), &CommonSignals::broadcastData, this, &HelperServer::onBroadcastData);
}

HelperServer::~HelperServer()
{

}

void HelperServer::onNewConnection()
{
    while (m_server->hasPendingConnections()) {
        QPointer<QLocalSocket> socket { m_server->nextPendingConnection() };
        QPointer<PipeConnection> conn = new PipeConnection(socket);
        qInfo() << "--------------new conn:" << conn->connName();
        m_connList.append(conn);

        connect(conn, &PipeConnection::notifyServerPipeDisconnected, this, [this] (const QString &connName) {
            removeConnection(connName);
        });

        Q_EMIT CommonSignals::getInstance()->pipeConnected();
    }
}

void HelperServer::onBroadcastData(const QByteArray &data)
{
    for (const auto &conn : m_connList) {
        conn->sendData(data);
    }
}

QStringList HelperServer::pasteProgressExePathList() const
{
    QStringList pathList;
    pathList << qApp->applicationDirPath() + "/image-paste-progress.exe";
    pathList << qApp->applicationDirPath() + "/../image-paste-progress/image-paste-progress.exe";
    return pathList;
}

void HelperServer::removeConnection(const QString &connName)
{
    for (auto itr = m_connList.begin(); itr != m_connList.end(); ++itr) {
        if (*itr && (*itr)->connName() == connName) {
            itr = m_connList.erase(itr);
            qInfo() << "------------remove one connection:" << connName;
            break;
        }
    }
}

void HelperServer::startServer(const QString &serverName)
{
    if (m_server) {
        m_server->listen(serverName);
        qInfo() << "--------------start server:" << m_server->fullServerName().toUtf8().constData();
    }
}
