#pragma once
#include <QObject>
#include <QPointer>
#include <QLocalSocket>
#include "common_signals.h"
#include "common_utils.h"

class PipeConnection : public QObject
{
    Q_OBJECT
public:
    PipeConnection(QPointer<QLocalSocket> socket, QObject *parent = nullptr);
    ~PipeConnection();

    QString connName() const;

    void setHashID(const QByteArray &hashIdVal);
    QByteArray hashID() const;

    void sendData(const QByteArray &data);

Q_SIGNALS:
    void notifyServerProcessData();
    void notifyServerPipeDisconnected();

private Q_SLOTS:
    void onReadyRead();
    void onDisconnected();

private:
    QPointer<QLocalSocket> m_socket;
    QString m_connName;
    Buffer m_buffer;

    static std::atomic<int64_t> s_connIndex;
};
