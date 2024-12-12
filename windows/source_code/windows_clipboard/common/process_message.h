#pragma once
#include "common_utils.h"
#include "common_signals.h"
#include "namedpipe_client.h"

class ProcessMessage : public QObject
{
    Q_OBJECT
public:
    ~ProcessMessage();
    static ProcessMessage *getInstance();

private Q_SLOTS:
    void onRecvServerData(const QByteArray &data);
    void onSendDataToServer(const QByteArray &data);

private:
    ProcessMessage(QObject *parent = nullptr);
    static ProcessMessage *m_instance;
    Buffer m_buffer;
    QPointer<QTimer> m_timer;
    QPointer<NamedPipeClient> m_pipeClient;
};
