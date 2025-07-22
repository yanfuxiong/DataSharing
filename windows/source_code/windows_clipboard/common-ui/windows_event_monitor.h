#pragma once
#include <QObject>
#include <QDebug>
#include <QWidget>
#include <QTimer>
#include <Windows.h>
#include <atomic>
#include <dbt.h>
#include <devguid.h>

class WindowsEventMonitor : public QObject
{
    Q_OBJECT
public:
    ~WindowsEventMonitor();
    static WindowsEventMonitor *getInstance();
    static QStringList getSelectedPathList();

Q_SIGNALS:
    void clickedPos(const QPoint &pt);

private:
    static LRESULT CALLBACK hookProc_mouse(int nCode, WPARAM wParam, LPARAM lParam);
    static LRESULT CALLBACK hookProc_keyboard(int nCode, WPARAM wParam, LPARAM lParam);

private:
    WindowsEventMonitor();

    static WindowsEventMonitor *m_instance;
    std::vector<HHOOK> m_hookVec;
};

// Used to detect display plug and unplug events
class MonitorPlugEvent : public QWidget
{
    Q_OBJECT
public:
    struct MonitorData
    {
        HANDLE hPhysicalMonitor { nullptr };
        QString desc;
        std::atomic<bool> isDIAS { false };
        std::string macAddress;

        MonitorData() = default;
        ~MonitorData() = default;
        MonitorData(const MonitorData &other)
            : hPhysicalMonitor(other.hPhysicalMonitor)
            , desc(other.desc)
            , isDIAS(other.isDIAS.load())
            , macAddress(other.macAddress)
        {}
        MonitorData &operator = (const MonitorData &other)
        {
            if (this == &other) {
                return *this;
            }
            hPhysicalMonitor = other.hPhysicalMonitor;
            desc = other.desc;
            isDIAS = other.isDIAS.load();
            macAddress = other.macAddress;
            return *this;
        }
    };
    ~MonitorPlugEvent();
    static MonitorPlugEvent *getInstance();
    void initData();
    void clearData();

    const QList<MonitorData> &getMonitorData() const { return m_monitorDataList; }
    MonitorData &getCacheMonitorData() { return m_cacheMonitorData; }

    static bool getSmartMonitorMacAddr(HANDLE hPhysicalMonitor, std::string &macAddr);
    static bool querySmartMonitorAuthStatus(HANDLE hPhysicalMonitor, uint32_t index, uint8_t &authResult);
    static bool getConnectedPortInfo(HANDLE hPhysicalMonitor, uint8_t &source, uint8_t &port);
    static bool updateMousePos(HANDLE hPhysicalMonitor, unsigned short width, unsigned short hight, short posX, short posY);
    static bool isDIASMonitor(uint8_t authResult);

Q_SIGNALS:
    void statusChanged(bool status);

private:
    MonitorPlugEvent();
    bool registerDeviceNotification();
    void unregisterDeviceNotification();
    bool nativeEvent(const QByteArray &eventType, void *message, long *result) override;
    void updateMonitorDataList();
    void debugDeviceInfo(DWORD dwDevType, const QString &devPath);
    void getCurrentAllMonitroData(QList<MonitorData> &data);
    static BOOL monitorEnumProc(HMONITOR hMonitor, HDC hdcMonitor, LPRECT lprcMonitor, LPARAM dwData);
    void processDDCCI();
    void initDataImpl();
    void stopProcessDDCCI();
    void restartProcessDDCCI();

    HDEVNOTIFY m_hDevNotify { nullptr };
    QList<MonitorData> m_monitorDataList;
    MonitorData m_cacheMonitorData;
    std::atomic<int> m_currentTimeout { 0 };

    static const int DDCCI_TIMEOUT = 90 * 1000; // 90s
    static MonitorPlugEvent *m_instance;
};
