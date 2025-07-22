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
    QString m_dbName { CommonUtils::desktopDirectoryPath() + "/cross_share_client_test.db" };

    FileOperationRecordContainer m_container;

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
};

QTEST_GUILESS_MAIN(TestSqliteDB)

#include "test_sqlite_db.moc"

