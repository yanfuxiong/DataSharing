#include "file_explorer.h"
#include "ui_file_explorer.h"
#include "device_info.h"
#include "windows_event_monitor.h"
#include "event_filter_process.h"
#include <QDesktopServices>
#include <QStandardPaths>
#include <QFileIconProvider>
#include <QDir>
#include <QPainter>
#include <QStyle>
#include <QHeaderView>

namespace {

void clickLeftMouse(const QPoint &pos)
{
    int targetX = pos.x();
    int targetY = pos.y();

    int screenWidth = ::GetSystemMetrics(SM_CXSCREEN);
    int screenHeight = ::GetSystemMetrics(SM_CYSCREEN);
    DWORD dx = targetX * 65535 / (screenWidth - 1);
    DWORD dy = targetY * 65535 / (screenHeight - 1);
    ::mouse_event(MOUSEEVENTF_ABSOLUTE | MOUSEEVENTF_MOVE | MOUSEEVENTF_LEFTDOWN, dx, dy, 0, 0);
    ::mouse_event(MOUSEEVENTF_LEFTUP, 0, 0, 0, 0);
}

void releaseLeftMouse()
{
    ::mouse_event(MOUSEEVENTF_LEFTUP, 0, 0, 0, 0);
}

void releaseShiftAndCtrlKey()
{
    ::keybd_event(VK_LSHIFT, 0, KEYEVENTF_KEYUP, 0);
    ::keybd_event(VK_LCONTROL, 0, KEYEVENTF_KEYUP, 0);
}

QString getWindowsDriveDisplayName(const QString &drivePath)
{
    QString normalizedPath = drivePath;
    if (!normalizedPath.endsWith("\\")) {
        normalizedPath += "\\";
    }

    const wchar_t *winPath = reinterpret_cast<const wchar_t*>(normalizedPath.utf16());

    SHFILEINFOW sfi;
    std::memset(&sfi, 0, sizeof (sfi));
    DWORD_PTR result = ::SHGetFileInfoW(winPath, 0, &sfi, sizeof(sfi), SHGFI_DISPLAYNAME | SHGFI_USEFILEATTRIBUTES);

    if (result != 0) {
        return QString::fromWCharArray(sfi.szDisplayName);
    } else {
        QString driveLetter = normalizedPath.left(2).toUpper();
        return QString("Local Disk (%1)").arg(driveLetter);
    }
}

std::vector<std::pair<QString, QString> > getAllDrives()
{
    std::vector<std::pair<QString, QString> > allDrivesVec;

    wchar_t drives[MAX_PATH] = { 0 };
    DWORD len = ::GetLogicalDriveStringsW(MAX_PATH, drives);

    if (len == 0) {
        qWarning() << "Unable to retrieve the list of drives";
        return allDrivesVec;
    }

    const wchar_t *p = drives;
    while (*p != L'\0') {
        QString drivePath = QString::fromWCharArray(p); // eg: "C:\\"
        QString displayName = getWindowsDriveDisplayName(drivePath);
        //qDebug() << "; displayName:" << displayName << "drivePath:" << drivePath;
        allDrivesVec.push_back({ displayName, drivePath });

        p += wcslen(p) + 1;
    }

    return allDrivesVec;
}

QString resolveLnkFile(const QString &lnkFilePath)
{
    return QFileInfo(lnkFilePath).symLinkTarget();
}

}

//----------------------------------------------------- TableViewEx

TableViewEx::TableViewEx(QWidget *parent)
    : QTableView(parent)
{
    setSelectionMode(QAbstractItemView::ExtendedSelection);
    setSelectionBehavior(QAbstractItemView::SelectRows);
    setMouseTracking(true);
    setShowGrid(false);
    setFrameShape(QFrame::NoFrame);
    viewport()->setAttribute(Qt::WA_Hover);
    {
        HoverItemDelegate *delegate = new HoverItemDelegate(this, this);
        setItemDelegate(delegate);
    }
    setStyleSheet(
        "QTableView {"
        "  selection-background-color: #cce8ff;"
        "  selection-color: black;"
        "}"
    );
}

int TableViewEx::rowAtPosition(int y) const
{
    const int singleRowHeight = rowHeight(0);
    if (singleRowHeight <= 0) {
        return -1;
    }
    return qBound(0, y / singleRowHeight, model()->rowCount(rootIndex()) - 1) + indexAt(QPoint(1, singleRowHeight / 2)).row();
}

QModelIndex TableViewEx::validIndexAt(const QPoint &pos) const
{
    const QModelIndex parent = rootIndex();
    const int row = rowAtPosition(pos.y());
    if (row < 0) {
        return QModelIndex();
    }

    return model()->index(row, 0, parent);
}

