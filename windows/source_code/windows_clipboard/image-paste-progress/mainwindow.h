#ifndef MAINWINDOW_H
#define MAINWINDOW_H

#include <QMainWindow>
#include <QPointer>
#include <QFileDialog>
#include <QTimer>
#include <QCloseEvent>
#include <QMouseEvent>
#include <QSystemTrayIcon>
#include "common_utils.h"
#include "common_signals.h"
#include "progress_bar_widget.h"
#include "window_move_handler.h"

QT_BEGIN_NAMESPACE
namespace Ui { class MainWindow; }
QT_END_NAMESPACE

extern uint64_t g_timeStamp;

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
    std::unique_ptr<WindowMoveHandler> m_moveHandler;
    QPointer<ProgressBarWidget> m_progressBarWidget;
};
#endif // MAINWINDOW_H
