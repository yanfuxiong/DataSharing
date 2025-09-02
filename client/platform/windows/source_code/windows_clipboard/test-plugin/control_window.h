#ifndef CONTROL_WINDOW_H
#define CONTROL_WINDOW_H

#include <QMainWindow>
#include "worker_thread.h"

namespace Ui {
class ControlWindow;
}

class ControlWindow : public QMainWindow
{
    Q_OBJECT

public:
    explicit ControlWindow(QWidget *parent = nullptr);
    ~ControlWindow();
    void appendLogInfo(const QString &logString);
    void on_SetCancelFileTransfer(const char* ipPort, const char* clientID, uint64_t timeStamp);

private slots:
    void on_UpdateClientStatus_btn_clicked();
    void on_UpdateSystemInfo_btn_clicked();
    void on_GetDownloadPath_btn_clicked();
    void on_GetDeviceName_btn_clicked();
    void on_StartClipboardMonitor_btn_clicked();
    void on_StopClipboardMonitor_btn_clicked();
    void on_AuthViaIndex_btn_clicked();
    void on_RequestSourceAndPort_btn_clicked();
    void on_DIASStatus_clicked();
    void on_DIAS_Connected_clicked();
    void on_CleanClipboard_btn_clicked();
    void on_NotiMessage_btn_clicked();
    void on_DragFileListNotify_btn_clicked();
    void on_MultiFilesDropNotify_btn_clicked();
    void on_UpdateMultipleProgress_btn_clicked();
    void on_UpdateClientVersion_btn_clicked();
    void on_NotifyErrorEvent_btn_clicked();

private:
    Ui::ControlWindow *ui;
    WorkerThread m_workerThread;
};

#endif // CONTROL_WINDOW_H
