#ifndef FILE_OPT_INFO_LIST_H
#define FILE_OPT_INFO_LIST_H

#include <QWidget>
#include "file_record_widget.h"
#include "common_utils.h"
#include "common_signals.h"
#include <QLabel>
#include <QVBoxLayout>
#include <QHBoxLayout>
#include <QScrollArea>

namespace Ui {
class FileOptInfoList;
}

class FileOptInfoList : public QWidget
{
    Q_OBJECT

public:
    explicit FileOptInfoList(QWidget *parent = nullptr);
    ~FileOptInfoList();

    static void updateCacheFileOptRecord(const QByteArray &hashID, UpdateProgressMsgPtr ptrMsg);
    static void updateCacheFileOptRecord(const QByteArray &hashID, int optStatus);

private Q_SLOTS:
    void onUpdateFileOptInfoList();
    void onUpdateProgressInfoWithID(int currentVal, const QByteArray &hashID);
    void onUpdateOptRecordStatus(const QByteArray &hashID, int optStatus);

private:
    Ui::FileOptInfoList *ui;
    QList<QPointer<FileRecordWidget> > m_recordWidgetList;
};

#endif // FILE_OPT_INFO_LIST_H
