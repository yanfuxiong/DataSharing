#include "about_view.h"
#include "ui_about_view.h"
#include "menu_manager.h"
#include "common_utils.h"
#include <QDebug>
#include <QFrame>

#define COPYRIGHT_INFO_STR \
"© 2025 Realtek Semiconductor Corp. All rights reserved"

AboutView::AboutView(QWidget *parent)
    : QWidget(parent)
    , ui(new Ui::AboutView)
{
    ui->setupUi(this);
    setWindowFlags(windowFlags() | (Qt::Popup | Qt::FramelessWindowHint));
    setAttribute(Qt::WA_DeleteOnClose, true);
    ui->about_title_label->setText("About");

    {
        {
            QFrame *hLine = new QFrame;
            hLine->setObjectName("about_view_h_line");
            hLine->setFixedHeight(1);
            hLine->setFrameShape(QFrame::HLine);
            ui->aboutViewLayout->insertWidget(1, hLine);
        }
        {
            QLabel *emptyLabel = new QLabel;
            emptyLabel->setStyleSheet("QLabel{background-color:white;}");
            emptyLabel->setSizePolicy(QSizePolicy::Expanding, QSizePolicy::Expanding);
            emptyLabel->setFixedWidth(10);
            ui->topBox_layout->insertWidget(0, emptyLabel);
        }

        QTimer::singleShot(0, this, [this] {
            ui->about_icon_label->setPixmap(SystemInfoMenuManager::createPixmap(":/resource/menu_icon/info_about.png",
                                                                                static_cast<int>(ui->about_icon_label->width() * 4 / 5.0)));
        });
    }

    {
        {
            QFont font;
            font.setPointSizeF(8);
            ui->version_info_key_label->setFont(font);
            ui->version_info_label->setFont(font);
            ui->copyright_info_label->setFont(font);
        }

        {
            QFont font;
            font.setPointSizeF(9);
            ui->content_info_label->setFont(font);
        }
        ui->version_info_key_label->setText(getVersionInfoKey());
        ui->version_info_label->setText(getVersionInfo());
        ui->content_info_label->setWordWrap(true);
        ui->content_info_label->setText(getReadmeInfo());
        ui->copyright_info_label->setText(COPYRIGHT_INFO_STR);
    }

    {
        setProperty(PR_ADJUST_WINDOW_X_SIZE, true);
        setProperty(PR_ADJUST_WINDOW_Y_SIZE, true);

        ui->about_top_box->setProperty(PR_ADJUST_WINDOW_Y_SIZE, true);
        ui->about_icon_label->setProperty(PR_ADJUST_WINDOW_X_SIZE, true);

        ui->version_info_key_label->setProperty(PR_ADJUST_WINDOW_X_SIZE, true);
        ui->version_info_key_label->setProperty(PR_ADJUST_WINDOW_Y_SIZE, true);

        ui->version_info_label->setProperty(PR_ADJUST_WINDOW_Y_SIZE, true);
        ui->about_bottom_box->setProperty(PR_ADJUST_WINDOW_Y_SIZE, true);
    }
}

AboutView::~AboutView()
{
    qDebug() << "------------------------------------delete AboutView;";
    delete ui;
}

QString AboutView::getVersionInfoKey() const
{
    QString appKey = "App Version";
    QString serviceVersionKey = "Service Version";
    QString nameKey = "Name";
    QString ipKey = "IP";
    QString infoString = appKey + "\n" +
                        serviceVersionKey + "\n" +
                        nameKey + "\n" +
                        ipKey;
    return infoString;
}

QString AboutView::getVersionInfo() const
{
    QString appVersion = qApp->applicationVersion();
    QString goServerVersion = g_getGlobalData()->systemConfig.serverVersionStr;
    QString machineName = QSysInfo::machineHostName();
    QString ipAddress = g_getGlobalData()->systemConfig.localIpAddress;

    QString infoString = appVersion + "\n" +
                             goServerVersion + "\n" +
                             machineName + "\n" +
                             ipAddress;
    return infoString;
}

QString AboutView::getReadmeInfo() const
{
    return CommonUtils::getFileContent(":/resource/about_view_content.txt");
}
