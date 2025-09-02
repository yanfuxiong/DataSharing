#include "mainwindow.h"
#include <QApplication>
#include <QProcess>
#include <QSharedMemory>
#include <QMessageBox>
#include <iostream>
#include "common_utils.h"
#include "common_proxy_style.h"
#include "common_ui_process.h"
#include "process_message.h"

uint64_t g_timeStamp { 0 };

int main(int argc, char *argv[])
{
    if (argc < 2) {
        std::cerr << argv[0] << "------------missing parameters hashId" << std::endl;
        return 1;
    }
    g_timeStamp = QByteArray::fromHex(argv[1]).toULongLong();

    qInstallMessageHandler(CommonUtils::commonMessageOutput);
    QApplication a(argc, argv);

    a.setStyleSheet(CommonUtils::getFileContent(":/resource/my.qss"));
    a.setStyle(new CustomProxyStyle);

    {
        g_globalRegister();
        g_namedPipeServerName = g_helperServerName; // The connected server has changed
        CommonUiProcess::getInstance();
        ProcessMessage::getInstance();
    }

    MainWindow w;
    w.setWindowFlag(Qt::WindowType::WindowMinMaxButtonsHint, false);
    w.setWindowFlags(w.windowFlags() | (Qt::WindowStaysOnTopHint | Qt::Tool | Qt::FramelessWindowHint));
    w.setWindowTitle("CrossShare");
    w.show();
    qInfo() << qApp->applicationFilePath().toUtf8().constData();

    {
        int windowCount = static_cast<int>(QRandomGenerator::global()->generate() % 5);
        int delta_x = 30 * (windowCount - 1);
        int delta_y = 30 * (windowCount - 1);
        w.move(delta_x + w.x(), delta_y + w.y());
    }

    return a.exec();
}