int TableViewEx::getIconColumnWidth() const
{
    QStyleOptionViewItem option;
    option.initFrom(this);

    // Fetch the system's default icon size, factoring in system DPI scaling.
    const int iconSize = style()->pixelMetric(QStyle::PM_SmallIconSize, &option, this);
    const int horizontalSpacing = style()->pixelMetric(QStyle::PM_LayoutHorizontalSpacing, &option, this);
    const int totalWidth = iconSize + horizontalSpacing + (option.decorationAlignment.testFlag(Qt::AlignLeft) ? 1 : 0) * 2;
    return totalWidth;
}

bool TableViewEx::checkClickedPostion(const QPoint &pos) const
{
    const int columnCount = 4;
    if (m_cacheSelectedIndexes.size() > columnCount) {
        return false;
    }
    QPoint newPos = viewport()->mapFromGlobal(pos);
    auto index = indexAt(viewport()->mapFromGlobal(pos));
    QString currentString = model()->data(index).toString();
    QFontMetrics fontInfo(font());
    if (index.column() != 1) {
        int minWidth = 0;
        for (int colIndex = 0; colIndex < index.column(); ++colIndex) {
            minWidth += columnWidth(colIndex);
        }

        minWidth += fontInfo.horizontalAdvance(currentString) + (index.column() == 0 ? getIconColumnWidth() : 0);
        return newPos.x() > minWidth;
    } else {
        int minWidth = columnWidth(0);
        int maxWidth = 0;
        maxWidth += columnWidth(0) + columnWidth(1) - fontInfo.horizontalAdvance(currentString);
        return newPos.x() > minWidth && newPos.x() < maxWidth;
    }
}

void TableViewEx::updateHoverRow(int row)
{
    if (row != m_hoverRow) {
        const int previousRow = m_hoverRow;
        m_hoverRow = row;

        auto updateRow = [this](int row) {
            if (row == -1 || model() == nullptr) {
                return;
            }
            const int cols = model()->columnCount();
            for (int c = 0; c < cols; ++c) {
                update(model()->index(row, c));
            }
        };

        updateRow(previousRow);
        updateRow(m_hoverRow);
    }
}

void TableViewEx::mousePressEvent(QMouseEvent *event)
{
    m_clickedPos = event->globalPos();
    if (event->button() == Qt::LeftButton) {
        m_dragStartPos = event->pos();
        m_initialIndex = validIndexAt(m_dragStartPos);
        m_isDragging = false;

        // Allow dragging to start in empty regions.
        if (!m_initialIndex.isValid()) {
            const int row = rowAtPosition(m_dragStartPos.y());
            m_initialIndex = model()->index(row, 0, rootIndex());
        }
    }

    QTableView::mousePressEvent(event);

    if (event->button() == Qt::LeftButton) {
        m_cacheSelectedIndexes = selectedIndexes();
    }
}

void TableViewEx::mouseReleaseEvent(QMouseEvent *event)
{
    if (m_isDragging && m_rubberBand) {
        m_rubberBand->hide();
    }

    m_isDragging = false;
    m_initialIndex = QModelIndex();
    m_cacheSelectedIndexes.clear();
    QTableView::mouseReleaseEvent(event);
}

void TableViewEx::mouseMoveEvent(QMouseEvent *event)
{
    {
        const QModelIndex index = indexAt(event->pos());
        updateHoverRow(index.isValid() ? index.row() : -1);
    }

    if ((event->buttons() & Qt::LeftButton)) {
        const int dragDistance = (event->pos() - m_dragStartPos).manhattanLength();

        if (!m_isDragging && dragDistance >= QApplication::startDragDistance()) {
            m_isDragging = true;
            if (!m_rubberBand) {
                m_rubberBand = new QRubberBand(QRubberBand::Rectangle, viewport());
                m_rubberBand->setAttribute(Qt::WA_TranslucentBackground);
                m_rubberBand->setAttribute(Qt::WA_TransparentForMouseEvents);

                m_rubberBand->setStyleSheet(
                    "QRubberBand {"
                    "  border: 1px solid #3d8de0;"
                    "  background-color: rgba(61, 141, 224, 40);"
                    "  border-radius: 1px;"
                    "}"
                );
            }
            m_rubberBand->setGeometry(QRect(m_dragStartPos, QSize()));
            if (checkClickedPostion(m_clickedPos)) {
                m_rubberBand->show();
            } else {
                m_rubberBand->hide();
            }
        }
    }

    if (m_isDragging) {
        QRect rubberRect = QRect(m_dragStartPos, event->pos()).normalized();
        m_rubberBand->setGeometry(rubberRect);
        updateSelection(event->pos(), event->modifiers());
    } else {
        QTableView::mouseMoveEvent(event);
    }
}

