#include "common_ui_process.h"
#include <QMessageBox>
#include <QApplication>
#include <QPushButton>
#include <QTimer>

CommonUiProcess *CommonUiProcess::m_instance = nullptr;

CommonUiProcess::CommonUiProcess()
    : QObject(nullptr)
{
    initConnect();
}

CommonUiProcess::~CommonUiProcess()
{
    if (m_instance) {
        m_instance->deleteLater();
        m_instance = nullptr;
    }
}

void CommonUiProcess::initConnect()
{
    connect(CommonSignals::getInstance(), &CommonSignals::showInfoMessageBox, this, &CommonUiProcess::onShowInfoMessageBox);
    connect(CommonSignals::getInstance(), &CommonSignals::showWarningMessageBox, this, &CommonUiProcess::onShowWarningMessageBox);
}

CommonUiProcess *CommonUiProcess::getInstance()
{
    if (m_instance == nullptr) {
        m_instance = new CommonUiProcess;
    }
    return m_instance;
}

void CommonUiProcess::onShowInfoMessageBox(const QString &title, const QString &message)
{
    QMessageBox box;
    QPushButton *button = new QPushButton;
    {
        button->setFixedSize(85, 36);
        button->setText("OK");
    }

    box.addButton(button, QMessageBox::ButtonRole::AcceptRole);
    box.setText(message);
    box.setWindowTitle(title);
    box.setIcon(QMessageBox::Icon::Information);

    QFont font;
    font.setPixelSize(16);
    box.setFont(font);
    box.exec();
}

void CommonUiProcess::onShowWarningMessageBox(const QString &title, const QString &message)
{
    QMessageBox box;
    QPushButton *button = new QPushButton;
    {
        button->setFixedSize(85, 36);
        button->setText("OK");
    }

    box.addButton(button, QMessageBox::ButtonRole::AcceptRole);
    box.setText(message);
    box.setWindowTitle(title);
    box.setIcon(QMessageBox::Icon::Warning);

    QFont font;
    font.setPixelSize(16);
    box.setFont(font);
    box.exec();
}
