#include "common_proxy_style.h"
#include <QLabel>
#include <QComboBox>
#include <QDialog>
#include <QTableWidget>
#include <QHeaderView>
#include <QMenu>
#include <QPainter>
#include <QPen>
#include <QApplication>
#include <QScreen>
#include <QPushButton>
#include <QHBoxLayout>
#include <QVBoxLayout>
#include <QGraphicsDropShadowEffect>
#include <windows.h>

QMainWindow *g_mainWindow = nullptr;

int g_getMenuItemHeight()
{
    return CUSTOM_MENU_ITEM_HEIGHT;
}

void CustomProxyStyle::polish(QWidget *widget)
{
    QProxyStyle::polish(widget);

    do {
        if (auto label = qobject_cast<QLabel*>(widget)) {
            if (label->property(PR_TEXT_BROWSER_INTERACTION_DISABLE).toBool() == false) {
                label->setTextInteractionFlags(Qt::TextInteractionFlag::TextBrowserInteraction);
            }
            break;
        }

        if (auto comboBox = qobject_cast<QComboBox*>(widget)) {
            comboBox->setView(createListView());
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
            tableView->setAlternatingRowColors(false);
            tableView->setContextMenuPolicy(Qt::ContextMenuPolicy::CustomContextMenu);
            tableView->setSelectionBehavior(QTableWidget::SelectionBehavior::SelectRows);
            tableView->horizontalHeader()->setStretchLastSection(true);
            tableView->horizontalHeader()->setDefaultAlignment(Qt::AlignVCenter | Qt::AlignLeft);
            tableView->horizontalHeader()->setSectionResizeMode(QHeaderView::ResizeMode::Interactive);
            tableView->verticalHeader()->setVisible(false);
            tableView->setSizePolicy(QSizePolicy::Expanding, QSizePolicy::Expanding);
            break;
        }

        if (auto menu = qobject_cast<QMenu*>(widget)) {
            menu->setWindowFlag(Qt::FramelessWindowHint, true);
            menu->setWindowFlag(Qt::NoDropShadowWindowHint, true);
            menu->setAttribute(Qt::WA_TranslucentBackground, true);
            auto func = [menu] {
                QGraphicsDropShadowEffect *effect = new QGraphicsDropShadowEffect(menu);
                effect->setBlurRadius(8);
                effect->setOffset(2, 2);
                effect->setColor(QColor(0, 0, 0, 40));
                menu->setGraphicsEffect(effect);
            };

            if (menu->property("IsSubMenu").toBool() == false) {
                func();
            } else {
                Q_ASSERT(g_mainWindow);
            }
            break;
        }
    } while (false);
}

QSize CustomProxyStyle::sizeFromContents(ContentsType type, const QStyleOption* option, const QSize& size, const QWidget* widget) const
{
    if (type == CT_MenuItem) {
        if (auto menu = qobject_cast<const QMenu*>(widget)) {
            if (menu->property("SystemInfoMenuManager").toBool()) {
                QSize newSize = QProxyStyle::sizeFromContents(type, option, size, widget);
                newSize.setHeight(g_getMenuItemHeight());
                return newSize;
            }

            if (menu->property("FileExplorerMenu").toBool()) {
                QSize newSize = QProxyStyle::sizeFromContents(type, option, size, widget);
                newSize.setHeight(35);
                return newSize;
            }
        }
    }
    return QProxyStyle::sizeFromContents(type, option, size, widget);
}

int CustomProxyStyle::pixelMetric(PixelMetric metric, const QStyleOption *option, const QWidget *widget) const
{
    if (auto menu = qobject_cast<const QMenu*>(widget)) {
        if (menu->property("FileExplorerMenu").toBool()) {
            if (metric == PM_SmallIconSize) {
                return 31;
            }
        }
    }
    return QProxyStyle::pixelMetric(metric, option, widget);
}

void CustomProxyStyle::drawControl(ControlElement element, const QStyleOption *option, QPainter *painter, const QWidget *widget) const
{
    if (element == CE_HeaderLabel) {
        if (const QStyleOptionHeader *headerOption = qstyleoption_cast<const QStyleOptionHeader*>(option)) {
            QStyleOptionHeader modifiedOption = *headerOption;
            modifiedOption.state &= ~State_On;
            QProxyStyle::drawControl(element, &modifiedOption, painter, widget);
            return;
        }
    }

    if (element == CE_HeaderSection) {
        if (const QStyleOptionHeader *headerOption = qstyleoption_cast<const QStyleOptionHeader*>(option)) {
            painter->save();
            painter->setRenderHint(QPainter::Antialiasing, true);
            if (headerOption->state & State_MouseOver) {
                painter->fillRect(headerOption->rect, QColor("#C0C0C0"));
            } else {
                painter->fillRect(headerOption->rect, QColor("#F0F0F0"));
                QPen pen(Qt::lightGray, 1, Qt::SolidLine);
                painter->setPen(pen);
                painter->drawLine(headerOption->rect.topRight(), headerOption->rect.bottomRight());
            }
            painter->restore();
            return;
        }
    }
    QProxyStyle::drawControl(element, option, painter, widget);
}

void CustomProxyStyle::drawPrimitive(PrimitiveElement element, const QStyleOption *option, QPainter *painter, const QWidget *widget) const
{
    if (element == PE_FrameMenu) {
        if (widget && widget->property("createClientInfoListMenu").toBool()) {
            painter->save();
            painter->setRenderHint(QPainter::Antialiasing, true);
            painter->setBrush(Qt::NoBrush);
            // CUSTOM_UI
            bool isRogTheme = widget->property("g_is_ROG_Theme").toBool();
            QPen pen(isRogTheme ? QColor("#E41A2B") : QColor(Qt::lightGray), 2, Qt::SolidLine);
            painter->setPen(pen);
            painter->drawRoundedRect(option->rect.adjusted(1, 1, -1, -1), 6, 6);
            painter->restore();
            return;
        }
    }
    QProxyStyle::drawPrimitive(element, option, painter, widget);
}

int CustomProxyStyle::styleHint(StyleHint hint, const QStyleOption *option, const QWidget *widget, QStyleHintReturn *returnData) const
{
    return QProxyStyle::styleHint(hint, option, widget, returnData);
}

QListView *CustomProxyStyle::createListView() const
{
    auto ptr = new QListView;
    QFont font;
    font.setPixelSize(16);
    ptr->setFont(font);
    return ptr;
}
