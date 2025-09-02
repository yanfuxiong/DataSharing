#include "menu_manager.h"
#include "license_view.h"
#include "about_view.h"
#include "common_utils.h"
#include "common_signals.h"
#include "common_proxy_style.h"
#include "event_filter_process.h"
#include "device_info.h"
#include <QScreen>
#include <QApplication>

namespace {

struct LicenseItemInfo
{
    QString name;
    QString licenseFilePath;
};

}

SystemInfoMenuManager::SystemInfoMenuManager(QObject *parent)
    : QObject(parent)
{
    Q_EMIT CommonSignals::getInstance()->updateCurrentMenuPosData();
}

SystemInfoMenuManager::~SystemInfoMenuManager()
{

}

QPixmap SystemInfoMenuManager::createPixmap(const QString &imgPath, int size)
{
    QPixmap pix(imgPath);
    pix = pix.scaled(size, size, Qt::KeepAspectRatio, Qt::SmoothTransformation);
    return pix;
}

void SystemInfoMenuManager::updateMenuPos(QMenu *menu)
{
    QTimer::singleShot(0, menu, [menu] {
        auto point = g_menuPos;
        point -= QPoint(menu->width() + 2, 0);
        menu->move(point);
    });
}

QMenu *SystemInfoMenuManager::createMainMenu()
{
    QMenu *mainMenu = new QMenu;
    mainMenu->setProperty("SystemInfoMenuManager", true);

    QFont font;
    font.setPointSizeF(9);

    {
        auto subMenu = mainMenu->addMenu(QIcon(":/resource/menu_icon/settings.png"), "Settings");
        subMenu->setProperty("SystemInfoMenuManager", true);
        subMenu->menuAction()->setFont(font);
        {
            auto action = new DownloadPathItemAction;
            subMenu->addAction(action);

            connect(action, &DownloadPathItemAction::triggered, this, [] {
                Q_EMIT CommonSignals::getInstance()->modifyDefaulutDownloadPath();
            });
        }

        {
            auto action = new BugReportItemAction;
            subMenu->addAction(action);

            connect(action, &BugReportItemAction::triggered, this, [] {
                Q_EMIT CommonSignals::getInstance()->bugReport();
            });
        }
    }

    {
        auto subMenu = mainMenu->addMenu(QIcon(":/resource/menu_icon/info.png"), "Info");
        subMenu->setProperty("SystemInfoMenuManager", true);
        subMenu->menuAction()->setFont(font);
        subMenu->addAction(QIcon(":/resource/menu_icon/info_about.png"), "About", [] {
            auto *viewWindow = new AboutView;
            viewWindow->show();
            QTimer::singleShot(0, viewWindow, [viewWindow] {
                viewWindow->move(g_menuPos - QPoint(viewWindow->width() + 4, -2));
            });
        })->setFont(font);

        subMenu->addAction(QIcon(":/resource/menu_icon/info_license.png"), "Open source license", [] {
            Q_EMIT CommonSignals::getInstance()->showOpenSourceLicenseMenu();
        })->setFont(font);
    }

    return mainMenu;
}

QMenu *SystemInfoMenuManager::createOpenSourceLicenseMenu()
{
    QMenu *menu = new QMenu;
    menu->setProperty("SystemInfoMenuManager", true);
    QFont font;
    font.setPointSizeF(9);
    menu->menuAction()->setFont(font);

    {
        {
            auto action = new LicenseTitleItemAction;
            menu->addAction(action);
        }

        std::vector<LicenseItemInfo> licenseItemInfoVec = {
                { "Qt 5.15.17", qApp->applicationDirPath() + "/licenses/Qt_LGPLv3.txt" },
                { "Boost 1.88", qApp->applicationDirPath() + "/licenses/Boost_LICENSE_1_0.txt" },
                { "nlohmann json", qApp->applicationDirPath() + "/licenses/nlohmann_json_MIT.txt" }
        };

        for (int index = 0; index < static_cast<int>(licenseItemInfoVec.size()); ++index) {
            const LicenseItemInfo &licenseItem = licenseItemInfoVec[index];
            auto action = new LicenseItemAction(licenseItem.name);
            connect(action, &LicenseItemAction::triggered, this, [licenseItem] {
                auto *viewWindow = new LicenseView;
                viewWindow->setTitle(licenseItem.name);
                viewWindow->setDisplayInfo(CommonUtils::getFileContent(licenseItem.licenseFilePath));
                viewWindow->show();
                QTimer::singleShot(0, viewWindow, [viewWindow] {
                    viewWindow->move(g_menuPos - QPoint(viewWindow->width() + 4, -2));
                });
            });
            menu->addAction(action);

            if (index + 1 < static_cast<int>(licenseItemInfoVec.size())) {
                auto action = new HLineItemAction;
                menu->addAction(action);
            }
        }
    }

    return menu;
}