void TableViewEx::leaveEvent(QEvent *event)
{
    updateHoverRow(-1);
    QTableView::leaveEvent(event);
}

void TableViewEx::updateSelection(const QPoint &currentPos, Qt::KeyboardModifiers modifiers)
{
    if (m_rubberBand == nullptr || m_rubberBand->isHidden()) {
        return;
    }
    const QModelIndex parent = rootIndex();
    const int rowCount = model()->rowCount(parent);

    // Handle dragging in empty regions
    const int startRow = qBound(0, rowAtPosition(m_dragStartPos.y()), rowCount - 1);
    const int endRow = qBound(0, rowAtPosition(currentPos.y()), rowCount - 1);

    QModelIndex topLeft = model()->index(qMin(startRow, endRow), 0, parent);
    QModelIndex bottomRight = model()->index(qMax(startRow, endRow), model()->columnCount(parent) - 1, parent);

    QItemSelection selection = createSelection(topLeft, bottomRight);
    QItemSelectionModel::SelectionFlags flags = QItemSelectionModel::Rows;

    if (modifiers & Qt::ControlModifier) {
        flags |= QItemSelectionModel::Toggle;
    } else if (modifiers & Qt::ShiftModifier) {
        flags |= QItemSelectionModel::Select;
    } else {
        flags |= QItemSelectionModel::ClearAndSelect;
    }

    if (selection.isEmpty()) {
        return;
    }
    selectionModel()->select(selection, flags);
}

QItemSelection TableViewEx::createSelection(const QModelIndex &topLeft, const QModelIndex &bottomRight)
{
    QItemSelection selection;

    if (!topLeft.isValid() || !bottomRight.isValid()) {
        return selection;
    }

    const QModelIndex parent = rootIndex();
    const int topRow = qMax(0, qMin(topLeft.row(), bottomRight.row()));
    const int bottomRow = qMin(model()->rowCount(parent) - 1, qMax(topLeft.row(), bottomRight.row()));

    if (qobject_cast<QFileSystemModel*>(model())) {
        for (int row = topRow; row <= bottomRow; ++row) {
            QModelIndex left = model()->index(row, 0, parent);
            QModelIndex right = model()->index(row, model()->columnCount(parent) - 1, parent);
            if (left.isValid() && right.isValid()) {
                selection.merge(QItemSelection(left, right), QItemSelectionModel::Select);
            }
        }
    } else {
        selection.select(topLeft, bottomRight);
    }

    return selection;
}

//---------------------------------------------------------------------------
HoverItemDelegate::HoverItemDelegate(TableViewEx *view, QObject *parent)
    : QStyledItemDelegate(parent), m_view(view)
{
}

void HoverItemDelegate::paint(QPainter *painter, const QStyleOptionViewItem &option, const QModelIndex &index) const
{
    QStyleOptionViewItem opt = option;
    initStyleOption(&opt, index);

    const bool isSelected = opt.state & QStyle::State_Selected;
    const bool isHovered = (index.row() == m_view->hoverRow()) && !isSelected;

    if (isHovered) {
        painter->save();
        painter->fillRect(opt.rect, QColor(225, 240, 255));
        painter->restore();
    }

    QStyledItemDelegate::paint(painter, opt, index);
}

//------------------------------------------------------- DirectoryJumpWidget

DirectoryJumpWidget::DirectoryJumpWidget(QWidget *parent)
    : QWidget(parent)
{
    m_stackedWidget = new QStackedWidget;
    m_stackedWidget->setSizePolicy(QSizePolicy::Expanding, QSizePolicy::Expanding);
    QHBoxLayout *pHBoxLayout = new QHBoxLayout;
    pHBoxLayout->setSpacing(0);
    pHBoxLayout->setMargin(0);
    setLayout(pHBoxLayout);
    pHBoxLayout->addWidget(m_stackedWidget);
}

DirectoryJumpWidget::~DirectoryJumpWidget()
{

}

void DirectoryJumpWidget::setCurrentPath(const QString &path)
{
    Q_ASSERT(QFileInfo(path).exists() == true);
    m_currentPath = QDir::toNativeSeparators(path);
    QTimer::singleShot(0, this, [this] {
        updateUI();
    });
}

