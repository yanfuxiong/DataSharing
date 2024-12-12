#include "common_proxy_style.h"
#include <QLabel>
#include <QComboBox>
#include <QDialog>
#include <QTableWidget>
#include <QHeaderView>
#include <QPushButton>
#include <QHBoxLayout>
#include <QVBoxLayout>

void CustomProxyStyle::polish(QWidget *widget)
{
    QProxyStyle::polish(widget);

    do {
        if (auto label = qobject_cast<QLabel*>(widget)) {
            label->setTextInteractionFlags(Qt::TextInteractionFlag::TextBrowserInteraction);
            break;
        }

        if (auto comboBox = qobject_cast<QComboBox*>(widget)) {
            comboBox->setView(createListView()); // 配合下拉框的界面美化
            break;
        }

        if (auto dialog = qobject_cast<QDialog*>(widget)) {
            dialog->setWindowFlag(Qt::WindowType::WindowContextHelpButtonHint, false);
            break;
        }

        if (auto tableWidget = qobject_cast<QTableWidget*>(widget)) {
            tableWidget->setAlternatingRowColors(true);
            tableWidget->setContextMenuPolicy(Qt::ContextMenuPolicy::CustomContextMenu);
            tableWidget->setSelectionBehavior(QTableWidget::SelectionBehavior::SelectRows);
            tableWidget->horizontalHeader()->setStretchLastSection(true);
            tableWidget->horizontalHeader()->setDefaultAlignment(Qt::AlignVCenter | Qt::AlignLeft);
            //m_tableWidget->horizontalHeader()->setSectionsMovable(false);
            tableWidget->horizontalHeader()->setSectionResizeMode(QHeaderView::ResizeMode::Fixed);
            tableWidget->verticalHeader()->setVisible(false);
            tableWidget->setSizePolicy(QSizePolicy::Expanding, QSizePolicy::Expanding);
            break;
        }

        if (auto tableView = qobject_cast<QTableView*>(widget)) {
            tableView->setAlternatingRowColors(true);
            tableView->setContextMenuPolicy(Qt::ContextMenuPolicy::CustomContextMenu);
            tableView->setSelectionBehavior(QTableWidget::SelectionBehavior::SelectRows);
            tableView->horizontalHeader()->setStretchLastSection(true);
            tableView->horizontalHeader()->setDefaultAlignment(Qt::AlignVCenter | Qt::AlignLeft);
            tableView->horizontalHeader()->setSectionResizeMode(QHeaderView::ResizeMode::Fixed);
            tableView->verticalHeader()->setVisible(false);
            tableView->setSizePolicy(QSizePolicy::Expanding, QSizePolicy::Expanding);

            tableView->verticalHeader()->setDefaultSectionSize(36);
            tableView->horizontalHeader()->setDefaultSectionSize(120);
            break;
        }
    } while (false);
}

QListView *CustomProxyStyle::createListView() const
{
    auto ptr = new QListView;
    QFont font;
    font.setPixelSize(16);
    ptr->setFont(font);
    return ptr;
}
