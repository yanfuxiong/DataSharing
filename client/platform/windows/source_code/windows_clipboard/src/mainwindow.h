#ifndef MAINWINDOW_H
#define MAINWINDOW_H

#include <QMainWindow>
#include <QPointer>
#include <QFileDialog>
#include <QTimer>
#include <QCloseEvent>

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
    void onSystemConfigChanged();
    void on_select_file_clicked();
    void onUpdateCurrentMenPosData();
    void onShowOpenSourceLicenseMenu();
    void onModifyDefaulutDownloadPath();
    void onBugReport();

private:
    void closeEvent(QCloseEvent *event) override;
    void moveEvent(QMoveEvent *event) override;
    void resizeEvent(QResizeEvent *event) override;
    void showEvent(QShowEvent *event) override;
    void changeEvent(QEvent *event) override;
    void clearAllUserOptRecord();

private:
    Ui::MainWindow *ui;
    QPointer<QTimer> m_testTimer;
    int m_currentProgressVal;
    Qt::WindowStates m_lastNormalState;
};
#endif // MAINWINDOW_H
