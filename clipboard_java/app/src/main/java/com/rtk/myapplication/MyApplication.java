package com.rtk.myapplication;

import android.app.Application;
import android.util.Log;

import android.content.Context;
import com.tencent.mmkv.MMKV;

public class MyApplication extends Application {

    private static Context context;

    @Override
    public void onCreate() {
        super.onCreate();

        context=this;
        String rootDir = MMKV.initialize(this);
        Log.i("MMKV", "mmkv root: " + rootDir);
    }

    public static Context getContext() {
        return context;
    }
}
