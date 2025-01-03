#include "pipe-connection.h"
#define PROPERY_NAME_HASH_ID "__conn_hash_id__"

std::atomic<int64_t> PipeConnection::s_connIndex { 0 };

PipeConnection::PipeConnection(QPointer<QLocalSocket> socket, QObject *parent)
    : QObject(parent)
    , m_socket(socket)
{
    {
        connect(socket, &QLocalSocket::readyRead, this, &PipeConnection::onReadyRead);
        connect(socket, &QLocalSocket::disconnected, this, &PipeConnection::onDisconnected);
    }

    m_connName = QString("%1_conn#%2").arg(g_helperServerName).arg(++s_connIndex);
}

PipeConnection::~PipeConnection()
{
    if (m_socket) {
        m_socket->deleteLater();
    }
}

void PipeConnection::onReadyRead()
{
    QByteArray data = m_socket->readAll();
    m_buffer.append(data);

    do {
        if (m_buffer.readableBytes() <= sizeof (uint32_t)) {
            break;
        }

        uint32_t messageLength = m_buffer.peekUInt32();
        if (messageLength > m_buffer.readableBytes() - sizeof (uint32_t)) {
            break;
        }

        m_buffer.retrieve(sizeof (uint32_t));
        QByteArray messageData = m_buffer.retrieveAsByteArray(messageLength);

        try {
            nlohmann::json dataJson = nlohmann::json::parse(messageData.toStdString());
            // FIXME: This needs improvement and can be distinguished based on message IDs
            QByteArray hashIdValue = QByteArray::fromHex(dataJson.at("hash_id").get<std::string>().c_str());
            setHashID(hashIdValue);
            qInfo() << dataJson.dump(4).c_str();
        } catch (const std::exception &e) {
            qWarning() << e.what();
        }
    } while (true);

    Q_EMIT notifyServerProcessData();
}

void PipeConnection::onDisconnected()
{
    qInfo() << "----------------disconnect:" << connName();
    Q_EMIT notifyServerPipeDisconnected();
    deleteLater();
}

QString PipeConnection::connName() const
{
    return m_connName;
}

void PipeConnection::setHashID(const QByteArray &hashIdVal)
{
    setProperty(PROPERY_NAME_HASH_ID, hashIdVal);
}

QByteArray PipeConnection::hashID() const
{
    return property(PROPERY_NAME_HASH_ID).toByteArray();
}

void PipeConnection::sendData(const QByteArray &data)
{
    m_socket->write(data);
    m_socket->flush();
}
