#ifndef FILE_RECORD_WIDGET_H
#define FILE_RECORD_WIDGET_H

#include <QWidget>
#include "common_signals.h"
#include "common_utils.h"

namespace Ui {
class FileRecordWidget;
}

class FileRecordWidget : public QWidget
{
    Q_OBJECT

public:
    explicit FileRecordWidget(QWidget *parent = nullptr);
    ~FileRecordWidget();

    void setFileOptInfo(const FileOperationRecord &record);
    void updateStatusInfo();
    std::string getClientID() const;
    QByteArray getHashID() const;

private:
    Ui::FileRecordWidget *ui;
    FileOperationRecord m_fileOptRecord;
};

#endif // FILE_RECORD_WIDGET_H