void DirectoryJumpWidget::updateUI()
{
    while (m_stackedWidget->count() > 0) {
        if (auto widget = m_stackedWidget->widget(0)) {
            m_stackedWidget->removeWidget(widget);
            widget->deleteLater();
        }
    }

    QStringList dirList = m_currentPath.split('\\');
    if (dirList.empty() == false && dirList.back().isEmpty()) {
        dirList.pop_back();
    }
    QWidget *widget = new QWidget;
    m_stackedWidget->addWidget(widget);
    QHBoxLayout *pHBoxLayout = new QHBoxLayout;
    pHBoxLayout->setSpacing(0);
    pHBoxLayout->setMargin(0);
    widget->setLayout(pHBoxLayout);

    QString cacheCurrentPath;
    QFont buttonFont;
    buttonFont.setPointSize(10);
    for (const auto &itemName : dirList) {
        Q_ASSERT(itemName.isEmpty() == false);
        if (cacheCurrentPath.isEmpty()) {
            cacheCurrentPath = itemName;
        } else {
            cacheCurrentPath += "\\" + itemName;
        }
        QPushButton *itemButton = new QPushButton;
        itemButton->setFont(buttonFont);
        itemButton->setObjectName("fileExplorerButton");
        itemButton->setSizePolicy(QSizePolicy::Fixed, QSizePolicy::Expanding);
        itemButton->setFixedWidth(QFontMetrics(buttonFont).horizontalAdvance(itemName + QByteArray(4, 'Q')) + itemName.size());
        itemButton->setText(itemName);
        pHBoxLayout->addWidget(itemButton);

        connect(itemButton, &QPushButton::clicked, this, std::bind(&DirectoryJumpWidget::onClickedItem, this, cacheCurrentPath));
    }
    pHBoxLayout->addStretch();
}

void DirectoryJumpWidget::onClickedItem(const QString &currentPath)
{
    Q_EMIT CommonSignals::getInstance()->directoryJump(currentPath);
}

//------------------------------------------------------- FileTableView

FileTableView::FileTableView(QWidget *parent)
    : TableViewEx(parent)
{
    {
        setShowGrid(false);
        {
            setSortingEnabled(false);
            horizontalHeader()->setSectionsClickable(false);
            horizontalHeader()->setStyleSheet(
                "QHeaderView::section {background-color:#F0F0F0;border:none;border-right:1px solid lightgrey;}"
            );
        }
        setContextMenuPolicy(Qt::ContextMenuPolicy::CustomContextMenu);
    }

    connect(this, &QTableView::doubleClicked, this, [this] (const QModelIndex &index) {
        auto fsModel = qobject_cast<QFileSystemModel*>(model());
        QString filePath = fsModel->filePath(index);
        if (QFileInfo(filePath).isFile() == false) {
            return;
        }
        QDesktopServices::openUrl(QUrl::fromLocalFile(filePath));
    });

    connect(CommonSignals::getInstance(), &CommonSignals::processLinkFile, this, [this] (const QString &filePath) {
        processLinkFile(filePath);
    });
}

FileTableView::~FileTableView()
{

}

int FileTableView::movingPixels()
{
    try {
        return g_getGlobalData()->localConfig.at("fileExplorer").at("movingPixels").get<int>();
    } catch (const std::exception &e) {
        qWarning() << e.what();
        return 30;
    }
}

void FileTableView::mousePressEvent(QMouseEvent *event)
{
    ++m_mouseClickIndex;
    m_clickedPos = event->globalPos();
    if (event->button() == Qt::MouseButton::LeftButton) {
        processLongPressTask();
    }

    do {
        if (event->pos().x() > 25) {
            break;
        }
        auto index = indexAt(event->pos());
        auto fsModel = qobject_cast<QFileSystemModel*>(model());
        QString filePath = fsModel->filePath(index);
        if (processLinkFile(filePath)) {
            return;
        }
    } while (false);
    TableViewEx::mousePressEvent(event);

    if (event->button() == Qt::MouseButton::LeftButton) {
        m_cacheSelectedIndexes = selectedIndexes();
    }
}

void FileTableView::mouseReleaseEvent(QMouseEvent *event)
{
    ++m_mouseClickIndex;
    m_cacheSelectedIndexes.clear();
    TableViewEx::mouseReleaseEvent(event);
}

void FileTableView::mouseDoubleClickEvent(QMouseEvent *event)
{
    if (event->button() == Qt::MouseButton::RightButton) {
        event->ignore();
        return;
    }
    do {
        auto index = indexAt(event->pos());
        auto fsModel = qobject_cast<QFileSystemModel*>(model());
        QString filePath = fsModel->filePath(index);
        if (processLinkFile(filePath)) {
            return;
        }
    } while (false);
    TableViewEx::mouseDoubleClickEvent(event);
}

