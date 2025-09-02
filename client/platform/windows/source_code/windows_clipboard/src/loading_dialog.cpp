#include "loading_dialog.h"
#include "ui_loading_dialog.h"
#include "windows_event_monitor.h"
#include <QTimer>

LoadingDialog::LoadingDialog(QWidget *parent)
    : QDialog(parent)
    , ui(new Ui::LoadingDialog)
    , m_moveHandler(this)
{
    ui->setupUi(this);
    setWindowFlag(Qt::FramelessWindowHint, true);
    setWindowFlag(Qt::WindowStaysOnTopHint, true);
    setAttribute(Qt::WA_TranslucentBackground, true);
    setWindowFlag(Qt::Dialog, true);
    setMouseTracking(true);
    setAutoFillBackground(true);

    {
        try {
            m_timerThreshold = g_getGlobalData()->localConfig.at("UITheme").at("timerThreshold").get<int>();
        } catch (const std::exception &e) {
            qWarning() << e.what();
            m_timerThreshold = 3000;
        }

        {
            MonitorPlugEvent::getInstance()->initData();
            QEventLoop loop;
            QTimer::singleShot(300, Qt::TimerType::PreciseTimer, &loop, [&loop] {
                loop.quit();
            });
            loop.exec();
        }

        if (MonitorPlugEvent::getInstance()->getCacheMonitorData().macAddress.empty()) {
            m_showTooltipText = true;
            m_showLoadingGIF = true;
            connect(MonitorPlugEvent::getInstance(), &MonitorPlugEvent::statusChanged, this, &LoadingDialog::onDIASStatusChanged);
            m_elapsedTimer.restart();
        } else {
            m_showTooltipText = false;
            m_showLoadingGIF = true;
            m_elapsedTimer.restart();
            QTimer::singleShot(0, this, &LoadingDialog::onGetThemeCode);
        }
    }

    {
        QTimer *pTimer = new QTimer;
        pTimer->setInterval(250);
        m_crossShareTextImagePath = ":/resource/loading/1.png";
        connect(pTimer, &QTimer::timeout, this, [this] {
            static int s_index = 1;
            s_index = s_index >= 7 ? 1 : s_index + 1;
            m_crossShareTextImagePath = QString(":/resource/loading/%1.png").arg(s_index);
            update();
        });
        pTimer->start();
    }

    {
        m_movie = new QMovie(this);
        m_movie->setFileName(":/resource/loading/loading.gif");
        connect(m_movie, &QMovie::frameChanged, this, &LoadingDialog::onFrameChanged);
        m_movie->start();
    }

    {
        constexpr int btn_width = 23;
        constexpr int btn_height = 23;
        m_closeBtnRect.setX(width() - btn_width - 20);
        m_closeBtnRect.setY(20);
        m_closeBtnRect.setWidth(btn_width);
        m_closeBtnRect.setHeight(btn_height);
    }

    {
        m_textRect.setX(0);
        m_textRect.setY(height() - 150);
        m_textRect.setWidth(width());
        m_textRect.setHeight(40);
    }
}

LoadingDialog::~LoadingDialog()
{
    delete ui;
}

QPixmap LoadingDialog::createPixmap(const QString &imgPath, int size)
{
    QPixmap pix(imgPath);
    pix = pix.scaled(size, size, Qt::KeepAspectRatio, Qt::SmoothTransformation);
    return pix;
}

void LoadingDialog::paintEvent(QPaintEvent *event)
{
    QPainter painter(this);
    painter.setRenderHint(QPainter::Antialiasing, true);
    painter.drawPixmap(event->rect(), QPixmap(":/resource/loading/bg.png"));

    {
        painter.save();
        if (m_closeBtnStatus) {
            painter.drawPixmap(m_closeBtnRect.toRect(), QPixmap(":/resource/loading/close_hover.svg"));
        } else {
            painter.drawPixmap(m_closeBtnRect.toRect(), QPixmap(":/resource/loading/close.svg"));
        }
        painter.restore();
    }

    if (m_showLoadingGIF) {
        painter.save();
        QPixmap currentFrame = m_movie->currentPixmap();
        QRectF pixRect;
        pixRect.setWidth(134);
        pixRect.setHeight(134);
        pixRect.moveCenter(QPointF(0, 0));
        pixRect.moveCenter(QPointF(event->rect().center().x(), event->rect().height() / 3));
        if (!currentFrame.isNull()) {
            painter.drawPixmap(pixRect.toRect(), currentFrame);
        }
        painter.restore();
    }

    {
        painter.save();
        QRectF crossShareTextRect;
        // 890 x 250
        crossShareTextRect.setWidth(890 / 4.0);
        crossShareTextRect.setHeight(250 / 4.0);
        crossShareTextRect.moveCenter(QPointF(0, 0));
        crossShareTextRect.moveCenter(QPointF(width() / 2.0, 260));
        QPixmap pix(m_crossShareTextImagePath);
        {
            pix = pix.scaled(crossShareTextRect.width(), crossShareTextRect.height(), Qt::KeepAspectRatio, Qt::SmoothTransformation);
        }
        painter.drawPixmap(crossShareTextRect.toRect(), pix);
        painter.restore();
    }

    if (m_showTooltipText) {
        painter.save();
        QFont font;
        font.setFamily("Arial");
        font.setPixelSize(18);
        font.setBold(false);
        painter.setFont(font);
        const char *text = "Please plug in a CrossShare-compatible monitor.";
        painter.drawText(m_textRect, Qt::AlignCenter, text);
        painter.restore();
    }
}

