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
#include "worker_thread.h"

QT_BEGIN_NAMESPACE
namespace Ui { class MainWindow; }
QT_END_NAMESPACE

class MainWindow : public QMainWindow
{
    Q_OBJECT

public:
    MainWindow(QWidget *parent = nullptr);
    ~MainWindow();

private slots:
    void onLogMessage(const QString &message);

    void onDispatchMessage(const QVariant &data);
    void onSystemTrayIconActivated(QSystemTrayIcon::ActivationReason reason);

private:
    void closeEvent(QCloseEvent *event) override;
    void changeEvent(QEvent *event) override;
    QStringList statusTipsExePathList() const;
    int getNotifyMessageDuration() const;
    UpdateClientStatusMsgPtr getClientStatusMsgByClientID(const QByteArray &clientID) const;
    void processNotifyMessage(NotifyMessagePtr ptrMsg);

private:
    Ui::MainWindow *ui;
    QPointer<QTimer> m_testTimer;
    bool m_exitsStatus { false };
    QPointer<QSystemTrayIcon> m_systemTrayIcon;
    qint64 m_cacheIndex { 0 };
    int m_processIndex { 0 };
    int m_processCount { 0 };
    WorkerThread m_workerThread;
};
#endif // MAINWINDOW_H
