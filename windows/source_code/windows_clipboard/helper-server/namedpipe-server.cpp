#include "namedpipe-server.h"
#include <QProcess>

HelperServer::HelperServer(QObject *parent)
    : QObject(parent)
{
    m_server = new QLocalServer;
    connect(m_server, &QLocalServer::newConnection, this, &HelperServer::onNewConnection);
    connect(CommonSignals::getInstance(), &CommonSignals::updateProgressInfoWithMsg, this, &HelperServer::onUpdateProgressInfoWithMsg);
}

HelperServer::~HelperServer()
{

}

void HelperServer::onNewConnection()
{
    updateConnList();

    while (m_server->hasPendingConnections()) {
        QPointer<QLocalSocket> socket { m_server->nextPendingConnection() };
        QPointer<PipeConnection> conn = new PipeConnection(socket);
        qInfo() << "++++++++++++++++++new conn:" << conn->connName();
        m_connList.append(conn);

        connect(conn, &PipeConnection::notifyServerProcessData, this, [this] {
            qInfo() << "---------------HelperServer recv notifyServerProcessData";
            processMessageData();
        });

        connect(conn, &PipeConnection::notifyServerPipeDisconnected, this, [this] {
            PipeConnection *conn = qobject_cast<PipeConnection*>(sender());
            for (auto itr = m_cacheHashIdList.begin(); itr != m_cacheHashIdList.end(); ++itr) {
                if (*itr == conn->hashID()) {
                    m_cacheHashIdList.erase(itr);
                    qInfo() << "---------------remove hash id:" << *itr;
                    break;
                }
            }

            updateConnList();
        });
    }

    processMessageData();
}

void HelperServer::processMessageData()
{
    updateConnList();

    // process image-paste-progress
    {
        for (auto itr = m_cacheMessageList.begin(); itr != m_cacheMessageList.end();) {
            const QVariant &data = *itr;
            if (data.canConvert<UpdateImageProgressMsgPtr>() == false) {
                ++itr;
                continue;
            }

            UpdateImageProgressMsgPtr ptr_msg = data.value<UpdateImageProgressMsgPtr>();
            auto hashIdValue = ptr_msg->toRecordData().getHashID();

            bool exists = false;
            for (auto conn : m_connList) {
                if (conn->hashID() == hashIdValue) {
                    conn->sendData(UpdateImageProgressMsg::toByteArray(*ptr_msg));
                    qInfo() << "HelperServer send message [UpdateImageProgressMsg]: sentSize=" << ptr_msg->sentSize;
                    exists = true;
                    break;
                }
            }

            if (exists == true) {
                itr = m_cacheMessageList.erase(itr);
            } else {
                ++itr;
            }
        }
    }
}

void HelperServer::onUpdateProgressInfoWithMsg(const QVariant &msgData)
{
    m_cacheMessageList.push_back(msgData);
    do {
        if (msgData.canConvert<UpdateImageProgressMsgPtr>()) {
            UpdateImageProgressMsgPtr ptr_msg = msgData.value<UpdateImageProgressMsgPtr>();
            bool exists = false;
            QByteArray hashIdValue = ptr_msg->toRecordData().getHashID();
            for (auto itr = m_cacheHashIdList.begin(); itr != m_cacheHashIdList.end(); ++itr) {
                if (hashIdValue == *itr) {
                    exists = true;
                    break;
                }
            }
            if (exists == false) {
                m_cacheHashIdList.push_back(hashIdValue);

                for (const auto &exePath : pasteProgressExePathList()) {
                    if (QFile::exists(exePath)) {
                        QTimer::singleShot(500, Qt::TimerType::PreciseTimer, this, [hashIdValue, exePath, this] {
                            QProcess process;
                            process.startDetached(exePath, { hashIdValue.toHex().toUpper().constData(), QString::number(m_cacheHashIdList.size()) });
                        });
                        break;
                    }
                }
            }
        }
    } while (false);

    processMessageData();
}

QStringList HelperServer::pasteProgressExePathList() const
{
    QStringList pathList;
    pathList << qApp->applicationDirPath() + "/image-paste-progress.exe";
    pathList << qApp->applicationDirPath() + "/../image-paste-progress/image-paste-progress.exe";
    return pathList;
}

void HelperServer::updateConnList()
{
    for (auto itr = m_connList.begin(); itr != m_connList.end();) {
        if ((*itr) == nullptr) {
            itr = m_connList.erase(itr);
            qInfo() << "xxxxxxxxxxxxxxxxx remove one connection......";
            continue;
        }
        ++itr;
    }
}

void HelperServer::startServer(const QString &serverName)
{
    if (m_server) {
        m_server->listen(serverName);
        qInfo() << "--------------start server:" << m_server->fullServerName().toUtf8().constData();
    }
}
