#include "mainwindow.h"
#include <QApplication>
#include <iostream>
#include <QProcess>
#include <QMessageBox>
#include "common_utils.h"
#include "common_proxy_style.h"
#include "common_ui_process.h"
#include "process_message.h"
#include <boost/interprocess/windows_shared_memory.hpp>
#define CROSS_SHARE_SERV_NAME "cross_share_serv.exe"

using namespace boost::interprocess;

int main(int argc, char *argv[])
{
    qInstallMessageHandler(CommonUtils::commonMessageOutput);
    //QCoreApplication::setAttribute(Qt::AA_EnableHighDpiScaling);
    //QCoreApplication::setAttribute(Qt::AA_DisableHighDpiScaling, true);
    QApplication a(argc, argv);
    windows_shared_memory shm;
    try {
        shm = windows_shared_memory(create_only, "cross_share_instance", read_write, 1);
    } catch (const std::exception &e) {
        std::cerr << e.what() << std::endl;
        QMessageBox::warning(nullptr, "warning", "An instance is already running");
        return 1;
    }

    a.setStyleSheet(CommonUtils::getFileContent(":/resource/my.qss"));
    a.setStyle(new CustomProxyStyle);

    {
        g_globalRegister();
        // init DB
        {
            g_getGlobalData()->sqlite_db = QSqlDatabase::addDatabase("QSQLITE", SQLITE_CONN_NAME);
            g_getGlobalData()->sqlite_db.setDatabaseName(g_sqliteDbPath());
            g_getGlobalData()->sqlite_db.open();
            Q_ASSERT(g_getGlobalData()->sqlite_db.isOpen() == true);
        }
        CommonUiProcess::getInstance();
        ProcessMessage::getInstance();

        {
            //g_getGlobalData()->systemConfig.localIpAddress = CommonUtils::localIpAddress();
            g_getGlobalData()->systemConfig.clientVersionStr = VERSION_STR;
            g_loadLocalConfig();
            qInfo() << g_getGlobalData()->localConfig.dump(4).c_str();
        }
    }

    MainWindow w;
    w.setWindowTitle("CrossShare Client");
    // This window requires a minimum size; otherwise, it won't be able to display all the content properly.
    w.setMinimumWidth(1220);
    w.setMinimumHeight(825);
    w.show();
    qInfo() << qApp->applicationFilePath().toUtf8().constData();

    for (const auto &path : g_getPipeServerExePathList()) {
        qInfo() << path;
    }

    CommonUtils::setAutoRun(false);
    CommonUtils::removeAutoRunByName(PIPE_SERVER_EXE_NAME);

    {
        QString crossShareServOldPath = CommonUtils::getValueByRegKeyName(CROSS_SHARE_SERV_NAME);
        qInfo() << "crossShareServ old path:" << QDir::toNativeSeparators(crossShareServOldPath).toUtf8().constData();
        QString currentPath = qApp->applicationDirPath();
        QString crossShareServPath = currentPath + "/" + CROSS_SHARE_SERV_NAME;
        if (CommonUtils::processIsRunning(crossShareServPath) == false && QFile::exists(crossShareServPath)) {
            QProcess process;
            process.setProcessChannelMode(QProcess::ProcessChannelMode::MergedChannels);
            process.startDetached(crossShareServPath, {});
        }
    }

    qApp->setStartDragDistance(1);

    return a.exec();
}
