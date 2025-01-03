#include <QTest>
#include "common_utils.h"
#include "worker_thread.h"
#include <QSqlDatabase>
#include <QSqlQuery>
#include <QSqlRecord>
#include <QSqlError>
#include <boost/filesystem.hpp>

class TestSqliteDB : public QObject
{
    Q_OBJECT

private:
    QSqlDatabase m_db;
    QString m_dbName { R"(C:\Users\TS\Desktop\cross_share_client_test.db)" };

    FileOperationRecordContainer m_container;
    //QString m_dbName { R"(C:\Users\TS\Desktop\cross_share_client.db)" };

private Q_SLOTS:
    void initTestCase()
    {
        qInfo() << QSqlDatabase::drivers();
        m_db = QSqlDatabase::addDatabase("QSQLITE", SQLITE_CONN_NAME);
        m_db.setDatabaseName(m_dbName);
        QVERIFY(m_db.open() == true);

        {
            QSqlQuery query(m_db);
            QVERIFY(query.exec(QString(g_drop_table_sql).arg("opt_record")) == true);
            QVERIFY(query.exec(g_create_opt_record) == true);
            //QVERIFY(query.exec(sql_new) == true);
        }
    }
    void cleanupTestCase()
    {

    }

    QString getSqlFilePath() const
    {
        return qApp->applicationDirPath() + "/cross_share_client.sql.txt";
    }

    void testWriteData()
    {
        {
            QString sql = QString("INSERT INTO opt_record (file_name, file_size, timestamp, progress_value, client_name, client_id, ip, direction, uuid) "
                                  "VALUES('%1', '%2', '%3', '%4', '%5', '%6', '%7', '%8', '%9')")
                                  .arg(QDir::fromNativeSeparators(__FILE__))
                                  .arg(boost::filesystem::file_size(__FILE__))
                                  .arg(QDateTime::currentDateTime().toMSecsSinceEpoch())
                                  .arg(100)
                                  .arg("测试客户端_1")
                                  .arg("QmQ7obXFx1XMFr6hCYXtovn9zREFqSXEtH5hdtpBDLjrAz")
                                  .arg("127.0.0.1")
                                  .arg(1)
                                  .arg(CommonUtils::createUuid())
                                  ;
            qInfo() << sql.toUtf8().constData();
            QSqlQuery query(m_db);
            QVERIFY(query.exec(sql) == true);
        }

        QThread::msleep(100);

        {
            QString sql = QString("INSERT INTO opt_record (file_name, file_size, timestamp, progress_value, client_name, client_id, ip, direction, uuid) "
                                  "VALUES('%1', '%2', '%3', '%4', '%5', '%6', '%7', '%8', '%9')")
                                  .arg(QDir::fromNativeSeparators(__FILE__))
                                  .arg(boost::filesystem::file_size(__FILE__))
                                  .arg(QDateTime::currentDateTime().toMSecsSinceEpoch())
                                  .arg(99)
                                  .arg("测试客户端_2")
                                  .arg("QmQ7obXFx1XMFr6hCYXtovn9zREFqSXEtH5hdtpBDLjrAz")
                                  .arg("192.168.0.1")
                                  .arg(0)
                                  .arg(CommonUtils::createUuid())
                                  ;
            qInfo() << sql.toUtf8().constData();
            QSqlQuery query(m_db);
            QVERIFY(query.exec(sql) == true);
        }

        {
            QSqlQuery query(m_db);
            QString sql = QString("SELECT file_name, file_size, timestamp, progress_value, client_name, client_id, ip, direction, uuid FROM opt_record");
            query.exec(sql);
            while (query.next()) {
                const auto &record = query.record();
                qInfo() << ">>>>>> " << record.value(0).toString().toUtf8().constData();

                FileOperationRecord optRecord;
                optRecord.fileName = record.value(0).toString().toStdString();
                optRecord.fileSize = record.value(1).toULongLong();
                optRecord.timeStamp = record.value(2).toULongLong();
                optRecord.progressValue = record.value(3).toInt();
                optRecord.clientName = record.value(4).toString().toStdString();
                optRecord.clientID = record.value(5).toString().toStdString();
                optRecord.ip = record.value(6).toString();
                optRecord.direction = record.value(7).toInt();
                optRecord.uuid = record.value(8).toString();

                m_container.push_back(optRecord);
                QVERIFY(optRecord.toRecordData().getHashID() == QByteArray::number(optRecord.toStdHashID()));
            }
        }
    }

    void testContainer()
    {
        const auto &tm_index = m_container.get<tag_db_timestamp>();
        QVERIFY(tm_index.size() == 2);
        for (auto itr = tm_index.begin(); itr != tm_index.end(); ++itr) {
            qInfo() << itr->fileName.c_str() << " => " << QDateTime::fromMSecsSinceEpoch(itr->timeStamp).toString("yyyy-MM-dd hh:mm:ss.zzz").toUtf8().constData();
        }

        QVERIFY(tm_index.begin()->progressValue == 99);
        QVERIFY(tm_index.begin()->direction == 0);
        QVERIFY(tm_index.begin()->clientName == "测试客户端_2");

        qInfo() << CommonUtils::localDataDirectory();
        qInfo() << CommonUtils::localIpAddress();
    }
};

QTEST_GUILESS_MAIN(TestSqliteDB)

#include "test_sqlite_db.moc"

