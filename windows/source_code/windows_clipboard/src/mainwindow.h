#ifndef MAINWINDOW_H
#define MAINWINDOW_H

#include <QMainWindow>
#include <QPointer>
#include <QFileDialog>
#include <QTimer>
#include <QCloseEvent>
#include "common_utils.h"
#include "device_list_dialog.h"
#include "common_signals.h"
#include "progress_bar_dialog.h"
#include "accept_file_dialog.h"
#include "file_opt_info_list.h"

QT_BEGIN_NAMESPACE
namespace Ui { class MainWindow; }
QT_END_NAMESPACE

class MainWindow : public QMainWindow
{
    Q_OBJECT

public:
    MainWindow(QWidget *parent = nullptr);
    ~MainWindow();

    void startTestTimer();

private slots:
    void onLogMessage(const QString &message);

    void onDispatchMessage(const QVariant &data);
    void onSystemConfigChanged();
    void on_settings_btn_clicked();
    void on_select_file_clicked();
    void onUpdateClientList();

private:
    void closeEvent(QCloseEvent *event) override;
    void processTopTitleLeftClicked();

private:
    Ui::MainWindow *ui;
    QPointer<QTimer> m_testTimer;
    int m_currentProgressVal;
    //QPointer<ProgressBarDialog> m_progressDialog;
    //QPointer<AcceptFileDialog> m_acceptFileDialog;
};
#endif // MAINWINDOW_H