QMenu *SystemInfoMenuManager::createClientInfoListMenu()
{
    QMenu *clientInfoListMenu = new QMenu;
    clientInfoListMenu->setProperty("SystemInfoMenuManager", true);

    for (const auto &data : g_getGlobalData()->m_clientVec) {
        QString info;
        info += data->clientName + "\n";
        info += data->ip + ":" + QString::number(data->port);

        auto action = new ClientInfoListItemAction;
        action->setIconPath(DeviceInfo::getIconPathByDeviceType(data->deviceType));
        action->setInfo(info);
        clientInfoListMenu->addAction(action);
    }

    return clientInfoListMenu;
}

QMenu *SystemInfoMenuManager::createTextInfoMenu(const QString &textInfo)
{
    QMenu *textInfoMenu = new QMenu;
    textInfoMenu->setFixedWidth(260);
    textInfoMenu->setProperty("SystemInfoMenuManager", true);
    auto action = new TextInfoItemAction;
    action->setInfo(textInfo);
    textInfoMenu->addAction(action);
    return textInfoMenu;
}

//----------------------------------
LicenseItem::LicenseItem(const QString &name, QWidget *parent)
    : QWidget(parent)
    , m_name(name)
{
    QHBoxLayout *pHBoxLayout = new QHBoxLayout;
    pHBoxLayout->setMargin(0);
    pHBoxLayout->setSpacing(0);
    setLayout(pHBoxLayout);
    QFont font;
    font.setPointSizeF(9);
    {
        QLabel *nameLabel = new QLabel;
        nameLabel->setFont(font);
        nameLabel->setAttribute(Qt::WA_TransparentForMouseEvents, true);
        nameLabel->setSizePolicy(QSizePolicy::Expanding, QSizePolicy::Expanding);
        nameLabel->setText(m_name);
        nameLabel->setContentsMargins(16, 0, 0, 0);
        pHBoxLayout->addWidget(nameLabel);
    }

    {
        QLabel *arrowLabel = new QLabel;
        arrowLabel->setAttribute(Qt::WA_TransparentForMouseEvents, true);
        arrowLabel->setMargin(0);
        arrowLabel->setObjectName("rightArrow");
        arrowLabel->setFixedWidth(24);
        arrowLabel->setPixmap(SystemInfoMenuManager::createPixmap(":/resource/menu_icon/right-arrow.png", 10));

        pHBoxLayout->addWidget(arrowLabel);
    }
}

LicenseItem::~LicenseItem()
{

}

void LicenseItem::enterEvent(QEvent *event)
{
    setStyleSheet("background-color: rgb(0,120, 215);color:white;");
    QWidget::enterEvent(event);
}

void LicenseItem::leaveEvent(QEvent *event)
{
    setStyleSheet("background-color: rgb(240, 240, 240);color:black;");
    QWidget::leaveEvent(event);
}

void LicenseItem::mousePressEvent(QMouseEvent *event)
{
    Q_EMIT actionTriggered();
    QWidget::mousePressEvent(event);
}

//----------------------------------------
LicenseItemAction::LicenseItemAction(const QString &name, QObject *parent)
    : QWidgetAction(parent)
    , m_name(name)
{
}

