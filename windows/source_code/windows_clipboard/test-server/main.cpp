#include <QApplication>
#include "common_utils.h"
#include "common_proxy_style.h"
#include "common_ui_process.h"
#include "mainwindow.h"

int main(int argc, char *argv[])
{
    qInstallMessageHandler(CommonUtils::commonMessageOutput);
    QCoreApplication::setAttribute(Qt::AA_EnableHighDpiScaling);
    QApplication a(argc, argv);
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
