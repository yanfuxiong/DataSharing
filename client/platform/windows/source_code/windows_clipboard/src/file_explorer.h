#ifndef FILE_EXPLORER_H
#define FILE_EXPLORER_H

#include <QWidget>
#include <QStandardItemModel>
#include <QFileSystemModel>
#include <QListView>
#include <QTreeView>
#include <QMimeData>
#include <QDrag>
#include <QUrl>
#include <QElapsedTimer>
#include <QTableView>
#include <QMenu>
#include <QLineEdit>
#include <QPushButton>
#include <QStackedWidget>
#include <QDropEvent>
#include <QMouseEvent>
#include <QContextMenuEvent>
#include <QRubberBand>
#include <QStyledItemDelegate>
#include <boost/bimap.hpp>
#include "common_utils.h"
#include "common_signals.h"

namespace Ui {
class FileExplorer;
}

class TableViewEx : public QTableView
{
    Q_OBJECT
public:
    explicit TableViewEx(QWidget *parent = nullptr);
    int hoverRow() const { return m_hoverRow; }

protected:
    void mousePressEvent(QMouseEvent *event) override;
    void mouseReleaseEvent(QMouseEvent *event) override;
    void mouseMoveEvent(QMouseEvent *event) override;
    void leaveEvent(QEvent *event) override;
    bool checkClickedPosForRubberBand(const QPoint &pos) const;

private:
    QModelIndex validIndexAt(const QPoint &pos) const;
    void updateSelection(const QPoint &currentPos, Qt::KeyboardModifiers modifiers);
    QItemSelection createSelection(const QModelIndex &topLeft, const QModelIndex &bottomRight);
    int rowAtPosition(int y) const;
    int getIconColumnWidth() const;
    void updateHoverRow(int row);

private:
    QPointer<QRubberBand> m_rubberBand { nullptr };
    QPoint m_dragStartPos;
    bool m_isDragging { false };
    QPoint m_clickedPos { 0, 0 };
    int m_hoverRow { -1 };
    QModelIndexList m_cacheSelectedIndexes;
};

class HoverItemDelegate : public QStyledItemDelegate
{
public:
    explicit HoverItemDelegate(TableViewEx *view, QObject *parent = nullptr);
    void paint(QPainter *painter, const QStyleOptionViewItem &option, const QModelIndex &index) const override;

private:
    TableViewEx *m_view;
};


class DirectoryJumpWidget : public QWidget
{
    Q_OBJECT
public:
    DirectoryJumpWidget(QWidget *parent = nullptr);
    ~DirectoryJumpWidget();

    void setCurrentPath(const QString &path);
    QString currentPath() const { return m_currentPath; }

private:
    void updateUI();
    void onClickedItem(const QString &currentPath);
    bool eventFilter(QObject *watched, QEvent *event) override;

private:
    QString m_currentPath;
    QPointer<QStackedWidget> m_stackedWidget;
    boost::bimap<QPushButton*, QPushButton*> m_buttonMap;
};

class FileTableView : public TableViewEx
{
    Q_OBJECT
public:
    FileTableView(QWidget *parent = nullptr);
    ~FileTableView();
    static int movingPixels();

Q_SIGNALS:
    void dragFiles(const QStringList &pathList, const QPoint &pos);

protected:
    void mousePressEvent(QMouseEvent *event) override;
    void mouseReleaseEvent(QMouseEvent *event) override;
    void mouseDoubleClickEvent(QMouseEvent *event) override;
    void mouseMoveEvent(QMouseEvent *event) override;

private:
    void processLongPressTask();
    void processMouseMovingTask();
    bool processLinkFile(const QString &filePath);

private:
    qint64 m_mouseClickIndex { 0 };
    QPoint m_clickedPos;
    QModelIndexList m_cacheSelectedIndexes;
};

class FileExplorer : public QWidget
{
    Q_OBJECT

public:
    explicit FileExplorer(QWidget *parent = nullptr);
    ~FileExplorer();

    QTreeView *createNaviWindow();

private Q_SLOTS:
    void onDragFiles(const QStringList &pathList, const QPoint &pos);
    void onClickedNaviItem(const QModelIndex &itemIndex);
    void onDoubleClickedTableItem(const QModelIndex &index);
    void onClickedBackLabel();
    void onDirectoryJump(const QString &currentPath);
    void onCustomContextMenu(const QPoint &pos);

private:
    QStandardItemModel *navigationListModel();
    void setCurrentPath(const QString &path);

private:
    Ui::FileExplorer *ui;
    QPointer<FileTableView> m_tableView;
    QStandardItemModel *m_naviModel;
    QPointer<QTreeView> m_naviView;
    DirectoryJumpWidget *m_dirJumpWidget;
};

#endif // FILE_EXPLORER_H
