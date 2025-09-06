package com.realtek.crossshare;

import android.app.Application;
import android.util.Log;

import android.content.Context;
import com.tencent.mmkv.MMKV;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

public class MyApplication extends Application {

    private static Context context;
    private static boolean isDialogShown = false;
    private static FileTransferAdapter myAdapter;
    private static List<FileTransferItem> fileTransferList;
    private List<FileTransferItem> allFileItems = new ArrayList<>();
    public List<FileTransferItem> getAllFileItems() { return allFileItems; }
    public void setAllFileItems(List<FileTransferItem> list) { allFileItems = list; }
    private static List<Server> serverList = new ArrayList<>();
    private static List<Device> deviceList = new ArrayList<>();
    private static Map<String, String> deviceNameIdMap = new HashMap<>();
    private static List<String> filePaths = new ArrayList<>();

    public static List<String> getFilePaths() {
        return filePaths;
    }
    public static void setFilePaths(List<String> list) {
        filePaths = list;
    }

    @Override
    public void onCreate() {
        super.onCreate();

        context=this;
        String rootDir = MMKV.initialize(this);
        LogUtils.init(context);
        myAdapter = new FileTransferAdapter(fileTransferList, new FileTransferAdapter.OnItemClickListener() {
            @Override
            public void onDeleteClick(int position) {

            }

            @Override
            public void onCancelClick(boolean isallfile, String filename) {

            }

            @Override
            public void onOpenFileClick(boolean isallfile,String filepath) {

            }

        });
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
    public static List<Server> getServerList() {
        return serverList;
    }
    public static void setServerList(List<Server> list) {
        Map<String, Server> uniqueMap = new LinkedHashMap<>();
        if (list != null) {
            for (Server s : list) {
                uniqueMap.put(s.instance, s);
            }
        }
        serverList.clear();
        serverList.addAll(uniqueMap.values());
        LogUtils.i("MyApplication","setServerList serverList.size()"+serverList.size());
    }
    public static void clearServerList() {
        serverList.clear();
    }

    public static List<Device> getDeviceList() {
        return deviceList;
    }
    public static void setDeviceList(List<Device> list) {
        deviceList = list;
    }
    public static Map<String, String> getDeviceNameIdMap() {
        return deviceNameIdMap;
    }
    public static void setDeviceNameIdMap(Map<String, String> map) {
        deviceNameIdMap = map;
    }
}
