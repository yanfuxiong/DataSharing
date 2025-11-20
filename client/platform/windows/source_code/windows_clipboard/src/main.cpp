#include "mainwindow.h"
#include <QApplication>
#include <iostream>
#include <QProcess>
#include <QMessageBox>
#include <QScreen>
#include "common_utils.h"
#include "common_proxy_style.h"
#include "common_ui_process.h"
#include "loading_dialog.h"
#include "process_message.h"
#include <boost/interprocess/windows_shared_memory.hpp>
#include "boost_log_setup.h"
#include <QDesktopWidget>

using namespace boost::interprocess;

int main(int argc, char *argv[])
{
    CommonUtils::supportForHighDPIDisplays();
    //QCoreApplication::setAttribute(Qt::AA_EnableHighDpiScaling);
    //QCoreApplication::setAttribute(Qt::AA_DisableHighDpiScaling, true);
    QApplication a(argc, argv);
    CommonUtils::ensureDirectoryExists();
    // Initialize logging operations after QApplication to prevent path-related information retrieval exceptions.
    qInstallMessageHandler(g_commonMessageOutput);
    g_boostLogSetup("windows_clipboard");
#ifdef NDEBUG
    g_setBoostLogLevel(QtSeverity::QLOG_INFO);
#else
    g_setBoostLogLevel(QtSeverity::QLOG_DEBUG);
#endif
    windows_shared_memory shm;
    try {
        shm = windows_shared_memory(create_only, "cross_share_instance", read_write, 1);
    } catch (const std::exception &e) {
        std::cerr << e.what() << std::endl;
        QMessageBox::warning(nullptr, "warning", "An instance is already running");
        return 1;
    }

    QDir::setCurrent(qApp->applicationDirPath());
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
            g_getGlobalData()->systemConfig.clientVersionStr = qApp->applicationVersion();
            g_loadLocalConfig();
            qInfo() << g_getGlobalData()->localConfig.dump(4).c_str();
            qInfo() << qApp->applicationFilePath().toUtf8().constData() << "; version:" << qApp->applicationVersion();
        }
    }

    do {
        bool isValid = false;
        uint32_t customerID = g_getCustomerIDForUITheme(isValid);
        Q_UNUSED(customerID)
        if (isValid) {
            break;
        }
        CommonUtils::killServer();
        QMainWindow mainWindow;
        LoadingDialog dialog(&mainWindow);
        if (dialog.exec() == QDialog::Rejected) {
            return 1;
        } else {
            LoadingDialog::updateThemeCode(dialog.getThemeCode());
        }
    } while (false);

    if (g_is_ROG_Theme()) {
        a.setStyleSheet(CommonUtils::getFileContent(":/resource/rog/rog.qss"));
        a.setWindowIcon(QIcon(":/resource/application.ico"));
    } else {
        a.setStyleSheet(CommonUtils::getFileContent(":/resource/my.qss"));
        a.setWindowIcon(QIcon(":/resource/application.ico"));
    }
    a.setStyle(new CustomProxyStyle);

    MainWindow w;
    w.setWindowTitle("CrossShare Client");
    // This window requires a minimum size; otherwise, it won't be able to display all the content properly.
    w.setMinimumWidth(1100);
    // w.setMinimumHeight(825);
    w.show();

    for (const auto &path : g_getPipeServerExePathList()) {
        qInfo() << path;
    }

    CommonUtils::setAutoRun(false);
    CommonUtils::removeAutoRunByName(PIPE_SERVER_EXE_NAME);

    QTimer::singleShot(500, qApp, [] {
        {
            QString crossShareServOldPath = CommonUtils::getValueByRegKeyName(CROSS_SHARE_SERV_NAME);
            qInfo() << "crossShareServ old path:" << QDir::toNativeSeparators(crossShareServOldPath).toUtf8().constData();
            QString currentPath = qApp->applicationDirPath();
            QString crossShareServPath = currentPath + "/" + CROSS_SHARE_SERV_NAME;
            if (CommonUtils::crossShareServerIsRunning() == false && QFile::exists(crossShareServPath)) {
                CommonUtils::startDetachedWithoutInheritance(crossShareServPath, {});
            }
        }

        {
            std::thread([] {
                QThread::msleep(2000);
                QString crossShareExePath = qApp->applicationDirPath() + "/" + CROSS_SHARE_SERV_NAME;
                if (QFile::exists(crossShareExePath) == false) {
                    return;
                }
                while (true) {
                    if (CommonUtils::crossShareServerIsRunning()) {
                        QThread::msleep(1000); // 1000ms
                        continue;
                    }
                    qApp->exit();
                    break;
                }
            }).detach();
        }
    });

    qApp->setStartDragDistance(1);
    qInfo() << "screen size info:" << qApp->primaryScreen()->size();

    return a.exec();
}
