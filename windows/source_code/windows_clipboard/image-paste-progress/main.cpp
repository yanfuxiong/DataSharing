#include "mainwindow.h"
#include <QApplication>
#include <QProcess>
#include <QSharedMemory>
#include <QMessageBox>
#include <iostream>
#include "common_utils.h"
#include "common_proxy_style.h"
#include "common_ui_process.h"
#include "worker_thread.h"
#include "process_message.h"

QByteArray g_hashIdValue{};

int main(int argc, char *argv[])
{
    if (argc < 3) {
        std::cerr << argv[0] << "------------missing parameters hashId" << std::endl;
        return 1;
    }
    g_hashIdValue = QByteArray::fromHex(argv[1]);

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
        srand(time(nullptr));
        int windowCount = QString(argv[2]).toInt();
        //w.setWindowTitle(QString("---CrossShare %1").arg(windowCount));
        if (windowCount > 1) {
            windowCount += qrand() % 5;
            int delta_x = 30 * (windowCount - 1);
            int delta_y = 30 * (windowCount - 1);
            w.move(delta_x + w.x(), delta_y + w.y());
        }
    }

    return a.exec();
}
