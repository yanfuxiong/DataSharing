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
        // This is only for testing purposes and does not carry any parameters
        Q_EMIT CommonSignals::getInstance()->connectdForTestServer();

        QTimer::singleShot(50, this, [] {
            {
                UpdateSystemInfoMsg msg;
                msg.ip = "192.168.0.123";
                msg.port = 12345;
                msg.serverVersion = R"(server_v1.0.1)";

                QByteArray data = UpdateSystemInfoMsg::toByteArray(msg);
                Q_EMIT CommonSignals::getInstance()->sendDataForTestServer(data);
            }

            {
                QByteArray clientStatusMsgData;
                UpdateClientStatusMsg msg;
                msg.status = 1;
                msg.ip = "192.168.30.1";
                msg.port = 12345;
                msg.clientID = "QmQ7obXFx1XMFr6hCYXtovn9zREFqSXEtH5hdtpBDLjrAz";
                msg.clientName = QString("HDMI1_%1").arg(1);

                clientStatusMsgData = UpdateClientStatusMsg::toByteArray(msg);
                Q_EMIT CommonSignals::getInstance()->sendDataForTestServer(clientStatusMsgData);
            }
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

    for (auto itr = m_clientList.begin(); itr != m_clientList.end(); ++itr) {
        if (*itr == socket) {
            qInfo() << "remove socket: " << socket->fullServerName().toUtf8().constData();
            m_clientList.erase(itr);
            socket->deleteLater();
            Q_EMIT CommonSignals::getInstance()->pipeDisconnected();
            break;
        }
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
    static const char *s_nameArry[] = {"HDMI1-\nTEST", "HDMI2", "Miracast", "USBC"};
    Q_ASSERT(sizeof(s_nameArry) / sizeof (s_nameArry[0]) == 4);
    for (auto client : m_clientList) {
        QByteArray clientStatusMsgData;
        {
            UpdateClientStatusMsg msg;
            msg.status = 1;
            msg.ip = "192.168.30.1";
            msg.port = 12345;
            //msg.clientID = "QmQ7obXFx1XMFr6hCYXtovn9zREFqSXEtH5hdtpBDLjrAz";
            msg.clientID = QByteArray(46, char(s_index));
            msg.clientName = QString("%1_%2").arg(s_nameArry[qrand() % 4]).arg(s_index);

            clientStatusMsgData = UpdateClientStatusMsg::toByteArray(msg);
        }

        Q_EMIT CommonSignals::getInstance()->sendDataForTestServer(clientStatusMsgData);

        ++s_index;
    }
}
