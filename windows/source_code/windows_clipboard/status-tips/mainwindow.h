#ifndef MAINWINDOW_H
#define MAINWINDOW_H

#include <QMainWindow>
#include <QPointer>
#include <QFileDialog>
#include <QTimer>
#include <QCloseEvent>
#include <QSystemTrayIcon>
#include "common_utils.h"
#include "common_signals.h"

QT_BEGIN_NAMESPACE
namespace Ui { class MainWindow; }
QT_END_NAMESPACE

extern NotifyMessage g_notifyMessage;

class MainWindow : public QMainWindow
{
    Q_OBJECT

public:
    MainWindow(QWidget *parent = nullptr);
    ~MainWindow();

private slots:
    void onLogMessage(const QString &message);
    void onDispatchMessage(const QVariant &data);

private:
    void closeEvent(QCloseEvent *event) override;

private:
    Ui::MainWindow *ui;
    QPointer<QTimer> m_testTimer;
};
#endif // MAINWINDOW_H
