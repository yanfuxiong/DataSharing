#include "namedpipe_client.h"

NamedPipeClient::NamedPipeClient(QObject *parent)
    : QObject(parent)
{
    m_client = new QLocalSocket;
    m_client->setServerName("CrossSharePipe");

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
    //Q_EMIT CommonSignals::getInstance()->showInfoMessageBox("提示", "成功连接服务器......");
    qInfo() << "--------------------成功连接服务器";
    Q_EMIT CommonSignals::getInstance()->logMessage("------------successfully connected to the server");
}

void NamedPipeClient::onDisconnected()
{
    g_getGlobalData()->namedPipeConnected = false;
    //Q_EMIT CommonSignals::getInstance()->showWarningMessageBox("警告", "与服务器断开连接......");
    Q_EMIT CommonSignals::getInstance()->logMessage("------------disconnected from the server");
    // 清空设备列表信息
    g_getGlobalData()->m_clientVec.clear();

    Q_EMIT CommonSignals::getInstance()->pipeDisconnected();
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
        m_client->write(data);
        m_client->flush();
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
