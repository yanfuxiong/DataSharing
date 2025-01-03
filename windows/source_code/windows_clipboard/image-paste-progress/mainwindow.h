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

QT_BEGIN_NAMESPACE
namespace Ui { class MainWindow; }
QT_END_NAMESPACE

extern QByteArray g_hashIdValue;

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

    void mousePressEvent(QMouseEvent *event) override;
    void mouseReleaseEvent(QMouseEvent *event) override;
    void mouseMoveEvent(QMouseEvent *event) override;

private:
    Ui::MainWindow *ui;
    QPointer<QTimer> m_testTimer;
    QPointer<ProgressBarWidget> m_progressBarWidget;

    QPoint m_clickedPos;
    bool m_clickedStatus { false };
};
#endif // MAINWINDOW_H