LicenseItemAction::~LicenseItemAction() = default;

QWidget *LicenseItemAction::createWidget(QWidget *parent)
{
    LicenseItem *itemWidget = new LicenseItem(m_name, parent);
    itemWidget->setSizePolicy(QSizePolicy::Expanding, QSizePolicy::Expanding);
    itemWidget->setFixedHeight(g_getMenuItemHeight());
    connect(itemWidget, &LicenseItem::actionTriggered, this, [this] {
        Q_EMIT triggered();
    });
    return itemWidget;
}

//---------------------------------- LicenseTitleItem
LicenseTitleItem::LicenseTitleItem(QWidget *parent)
    : QWidget(parent)
{
    QHBoxLayout *pHBoxLayout = new QHBoxLayout;
    pHBoxLayout->setMargin(0);
    pHBoxLayout->setSpacing(0);
    setLayout(pHBoxLayout);
    QFont font;
    font.setPointSizeF(9);

    pHBoxLayout->addSpacing(5);

    {
        QLabel *iconLabel = new QLabel;
        iconLabel->setAttribute(Qt::WA_TransparentForMouseEvents, true);
        iconLabel->setMargin(0);
        iconLabel->setFixedWidth(24);
        iconLabel->setPixmap(SystemInfoMenuManager::createPixmap(":/resource/menu_icon/info_license.png", 20));

        pHBoxLayout->addWidget(iconLabel);
    }

    {
        QLabel *nameLabel = new QLabel;
        nameLabel->setFont(font);
        nameLabel->setAttribute(Qt::WA_TransparentForMouseEvents, true);
        nameLabel->setSizePolicy(QSizePolicy::Expanding, QSizePolicy::Expanding);
        nameLabel->setText("Open source license");
        nameLabel->setContentsMargins(12, 0, 50, 0);
        pHBoxLayout->addWidget(nameLabel);
    }
}

LicenseTitleItem::~LicenseTitleItem() = default;

//----------------------------- LicenseTitleItemAction

LicenseTitleItemAction::LicenseTitleItemAction(QObject *parent)
    : QWidgetAction(parent)
{

}

LicenseTitleItemAction::~LicenseTitleItemAction() = default;

QWidget *LicenseTitleItemAction::createWidget(QWidget *parent)
{
    LicenseTitleItem *itemWidget = new LicenseTitleItem(parent);
    itemWidget->setFixedHeight(g_getMenuItemHeight());
    return itemWidget;
}

// -------------------------------------- HLineItem
HLineItem::HLineItem(QWidget *parent)
    : QWidget(parent)
{
    QHBoxLayout *pHBoxLayout = new QHBoxLayout;
    pHBoxLayout->setMargin(0);
    pHBoxLayout->setSpacing(0);
    setLayout(pHBoxLayout);

    {
        QLabel *hLine = new QLabel;
        hLine->setStyleSheet("background-color:rgb(200,200,200);");
        hLine->setSizePolicy(QSizePolicy::Expanding, QSizePolicy::Expanding);
        hLine->setFixedHeight(1);
        pHBoxLayout->addSpacing(5);
        pHBoxLayout->addWidget(hLine);
        pHBoxLayout->addSpacing(5);
    }
}

HLineItem::~HLineItem() = default;


HLineItemAction::HLineItemAction(QObject *parent)
    : QWidgetAction(parent)
{

}

HLineItemAction::~HLineItemAction() = default;

QWidget *HLineItemAction::createWidget(QWidget *parent)
{
    HLineItem *itemWidget = new HLineItem(parent);
    itemWidget->setFixedHeight(1);
    return itemWidget;
}


