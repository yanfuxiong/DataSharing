#include "mainwindow.h"
#include <QApplication>
#include <QProcess>
#include <QSharedMemory>
#include <QMessageBox>
#include "common_utils.h"
#include "common_proxy_style.h"
#include "common_ui_process.h"
#include "namedpipe_server.h"
#include "load_plugin.h"
#include <windows.h>

int main(int argc, char *argv[])
{
    qInstallMessageHandler(CommonUtils::commonMessageOutput);
    QApplication a(argc, argv);
    {
        static QSharedMemory s_sharedMemeory { "__hyperdrop_helper_server__" };
        if (s_sharedMemeory.create(1) == false) {
            qApp->quit();
            qWarning() << "The process has started, please check the tray in the bottom right corner......";
            return 1;
        }
    }

    SetCurrentDirectoryW(QDir::toNativeSeparators(qApp->applicationDirPath()).toStdWString().c_str());
    a.setStyleSheet(CommonUtils::getFileContent(":/resource/my.qss"));
    a.setStyle(new CustomProxyStyle);

    if (0)
    {
        g_logFile.reset(new QFile(qApp->applicationDirPath() + "/cross_share_serv.log"));
        if (g_logFile->open(QFile::WriteOnly | QFile::Append) == false) {
            g_logFile.reset();
        }
    }

    {
        g_globalRegister();
        g_loadLocalConfig();
        qInfo() << g_getGlobalData()->localConfig.dump(4).c_str();
        CommonUiProcess::getInstance();

        QPointer<HelperServer> helperServer = new HelperServer;
        helperServer->startServer(g_helperServerName);

        QTimer::singleShot(0, qApp, [] {
            LoadPlugin::getInstance()->initPlugin();
        });
    }

    QTimer timer;
    timer.setInterval(1000);
    QObject::connect(&timer, &QTimer::timeout, qApp, [] {
        QCoreApplication::processEvents();
    });
    timer.start();

    MainWindow w;
    w.setWindowTitle("helper server");
    w.hide();
    qInfo() << qApp->applicationFilePath().toUtf8().constData();
    CommonUtils::setAutoRun(true);

    return a.exec();
}
