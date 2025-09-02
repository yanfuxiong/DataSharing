#include "license_view.h"
#include "ui_license_view.h"
#include "event_filter_process.h"
#include "common_signals.h"
#include "menu_manager.h"
#include <QDebug>

LicenseView::LicenseView(QWidget *parent)
    : QWidget(parent)
    , ui(new Ui::LicenseView)
{
    ui->setupUi(this);
    setWindowFlags(windowFlags() | (Qt::Popup | Qt::FramelessWindowHint));
    setAttribute(Qt::WA_DeleteOnClose, true);
    ui->license_icon_label->clear();
    ui->view_top_box->setFlat(true);
    QTimer::singleShot(0, this, [this] {
        ui->license_icon_label->setPixmap(SystemInfoMenuManager::createPixmap(":/resource/menu_icon/left-arrow.png",
                                                                              static_cast<int>(ui->license_icon_label->width() * 3 / 5.0)));
    });

    {
        {
            QFont font;
            font.setPointSizeF(9);
            ui->license_info->setFont(font);
        }

        {
            QFont font;
            font.setPointSizeF(12);
            ui->license_title_label->setFont(font);
        }

        ui->view_top_box->setProperty(PR_ADJUST_WINDOW_Y_SIZE, true);
        ui->license_icon_label->setProperty(PR_ADJUST_WINDOW_X_SIZE, true);

        setProperty(PR_ADJUST_WINDOW_X_SIZE, true);
    }

    EventFilterProcess::getInstance()->registerFilterEvent({ ui->license_icon_label, std::bind(&LicenseView::clickedLeftArrow, this) });
}

LicenseView::~LicenseView()
{
    qDebug() << "------------------------------------delete LicenseView;" << ui->license_title_label->text();
    delete ui;
}

void LicenseView::setTitle(const QString &title)
{
    ui->license_title_label->setText(title);
}

void LicenseView::setDisplayInfo(const QString &info)
{
    ui->license_info->setText(info);
}

void LicenseView::clickedLeftArrow()
{
    deleteLater();
    Q_EMIT CommonSignals::getInstance()->showOpenSourceLicenseMenu();
}
