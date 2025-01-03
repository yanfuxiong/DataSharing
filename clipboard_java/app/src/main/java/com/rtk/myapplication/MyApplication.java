package com.rtk.myapplication;

import android.app.Application;
import android.util.Log;

import android.content.Context;
import com.tencent.mmkv.MMKV;

import java.util.List;

public class MyApplication extends Application {

    private static Context context;
    private static boolean isDialogShown = false;
    private static FileTransferAdapter myAdapter;
    private static List<FileTransferItem> fileTransferList;

    @Override
    public void onCreate() {
        super.onCreate();

        context=this;
        String rootDir = MMKV.initialize(this);
        myAdapter = new FileTransferAdapter(fileTransferList);
        Log.i("MMKV", "mmkv root: " + rootDir);
    }

    public static Context getContext() {
        return context;
    }

    public static synchronized boolean isDialogShown() {
        return isDialogShown;
    }

    public static synchronized void setDialogShown(boolean shown) {
        isDialogShown = shown;
    }

    public static void setMyAdapter(FileTransferAdapter Adapter) {
         myAdapter = Adapter ;
    }

    public static FileTransferAdapter getMyAdapter() {
        return myAdapter;
    }
}
