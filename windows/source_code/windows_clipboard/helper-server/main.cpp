#include "mainwindow.h"
#include <QApplication>
#include <QProcess>
#include <QSharedMemory>
#include <QMessageBox>
#include "common_utils.h"
#include "common_proxy_style.h"
#include "common_ui_process.h"
#include "worker_thread.h"
#include "process_message.h"
#include "namedpipe-server.h"

int main(int argc, char *argv[])
{
    qInstallMessageHandler(CommonUtils::commonMessageOutput);
    QApplication a(argc, argv);
    {
        static QSharedMemory s_sharedMemeory { "__cross_share_helper_server__" };
        if (s_sharedMemeory.create(1) == false) {
            qApp->quit();
            qWarning() << "The process has started, please check the tray in the bottom right corner......";
            return 1;
        }
    }

    a.setStyleSheet(CommonUtils::getFileContent(":/resource/my.qss"));
    a.setStyle(new CustomProxyStyle);

    {
        g_globalRegister();
        CommonUiProcess::getInstance();
        ProcessMessage::getInstance();

        QPointer<HelperServer> helperServer = new HelperServer;
        helperServer->startServer(g_helperServerName);
    }

    MainWindow w;
    w.setWindowTitle("helper server");
#ifndef NDEBUG
    w.show();
#endif
    qInfo() << qApp->applicationFilePath().toUtf8().constData();

    for (const auto &path : g_getPipeServerExePathList()) {
        qInfo() << path;
    }

    CommonUtils::setAutoRun(true);

    return a.exec();
}
