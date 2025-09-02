#ifndef LOADING_DIALOG_H
#define LOADING_DIALOG_H

#include <QDialog>
#include <QPainter>
#include <QPaintEvent>
#include <QImageReader>
#include <QMovie>
#include <QPointer>
#include <QMouseEvent>
#include <QElapsedTimer>
#include <QTimer>
#include "window_move_handler.h"

namespace Ui {
class LoadingDialog;
}

class LoadingDialog : public QDialog
{
    Q_OBJECT

public:
    explicit LoadingDialog(QWidget *parent = nullptr);
    ~LoadingDialog();

    QPixmap createPixmap(const QString &imgPath, int size);
    uint32_t getThemeCode() const { return m_themeCode; }

    static void updateThemeCode(uint32_t themeCode);

private:
    void paintEvent(QPaintEvent *event) override;
    void mouseMoveEvent(QMouseEvent *event) override;
    void mousePressEvent(QMouseEvent *event) override;

private Q_SLOTS:
    void onFrameChanged(int frameNumber);
    void onClickedCloseButton();
    void onDIASStatusChanged(bool status);
    void onGetThemeCode();

private:
    Ui::LoadingDialog *ui;
    WindowMoveHandler m_moveHandler;
    QPointer<QMovie> m_movie;
    QRectF m_closeBtnRect;
    QRectF m_textRect;
    bool m_closeBtnStatus { false };
    QString m_crossShareTextImagePath;
    bool m_showLoadingGIF { false };
    bool m_showTooltipText { false };
    int m_timerThreshold { 3000 };
    QPointer<QTimer> m_timer;
    QElapsedTimer m_elapsedTimer;
    uint32_t m_themeCode { 0 };
};

#endif // LOADING_DIALOG_H