void LoadingDialog::onFrameChanged(int frameNumber)
{
    Q_UNUSED(frameNumber)
    update();
}

void LoadingDialog::onClickedCloseButton()
{
    Q_EMIT CommonSignals::getInstance()->quitAllEventLoop();
    reject();
}

void LoadingDialog::onDIASStatusChanged(bool status)
{
    if (status) {
        auto cacheData = MonitorPlugEvent::getInstance()->getCacheMonitorData();
        if (cacheData.macAddress.empty() == false) {
            qInfo() << "macAddress:" << QByteArray::fromStdString(cacheData.macAddress).toHex().toUpper().constData();
            if (m_elapsedTimer.hasExpired(m_timerThreshold)) {
                m_showLoadingGIF = true;
                m_showTooltipText = false;
                m_elapsedTimer.restart();
                QTimer::singleShot(0, this, &LoadingDialog::onGetThemeCode);
            } else {
                QTimer::singleShot(std::abs(m_timerThreshold - m_elapsedTimer.elapsed()), Qt::TimerType::PreciseTimer, this, [this] {
                    m_showLoadingGIF = true;
                    m_showTooltipText = false;
                    m_elapsedTimer.restart();
                    QTimer::singleShot(0, this, &LoadingDialog::onGetThemeCode);
                });
            }
        }
    }
}

void LoadingDialog::onGetThemeCode()
{
    qInfo() << "-------------------------------------onGetThemeCode";
    uint32_t themeCode = 0;
    while (true) {
        MonitorPlugEvent::getInstance()->refreshCachedMonitorData();
        if (MonitorPlugEvent::getInstance()->getCacheMonitorData().macAddress.empty()) {
            MonitorPlugEvent::delayInEventLoop(50);
        } else {
            qInfo() << "[onGetThemeCode]: macAddress=" << QByteArray::fromStdString(MonitorPlugEvent::getInstance()->getCacheMonitorData().macAddress).toHex().toUpper().constData();
            break;
        }
    }
    bool retVal = MonitorPlugEvent::getCustomerThemeCode(MonitorPlugEvent::getInstance()->getCacheMonitorData().hPhysicalMonitor, themeCode);
    if (retVal) {
        m_themeCode = themeCode;
        if (m_elapsedTimer.hasExpired(m_timerThreshold)) {
            accept();
        } else {
            QTimer::singleShot(std::abs(m_timerThreshold - m_elapsedTimer.elapsed()), Qt::TimerType::PreciseTimer, this, [this] {
                accept();
            });
        }
    }
}

void LoadingDialog::mouseMoveEvent(QMouseEvent *event)
{
    if (m_closeBtnRect.contains(event->pos())) {
        if (m_closeBtnStatus == false) {
            m_closeBtnStatus = true;
            update();
        }
    } else {
        if (m_closeBtnStatus == true) {
            m_closeBtnStatus = false;
            update();
        }
    }
}

void LoadingDialog::mousePressEvent(QMouseEvent *event)
{
    if (m_closeBtnRect.contains(event->pos())) {
        onClickedCloseButton();
    }
}

void LoadingDialog::updateThemeCode(uint32_t themeCode)
{
    try {
        g_getGlobalData()->localConfig["UITheme"]["customerID"] = themeCode;
        g_getGlobalData()->localConfig["UITheme"]["isInited"] = true;
        g_updateLocalConfig();
    } catch (const std::exception &e) {
        qWarning() << e.what();
    }

    {
        UpdateLocalConfigInfoMsg message;
        message.configData = g_getGlobalData()->localConfig.dump().c_str();
        message.appFilePath = qApp->applicationFilePath();
        message.timeStamp = QDateTime::currentDateTime().toMSecsSinceEpoch();
        message.appVersion = qApp->applicationVersion();
        g_sendDataToServer(UpdateLocalConfigInfo_code, UpdateLocalConfigInfoMsg::toByteArray(message));
    }
}
