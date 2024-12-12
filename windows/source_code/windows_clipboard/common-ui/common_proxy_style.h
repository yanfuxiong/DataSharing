#pragma once
#include "common_utils.h"
#include <QTableWidget>
#include <QProxyStyle>
#include <QComboBox>
#include <QListView>
#include <QFont>

class CustomProxyStyle : public QProxyStyle
{
public:
    CustomProxyStyle() = default;

    void polish(QWidget *widget) override;

private:
    QListView *createListView() const;
};