bool FileTableView::processLinkFile(const QString &filePath)
{
    QFileInfo fileInfo(filePath);
    if (fileInfo.suffix().toLower() == "lnk") {
        auto realFilePath = resolveLnkFile(fileInfo.absoluteFilePath());
        if (fileInfo.isDir()) {
            auto fsModel = qobject_cast<QFileSystemModel*>(model());
            fsModel->setRootPath(realFilePath);
            setRootIndex(fsModel->index(realFilePath));
            Q_EMIT CommonSignals::getInstance()->directoryJump(realFilePath);
            return true;
        }
    }
    return false;
}

void FileTableView::mouseMoveEvent(QMouseEvent *event)
{
    if ((event->globalPos() - m_clickedPos).manhattanLength() >= movingPixels()) {
        const int minMovingDistance = 20;
        if ((event->globalPos() - m_clickedPos).manhattanLength() >= minMovingDistance) {
            ++m_mouseClickIndex;
        }
        TableViewEx::mouseMoveEvent(event);
        if (movingPixels() > 0) {
            processMouseMovingTask();
        }
    } else {
        event->ignore();
    }
}

void FileTableView::processMouseMovingTask()
{
    if (m_cacheSelectedIndexes.isEmpty() ||
        m_cacheSelectedIndexes != selectedIndexes() ||
        checkClickedPostion(m_clickedPos)) {
        return;
    }
    const int columnCount = 4;
    if (m_cacheSelectedIndexes.size() == columnCount && indexAt(viewport()->mapFromGlobal(m_clickedPos)).column() != 0) {
        return;
    }
    auto fsModel = qobject_cast<QFileSystemModel*>(model());
    auto cacheSelection = selectionModel()->selection();
    QStringList pathList;
    if (fsModel) {
        for (const auto &index : selectedIndexes()) {
            if (index.column() == 0) {
                if (QFileInfo(fsModel->filePath(index)).isFile() || QFileInfo(fsModel->filePath(index)).isDir()) {
                    qInfo() << "drag:" << fsModel->filePath(index);
                    pathList << fsModel->filePath(index);
                }
            }
        }
    }

    if (pathList.isEmpty() == false) {
        if (MonitorPlugEvent::getInstance()->getCacheMonitorData().macAddress.empty()) {
            Q_EMIT CommonSignals::getInstance()->showWarningMessageBox("warning", "Invalid monitor!!!");
            return;
        }

        QPoint currentPos = QCursor::pos();
        QTimer::singleShot(0, this, [this, pathList, currentPos] {
            Q_EMIT dragFiles(pathList, currentPos);
        });
    }

    if (auto currentSelectionModel = selectionModel()) {
        currentSelectionModel->clear();
    }
    releaseLeftMouse();
    releaseShiftAndCtrlKey();

    // To avoid interference from the event of releasing the left mouse button,
    // a delay of 500ms is applied to execute the reselection behavior.
    QTimer::singleShot(500, this, [this, cacheSelection] {
        selectionModel()->select(cacheSelection, QItemSelectionModel::ClearAndSelect);
    });
}

void FileTableView::processLongPressTask()
{
    auto currentIndex = m_mouseClickIndex;
    auto getMousePressTimeout = [] {
        try {
            return g_getGlobalData()->localConfig.at("fileExplorer").at("mousePressTimeout").get<int>();
        } catch (const std::exception &e) {
            qWarning() << "getMousePressTimeout failed:" << e.what();
            const int mousePressTimeout = 1000; // 1000ms
            return mousePressTimeout;
        }
    };
    QTimer::singleShot(getMousePressTimeout(), this, [this, currentIndex] {
        if (currentIndex != m_mouseClickIndex) { // Used to detect when the user holds down the mouse for less than a second
            return;
        }
        auto fsModel = qobject_cast<QFileSystemModel*>(model());
        auto cacheSelection = selectionModel()->selection();
        QStringList pathList;
        if (fsModel) {
            for (const auto &index : selectedIndexes()) {
                if (index.column() == 0) {
                    if (QFileInfo(fsModel->filePath(index)).isFile() || QFileInfo(fsModel->filePath(index)).isDir()) {
                        qInfo() << "drag:" << fsModel->filePath(index);
                        pathList << fsModel->filePath(index);
                    }
                }
            }
        }

        if (pathList.isEmpty() == false) {
            if (MonitorPlugEvent::getInstance()->getCacheMonitorData().macAddress.empty()) {
                Q_EMIT CommonSignals::getInstance()->showWarningMessageBox("warning", "Invalid monitor!!!");
                return;
            }

            QPoint currentPos = QCursor::pos();
            QTimer::singleShot(0, this, [this, pathList, currentPos] {
                Q_EMIT dragFiles(pathList, currentPos);
            });
        }

        if (auto currentSelectionModel = selectionModel()) {
            currentSelectionModel->clear();
        }
        releaseLeftMouse();
        releaseShiftAndCtrlKey();

        // To avoid interference from the event of releasing the left mouse button,
        // a delay of 500ms is applied to execute the reselection behavior.
        QTimer::singleShot(500, this, [this, cacheSelection] {
            selectionModel()->select(cacheSelection, QItemSelectionModel::ClearAndSelect);
        });
    });

}

