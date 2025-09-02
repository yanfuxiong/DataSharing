#include "navibar_widget.h"
#include "ui_navibar_widget.h"
#include "common_signals.h"
#include "common_utils.h"
#include "menu_manager.h"
#include "event_filter_process.h"
#include "common_ui_process.h"
#include <QPainter>


OnlineStatusWidget::OnlineStatusWidget(QWidget *parent)
    : QWidget(parent)
{
    m_imagePath = ":/resource/conn_status/connStatus1.png";
    m_onlineDevicesCount = 0;

    {
        connect(CommonSignals::getInstance(), &CommonSignals::pipeDisconnected, this, &OnlineStatusWidget::onPipeDisconnected);
        connect(CommonSignals::getInstance(), &CommonSignals::updateClientList, this, &OnlineStatusWidget::onUpdateClientList);
        connect(CommonSignals::getInstance(), &CommonSignals::dispatchMessage, this, &OnlineStatusWidget::onDispatchMessage);
    }
}

OnlineStatusWidget::~OnlineStatusWidget()
{

}

void OnlineStatusWidget::paintEvent(QPaintEvent *event)
{
    QPainter painter(this);
    painter.setRenderHint(QPainter::Antialiasing, true);
    {
        painter.save();
        if (m_enterStatus && isEnabled()) {
            painter.fillRect(event->rect(), QColor(192, 192, 192, 160));
        }
        painter.setBrush(Qt::NoBrush);
        if (g_is_ROG_Theme()) {
            std::string imagePath = m_imagePath.toStdString();
            boost::replace_head(imagePath, NORMAL_RES_PATH_HEAD_LEN, ROG_RES_PATH_HEAD_STR);
            painter.drawImage(event->rect(), QImage(imagePath.c_str()));
        } else {
            painter.drawImage(event->rect(), QImage(m_imagePath));
        }
        painter.restore();
    }

    if (m_statusCode >= 6 && m_onlineDevicesCount > 0) {
        painter.save();
        QFont font;
        font.setPixelSize(12);
        font.setBold(true);
        QPen pen(Qt::white, Qt::SolidLine);
        painter.setFont(font);
        painter.setPen(pen);
        QPoint textPos = event->rect().bottomRight();
        textPos -= QPoint(12, 5);
        painter.drawText(textPos, QString::number(m_onlineDevicesCount));
        painter.restore();
    }
}

void OnlineStatusWidget::enterEvent(QEvent *event)
{
    Q_UNUSED(event)
    m_enterStatus = true;
    update();

    ++m_cacheIndex;
    constexpr int delayTime = 400;
    QTimer::singleShot(delayTime, this, [this, cacheIndex = m_cacheIndex] {
        if (m_currentMenu || cacheIndex != m_cacheIndex || isEnabled() == false) {
            return;
        }
        SystemInfoMenuManager manager;
        if (m_statusCode >= 6 && g_getGlobalData()->m_clientVec.size() > 0) {
            m_currentMenu = manager.createClientInfoListMenu();
        } else {
            if (m_statusMessage.isEmpty()) {
                return;
            }
            m_currentMenu = manager.createTextInfoMenu(m_statusMessage);
        }
        if (m_currentMenu->actions().isEmpty()) {
            m_currentMenu->deleteLater();
            return;
        }
        SystemInfoMenuManager::updateMenuPos(m_currentMenu.data());
        m_currentMenu->exec(g_menuPos);
        m_currentMenu->deleteLater();
    });
}

void OnlineStatusWidget::leaveEvent(QEvent *event)
{
    Q_UNUSED(event)
    m_enterStatus = false;
    update();
    ++m_cacheIndex;
}

