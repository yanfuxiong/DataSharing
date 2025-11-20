#pragma once
#include "common_utils.h"
#include <QTableWidget>
#include <QProxyStyle>
#include <QComboBox>
#include <QListView>
#include <QFont>
#include <QPainter>
#include <QMainWindow>
#define CUSTOM_MENU_ITEM_HEIGHT 35

int g_getMenuItemHeight();
extern QMainWindow *g_mainWindow;

class CustomProxyStyle : public QProxyStyle
{
public:
    CustomProxyStyle() = default;

    void polish(QWidget *widget) override;
    QSize sizeFromContents(ContentsType type, const QStyleOption* option,
                           const QSize& size, const QWidget* widget) const override;
    int pixelMetric(PixelMetric metric, const QStyleOption *option, const QWidget *widget) const override;
    void drawControl(ControlElement element, const QStyleOption *option, QPainter *painter, const QWidget *widget = nullptr) const override;
    void drawPrimitive(PrimitiveElement element, const QStyleOption *option, QPainter *painter, const QWidget *widget) const override;
    int styleHint(StyleHint hint, const QStyleOption *option, const QWidget *widget, QStyleHintReturn *returnData) const override;

private:
    QListView *createListView() const;
};