// ---------------------------- DownloadPathItem
DownloadPathItem::DownloadPathItem(QWidget *parent)
    : QWidget(parent)
{
    QHBoxLayout *pHBoxLayout = new QHBoxLayout;
    pHBoxLayout->setMargin(0);
    pHBoxLayout->setSpacing(0);
    setLayout(pHBoxLayout);
    QFont font;
    font.setPointSizeF(9);

    {
        QLabel *iconLabel = new QLabel;
        //iconLabel->setAttribute(Qt::WA_TransparentForMouseEvents, true);
        iconLabel->setMargin(0);
        iconLabel->setFixedWidth(24);
        iconLabel->setPixmap(SystemInfoMenuManager::createPixmap(":/resource/menu_icon/settings_downloadLocation.png", 20));

        pHBoxLayout->addWidget(iconLabel);
    }

    {
        QLabel *nameLabel = new QLabel;
        nameLabel->setFont(font);
        //nameLabel->setAttribute(Qt::WA_TransparentForMouseEvents, true);
        nameLabel->setSizePolicy(QSizePolicy::Expanding, QSizePolicy::Expanding);
        nameLabel->setText("Download path");
        nameLabel->setContentsMargins(12, 0, 20, 0);
        pHBoxLayout->addWidget(nameLabel);
    }

    pHBoxLayout->addStretch();

    {
        QPushButton *button = new QPushButton(QIcon(":/resource/menu_icon/settings_folder.png"), "  Default");
        button->setObjectName("customMenuButton");
        button->setFixedSize(100, g_getMenuItemHeight() - 6);
        button->setProperty(PR_ADJUST_WINDOW_X_SIZE, true);
        pHBoxLayout->addWidget(button);
        button->setFont(font);
        connect(button, &QPushButton::clicked, this, &DownloadPathItem::actionTriggered);
    }

    pHBoxLayout->addSpacing(8);
}

DownloadPathItem::~DownloadPathItem() = default;


DownloadPathItemAction::DownloadPathItemAction(QObject *parent)
    : QWidgetAction(parent)
{

}

DownloadPathItemAction::~DownloadPathItemAction() = default;

QWidget *DownloadPathItemAction::createWidget(QWidget *parent)
{
    DownloadPathItem *itemWidget = new DownloadPathItem(parent);
    itemWidget->setFixedHeight(g_getMenuItemHeight());
    connect(itemWidget, &DownloadPathItem::actionTriggered, this, [this] {
        Q_EMIT triggered();
    });
    return itemWidget;
}

//-------------------------

BugReportItem::BugReportItem(QWidget *parent)
    : QWidget(parent)
{
    QHBoxLayout *pHBoxLayout = new QHBoxLayout;
    pHBoxLayout->setMargin(0);
    pHBoxLayout->setSpacing(0);
    setLayout(pHBoxLayout);
    QFont font;
    font.setPointSizeF(9);

    {
        QLabel *iconLabel = new QLabel;
        //iconLabel->setAttribute(Qt::WA_TransparentForMouseEvents, true);
        iconLabel->setMargin(0);
        iconLabel->setFixedWidth(24);
        iconLabel->setPixmap(SystemInfoMenuManager::createPixmap(":/resource/menu_icon/settings_bugReport.png", 20));

        pHBoxLayout->addWidget(iconLabel);
    }

    {
        QLabel *nameLabel = new QLabel;
        nameLabel->setFont(font);
        //nameLabel->setAttribute(Qt::WA_TransparentForMouseEvents, true);
        nameLabel->setSizePolicy(QSizePolicy::Expanding, QSizePolicy::Expanding);
        nameLabel->setText("Bug report");
        nameLabel->setContentsMargins(12, 0, 20, 0);
        pHBoxLayout->addWidget(nameLabel);
    }

    pHBoxLayout->addStretch();

    {
        QPushButton *button = new QPushButton(QIcon(":/resource/menu_icon/settings_fileExport.png"), "  Export");
        button->setObjectName("customMenuButton");
        button->setFixedSize(100, g_getMenuItemHeight() - 6);
        button->setProperty(PR_ADJUST_WINDOW_X_SIZE, true);
        pHBoxLayout->addWidget(button);
        button->setFont(font);
        connect(button, &QPushButton::clicked, this, &BugReportItem::actionTriggered);
    }

    pHBoxLayout->addSpacing(8);
}