void OnlineStatusWidget::onDispatchMessage(const QVariant &data)
{
    if (data.canConvert<StatusInfoNotifyMsgPtr>() == true) {
        StatusInfoNotifyMsgPtr ptr_msg = data.value<StatusInfoNotifyMsgPtr>();
        QString logMessage;
        switch (ptr_msg->statusCode) {
        case 1: {
            logMessage = "Wait for connecting to the monitor...";
            m_statusMessage = "Detecting monitor...";
            m_imagePath = ":/resource/conn_status/connStatus1.png";
            break;
        }
        case 2: {
            logMessage = "Searching the service in LAN...";
            m_statusMessage = "Searching for service...";
            m_imagePath = ":/resource/conn_status/connStatus2.png";
            break;
        }
        case 3: {
            logMessage = "Checking authorization...";
            m_statusMessage = "Checking authorization...";
            m_imagePath = ":/resource/conn_status/connStatus3.png";
            break;
        }
        case 4: {
            logMessage = "Wait for screen casting...";
            m_statusMessage = logMessage;
            break;
        }
        case 5: {
            logMessage = "Failed! Authorization not available";
            m_statusMessage = "Authorization failed!";
            m_imagePath = ":/resource/conn_status/connStatus4.png";
            break;
        }
        case 6: {
            logMessage = "Connected. Wait for other clients";
            m_statusMessage = "Connected";
            m_imagePath = ":/resource/conn_status/connStatus5.png";
            break;
        }
        case 7: {
            logMessage = "Connected";
            m_statusMessage = "Connected";
            m_imagePath = ":/resource/conn_status/connStatus5.png";
            break;
        }
        default: {
            break;
        }
        }

        qInfo() << "[STATUS INFO]:" << logMessage << "; status_code=" << ptr_msg->statusCode;
        m_statusCode = static_cast<int>(ptr_msg->statusCode);
        update();
    }
}

void OnlineStatusWidget::onUpdateClientList()
{
    m_onlineDevicesCount = static_cast<int>(g_getGlobalData()->m_clientVec.size());
    QString infoText = QString("Online devices: %1; status_code: %2").arg(m_onlineDevicesCount).arg(m_statusCode);
    qInfo() << infoText;
    update();
}

void OnlineStatusWidget::onPipeDisconnected()
{
    m_statusCode = 1;
    m_onlineDevicesCount = static_cast<int>(g_getGlobalData()->m_clientVec.size());
    m_imagePath = ":/resource/rog/conn_status/connStatus1.png";
    update();
}

//-------------------------------------------------------------------------------------

NaviBarWidget::NaviBarWidget(QWidget *parent)
    : QWidget(parent)
    , ui(new Ui::NaviBarWidget)
    , m_onlineStatusWidget(nullptr)
{
    ui->setupUi(this);

    {
        ui->more_icon->clear();
        ui->title_logo->clear();
        ui->title_label->setProperty(PR_TEXT_BROWSER_INTERACTION_DISABLE, true);
        if (g_is_ROG_Theme()) {
            QPixmap pix(":/resource/rog/rog_logo.png");
            pix = pix.scaled(320, 100, Qt::KeepAspectRatio, Qt::SmoothTransformation);
            QLabel *logoLabel = new QLabel;
            logoLabel->setPixmap(pix);
            ui->left_layout->insertWidget(0, logoLabel);
        }

        {
            m_onlineStatusWidget = new OnlineStatusWidget;
            m_onlineStatusWidget->setFixedSize(50, 45);
            ui->right_layout->insertSpacing(1, 20);
            ui->right_layout->insertWidget(1, m_onlineStatusWidget);
        }

        {
            ui->navibar_box_middle->setProperty(PR_ADJUST_WINDOW_X_SIZE, true);
        }

        {
            std::string iconPath = ":/resource/menu_icon/more.png";
            if (g_is_ROG_Theme()) {
                boost::replace_head(iconPath, NORMAL_RES_PATH_HEAD_LEN, ROG_RES_PATH_HEAD_STR);
            }
            QPixmap pix(iconPath.c_str());
            {
                int targetSize = 20;
                pix = pix.scaled(targetSize, targetSize, Qt::KeepAspectRatio, Qt::SmoothTransformation);
                ui->more_icon->setPixmap(pix);
                ui->more_icon->setAlignment(Qt::AlignCenter);
            }
        }
    }

    {
        EventFilterProcess::getInstance()->registerFilterEvent({ ui->more_icon, std::bind(&NaviBarWidget::processMoreIconClicked, this) });
    }
}

NaviBarWidget::~NaviBarWidget()
{
    delete ui;
}

void NaviBarWidget::processMoreIconClicked()
{
    SystemInfoMenuManager manager;
    QMenu *mainMenu = manager.createMainMenu();
    SystemInfoMenuManager::updateMenuPos(mainMenu);
    mainMenu->exec(g_menuPos);
    mainMenu->deleteLater();
}
