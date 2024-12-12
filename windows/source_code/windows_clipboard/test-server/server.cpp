#include "server.h"
#include "common_utils.h"
#include "common_signals.h"

NamedPipeServer::NamedPipeServer(QObject *parent)
    : QObject(parent)
{
    m_server = new QLocalServer;
    connect(m_server, &QLocalServer::newConnection, this, &NamedPipeServer::onNewConnection);
    connect(CommonSignals::getInstance(), &CommonSignals::addTestClient, this, &NamedPipeServer::onAddTestClient);
}

NamedPipeServer::~NamedPipeServer()
{

}

void NamedPipeServer::onNewConnection()
{
    while (m_server->hasPendingConnections()) {
        auto socket = m_server->nextPendingConnection();
        connect(socket, &QLocalSocket::readyRead, this, &NamedPipeServer::onReadyRead);
        connect(socket, &QLocalSocket::disconnected, this, &NamedPipeServer::onDisconnected);

        connect(CommonSignals::getInstance(), &CommonSignals::sendDataForTestServer, this, &NamedPipeServer::onSendDataForTestServer);

        m_clientList.append(socket);
        // 这里只用于测试, 不携带参数
        Q_EMIT CommonSignals::getInstance()->connectdForTestServer();

        QTimer::singleShot(50, this, [] {
            QByteArray clientStatusMsgData;
            {
                UpdateClientStatusMsg msg;
                msg.status = 1;
                msg.ip = "192.168.30.1";
                msg.port = 12345;
                msg.clientID = "QmQ7obXFx1XMFr6hCYXtovn9zREFqSXEtH5hdtpBDLjrAz";
                msg.clientName = QString("测试电脑_%1").arg(1);

                clientStatusMsgData = UpdateClientStatusMsg::toByteArray(msg);
            }

            Q_EMIT CommonSignals::getInstance()->sendDataForTestServer(clientStatusMsgData);
        });
    }
}

void NamedPipeServer::onReadyRead()
{
    QPointer<QLocalSocket> socket = qobject_cast<QLocalSocket*>(sender());
    Q_ASSERT(socket != nullptr);
    QByteArray data = socket->readAll();
    //qInfo() << "[RECV]:" << data.constData();
    //socket->write(data);
    Q_EMIT recvData(data);
}

void NamedPipeServer::onDisconnected()
{
    QPointer<QLocalSocket> socket = qobject_cast<QLocalSocket*>(sender());
    Q_ASSERT(socket != nullptr);
    if (m_clientList.removeOne(socket)) {
        qInfo() << "remove socket: " << socket->fullServerName().toUtf8().constData();
        socket->deleteLater();

        Q_EMIT CommonSignals::getInstance()->pipeDisconnected();
    }
}

void NamedPipeServer::startServer(const QString &serverName)
{
    if (m_server) {
        m_server->listen(serverName);
        qInfo() << "--------------start server:" << m_server->fullServerName().toUtf8().constData();
    }
}

void NamedPipeServer::onSendDataForTestServer(const QByteArray &data)
{
    for (const auto &socket : m_clientList) {
        socket->write(data);
        socket->flush();
    }
}

void NamedPipeServer::onAddTestClient()
{
    static int s_index = 2;
    for (auto client : m_clientList) {
        QByteArray clientStatusMsgData;
        {
            UpdateClientStatusMsg msg;
            msg.status = 1;
            msg.ip = "192.168.30.1";
            msg.port = 12345;
            //msg.clientID = "QmQ7obXFx1XMFr6hCYXtovn9zREFqSXEtH5hdtpBDLjrAz";
            msg.clientID = QByteArray(46, char(s_index));
            msg.clientName = QString("测试电脑_%1").arg(s_index);

            clientStatusMsgData = UpdateClientStatusMsg::toByteArray(msg);
        }

        Q_EMIT CommonSignals::getInstance()->sendDataForTestServer(clientStatusMsgData);

        ++s_index;
    }
}