BugReportItem::~BugReportItem() = default;


BugReportItemAction::BugReportItemAction(QObject *parent)
    : QWidgetAction(parent)
{

}

BugReportItemAction::~BugReportItemAction() = default;

QWidget *BugReportItemAction::createWidget(QWidget *parent)
{
    BugReportItem *itemWidget = new BugReportItem(parent);
    itemWidget->setFixedHeight(g_getMenuItemHeight());
    connect(itemWidget, &BugReportItem::actionTriggered, this, [this] {
        Q_EMIT triggered();
    });
    return itemWidget;
}

//------------------------------------------------------------------
ClientInfoListItem::ClientInfoListItem(const QString &iconPath, const QString &info, QWidget *parent)
    : QWidget(parent)
{
    QHBoxLayout *pHBoxLayout = new QHBoxLayout;
    pHBoxLayout->setMargin(0);
    pHBoxLayout->setSpacing(0);
    setLayout(pHBoxLayout);
    pHBoxLayout->addSpacing(10);
    {
        QLabel *iconLabel = new QLabel;
        iconLabel->setSizePolicy(QSizePolicy::Expanding, QSizePolicy::Expanding);
        iconLabel->setAttribute(Qt::WA_TransparentForMouseEvents, true);
        iconLabel->setMargin(0);
        iconLabel->setFixedSize(40, 40);
        iconLabel->setPixmap(SystemInfoMenuManager::createPixmap(iconPath, 40));

        pHBoxLayout->addWidget(iconLabel);
    }

    pHBoxLayout->addSpacing(15);

    {
        QLabel *infoLabel = new QLabel;
        infoLabel->setSizePolicy(QSizePolicy::Expanding, QSizePolicy::Expanding);
        infoLabel->setAttribute(Qt::WA_TransparentForMouseEvents, true);
        infoLabel->setMargin(0);
        infoLabel->setText(info);

        pHBoxLayout->addWidget(infoLabel);
    }
}

ClientInfoListItem::~ClientInfoListItem() = default;

ClientInfoListItemAction::ClientInfoListItemAction(QObject *parent)
    : QWidgetAction(parent)
{

}

ClientInfoListItemAction::~ClientInfoListItemAction() = default;

QWidget *ClientInfoListItemAction::createWidget(QWidget *parent)
{
    ClientInfoListItem *itemWidget = new ClientInfoListItem(m_iconPath, m_info, parent);
    itemWidget->setFixedSize(250, 50);
    connect(itemWidget, &ClientInfoListItem::actionTriggered, this, [this] {
        Q_EMIT triggered();
    });
    return itemWidget;
}

//--------------------------------------------------
TextInfoItem::TextInfoItem(const QString &info, QWidget *parent)
    : QWidget(parent)
{
    QHBoxLayout *pHBoxLayout = new QHBoxLayout;
    pHBoxLayout->setMargin(0);
    pHBoxLayout->setSpacing(0);
    setLayout(pHBoxLayout);
    {
        QLabel *textLabel = new QLabel;
        textLabel->setAlignment(Qt::AlignVCenter | Qt::AlignLeft);
        textLabel->setSizePolicy(QSizePolicy::Expanding, QSizePolicy::Expanding);
        textLabel->setText(info);
        pHBoxLayout->addWidget(textLabel);
    }
}

TextInfoItem::~TextInfoItem() = default;

TextInfoItemAction::TextInfoItemAction(QObject *parent)
    : QWidgetAction(parent)
{

}

TextInfoItemAction::~TextInfoItemAction() = default;

QWidget *TextInfoItemAction::createWidget(QWidget *parent)
{
    TextInfoItem *itemWidget = new TextInfoItem(m_info, parent);
    itemWidget->setFixedHeight(g_getMenuItemHeight());
    connect(itemWidget, &TextInfoItem::actionTriggered, this, [this] {
        Q_EMIT triggered();
    });
    return itemWidget;
}
