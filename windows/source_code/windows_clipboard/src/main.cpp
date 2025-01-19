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
        CommonUiProcess::getInstance();
        ProcessMessage::getInstance();
    }

    MainWindow w;
    w.setWindowTitle("CrossShare Client");
    // FIXME: 特殊处理一下:
    //w.resize(1185, 517);
    w.show();
    qInfo() << qApp->applicationFilePath().toUtf8().constData();

    for (const auto &path : g_getPipeServerExePathList()) {
        qInfo() << path;
    }

    // FIXME: 开机自启动设置, 目前屏蔽
    CommonUtils::setAutoRun(false);

// FIXME: 暂时屏蔽
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