//------------------------------------------------------- FileExplorer
FileExplorer::FileExplorer(QWidget *parent) :
    QWidget(parent),
    ui(new Ui::FileExplorer),
    m_naviModel(nullptr),
    m_naviView(nullptr)
{
    ui->setupUi(this);
    {
        ui->current_path_box->setProperty(PR_ADJUST_WINDOW_Y_SIZE, true);
        ui->back_label->setProperty(PR_ADJUST_WINDOW_X_SIZE, true);
        ui->back_label->setProperty(PR_ADJUST_WINDOW_Y_SIZE, true);
    }

    {
        ui->back_label->clear();
        m_dirJumpWidget = new DirectoryJumpWidget;
        m_dirJumpWidget->setSizePolicy(QSizePolicy::Expanding, QSizePolicy::Expanding);
        QHBoxLayout *pHBoxLayout = new QHBoxLayout;
        pHBoxLayout->setMargin(0);
        pHBoxLayout->setSpacing(0);
        ui->current_path_box->setLayout(pHBoxLayout);
        pHBoxLayout->addWidget(m_dirJumpWidget);
    }

    {
        QHBoxLayout *pHBoxLayout = new QHBoxLayout;
        pHBoxLayout->setMargin(0);
        pHBoxLayout->setSpacing(0);
        ui->view_box->setLayout(pHBoxLayout);

        m_tableView = new FileTableView;
        m_tableView->setSizePolicy(QSizePolicy::Expanding, QSizePolicy::Expanding);
        pHBoxLayout->addWidget(m_tableView);

        QFileSystemModel *model = new QFileSystemModel(this);
        model->setOption(QFileSystemModel::Option::DontResolveSymlinks, true);
        m_tableView->setModel(model);

        m_tableView->setColumnWidth(0, 360);
        qDebug() << "FileExplorer font size:" << m_tableView->font().pointSizeF();

        model->setRootPath(QString());
        setCurrentPath(CommonUtils::desktopDirectoryPath());
    }

    {
        connect(CommonSignals::getInstance(), &CommonSignals::directoryJump, this, &FileExplorer::onDirectoryJump);
        connect(m_tableView, &QTableView::doubleClicked, this, &FileExplorer::onDoubleClickedTableItem);
        connect(m_tableView, &FileTableView::dragFiles, this, &FileExplorer::onDragFiles);
        connect(m_tableView, &QTableView::customContextMenuRequested, this, &FileExplorer::onCustomContextMenu);
        connect(MonitorPlugEvent::getInstance(), &MonitorPlugEvent::statusChanged, this, [this] (bool status) {
            if (status == false) {
                MonitorPlugEvent::getInstance()->clearData();
            }
        });
        MonitorPlugEvent::getInstance()->initData();
    }

    EventFilterProcess::getInstance()->registerFilterEvent({ ui->back_label, std::bind(&FileExplorer::onClickedBackLabel, this) });
}

FileExplorer::~FileExplorer()
{
    delete ui;
}

QTreeView *FileExplorer::createNaviWindow()
{
    if (m_naviView) {
        return m_naviView.data();
    }
    m_naviView = new QTreeView;
    m_naviView->setEditTriggers(QAbstractItemView::NoEditTriggers);
    m_naviView->setMaximumWidth(220);
    m_naviView->setSizePolicy(QSizePolicy::MinimumExpanding, QSizePolicy::Preferred);
    auto model = navigationListModel();
    model->setParent(m_naviView);
    m_naviView->setModel(model);
    m_naviView->header()->setVisible(false);
    m_naviView->header()->setSectionResizeMode(QHeaderView::Fixed);
    m_naviView->header()->setDefaultSectionSize(0);

    connect(m_naviView, &QTreeView::clicked, this, &FileExplorer::onClickedNaviItem);
    // Detect changes in the system drive letter
    std::thread([this] {
        static std::vector<std::pair<QString, QString> > s_allDrives = getAllDrives();
        while (true) {
            auto currentDrives = getAllDrives();
            if (currentDrives != s_allDrives) {
                s_allDrives = std::move(currentDrives);
                // Transfer to the main thread for execution
                QTimer::singleShot(0, this, [this] {
                    QStandardItemModel *model = qobject_cast<QStandardItemModel*>(m_naviView->model());
                    model->deleteLater();
                    m_naviModel = nullptr;
                    model = navigationListModel();
                    model->setParent(m_naviView);
                    m_naviView->setModel(model);
                });
            }
            // Delay by 1000ms to avoid consuming system CPU resources for too fast detection.
            QThread::msleep(1000);
        }
    }).detach();
    return m_naviView;
}

