package com.rtk.myapplication;

import android.accessibilityservice.AccessibilityService;
import android.app.Activity;
import android.app.Application;
import android.content.ClipData;
import android.os.Bundle;
import android.os.Handler;
import android.util.Log;
import android.view.KeyEvent;
import android.view.View;
import android.view.accessibility.AccessibilityEvent;
import android.os.Build;
import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.content.Intent;
import android.app.PendingIntent;
import android.content.Context;
import android.app.Notification;
import android.content.pm.ServiceInfo;
import android.content.ClipboardManager;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

public class TestService extends AccessibilityService  implements View.OnKeyListener{

    private static final String TAG = "lszz";
    boolean isNullData;
    private int testCount = 0;
    ClipData mclipData;
    private Notification notification;


    private ClipboardManager clipboardManager;
    private ClipboardManager.OnPrimaryClipChangedListener clipChangedListener;

    @Override
    public void onAccessibilityEvent(AccessibilityEvent accessibilityEvent) {
        Log.d(TAG,"onAccessibilityEvent 11");



        if (accessibilityEvent.getEventType() == AccessibilityEvent.TYPE_WINDOW_STATE_CHANGED) {
            // 检查是否是我们关心的窗口
            Log.d(TAG,"START");
            if (accessibilityEvent.getClassName().equals("com.example.MyActivity")) {
                // 执行操作
            }

        }
        if (accessibilityEvent.getEventType() == AccessibilityEvent.TYPE_VIEW_TEXT_CHANGED
        || accessibilityEvent.getEventType() == AccessibilityEvent.TYPE_VIEW_TEXT_SELECTION_CHANGED) {

            Log.d(TAG,"TYPE_VIEW_TEXT_CHANGED or TYPE_VIEW_TEXT_SELECTION_CHANGED ");
        }

    }

    @Override
    public void onInterrupt() {

    }

    @Override
    public void onCreate() {
        super.onCreate();
        Log.d(TAG,"lsz onCreate service!!!!!ab");


       // createNotification();

        setClibMessageLoop();

        //ClipboardManager clipboardManager = (ClipboardManager) getSystemService(Context.CLIPBOARD_SERVICE);
        //mclipData = clipboardManager.getPrimaryClip();

        clipboardManager = (ClipboardManager) getSystemService(Context.CLIPBOARD_SERVICE);
        clipChangedListener = new ClipboardManager.OnPrimaryClipChangedListener() {
            @Override
            public void onPrimaryClipChanged() {
                ClipData clipData = clipboardManager.getPrimaryClip();
                ClipData.Item item = clipData.getItemAt(0);
                String clipContent = item.getText().toString();
                // 处理剪切板内容
                // ...
                Log.d(TAG,"AAAAAb+====get"+clipContent);
            }
        };

        clipboardManager.addPrimaryClipChangedListener(clipChangedListener);
    }


    private void setClibMessageLoop() {
        new Handler().postDelayed(new Runnable() {
            @Override
            public void run() {
                setClibMessage();


                setClibMessageLoop();


            }
        }, 3000);
    }

    private void setClibMessage() {
        ClipData clipData = ClipData.newPlainText(null, "编辑后的文本数据+" + testCount);
        clipboardManager.setPrimaryClip(clipData);
        Log.d(TAG,"set AAAAAb+"+testCount);
        testCount++;


    }

    private void createNotification() {
        // 创建通知渠道（API 26+）


        // 启动前台服务
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            startForeground(1, notification, ServiceInfo.FOREGROUND_SERVICE_TYPE_SPECIAL_USE);
        }else{
            startForeground(1,notification);
        }
    }


    @Override
    protected boolean onKeyEvent(KeyEvent event) {
        Log.d(TAG,"lsz kecode=="+event.getKeyCode());

        int keyCode = event.getKeyCode();
        int action = event.getAction();
        if (keyCode == KeyEvent.KEYCODE_VOLUME_UP || keyCode == KeyEvent.KEYCODE_VOLUME_DOWN) {

            Log.d(TAG,"lsz kecode=="+event.getKeyCode());
        }
        return super.onKeyEvent(event);
    }


/*
    @Override
    public void onActivityStarted(@NonNull Activity activity) {
        Log.i(TAG, " lsz listenClipboard:  onActivityStarted ");
        if(mclipData == null) {
            isNullData = true;
            // 这里获取剪贴板数据
            if (mclipData != null && mclipData.getItemCount() > 0) {
                ClipData.Item item = mclipData.getItemAt(0);
                String text = item.getText().toString();
                Log.i(TAG, " lsz listenClipboard: " + text);
//                updateTextView(text);
            }

        } else {
            isNullData = false;
        }
    }
*/



    @Override
    public boolean onKey(View view, int i, KeyEvent keyEvent) {

        int acode = keyEvent.getKeyCode();
        Log.d("lsz","acode=="+acode);
        return false;
    }
}
