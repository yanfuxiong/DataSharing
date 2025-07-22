#include "mainwindow.h"
#include <iostream>
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

NotifyMessage g_notifyMessage{};

int main(int argc, char *argv[])
{
    if (argc <= 2) {
        std::cerr << "----------------param error !!!" << std::endl;
        return 1;
    }
    NotifyMessage::fromByteArray(QByteArray::fromHex(argv[1]), g_notifyMessage);
    const int processIndex = QString(argv[2]).toInt();

    qInstallMessageHandler(CommonUtils::commonMessageOutput);
    QApplication a(argc, argv);

    a.setStyleSheet(CommonUtils::getFileContent(":/resource/my.qss"));
    a.setStyle(new CustomProxyStyle);

    {
        g_globalRegister();
        g_namedPipeServerName = g_helperServerName; // The connected server has changed
        CommonUiProcess::getInstance();
        //ProcessMessage::getInstance();
    }

    MainWindow w;
    w.setWindowFlags(w.windowFlags() | (Qt::WindowStaysOnTopHint | Qt::FramelessWindowHint | Qt::Tool));
    w.setWindowTitle("StatusTips");
    w.setProcessIndex(processIndex);
    w.show();
    qInfo() << qApp->applicationFilePath().toUtf8().constData();
    w.updateWindowPos();

    return a.exec();
}
