#include <QTest>
#include "common_utils.h"
#include "monitor_info.h"
#include <boost/filesystem.hpp>
#include <boost/container/stable_vector.hpp>

class TestQString : public QObject
{
    Q_OBJECT

private Q_SLOTS:
    void initTestCase() {}
    void cleanupTestCase() {}

    void test_utf16()
    {
        {
            QString str(R"(D:\jack_huang\Downloads\新增資料夾\測試.mp4)"); // UTF8
            qInfo() << CommonUtils::toUtf16LE(str).toHex().toUpper().constData();
        }

        // 从 utf16 LE 转 utf8
        {
            QByteArray data;
            {
                QString str(R"(D:\jack_huang\Downloads\新增資料夾\測試.mp4)"); // UTF8
                data = CommonUtils::toUtf16LE(str).toHex().toUpper();
            }

            QByteArray filePath = CommonUtils::toUtf8(QByteArray::fromHex(data));
            qInfo() << filePath.constData();
        }
    }

    void test_file_path_parse()
    {
        {
            QVERIFY(CommonUtils::getFileNameByPath("D:/aa/bb/name.txt") == "name.txt");
            QVERIFY(CommonUtils::getFileNameByPath("D:\\aa\\测试\\測試.mp4") == "測試.mp4");
        }
    }

    void test_boost()
    {
        QVERIFY(boost::filesystem::exists(__FILE__) == true);
        boost::container::stable_vector<std::string> test_vec;
        test_vec.push_back("aa");
        test_vec.push_back("bb");
        test_vec.push_back("cc");
        QVERIFY(test_vec.size() == 3);
        QVERIFY(test_vec.front() == "aa");
        auto itr = test_vec.nth(1);
        QVERIFY(*itr == "bb");
        auto itr_last = test_vec.nth(2);
        QVERIFY(*itr_last == "cc");
        test_vec.erase(itr);
        QVERIFY(*itr_last == "cc");
        for (int index = 0; index < 10000; ++index) {
            test_vec.insert(test_vec.begin(), std::to_string(index + 1));
        }
        QVERIFY(*itr_last == "cc");
    }

    void test_edid_parse()
    {
        qInfo() << "-------------------------------------------";
        MonitorInfoUtils utils;
        for (const auto &info : utils.parseAllEdidInfo()) {
            nlohmann::json infoJson;
            infoJson["manufacturer"] = info.manufacturer;
            infoJson["productName"] = info.productName;
            if (info.productSerialNumber.empty() == false) {
                infoJson["productSerialNumber"] = info.productSerialNumber;
            }
            infoJson["edidVersion"] = info.edidVersion;
            if (info.gamma.has_value()) {
                infoJson["gamma"] = info.gamma.value();
            }

            if (info.manufactureWeek.has_value()) {
                infoJson["manufactureDate"] = QString("year %1 (week %2)").arg(info.manufactureYear).arg(info.manufactureWeek.value()).toStdString();
            } else {
                infoJson["manufactureDate"] = QString("year %1").arg(info.manufactureYear).toStdString();
            }
            infoJson["maxHSize"] = std::to_string(info.maxHSize) + "cm";
            infoJson["maxVSize"] = std::to_string(info.maxVSize) + "cm";
            infoJson["size"] = QString::asprintf("%.2f inch", info.size).toStdString();

            qInfo() << infoJson.dump(4).c_str();
        }

        nlohmann::json vcpCodeInfoJson;
        for (const auto &vcpCode : { 0xCC, 0x02, 0xFD }) {
            nlohmann::json infoJson;
            infoJson["vcpCode"] = static_cast<int>(vcpCode);
            uint32_t currentVal = 0;
            uint32_t maxVal = 0;
            if (MonitorController().control(vcpCode, nullptr, &currentVal, &maxVal)) {
                infoJson["currentValue"] = currentVal;
                infoJson["maxValue"] = maxVal;
            }
            vcpCodeInfoJson["infoArry"].push_back(infoJson);
        }
        qInfo() << vcpCodeInfoJson.dump(4).c_str();
    }
};

QTEST_GUILESS_MAIN(TestQString)

#include "test_qstring.moc"

