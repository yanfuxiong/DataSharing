#ifndef FILE_RECORD_WIDGET_H
#define FILE_RECORD_WIDGET_H

#include <QWidget>
#include <QLabel>
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
    const FileOperationRecord &recordData() const { return m_fileOptRecord; }
    void updateStatusInfo();
    std::string getClientID() const;
    QByteArray getHashID() const;

private:
    void processClickedOpenFileIcon();
    void processClickedRightIcon();
    void updateUI(const FileOperationRecord &record);
    bool showCancelAllTransferDialog() const;
    bool showTerminateSingleFileTransfer(const QString &fileName) const;
    bool cancelTransferFunctionIsEnabled() const;
    static bool isFullPath(const QString &path);

private:
    Ui::FileRecordWidget *ui;
    FileOperationRecord m_fileOptRecord;
    QLabel *m_openFileLabel;
};

#endif // FILE_RECORD_WIDGET_H