QStandardItemModel *FileExplorer::navigationListModel()
{
    if (m_naviModel) {
        return m_naviModel;
    }
    m_naviModel = new QStandardItemModel;

    std::vector<std::pair<QString, QString> > systemFoldersInfo;
    systemFoldersInfo.push_back({ "Videos", QStandardPaths::writableLocation(QStandardPaths::StandardLocation::MoviesLocation) });
    systemFoldersInfo.push_back({ "Pictures", QStandardPaths::writableLocation(QStandardPaths::StandardLocation::PicturesLocation) });
    systemFoldersInfo.push_back({ "Documents", QStandardPaths::writableLocation(QStandardPaths::StandardLocation::DocumentsLocation) });
    systemFoldersInfo.push_back({ "Downloads", QStandardPaths::writableLocation(QStandardPaths::StandardLocation::DownloadLocation) });
    systemFoldersInfo.push_back({ "Music", QStandardPaths::writableLocation(QStandardPaths::StandardLocation::MusicLocation) });
    systemFoldersInfo.push_back({ "Desktop", QStandardPaths::writableLocation(QStandardPaths::StandardLocation::DesktopLocation) });

    for (const auto &info : getAllDrives()) {
        QString displayName = info.first;
        QString rootPath = info.second;
        systemFoldersInfo.push_back({ displayName, rootPath });
    }


    QFileIconProvider iconProvider;
    for (const auto &data : systemFoldersInfo) {
        QString name = data.first;
        QString path = data.second;

        QStandardItem *item = new QStandardItem(name);
        item->setData(path, Qt::UserRole);
        item->setIcon(iconProvider.icon(QFileIconProvider::Folder));

        m_naviModel->appendRow(item);
    }
    return m_naviModel;
}

void FileExplorer::onDragFiles(const QStringList &pathList, const QPoint &pos)
{
    Q_ASSERT(pathList.isEmpty() == false);
    DragFilesMsg message;
    message.functionCode = DragFilesMsg::FuncCode::MultiFiles;
    message.timeStamp = QDateTime::currentDateTime().toMSecsSinceEpoch();
    message.rootPath = QFileInfo(m_dirJumpWidget->currentPath()).absoluteFilePath();
    for (const auto &path : pathList) {
        message.filePathVec.push_back(path);
    }

    qInfo() << "[DIAS]: Send data to the CrossShareServ......";
    QByteArray data = DragFilesMsg::toByteArray(message);
    Q_EMIT CommonSignals::getInstance()->sendDataToServer(data);

    {
        int width = GetSystemMetrics(SM_CXSCREEN);
        int height = GetSystemMetrics(SM_CYSCREEN);
        qInfo() << "screen width=" << width << "; screen height=" << height << ";x=" << pos.x() << ";y=" << pos.y();
        auto func = [&] {
            bool retVal = MonitorPlugEvent::updateMousePos(MonitorPlugEvent::getInstance()->getCacheMonitorData().hPhysicalMonitor,
                                                    static_cast<uint16_t>(width),
                                                    static_cast<uint16_t>(height),
                                                    static_cast<int16_t>(pos.x()),
                                                    static_cast<int16_t>(pos.y()));
            if (retVal) {
                qInfo() << "[updateMousePos]: success !!!";
            }
            return retVal;
        };
        auto retVal = func();
        if (retVal == false) {
            MonitorPlugEvent::getInstance()->initData();
            retVal = func();
        }
        if (retVal == false) {
            Q_EMIT CommonSignals::getInstance()->showWarningMessageBox("warning", "UpdateMousePos failed !!!");
        }
    }
}

void FileExplorer::setCurrentPath(const QString &path)
{
    auto model = qobject_cast<QFileSystemModel*>(m_tableView->model());
    if (model == nullptr) {
        return;
    }
    m_tableView->setRootIndex(model->index(path));
    m_dirJumpWidget->setCurrentPath(QDir::toNativeSeparators(path));
}

