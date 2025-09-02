#include "namedpipe_client.h"

NamedPipeClient::NamedPipeClient(QObject *parent)
    : QObject(parent)
{
    m_client = new QLocalSocket;
    m_client->setServerName(g_namedPipeServerName);

    connect(m_client, &QLocalSocket::connected, this, &NamedPipeClient::onConnected);
    connect(m_client, &QLocalSocket::disconnected, this, &NamedPipeClient::onDisconnected);
    connect(m_client, &QLocalSocket::readyRead, this, &NamedPipeClient::onReadyRead);
    connect(m_client, QOverload<QLocalSocket::LocalSocketError>::of(&QLocalSocket::error), this, &NamedPipeClient::onError);
    connect(m_client, &QLocalSocket::stateChanged, this, &NamedPipeClient::onStateChanged);

//    QTimer::singleShot(0, this, [this] {
//        m_client->connectToServer();
//    });
}

NamedPipeClient::~NamedPipeClient()
{

}

void NamedPipeClient::connectToServer()
{
    m_client->connectToServer();
}

void NamedPipeClient::onConnected()
{
    g_getGlobalData()->namedPipeConnected = true;
    qInfo() << "--------------------successfully connected to the server";
    Q_EMIT CommonSignals::getInstance()->logMessage("------------successfully connected to the server");
    Q_EMIT CommonSignals::getInstance()->pipeConnected();
}

void NamedPipeClient::onDisconnected()
{
    g_getGlobalData()->namedPipeConnected = false;
    Q_EMIT CommonSignals::getInstance()->logMessage("------------disconnected from the server");
    g_getGlobalData()->m_clientVec.clear();

    Q_EMIT CommonSignals::getInstance()->pipeDisconnected();
    Q_EMIT CommonSignals::getInstance()->updateClientList();
}

void NamedPipeClient::onReadyRead()
{
    QByteArray data = m_client->readAll();
    Q_EMIT CommonSignals::getInstance()->recvServerData(data);
}

bool NamedPipeClient::connectdStatus() const
{
    if (m_client == nullptr) {
        return false;
    }
    return m_client->state() == QLocalSocket::LocalSocketState::ConnectedState;
}

void NamedPipeClient::sendData(const QByteArray &data)
{
    if (connectdStatus()) {
        m_outputBuffer.append(data);
        if (m_runningStatus.load() == false) {
            m_runningStatus.store(true);
            QTimer::singleShot(0, this, [this] {
                processBufferData();
            });
        }
    }
}

void NamedPipeClient::onError(QLocalSocket::LocalSocketError socketError)
{
    Q_UNUSED(socketError)
    //qWarning() << socketError;
}

void NamedPipeClient::onStateChanged(QLocalSocket::LocalSocketState socketState)
{
    Q_UNUSED(socketState)
    //qWarning() << socketState;
}

void NamedPipeClient::processBufferData()
{
    static const int s_blockSize = 1024;
    if (m_outputBuffer.readableBytes() <= s_blockSize) {
        m_client->write(m_outputBuffer.retrieveAllAsByteArray());
    } else {
        m_client->write(QByteArray(m_outputBuffer.peek(), s_blockSize));
        m_outputBuffer.retrieve(s_blockSize);
    }
    m_client->flush();
    if (m_outputBuffer.readableBytes() > 0) {
        QTimer::singleShot(10, this, [this] {
            processBufferData();
        });
    } else {
        m_runningStatus.store(false);
    }
}
