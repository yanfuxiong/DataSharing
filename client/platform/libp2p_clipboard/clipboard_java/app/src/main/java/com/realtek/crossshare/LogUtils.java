package com.realtek.crossshare;

import android.content.Context;
import android.util.Log;
import java.io.File;
import java.io.FileWriter;
import java.io.IOException;
import java.text.SimpleDateFormat;
import java.util.Calendar;
import java.util.Date;

public class LogUtils {
    private static Context appContext;

    public static void init(Context context) {
        appContext = context.getApplicationContext();
        deleteOldLogs(3);
    }

    private static File getLogDir() {
        File logDir = new File(appContext.getExternalFilesDir(null), "Log");
        if (!logDir.exists()) {
            logDir.mkdirs();
        }
        return logDir;
    }

    private static File getLogFile() {
        String today = new SimpleDateFormat("yyyy-MM-dd").format(new Date());
        File logDir = getLogDir();
        return new File(logDir, "crossshare_app_log_" + today + ".txt");
    }

    private static void deleteOldLogs(int keepDays) {
        File logDir = getLogDir();
        SimpleDateFormat sdf = new SimpleDateFormat("yyyy-MM-dd");
        Calendar cal = Calendar.getInstance();
        cal.add(Calendar.DAY_OF_YEAR, -keepDays);
        String deleteDay = sdf.format(cal.getTime());
        File oldFile = new File(logDir, "crossshare_app_log_" + deleteDay + ".txt");
        if (oldFile.exists()) {
            oldFile.delete();
        }
    }

    private static void writeToFile(String msg) {
        File logFile = getLogFile();
        String time = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss").format(new Date());
        String logLine = "[" + time + "] " + msg;
        try (FileWriter writer = new FileWriter(logFile, true)) {
            writer.write(logLine + "\n");
        } catch (IOException e) {
            Log.e("LogUtils", "writeToFile err：" + e.getMessage());
        }
    }

    public static void i(String tag, String msg) {
        Log.i(tag, msg);
        writeToFile("[INFO] " + tag + ": " + msg);
    }

    public static void d(String tag, String msg) {
        Log.d(tag, msg);
        writeToFile("[DEBUG] " + tag + ": " + msg);
    }

    public static void e(String tag, String msg) {
        Log.e(tag, msg);
        writeToFile("[ERROR] " + tag + ": " + msg);
    }
}