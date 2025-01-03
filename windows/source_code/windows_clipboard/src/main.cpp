#include "mainwindow.h"
#include <QApplication>
#include <QProcess>
#include "common_utils.h"
#include "common_proxy_style.h"
#include "common_ui_process.h"
#include "worker_thread.h"
#include "process_message.h"

int main(int argc, char *argv[])
{
    qInstallMessageHandler(CommonUtils::commonMessageOutput);
    //QCoreApplication::setAttribute(Qt::AA_EnableHighDpiScaling);
    //QCoreApplication::setAttribute(Qt::AA_DisableHighDpiScaling, true);
    QApplication a(argc, argv);
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
        }
    }

    MainWindow w;
    w.setWindowTitle("CrossShare Client");
    // FIXME:
    //w.resize(1185, 517);
    w.show();
    qInfo() << qApp->applicationFilePath().toUtf8().constData();

    for (const auto &path : g_getPipeServerExePathList()) {
        qInfo() << path;
    }

    // FIXME:
    CommonUtils::setAutoRun(false);

// FIXME:
#if STABLE_VERSION_CONTROL > 0
    QPointer<WorkerThread> thread(new WorkerThread);
    thread->runInThread([] {
        QTimer *pTimer = new QTimer;
        pTimer->setTimerType(Qt::TimerType::PreciseTimer);
        pTimer->setInterval(1000);
        auto func = [] {
            bool exists = false;
            for (const auto &serverExePath : g_getPipeServerExePathList()) {
                if (CommonUtils::processIsRunning(serverExePath)) {
                    //qInfo() << "process is running:" << serverExePath;
                    return;
                }
            }
            for (const auto &serverExePath : g_getPipeServerExePathList()) {
                if (QFile::exists(serverExePath)) {
                    exists = true;
                    QProcess process;
                    process.startDetached(serverExePath);
                    break;
                }
            }
//            if (exists == false) {
//                Q_EMIT CommonSignals::getInstance()->showWarningMessageBox("warning",
//                                                                           QString("startup failed [%1]").arg(g_getPipeServerExePathList().front()));
//            }
        };

        QObject::connect(pTimer, &QTimer::timeout, pTimer, func);
        QTimer::singleShot(600, QThread::currentThread(), func);
        pTimer->start();
    });
#endif

    return a.exec();
}
