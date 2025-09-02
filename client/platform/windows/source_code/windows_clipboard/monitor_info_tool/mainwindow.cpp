#include "mainwindow.h"
#include "ui_mainwindow.h"
#include "monitor_info.h"
#include "common_signals.h"
#include "windows_event_monitor.h"
#include <QTimer>

MainWindow::MainWindow(QWidget *parent) :
    QMainWindow(parent),
    ui(new Ui::MainWindow)
{
    ui->setupUi(this);

    QTimer::singleShot(0, this, [this] {
        on_get_info_btn_clicked();
    });

    connect(CommonSignals::getInstance(), &CommonSignals::userSelectedFiles, this, [this] {
        ++m_cacheIndex;
        auto currentIndexVal = m_cacheIndex;
        QTimer::singleShot(800, this, [currentIndexVal, this] {
            if (currentIndexVal < m_cacheIndex) {
                return;
            }

            const auto &pathList = WindowsEventMonitor::getSelectedPathList();
            QString pathInfoString;
            if (pathList.empty() == false) {
                pathInfoString += QByteArray(100, '-') + "\n";
            }
            for (const auto &path : pathList) {
                pathInfoString += path + "\n";
            }
            ui->monitor_info->clear();
            ui->monitor_info->append(pathInfoString);
        });
    });

    connect(MonitorPlugEvent::getInstance(), &MonitorPlugEvent::statusChanged, this, [this] (bool status) {
        if (status) {
            ui->monitor_info->append("monitor connected........." + QDateTime::currentDateTime().toString("yyyy-MM-dd hh:mm:ss.zzz"));
        } else {
            ui->monitor_info->append("monitor disconnected........" + QDateTime::currentDateTime().toString("yyyy-MM-dd hh:mm:ss.zzz"));
            MonitorPlugEvent::getInstance()->clearData();
        }
    });

    {
        connect(WindowsEventMonitor::getInstance(), &WindowsEventMonitor::clickedPos, this, [this] (const QPoint &pt) {
            int width = GetSystemMetrics(SM_CXSCREEN);
            int height = GetSystemMetrics(SM_CYSCREEN);
            ui->monitor_info->append(QString("x=%1,y=%2, width=%3,height=%4").arg(pt.x()).arg(pt.y()).arg(width).arg(height));
        });
    }

    QTimer::singleShot(0, this, [] {
        MonitorPlugEvent::getInstance()->initData();
    });
}

MainWindow::~MainWindow()
{
    delete ui;
}

void MainWindow::on_get_info_btn_clicked()
{
    ui->monitor_info->clear();

    MonitorInfoUtils utils;
    for (const auto &info : utils.parseAllEdidInfo()) {
        nlohmann::json infoJson;
        infoJson["manufacturer"] = info.manufacturer;
        infoJson["productName"] = info.productName;
        if (info.productSerialNumber.empty() == false) {
            infoJson["productSerialNumber"] = info.productSerialNumber;
        }
        infoJson["edidVersion"] = info.edidVersion;
        if (info.gamma.has_value()) {
            infoJson["gamma"] = info.gamma.value();
        }

        if (info.manufactureWeek.has_value()) {
            infoJson["manufactureDate"] = QString("year %1 (week %2)").arg(info.manufactureYear).arg(info.manufactureWeek.value()).toStdString();
        } else {
            infoJson["manufactureDate"] = QString("year %1").arg(info.manufactureYear).toStdString();
        }
        infoJson["maxHSize"] = std::to_string(info.maxHSize) + "cm";
        infoJson["maxVSize"] = std::to_string(info.maxVSize) + "cm";
        infoJson["size"] = QString::asprintf("%.2f inch", info.size).toStdString();

        qInfo() << infoJson.dump(4).c_str();
        ui->monitor_info->append(infoJson.dump(4).c_str());
    }
}

void MainWindow::on_clear_btn_clicked()
{
    ui->monitor_info->clear();
}

