#include "process_message.h"

ProcessMessage *ProcessMessage::m_instance = nullptr;


ProcessMessage::ProcessMessage(QObject *parent)
    : QObject(parent)
{
    {
        connect(CommonSignals::getInstance(), &CommonSignals::recvServerData, this, &ProcessMessage::onRecvServerData);
        connect(CommonSignals::getInstance(), &CommonSignals::sendDataToServer, this, &ProcessMessage::onSendDataToServer);
    }

    {
        m_timer = new QTimer(this);
        m_timer->setTimerType(Qt::TimerType::PreciseTimer);
        m_timer->setInterval(1000);
        auto func = [this] {
            if (g_getGlobalData()->namedPipeConnected == false) {
                if (m_pipeClient) {
                    m_pipeClient->deleteLater();
                }
                m_pipeClient = new NamedPipeClient;
                m_pipeClient->connectToServer();
            }
        };
        QObject::connect(m_timer, &QTimer::timeout, this, func);
        m_timer->start();
        QTimer::singleShot(0, this, func);
    }
}

ProcessMessage::~ProcessMessage()
{

}

ProcessMessage *ProcessMessage::getInstance()
{
    if (m_instance == nullptr) {
        m_instance = new ProcessMessage;
    }
    return m_instance;
}

void ProcessMessage::onSendDataToServer(const QByteArray &data)
{
    qInfo() << data.toHex().toUpper().constData();
    m_pipeClient->sendData(data);
}

void ProcessMessage::onRecvServerData(const QByteArray &data)
{
    m_buffer.append(data);

    uint8_t typeValue = 0;
    uint8_t code = 0;
    // parse message
    while (g_getCodeFromByteArray(QByteArray(m_buffer.peek(), m_buffer.readableBytes()), typeValue, code)) {
        switch (code) {
        case 1: {
            GetConnStatusResponseMsg message;
            if (GetConnStatusResponseMsg::fromByteArray(QByteArray(m_buffer.peek(), m_buffer.readableBytes()), message)) {
                m_buffer.retrieve(message.getMessageLength());
                QVariant sendData = QVariant::fromValue<GetConnStatusResponseMsgPtr>(std::make_shared<GetConnStatusResponseMsg>(message));
                Q_ASSERT(sendData.canConvert<GetConnStatusResponseMsgPtr>() == true);
                Q_EMIT CommonSignals::getInstance()->dispatchMessage(sendData);
            }
            break;
        }
        case 2: {
            break;
        }
        case 3: {
            UpdateClientStatusMsg message;
            if (UpdateClientStatusMsg::fromByteArray(QByteArray(m_buffer.peek(), m_buffer.readableBytes()), message)) {
                m_buffer.retrieve(message.getMessageLength());
                QVariant sendData = QVariant::fromValue<UpdateClientStatusMsgPtr>(std::make_shared<UpdateClientStatusMsg>(message));
                Q_ASSERT(sendData.canConvert<UpdateClientStatusMsgPtr>() == true);
                Q_EMIT CommonSignals::getInstance()->dispatchMessage(sendData);
            }
            break;
        }
        case 4: {
            if (typeValue == PipeMessageType::Request) {
                SendFileRequestMsg message;
                if (SendFileRequestMsg::fromByteArray(QByteArray(m_buffer.peek(), m_buffer.readableBytes()), message)) {
                    m_buffer.retrieve(message.getMessageLength());
                    QVariant sendData = QVariant::fromValue<SendFileRequestMsgPtr>(std::make_shared<SendFileRequestMsg>(message));
                    Q_ASSERT(sendData.canConvert<SendFileRequestMsgPtr>() == true);
                    Q_EMIT CommonSignals::getInstance()->dispatchMessage(sendData);
                }
            } else {
                Q_ASSERT(false);
            }
            break;
        }
        case 5: {
            UpdateProgressMsg message;
            if (UpdateProgressMsg::fromByteArray(QByteArray(m_buffer.peek(), m_buffer.readableBytes()), message)) {
                m_buffer.retrieve(message.getMessageLength());
                QVariant sendData = QVariant::fromValue<UpdateProgressMsgPtr>(std::make_shared<UpdateProgressMsg>(message));
                Q_ASSERT(sendData.canConvert<UpdateProgressMsgPtr>() == true);
                Q_EMIT CommonSignals::getInstance()->dispatchMessage(sendData);
            }
            break;
        }
        case 6: {
            UpdateImageProgressMsg message;
            if (UpdateImageProgressMsg::fromByteArray(QByteArray(m_buffer.peek(), m_buffer.readableBytes()), message)) {
                m_buffer.retrieve(message.getMessageLength());
                QVariant sendData = QVariant::fromValue<UpdateImageProgressMsgPtr>(std::make_shared<UpdateImageProgressMsg>(message));
                Q_ASSERT(sendData.canConvert<UpdateImageProgressMsgPtr>() == true);
                Q_EMIT CommonSignals::getInstance()->dispatchMessage(sendData);
            }
            break;
        }
        case 7: {
            NotifyMessage message;
            if (NotifyMessage::fromByteArray(QByteArray(m_buffer.peek(), m_buffer.readableBytes()), message)) {
                m_buffer.retrieve(message.getMessageLength());
                QVariant sendData = QVariant::fromValue<NotifyMessagePtr>(std::make_shared<NotifyMessage>(message));
                Q_ASSERT(sendData.canConvert<NotifyMessagePtr>() == true);
                Q_EMIT CommonSignals::getInstance()->dispatchMessage(sendData);
            }
            break;
        }
        case 8: {
            UpdateSystemInfoMsg message;
            if (UpdateSystemInfoMsg::fromByteArray(QByteArray(m_buffer.peek(), m_buffer.readableBytes()), message)) {
                m_buffer.retrieve(message.getMessageLength());
                QVariant sendData = QVariant::fromValue<UpdateSystemInfoMsgPtr>(std::make_shared<UpdateSystemInfoMsg>(message));
                Q_ASSERT(sendData.canConvert<UpdateSystemInfoMsgPtr>() == true);
                Q_EMIT CommonSignals::getInstance()->dispatchMessage(sendData);
            }
            break;
        }
        default: {
            Q_ASSERT(false);
            break;
        }
        }
    }
}
