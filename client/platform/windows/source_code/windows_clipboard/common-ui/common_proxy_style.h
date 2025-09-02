#pragma once
#include "common_utils.h"
#include <QTableWidget>
#include <QProxyStyle>
#include <QComboBox>
#include <QListView>
#include <QFont>
#define CUSTOM_MENU_ITEM_HEIGHT 35

int g_getMenuItemHeight();

class CustomProxyStyle : public QProxyStyle
{
public:
    CustomProxyStyle() = default;

    void polish(QWidget *widget) override;
    QSize sizeFromContents(ContentsType type, const QStyleOption* option,
                           const QSize& size, const QWidget* widget) const override;

private:
    QListView *createListView() const;
};
