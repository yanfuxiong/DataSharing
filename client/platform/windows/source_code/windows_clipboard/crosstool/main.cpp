#include <QCoreApplication>
#include <QProcess>
#include <QSharedMemory>
#include <QMessageBox>
#include "common_utils.h"
#include "common_proxy_style.h"
#include "common_ui_process.h"
#include "hidden_message_window.h"
#include "boost_log_setup.h"
#include <windows.h>
#include <QDir>

int main(int argc, char *argv[])
{
    CommonUtils::supportForHighDPIDisplays();
    QCoreApplication a(argc, argv);
    CommonUtils::ensureDirectoryExists();
    // Initialize logging operations after QApplication to prevent path-related information retrieval exceptions.
    qInstallMessageHandler(g_commonMessageOutput);
    g_boostLogSetup("crosstool");
    g_setBoostLogLevel(QtSeverity::QLOG_INFO);
    {
        static QSharedMemory s_sharedMemeory { "__crosstool__" };
        if (s_sharedMemeory.create(1) == false) {
            qApp->quit();
            qWarning() << "The process has started !!!";
            return 1;
        }
    }
    SetCurrentDirectoryW(QDir::toNativeSeparators(qApp->applicationDirPath()).toStdWString().c_str());

    {
        g_globalRegister();
        g_loadLocalConfig();
        qInfo() << g_getGlobalData()->localConfig.dump(4).c_str();
        qInfo() << qApp->applicationFilePath().toUtf8().constData() << "; version:" << qApp->applicationVersion();
    }

    {
        QTimer *pTimer = new QTimer;
        pTimer->setTimerType(Qt::TimerType::PreciseTimer);
        pTimer->setInterval(2000);
        QObject::connect(pTimer, &QTimer::timeout, qApp, [] {
            QString crossShareServOldPath = CommonUtils::getValueByRegKeyName(CROSS_SHARE_SERV_NAME);
            qDebug() << "crossShareServ old path:" << QDir::toNativeSeparators(crossShareServOldPath).toUtf8().constData();
            QString currentPath = qApp->applicationDirPath();
            QString crossShareServPath = currentPath + "/" + CROSS_SHARE_SERV_NAME;
            if (CommonUtils::crossShareServerIsRunning() == false && QFile::exists(crossShareServPath)) {
                CommonUtils::startDetachedWithoutInheritance(crossShareServPath, {});
                qInfo() << "---------------startup cross_share_serv.exe !!!";
            }
        });
        pTimer->start();
    }

    std::thread([] {
        HiddenMessageWindow g_msgWindow;
        if (!g_msgWindow.create()) {
            qCritical() << "Failed to create hidden message window";
            _Exit(1);
        }
        MSG msg;
        while (true) {
            while (GetMessage(&msg, nullptr, 0, 0)) {
                TranslateMessage(&msg);
                DispatchMessage(&msg);
            }
        }
    }).detach();

    return a.exec();
}
