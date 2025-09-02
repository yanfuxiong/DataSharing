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
#include "boost_log_setup.h"
#include <windows.h>

int main(int argc, char *argv[])
{
    CommonUtils::supportForHighDPIDisplays();
    QApplication a(argc, argv);
    CommonUtils::ensureDirectoryExists();
    // Initialize logging operations after QApplication to prevent path-related information retrieval exceptions.
    qInstallMessageHandler(g_commonMessageOutput);
    g_boostLogSetup("cross_share_serv");
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

    {
        g_globalRegister();
        g_loadLocalConfig();
        qInfo() << g_getGlobalData()->localConfig.dump(4).c_str();
        qInfo() << qApp->applicationFilePath().toUtf8().constData() << "; version:" << qApp->applicationVersion();
        {
            bool isInited = false;
            uint32_t customerID = g_getCustomerIDForUITheme(isInited);
            Q_UNUSED(customerID)
            if (isInited == false) {
                qWarning() << "Because UITheme.isInited is false, the program exits.";
                return 1;
            }
        }
        CommonUiProcess::getInstance();

        QPointer<HelperServer> helperServer = new HelperServer;
        helperServer->startServer(g_helperServerName);
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
    CommonUtils::setAutoRun(true);

    return a.exec();
}
