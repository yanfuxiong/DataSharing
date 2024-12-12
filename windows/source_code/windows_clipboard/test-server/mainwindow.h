#ifndef MAINWINDOW_H
#define MAINWINDOW_H

#include <QMainWindow>
#include "server.h"
#include "common_utils.h"
#include "common_signals.h"

namespace Ui {
class MainWindow;
}

class MainWindow : public QMainWindow
{
    Q_OBJECT

public:
    explicit MainWindow(QWidget *parent = nullptr);
    ~MainWindow();

    // 这里只是模拟
    void sendProgressData();

private slots:
    void on_add_client_clicked();
    void onLogMessage(const QString &message);
    void onRecvClientData(const QByteArray &data);

    void on_send_file_clicked();

    void on_disconnect_client_clicked();

private:
    Ui::MainWindow *ui;
    NamedPipeServer m_server;
    Buffer m_buffer;
};

#endif // MAINWINDOW_H
