#include <QApplication>
#include <QSharedMemory>
#include "common_utils.h"
#include "common_proxy_style.h"
#include "common_ui_process.h"
#include "mainwindow.h"

int main(int argc, char *argv[])
{
    qInstallMessageHandler(CommonUtils::commonMessageOutput);
    QCoreApplication::setAttribute(Qt::AA_EnableHighDpiScaling);
    QApplication a(argc, argv);
    {
        static QSharedMemory s_sharedMemeory {"__cross_share_test_server__"};
        if (s_sharedMemeory.create(1) == false) {
            qApp->quit();
            qWarning() << "test-server已启动......";
            return 1;
        }
    }
    a.setStyleSheet(CommonUtils::getFileContent(":/resource/my.qss"));
    a.setStyle(new CustomProxyStyle);

    {
        g_globalRegister();
        CommonUiProcess::getInstance();
    }

    MainWindow mainWindow;
    mainWindow.setWindowTitle("测试namedpipe server");
    mainWindow.show();

    return a.exec();
}
