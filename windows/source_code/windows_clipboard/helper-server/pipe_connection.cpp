#include "pipe_connection.h"
#include "load_plugin.h"

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

    uint8_t typeValue = 0;
    uint8_t code = 0;
    // parse message
    while (g_getCodeFromByteArray(QByteArray(m_buffer.peek(), m_buffer.readableBytes()), typeValue, code)) {
        switch (code) {
        case SendFile_code: {
            if (typeValue == PipeMessageType::Request) {
                SendFileRequestMsg message;
                if (SendFileRequestMsg::fromByteArray(QByteArray(m_buffer.peek(), m_buffer.readableBytes()), message)) {
                    m_buffer.retrieve(message.getMessageLength());
                    Q_ASSERT(message.filePathVec.empty() == false);

                    QByteArray ipString = message.ip.toUtf8() + ":" + QByteArray::number(message.port);
                    QByteArray clientID = message.clientID;
                    uint64_t timeStamp = message.timeStamp;

                    std::vector<std::wstring> filePathVec;
                    for (const auto &filePath : message.filePathVec) {
                        filePathVec.push_back(filePath.toStdWString());
                    }
                    const wchar_t **filePathArry = new const wchar_t*[message.filePathVec.size()];
                    for (uint32_t index = 0; index < message.filePathVec.size(); ++index) {
                        filePathArry[index] = filePathVec[index].c_str();
                    }
                    LoadPlugin::getInstance()->multiFilesDropRequest(ipString.data(),
                                            clientID.data(),
                                            timeStamp,
                                            filePathArry,
                                            message.filePathVec.size());
                    delete []filePathArry;
                }
            }
            break;
        }
        case DragFilesMsg_code: {
            Q_ASSERT(typeValue == PipeMessageType::Notify);
            DragFilesMsg message;
            if (DragFilesMsg::fromByteArray(QByteArray(m_buffer.peek(), m_buffer.readableBytes()), message)) {
                m_buffer.retrieve(message.getMessageLength());

                if (message.functionCode == DragFilesMsg::FuncCode::CancelFileTransfer) {
                    QByteArray ipString = message.ip.toUtf8() + ":" + QByteArray::number(message.port);
                    QByteArray clientID = message.clientID;
                    LoadPlugin::getInstance()->cancelFileTransfer(ipString.data(), clientID.data(), message.timeStamp);
                    break;
                }

                Q_ASSERT(message.filePathVec.empty() == false);

                std::vector<std::wstring> filePathVec;
                for (const auto &filePath : message.filePathVec) {
                    filePathVec.push_back(filePath.toStdWString());
                }
                const wchar_t **filePathArry = new const wchar_t*[message.filePathVec.size()];
                for (uint32_t index = 0; index < message.filePathVec.size(); ++index) {
                    filePathArry[index] = filePathVec[index].c_str();
                }
                LoadPlugin::getInstance()->dragFileListRequest(filePathArry, message.filePathVec.size(), message.timeStamp);
                delete []filePathArry;
            }
            break;
        }
        default: {
            qInfo() << "--------------typeValue:" << typeValue << "; code:" << code;
            uint32_t messageLength = m_buffer.peekUInt32();
            m_buffer.retrieveUInt32();
            m_buffer.retrieve(messageLength);
            break;
        }
        }
    }
}

void PipeConnection::onDisconnected()
{
    Q_EMIT notifyServerPipeDisconnected(m_connName);
    deleteLater();
}

QString PipeConnection::connName() const
{
    return m_connName;
}

void PipeConnection::sendData(const QByteArray &data)
{
    if (m_socket) {
        m_socket->write(data);
        m_socket->flush();
    }
}
