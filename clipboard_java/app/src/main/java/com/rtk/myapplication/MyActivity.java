package com.rtk.myapplication;

import android.app.Activity;
import android.content.ComponentName;
import android.content.Context;
import android.content.Intent;
import android.content.ServiceConnection;
import android.graphics.Bitmap;
import android.os.Bundle;
import android.os.IBinder;
import android.util.Log;

import androidx.annotation.Nullable;
import androidx.recyclerview.widget.LinearLayoutManager;
import androidx.recyclerview.widget.RecyclerView;

import java.util.ArrayList;
import java.util.Calendar;
import java.util.Date;
import java.util.List;

public class MyActivity extends Activity {

    private FileTransferAdapter adapter;
    private List<FileTransferItem> fileTransferList = new ArrayList<>();
    private FloatClipboardService myService;
    private boolean isBound = false;

    String filename;



    private ServiceConnection connection2 = new ServiceConnection() {

        @Override
        public void onServiceConnected(ComponentName className, IBinder service) {
            FloatClipboardService.LocalBinder binder = (FloatClipboardService.LocalBinder) service;
            myService = binder.getService();
            isBound = true;

            // 设置回调
            myService.setCallback(new FloatClipboardService.DataCallback() {
                @Override
                public void onDataReceived(double data) {
                    // 在UI线程更新UI
                    runOnUiThread(() -> {
                        // 更新UI操作
                        Log.i("lsz","ServiceConnection Myactivity data="+data);

                         adapter.updateProgress(filename, (int)data);
                    });
                }

                @Override
                public void onBitmapReceived(Bitmap bitmap,String path) {

                      adapter.setBitmap(filename,bitmap);
                }
                @Override
                public void onCallbackMethodFileDone(String path) {
                }
            });
        }

        @Override
        public void onServiceDisconnected(ComponentName arg0) {
            isBound = false;
        }
    };


    @Override
    protected void onCreate(@Nullable Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        /*setContentView(R.layout.myactivity);


        RecyclerView recyclerView2 = findViewById(R.id.recycler_view);
        recyclerView2.setLayoutManager(new LinearLayoutManager(this));
        adapter = new FileTransferAdapter(fileTransferList);
        recyclerView2.setAdapter(adapter);


        boolean booleanValue = getIntent().getBooleanExtra("booleanKey", false); // 第二个参数是默认值，如果没找到键则使用默认值
        filename = getIntent().getStringExtra("filename");
        long filesize = getIntent().getLongExtra("filesize", -1L);
        String bitmappath = getIntent().getStringExtra("bitmappath");
        //countSize=filesize;
        Log.d("lszz", "lsz booleanValue booleanValue=" + booleanValue);
        Log.d("lszz", "lsz filename filename=" + filename);
        Log.d("lszz", "long filesize=" + filesize);
        Log.d("lszz", "String bitmappath=" + bitmappath);

        Intent intent = new Intent(MyApplication.getContext(), FloatClipboardService.class);
        bindService(intent, connection2, Context.BIND_AUTO_CREATE);

        if (booleanValue) {

            FileTransferItem item = new FileTransferItem(filename, filesize, BitmapHolder.getBitmap());
            fileTransferList.add(item);
            adapter.notifyItemInserted(fileTransferList.size() - 1);


        }*/
    }

}