void FileExplorer::onClickedNaviItem(const QModelIndex &itemIndex)
{
    auto model = qobject_cast<QFileSystemModel*>(m_tableView->model());
    if (model == nullptr) {
        return;
    }
    QString currentPath = itemIndex.data(Qt::UserRole).toString();
    qInfo() << "currentPath:" << currentPath.toUtf8().constData();
    setCurrentPath(currentPath);
}

void FileExplorer::onDoubleClickedTableItem(const QModelIndex &index)
{
    auto model = qobject_cast<QFileSystemModel*>(m_tableView->model());
    if (model == nullptr) {
        return;
    }
    QFileInfo fileInfo = model->fileInfo(index);
    if (fileInfo.isDir()) {
        setCurrentPath(fileInfo.absoluteFilePath());
    }
}

void FileExplorer::onClickedBackLabel()
{
    QDir currentDir(m_dirJumpWidget->currentPath());
    if (currentDir.cdUp()) {
        setCurrentPath(currentDir.absolutePath());
    }
}

void FileExplorer::onDirectoryJump(const QString &currentPath)
{
    setCurrentPath(currentPath);
}

void FileExplorer::onCustomContextMenu(const QPoint &pos)
{
    QModelIndex index = m_tableView->indexAt(pos);
    if (index.isValid() == false) {
        return;
    }
    QMenu menu;
    menu.addAction("open", [index, this] {
        auto fsModel = qobject_cast<QFileSystemModel*>(m_tableView->model());
        QFileInfo fileInfo = fsModel->fileInfo(index);
        qInfo() << "open:" << fileInfo.absoluteFilePath().toUtf8().constData();
        if (fileInfo.isShortcut()) {
            Q_EMIT CommonSignals::getInstance()->processLinkFile(fileInfo.absoluteFilePath());
        } else if (fileInfo.isDir()) {
            setCurrentPath(fileInfo.absoluteFilePath());
        } else if (fileInfo.isFile()) {
            QDesktopServices::openUrl(QUrl::fromLocalFile(fileInfo.absoluteFilePath()));
        }
    });

    menu.addSeparator();

    {
        auto subMenu = menu.addMenu("send to");
        for (const auto &client : g_getGlobalData()->m_clientVec) {
            QByteArray clientID = client->clientID;
            QString clientName = client->clientName;
            QByteArray deviceType = client->deviceType;
            QString iconPath = DeviceInfo::getIconPathByDeviceType(deviceType);
            subMenu->addAction(QIcon(iconPath), clientName, [clientID, clientName, deviceType, this] {
                qInfo() << "click:" << clientName << ";deviceType:" << deviceType.constData();
                UpdateClientStatusMsgPtr ptr_client = nullptr;
                for (const auto &data : g_getGlobalData()->m_clientVec) {
                    if (data->clientID == clientID) {
                        ptr_client = data;
                        break;
                    }
                }
                if (ptr_client == nullptr) {
                    return;
                }

                auto fsModel = qobject_cast<QFileSystemModel*>(m_tableView->model());
                QStringList pathList;
                if (fsModel) {
                    for (const auto &index : m_tableView->selectionModel()->selectedIndexes()) {
                        if (index.column() == 0) {
                            if (QFileInfo(fsModel->filePath(index)).isFile() || QFileInfo(fsModel->filePath(index)).isDir()) {
                                qInfo() << "send to:" << fsModel->filePath(index);
                                pathList << fsModel->filePath(index);
                            }
                        }
                    }
                }

                if (pathList.isEmpty() == false) {
                    if (MonitorPlugEvent::getInstance()->getCacheMonitorData().macAddress.empty()) {
                        Q_EMIT CommonSignals::getInstance()->showWarningMessageBox("warning", "Invalid monitor!!!");
                        return;
                    }
                }

                if (pathList.empty()) {
                    return;
                }

                SendFileRequestMsg msg;
                msg.ip = ptr_client->ip;
                msg.port = ptr_client->port;
                msg.clientID = ptr_client->clientID;
                msg.timeStamp = QDateTime::currentDateTime().toMSecsSinceEpoch();
                for (const auto &filePath : pathList) {
                    msg.filePathVec.push_back(filePath);
                }

                QByteArray send_data = SendFileRequestMsg::toByteArray(msg);
                Q_EMIT CommonSignals::getInstance()->sendDataToServer(send_data);
            });
        }
    }

    menu.exec(m_tableView->viewport()->mapToGlobal(pos));
}
