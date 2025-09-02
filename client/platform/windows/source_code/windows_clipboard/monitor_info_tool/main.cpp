#include "mainwindow.h"
#include <QApplication>
#include <QMessageBox>
#include "common_utils.h"
#include "common_proxy_style.h"
#include "common_ui_process.h"
#include "process_message.h"
#include "windows_event_monitor.h"
#include "worker_thread.h"

int main(int argc, char *argv[])
{
    CommonUtils::supportForHighDPIDisplays();
    qInstallMessageHandler(CommonUtils::commonMessageOutput);
    QApplication a(argc, argv);
    a.setStyleSheet(CommonUtils::getFileContent(":/resource/my.qss"));
    a.setStyle(new CustomProxyStyle);

    {
        g_globalRegister();
        CommonUiProcess::getInstance();
        ProcessMessage::getInstance();
    }

    {
        WorkerThread *pThread = new WorkerThread;
        pThread->runInThread([] {
            WindowsEventMonitor::getInstance();
        });
    }

    MainWindow w;
    w.setWindowTitle("monitor_info_tool");
    w.show();
    qInfo() << qApp->applicationFilePath().toUtf8().constData();
    return a.exec();
}
