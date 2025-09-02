#ifndef MAINWINDOW_H
#define MAINWINDOW_H

#include <QMainWindow>

namespace Ui {
class MainWindow;
}

class MainWindow : public QMainWindow
{
    Q_OBJECT

public:
    explicit MainWindow(QWidget *parent = nullptr);
    ~MainWindow();

private slots:
    void on_get_info_btn_clicked();
    void on_clear_btn_clicked();

private:
    Ui::MainWindow *ui;
    qint64 m_cacheIndex { 0 };
};

#endif // MAINWINDOW_H
