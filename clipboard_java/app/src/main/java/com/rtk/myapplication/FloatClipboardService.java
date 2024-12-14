package com.rtk.myapplication;

import android.app.AlertDialog;
import android.app.Notification;
import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.app.PendingIntent;
import android.app.Service;
import android.content.ClipData;
import android.content.ClipboardManager;
import android.content.Context;
import android.content.Intent;
import android.content.pm.ServiceInfo;
import android.content.res.Configuration;
import android.graphics.Bitmap;
import android.graphics.BitmapFactory;
import android.graphics.Color;
import android.graphics.PixelFormat;
import android.net.ConnectivityManager;
import android.net.LinkAddress;
import android.net.LinkProperties;
import android.net.Uri;
import android.net.wifi.WifiManager;
import android.os.Binder;
import android.os.Build;
import android.os.Environment;
import android.os.Handler;
import android.os.IBinder;
import android.os.Looper;
import android.provider.Settings;
import android.text.format.Formatter;
import android.util.Base64;
import android.util.Log;
import android.view.Gravity;
import android.view.LayoutInflater;
import android.view.MotionEvent;
import android.view.View;
import android.view.WindowManager;
import android.widget.Button;
import android.widget.CheckBox;
import android.widget.CompoundButton;
import android.widget.ImageView;
import android.widget.ProgressBar;
import android.widget.TextView;
import android.widget.Toast;
import android.graphics.Point;
import android.os.Message;
import android.view.Display;

import java.lang.reflect.InvocationTargetException;

import androidx.core.content.FileProvider;
import androidx.localbroadcastmanager.content.LocalBroadcastManager;

import com.tencent.mmkv.MMKV;

import java.io.ByteArrayOutputStream;
import java.io.File;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.Inet4Address;
import java.net.Inet6Address;
import java.net.NetworkInterface;
import java.net.ServerSocket;
import java.net.SocketException;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.util.Enumeration;
import java.util.List;
import java.util.concurrent.atomic.AtomicReference;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

import libp2p_clipboard.Callback;
import libp2p_clipboard.Libp2p_clipboard;

public class FloatClipboardService extends Service {

    private static final String TAG = FloatClipboardService.class.getSimpleName();
    private static final String SETTINGS_DEBUG_FLOAT_WINDOW = "debug_float_window";
    private WindowManager windowManager;
    private WindowManager windowManager2;
    private View floatView;
    private View floatViewTypec;
    private View floatViewForAdjust;
    private ClipboardManager clipboardManager;
    private Notification notification;
    private WindowManager.LayoutParams params;
    private WindowManager.LayoutParams paramsForFloatviewTypec;
    private WindowManager.LayoutParams paramsForAdjust;
    private int testCount = 0;
    private float initialX, initialY;

    String previousText = "";
    String previousImgMD5 = "";
    ImageView imageView;
    MMKV kv;
    boolean boxischeck;
    long countSize;
    long countSizebuf;
    double countbuf;
    ProgressBar progressBar;
    String filename;
    byte[] bitmapData;
    AlertDialog dialog;
    boolean mIsDebugfloatWindow = false;
    long filesize;
    String getnetname;
    int getindex;
    boolean mfloatViewTypecColor = true;
    int mAdjustViewX, mAdjustViewY;

    private final IBinder binder = new LocalBinder();

    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        return START_STICKY;
    }

    public class LocalBinder extends Binder {
        FloatClipboardService getService() {
            return FloatClipboardService.this;
        }
    }

    public interface DataCallback {
        void onDataReceived(double data);

        void onBitmapReceived(Bitmap bitmap, String path);
        void onCallbackMethodFileDone(String path);
    }

    private DataCallback callback;

    public void setCallback(DataCallback callback) {
        this.callback = callback;
    }

    public void sendData(double data) {
        if (callback != null) {
            callback.onDataReceived(data);
        }
    }

    public void onBitmapReceived(Bitmap bitmap, String path) {
        if (callback != null) {
            callback.onBitmapReceived(bitmap, path);
        }
    }

    public void onCallbackMethodFileDone(String path) {
        if (callback != null) {
            callback.onCallbackMethodFileDone(path);
        }
    }


    @Override
    public void onCreate() {
        super.onCreate();
        clipboardManager = (ClipboardManager) getSystemService(Context.CLIPBOARD_SERVICE);
        initService();

        //lsz add test
        //setClibMessageLoop();

        // testClipboardUtils();
        kv = MMKV.defaultMMKV();
        boxischeck = kv.decodeBool("ischeck", false);

        long port = findFreePort();
        Log.d(TAG, "get wifiip==" + getWifiIpAddress(MyApplication.getContext()));
        Log.d(TAG, "get GoLog findFreePort===" + port);
        new Thread(new Runnable() {
            @Override
            public void run() {
                try {
                    getNetInfoFromLocalIp();
                    Log.i(TAG, "getnetname=" + getnetname + "/ getindex=" + getindex);
                    Libp2p_clipboard.sendNetInterfaces(getnetname, getindex);
                } catch (SocketException e) {
                    throw new RuntimeException(e);
                }
                Libp2p_clipboard.sendAddrsFromJava(getIpAddres());

                Libp2p_clipboard.mainInit(getGolangCallBack(), getWifiIpAddress(MyApplication.getContext()), "aaa", getWifiIpAddress(MyApplication.getContext()), port);
            }
        }).start();


    }


    public void testClipboardUtils() {

        ClipboardUtils.setContext(getApplication());
        ClipboardUtils clipboardUtils = ClipboardUtils.getInstance();
        //Log.d("lsz","GoLog clipboardUtils.hasClip()=s=="+clipboardUtils.hasClip());
        if (clipboardUtils.hasClip()) {
            getClipFromClipboard();
        } else {
            Toast.makeText(this, "lsz get Clipboard is empty", Toast.LENGTH_SHORT).show();
        }

    }

    @Override
    public void onConfigurationChanged(Configuration newConfig) {
        Log.d(TAG, "onConfigurationChanged() newConfig: " + newConfig);
        super.onConfigurationChanged(newConfig);
        mCheckBoundshandler.sendEmptyMessageDelayed(0, 500);
    }

    //String a = copyFileToPublicDir("aa");
    private void initService() {
        if (Settings.Global.getInt(MyApplication.getContext().getContentResolver(), SETTINGS_DEBUG_FLOAT_WINDOW, 0) == 1) {
            mIsDebugfloatWindow = true;
        }
        Log.d(TAG, "mIsDebugfloatWindow:" + mIsDebugfloatWindow);

        // 创建悬浮窗视图
        windowManager = (WindowManager) getApplicationContext().getSystemService(WINDOW_SERVICE);
        if (mIsDebugfloatWindow) {
            LayoutInflater inflater = (LayoutInflater) getSystemService(LAYOUT_INFLATER_SERVICE);
            floatView = inflater.inflate(R.layout.float_window, null);
        } else {
            floatView = new View(this);
            floatView.setBackgroundColor(Color.TRANSPARENT);
        }

        floatView.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                Log.i(TAG, "lsz floatView onClick, copy");
                onFloatviewClicked();
            }
        });

        if (mIsDebugfloatWindow) {
            params = new WindowManager.LayoutParams(
                    WindowManager.LayoutParams.WRAP_CONTENT,
                    WindowManager.LayoutParams.WRAP_CONTENT,
                    WindowManager.LayoutParams.TYPE_APPLICATION_OVERLAY,
                    WindowManager.LayoutParams.FLAG_NOT_FOCUSABLE
                            | WindowManager.LayoutParams.FLAG_NOT_TOUCH_MODAL,
                    PixelFormat.TRANSLUCENT);
            params.gravity = Gravity.TOP | Gravity.START;
            params.x = 0;
            params.y = 100;
            windowManager.addView(floatView, params);
        } else {
            params = new WindowManager.LayoutParams();
            params.type = WindowManager.LayoutParams.TYPE_APPLICATION_OVERLAY;
            params.flags = WindowManager.LayoutParams.FLAG_NOT_TOUCH_MODAL
                    | WindowManager.LayoutParams.FLAG_LAYOUT_NO_LIMITS
                    | WindowManager.LayoutParams.FLAG_NOT_FOCUSABLE;
            params.format = PixelFormat.TRANSLUCENT;

            params.gravity = Gravity.CENTER;
            params.x = 0;
            params.y = 0;
            params.height = 6;
            params.width = 6;
            windowManager.addView(floatView, params);
        }

        floatViewForAdjust = new View(this);
        floatViewForAdjust.setBackgroundColor(Color.TRANSPARENT);
        paramsForAdjust = new WindowManager.LayoutParams();
        paramsForAdjust.type = WindowManager.LayoutParams.TYPE_APPLICATION_OVERLAY;
        paramsForAdjust.flags = WindowManager.LayoutParams.FLAG_NOT_TOUCH_MODAL
                | WindowManager.LayoutParams.FLAG_LAYOUT_NO_LIMITS
                | WindowManager.LayoutParams.FLAG_NOT_FOCUSABLE;
        paramsForAdjust.format = PixelFormat.TRANSLUCENT;
        paramsForAdjust.gravity = Gravity.CENTER;
        paramsForAdjust.x = 0;
        paramsForAdjust.y = 0;
        paramsForAdjust.height = 4;
        paramsForAdjust.width = 4;
        windowManager.addView(floatViewForAdjust, paramsForAdjust);
        mCheckBoundshandler.sendEmptyMessageDelayed(0, 500);

        //floatviewtypec
        floatViewTypec = new View(this);
        floatViewTypec.setBackgroundColor(Color.TRANSPARENT);
        floatViewTypec.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                Log.i(TAG, "lsz floatViewTypec onClick 4*4, copy");
                onFloatviewClicked();
            }
        });

        paramsForFloatviewTypec = new WindowManager.LayoutParams();
        paramsForFloatviewTypec.type = WindowManager.LayoutParams.TYPE_APPLICATION_OVERLAY;
        paramsForFloatviewTypec.flags = WindowManager.LayoutParams.FLAG_NOT_TOUCH_MODAL
                | WindowManager.LayoutParams.FLAG_LAYOUT_NO_LIMITS
                | WindowManager.LayoutParams.FLAG_NOT_FOCUSABLE;
        paramsForFloatviewTypec.format = PixelFormat.TRANSLUCENT;
        paramsForFloatviewTypec.gravity = Gravity.TOP | Gravity.LEFT;
        Point realSize = getDefaultDisplay();
        paramsForFloatviewTypec.x = 0;
        paramsForFloatviewTypec.y = 0;
        Log.d(TAG, "paramsForFloatviewTypec x: "+ paramsForFloatviewTypec.x+ ", paramsForFloatviewTypec.y:"+paramsForFloatviewTypec.y);
        paramsForFloatviewTypec.width = 4;
        paramsForFloatviewTypec.height = 4;
        windowManager.addView(floatViewTypec, paramsForFloatviewTypec);

        if (mIsDebugfloatWindow) {
            // 设置触摸事件监听
            floatView.setOnTouchListener(new View.OnTouchListener() {
                @Override
                public boolean onTouch(View v, MotionEvent event) {
                    switch (event.getAction()) {
                        case MotionEvent.ACTION_DOWN:
                            // 记录初始坐标
                            initialX = event.getRawX() - params.x;
                            initialY = event.getRawY() - params.y;
                            break;
                        case MotionEvent.ACTION_MOVE:
                            // 更新悬浮窗位置
                            params.x = (int) (event.getRawX() - initialX);
                            params.y = (int) (event.getRawY() - initialY);
                            windowManager.updateViewLayout(v, params);
                            break;
                        case MotionEvent.ACTION_UP:
                            // 可以考虑在这里处理抬起手指后的操作，例如隐藏悬浮窗
                            break;
                    }
                    return false;
                }
            });

        }
        // 创建通知
        createNotification();

        // 监听剪贴板变化


    }

    private void onFloatviewClicked() {
        //floatView.requestFocus();
        //floatView.setFocusable(true);
        updateFocus(true);
        //点击悬浮窗之后，先设置交点获取剪切版内容，之后失去焦点
        new Handler(Looper.getMainLooper()).postDelayed(new Runnable() {
            @Override
            public void run() {
                getClipboard();
            }
        }, 100);

        new Handler(Looper.getMainLooper()).postDelayed(new Runnable() {
            @Override
            public void run() {
                updateFocus(false);
            }
        }, 600);
    }

    private String getIpAddres() {
        ConnectivityManager connectivityManager = (ConnectivityManager) getApplicationContext().getSystemService(Service.CONNECTIVITY_SERVICE);
        LinkProperties linkProperties = connectivityManager.getLinkProperties(connectivityManager.getActiveNetwork());
        List<LinkAddress> addressList = linkProperties.getLinkAddresses();
        StringBuffer sbf = new StringBuffer();
        for (LinkAddress linkAddress : addressList) {
            sbf.append(linkAddress.toString()).append("#");
        }
        Log.d(TAG, "getIpAddres: " + sbf.toString());
        return sbf.toString();
    }

    private void updateFocus(boolean focusable) {
        if (focusable) {
            params.flags = WindowManager.LayoutParams.FLAG_NOT_TOUCH_MODAL
                    | WindowManager.LayoutParams.FLAG_LAYOUT_NO_LIMITS;
            windowManager.updateViewLayout(floatView, params);
        } else {
            params.flags = WindowManager.LayoutParams.FLAG_NOT_FOCUSABLE
                    | WindowManager.LayoutParams.FLAG_NOT_TOUCH_MODAL
                    | WindowManager.LayoutParams.FLAG_LAYOUT_NO_LIMITS;
            windowManager.updateViewLayout(floatView, params);
        }
    }

    private void createNotification() {
        // 创建通知渠道（API 26+）
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            CharSequence name = "Float Clipboard Service";
            String channelId = "float_clipboard_channel_id";
            NotificationChannel channel = new NotificationChannel(channelId, name, NotificationManager.IMPORTANCE_LOW);
            NotificationManager notificationManager = (NotificationManager) getSystemService(Context.NOTIFICATION_SERVICE);
            notificationManager.createNotificationChannel(channel);
        }

        // 创建通知
        Intent intent = new Intent(this, FloatClipboardService.class);
        PendingIntent pendingIntent = PendingIntent.getActivity(this, 0, intent, PendingIntent.FLAG_MUTABLE);
        notification = new Notification.Builder(this, "float_clipboard_channel_id")
                .setContentTitle("剪贴板悬浮窗服务")
                .setContentText("点击打开应用")
                .setSmallIcon(R.drawable.ic_launcher_foreground)
                .setContentIntent(pendingIntent)
                .build();

        // 启动前台服务
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            startForeground(1, notification, ServiceInfo.FOREGROUND_SERVICE_TYPE_SPECIAL_USE);
        } else {
            startForeground(1, notification);
        }
    }


    private void getClibMessage() {
        /*ClipData clip = clipboardManager.getPrimaryClip();
        Log.i(TAG, "lsz getClibMessage: clip=" + clip);
        if (clip != null && clip.getItemCount() > 0) {
            ClipData.Item item = clip.getItemAt(0);
            String text = item.getText().toString();
            Log.i(TAG, " 11 lsz getClibMessage: " + text);
            sendToPC(text);
            //    updateTextView(text);
        }*/
        // updateFocus(false);
    }

    private void sendToPC(String text) {

        new Handler().postDelayed(new Runnable() {
            @Override
            public void run() {
                String currentText = text;
                if (!currentText.equals(previousText)) {
                    Log.i(TAG, "lszz GoLog sendToPC: text:" + text);
                    Libp2p_clipboard.sendMessage(text);
                }
                previousText = currentText;
            }
        }, 100);


        /*Log.i(TAG, "sendToPC: text:" + text);
        new Thread(new Runnable() {
            @Override
            public void run() {
                Libp2p_clipboard.sendMessage(text);
            }
        }).start();*/
    }

    private void sendToPCIMG(byte[] value) {
        //Log.i(TAG, "sendToPC: img byte:" + value);
        /*new Thread(new Runnable() {
            @Override
            public void run() {
                String base64String = Base64.encodeToString(value, Base64.DEFAULT);
                String  clearbase64String=removeInvalidCharacters(base64String);
                Libp2p_clipboard.sendImage(clearbase64String);
            }
        }).start();*/

        new Handler().postDelayed(new Runnable() {
            @Override
            public void run() {
                String currentimgmd5;
                MessageDigest md = null;
                try {
                    md = MessageDigest.getInstance("MD5");
                } catch (NoSuchAlgorithmException e) {
                    throw new RuntimeException(e);
                }
                byte[] md5Bytes = md.digest(value);

                StringBuilder sb = new StringBuilder();
                for (byte b : md5Bytes) {
                    sb.append(String.format("%02X", b));
                }
                currentimgmd5 = sb.toString();
                if (!currentimgmd5.equals(previousImgMD5)) {
                    //Log.i(TAG, "lszz GoLog sendToPC: text:"+currentimgmd5 );
                    String base64String = Base64.encodeToString(value, Base64.DEFAULT);
                    String clearbase64String = removeInvalidCharacters(base64String);
                    Libp2p_clipboard.sendImage(clearbase64String);
                }

                previousImgMD5 = currentimgmd5;
            }
        }, 100);

    }

    private void setClibMessageLoop() {
        new Handler().postDelayed(new Runnable() {
            @Override
            public void run() {
                //setClibMessage();

                //  getClibMessage();

                //setClibMessageLoop();
                //testClipboardUtils();

            }
        }, 3000);
    }

    private void setClibMessage() {
        ClipData clipData = ClipData.newPlainText(null, "编辑后的文本数据+" + testCount);
        clipboardManager.setPrimaryClip(clipData);
        testCount++;
    }

    @Override
    public IBinder onBind(Intent intent) {
        Log.d(TAG, "lsz onBindonBindonBindonBind");
        return binder;
    }

    @Override
    public void onDestroy() {
        if (floatView != null) windowManager.removeView(floatView);
        if (notification != null) stopForeground(true);
        Log.d(TAG, "lsz onDestroy");

        super.onDestroy();
    }

    private final static int P2P_EVENT_SERVER_CONNEDTED = 0;
    private final static int P2P_EVENT_SERVER_CONNECT_FAIL = 1;
    private final static int P2P_EVENT_CLIENT_CONNEDTED = 2;
    private final static int P2P_EVENT_CLIENT_CONNECT_FAIL = 3;

    private Callback getGolangCallBack() {
        return new Callback() {
            @Override
            public void callbackMethod(String s) {
                Log.i(TAG, "lsz GoLog callmsg callbackMethod: callback调用 ==" + s);
                ClipData clipData = ClipData.newPlainText(null, s);
                if (!s.isEmpty()) {
                    Log.i(TAG, "lsz callbackMethod init");
                    clipboardManager.setPrimaryClip(clipData);
                }
            }

            @Override
            public void callbackMethodFileConfirm(String ipAddr ,String s, String name, long l) {
                Log.i(TAG, "lszz GoLog callbackMethodFileConfirm: amsg:String= " + s);
                Log.i(TAG, "lszz GoLog callbackMethodFileConfirm: amsg:long= " + l);
                boxischeck = kv.decodeBool("ischeck", false);

                countSize = l;
                new Handler(Looper.getMainLooper()).postDelayed(new Runnable() {
                    @Override
                    public void run() {
                        define_cancel_service(MyApplication.getContext(),ipAddr, s, name, l);
                        /*if (!boxischeck) {
                            Log.i(TAG, "CheckBox boxischeck======false");
                            define_cancel_service(MyApplication.getContext(), s, name, l);
                        } else {
                            Log.i(TAG, "CheckBox boxischeck======true");
                            Libp2p_clipboard.ifClipboardPasteFile(true);
                        }*/
                    }
                }, 100);


            }

            @Override
            public void callbackMethodFileDone(String s, long l) {
                Log.i(TAG, "callbackMethodFileDone: msg: " + s);
                new Handler(Looper.getMainLooper()).postDelayed(new Runnable() {
                    @Override
                    public void run() {
                        copyFileToPublicDir(s);
                        Bitmap mbitmap = getBitmap(s);
                        if (mbitmap != null) {
                            BitmapHolder.setBitmap(mbitmap);
                            onBitmapReceived(mbitmap, s);
                        }
                        countbuf = 0;
                        countSizebuf = 0;

                        onCallbackMethodFileDone(s);
                    }
                }, 100);


            }

            @Override
            public void callbackMethodFoundPeer() {
                Log.i(TAG, "lsz callbackMethodFoundPeer: ");
                new Handler(Looper.getMainLooper()).postDelayed(new Runnable() {
                    @Override
                    public void run() {
                        getClientList();
                    }
                }, 100);
            }

            @Override
            public void logMessageCallback(String msg) {
                Log.i(TAG, "logMessageCallback: msg: " + msg);
            }

            @Override
            public void callbackMethodImage(String msg) {
                Log.i(TAG, "lsz callmsg GoLog callbackMethodImage: msg: " + msg);

                new Handler(Looper.getMainLooper()).postDelayed(new Runnable() {
                    @Override
                    public void run() {
                        if (!msg.isEmpty()) {
                            Bitmap ba = base64ToBitmapa(msg);
                            //Log.i(TAG, "lszz GoLog ba.getHeight()" + ba.getHeight());
                            //Log.i(TAG, "lszz GoLog ba.getWidth()" + ba.getWidth());
                            //Log.i(TAG, "lszz GoLog ba.getByteCount()" + ba.getByteCount());
                            setBitmapToClipboard(MyApplication.getContext(), ba);
                            Toast.makeText(MyApplication.getContext(), "Image is saved to clipboard", Toast.LENGTH_SHORT).show();
                        }
                    }
                }, 100);


            }

            @Override
            public void callbackUpdateProgressBar(long l) {


                countSizebuf = l + countSizebuf;

                countbuf = (countSizebuf / (double) countSize) * 100;

                sendData(countbuf);


            }


            @Override
            public void eventCallback(long event) {
                Log.i(TAG, "eventCallBack: event2: " + event);
            }
        };
    }

    public void getImg(String msg) {
        byte[] value = Base64.decode(msg, Base64.DEFAULT);
        Log.d("lszzz", "GoLog sendToPC: img value:==" + value.length);
        Bitmap bitmap = BitmapFactory.decodeByteArray(value, 0, value.length);
        imageView.setImageBitmap(bitmap);
    }


    public static String removeInvalidCharacters(String base64String) {
        // 正则表达式，匹配Base64的有效字符
        String regex = "[^A-Za-z0-9+/=]";
        // 使用正则表达式替换掉非法字符
        String cleanString = base64String.replaceAll(regex, "");
        return cleanString;
    }


    private void getClipFromClipboard() {
//本地图片 取到剪切版
       /* ClipboardManager clipboard = (ClipboardManager) getSystemService(Context.CLIPBOARD_SERVICE);
        Log.d("lsz","clipa"+clipboard.hasPrimaryClip());
        // 檢查剪貼簿是否有內容
        if (clipboard.hasPrimaryClip()) {
            ClipData clip = clipboard.getPrimaryClip();
            Log.d("lsz","clip"+clip);
            // 檢查是否包含 URI 類型的資料
            if (clip != null && clip.getItemCount() > 0) {
                ClipData.Item item = clip.getItemAt(0);
                Uri imageUri = item.getUri();
                Log.d("lsz","clip imageUri"+imageUri);
                if (imageUri != null) {
                    // 將圖片 URI 設置到 ImageView 顯示圖片
                    imageView2.setImageURI(imageUri);
                    //imageView2.setImageBitmap(bitmap1);
                }
            }
        } else {        Toast.makeText(this, "Clipboard is empty", Toast.LENGTH_SHORT).show();    } */
//本地图片 取到剪切版end

        AtomicReference<ClipData> clipDataRef = new AtomicReference<>(null);
        ClipboardUtils clipboardUtils = ClipboardUtils.getInstance();
        clipboardUtils.getPrimaryClip(clipDataRef);
        //Log.e("clip", "lsz len===hasClip GoLog=clipboardUtils.hasClip()=" + clipboardUtils.hasClip());
        for (int i = 0; i < clipboardUtils.getItemCount(clipDataRef); i++) {
            Log.e("clip", "GoLog lsz len=getItemType 0：TETX 1：IMG=" + clipboardUtils.getItemType(clipDataRef, i));
            if (clipboardUtils.getItemType(clipDataRef, i) == clipboardUtils.CLIPBOARD_DATA_TYPE_TEXT) {
                String text = clipboardUtils.getTextItem(clipDataRef, i);
                if (mIsDebugfloatWindow) {
                    TextView textView = floatView.findViewById(R.id.text_id);
                    textView.setText("get=" + text);
                }
                sendToPC(text);
                //Log.e("clip", "lsz len===hasClip text GoLog=" + text);
            } else if (clipboardUtils.getItemType(clipDataRef, i) == clipboardUtils.CLIPBOARD_DATA_TYPE_IMAGE) {
                Bitmap bitmap1 = clipboardUtils.getImageItem(clipDataRef, i);
                //Log.e("clip", "lsz GoLog len===hasClip bitmap1==" + bitmap1);
                if (bitmap1 != null) {
                /*
                数组转bitmap
                */
                    //Bitmap drawableicon = BitmapFactory.decodeResource(getResources(), R.drawable.liu2);
                    //byte[] imageData = bitmapToByteArray(drawableicon); // 要转换的字节数组
                    //Bitmap bitmap3 = BitmapFactory.decodeByteArray(imageData, 0, imageData.length);
                    //imageView2.setImageBitmap(bitmap1);

                    //bitmap转byteArray
                    //int bytes = bitmap1.getByteCount();
                    //ByteBuffer buf = ByteBuffer.allocate(bytes);
                    //bitmap1.copyPixelsToBuffer(buf);
                    //byte[] byteArray = buf.array();

                    if (mIsDebugfloatWindow) {
                        imageView = floatView.findViewById(R.id.imageView);
                        imageView.setImageBitmap(bitmap1);
                    }

                    byte[] imageData = bitmapToByteArray(bitmap1);

                    sendToPCIMG(imageData);


                } else {
                    Toast.makeText(this, " Clipboard img is empty", Toast.LENGTH_SHORT).show();
                }


            } else {
                Log.e("clip", "not support format");
                Toast.makeText(this, "lsz111 Clipboard is empty", Toast.LENGTH_SHORT).show();
            }
        }


    }


    private void setClipToClipboard(String string, String msg) {
        ClipboardUtils clipboardUtils = ClipboardUtils.getInstance();
        //clipboardUtils.clearClip();
        AtomicReference<ClipData> clipDataRef = ClipboardUtils.createClipdataRef();
        if (!string.equals("") || string != null) {
            clipboardUtils.addTextItem(clipDataRef, "test text12");
        }

        if (!msg.equals("") || msg != null) {

            byte[] decodedBytes = Base64.decode(msg, Base64.DEFAULT);
            Bitmap bitmap = BitmapFactory.decodeByteArray(decodedBytes, 0, decodedBytes.length);

            clipboardUtils.addImageItem(clipDataRef, bitmap);
        }

        clipboardUtils.setPrimaryClip(clipDataRef);

    }


    public static byte[] bitmapToByteArray(Bitmap bitmap) {
        ByteArrayOutputStream outputStream = new ByteArrayOutputStream();
        bitmap.compress(Bitmap.CompressFormat.JPEG, 100, outputStream);
        //Log.d("lsz", "outputStream. imag toByteArray()=" + outputStream.toByteArray().toString());
        return outputStream.toByteArray();

    }


    public void getClipboard() {
        //Log.i("lsz", "getClipboard");
        ClipboardManager clipboard = (ClipboardManager) getSystemService(Context.CLIPBOARD_SERVICE);
        Log.i("lsz", "hasPrimaryClip" + clipboard.hasPrimaryClip());
        ClipData clip = clipboard.getPrimaryClip();
        Log.i("lsz", "clip=" + clip);
        if (clip != null && clip.getItemCount() > 0) {
            ClipData.Item item = clip.getItemAt(0);
            boolean isImage = item.getUri() != null && item.getUri().toString().startsWith("content://");
            boolean isText = item.getText() != null;
            Log.i("lsz", "isImage=" + isImage + "isText=" + isText);
            if (isImage) {
                Uri uri = item.getUri();
                Log.i("lsz", "uri=" + uri);
                try (InputStream inputStream = getContentResolver().openInputStream(uri)) {
                    Bitmap bitmap = BitmapFactory.decodeStream(inputStream);
                    if (mIsDebugfloatWindow) {
                        imageView = floatView.findViewById(R.id.imageView);
                        imageView.setImageBitmap(bitmap);
                    }
                    Toast.makeText(this, "Image has been put to clipboard", Toast.LENGTH_SHORT).show();
                    byte[] imageData = bitmapToByteArray(bitmap);
                    Log.i("lsz", "send img bitmap.getByteCount()=" + bitmap.getByteCount());
                    sendToPCIMG(imageData);
                    // 现在你可以使用bitmap了
                } catch (Exception e) {
                    e.printStackTrace();
                    // 处理找不到文件的情况
                }
            } else if (isText) {
                ClipData.Item item2 = clip.getItemAt(0);
                String text = item2.getText().toString();
                Log.i(TAG, "lsz get text: " + text);
                if (text.isEmpty()) {
                    Toast.makeText(this, "Get empty string", Toast.LENGTH_SHORT).show();
                } else {
                    Toast.makeText(this, "Text has been put to clipboard", Toast.LENGTH_SHORT).show();
                    if (mIsDebugfloatWindow) {
                        TextView textView = floatView.findViewById(R.id.text_id);
                        textView.setText(" " + text);
                    }
                    sendToPC(text);
                }
            }
        } else {
            Log.i("lsz", "no data");
            // 剪贴板中没有可用的图片数据
        }

    }

    public static Bitmap base64ToBitmapa(String base64String) {
        // 移除Base64编码的前缀（如果有的话）
        if (base64String.contains(",")) {
            base64String = base64String.split(",")[1];
        }

        // 对Base64字符串进行解码
        byte[] decodedBytes = Base64.decode(base64String, Base64.DEFAULT);

        Log.i(TAG, "lszz bitmap decodedBytes[] length" + decodedBytes.length);
        // 将字节数组转换为Bitmap
        return BitmapFactory.decodeByteArray(decodedBytes, 0, decodedBytes.length);
    }

    public void setBitmapToClipboard(Context context, Bitmap bitmap) {
        // 确保外部存储可用
        Log.i(TAG, "lsz setBitmapToClipboard init");
        if (!Environment.getExternalStorageState().equals(Environment.MEDIA_MOUNTED)) {
            return;
        }

        // 创建一个文件来保存Bitmap
        File file = new File(context.getExternalFilesDir(null), "shared_image.png");
        Log.i(TAG, "lsz getExternalStorageState imageFile getPath=" + file.getPath());
        //Uri imageUri = FileProvider.getUriForFile(context, "com.rtk.myapplication", file);

        try (FileOutputStream out = new FileOutputStream(file)) {
            bitmap.compress(Bitmap.CompressFormat.PNG, 100, out);
        } catch (IOException e) {
            e.printStackTrace();
            return;
        }

        // 获取文件的Uri
        Uri imageUri = FileProvider.getUriForFile(context, "com.rtk.myapplication", file);
        Log.i(TAG, "lsz getExternalStorageState imageFile imageUri=" + imageUri);
        // 创建ClipData
        ClipData clip = ClipData.newUri(context.getContentResolver(), "image/png", imageUri);

        // 获取ClipboardManager实例
        ClipboardManager clipboard = (ClipboardManager) context.getSystemService(Context.CLIPBOARD_SERVICE);

        // 将ClipData放入剪贴板
        Log.i(TAG, "lsz setimg to clipboard ");
        clipboard.setPrimaryClip(clip);
    }


    public static String bytekb(long bytes) {
//格式化小数
        int GB = 1024 * 1024 * 1024;
        int MB = 1024 * 1024;
        int KB = 1024;

        if (bytes / GB >= 1) {
            double gb = Math.round((double) bytes / 1024.0 / 1024.0 / 1024.0 * 100.0) / 100.0;
            return String.format("%.2f", gb) + "GB";
        } else if (bytes / MB >= 1) {
            double mb = Math.round((double) bytes / 1024.0 / 1024.0 * 100.0) / 100.0;

            //Log.i("lsz","1111=="+String.format("%.2f", mb));
            return String.format("%.2f", mb) + "MB";
        } else if (bytes / KB >= 1) {
            double kb = Math.round((double) bytes / 1024.0 * 100.0) / 100.0;
            return String.format("%.2f", kb) + "KB";
        } else {
            return bytes + "B";
        }
    }


    public Bitmap getBitmap(String privateFilePath) {
        File file = new File(privateFilePath);
        //Log.i("lsz", "ss name " + file.getName().substring(file.getName().length() - 3));
        if (file.getName().substring(file.getName().length() - 3).equals("png") ||
                file.getName().substring(file.getName().length() - 3).equals("jpg")) {
            //Log.i("lsz", "ss name  init");

            return BitmapFactory.decodeFile(file.getAbsolutePath());
        }
        return null;
    }

    public byte[] getbitmapData(String privateFilePath) {
        File file = new File(privateFilePath);
        Log.i("lsz", "getbitmapData name " + file.getName().substring(file.getName().length() - 3));
        if (file.getName().substring(file.getName().length() - 3).equals("png") ||
                file.getName().substring(file.getName().length() - 3).equals("jpg")) {
            //Log.i("lsz","ss name  init");
            Bitmap bitmap = BitmapFactory.decodeFile(file.getAbsolutePath());
            if (bitmap != null) {
                ByteArrayOutputStream byteArrayOutputStream = new ByteArrayOutputStream();
                bitmap.compress(Bitmap.CompressFormat.PNG, 100, byteArrayOutputStream);
                bitmapData = byteArrayOutputStream.toByteArray();
                Log.i("lsz", "getbitmapData  bitmapData" + bitmapData.length);
                return bitmapData;
            }
        }
        return null;
    }

    public String copyFileToPublicDir(String privateFilePath) {
        boolean isRetry = false;
        FileInputStream fis = null;
        FileOutputStream fos = null;
        File srcFile = new File(privateFilePath);
        File saveDir = Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DOWNLOADS);
        File destFile = new File(saveDir, srcFile.getName());
        //if (destFile.exists()) {

        //}
        try {
            fis = new FileInputStream(srcFile);
            fos = new FileOutputStream(destFile);

            byte[] buffer = new byte[1024];
            int bytesRead;
            long fileSize = srcFile.length(); // 获取源文件大小
            long totalBytesRead = 0; // 已读取字节数

            while ((bytesRead = fis.read(buffer)) != -1) {
                fos.write(buffer, 0, bytesRead);
                totalBytesRead += bytesRead; // 累加已读取字节数

                //Log.i("lszz", "get totalBytesRead =" + totalBytesRead);
                if (totalBytesRead >= fileSize) {
                    // 已读取字节数大于等于文件大小，认为文件已经拷贝完毕
                    Toast.makeText(this, "file has been save to Download of internal storage", Toast.LENGTH_SHORT).show();
                    Log.d("lszz", "storage private app file is exists,now remove");
                    srcFile.delete();
                    break;
                }

            }
            fis.close();
            fos.close();
        } catch (IOException e) {
            e.printStackTrace();
            isRetry = true;
        }
        if (isRetry) {
            File originalFile = new File(privateFilePath);
            File newFile = changeName(originalFile);
            Log.d(TAG, "originalFile: "+ originalFile.getAbsolutePath() + ", new: "+ newFile.getAbsolutePath());
            if (newFile != null) {
                copyFileToPublicDir(newFile.getAbsolutePath());
            } else {
                Log.d(TAG,  "do nothing due to rename fail");
            }
        }
        return null;
    }

    // ex. 15MB.mp4 => 15MB(0).mp4 or 15MB(0).mp4 => 15MB(1).mp4
    public static File changeName(File originalFile) {
        // Check if the original file exists
        if (!originalFile.exists()) {
            System.out.println("Original file does not exist.");
            return null;
        }

        // Extract file name and extension
        String fileName = originalFile.getName();
        String fileExtension = "";

        // Separate the extension if present
        int dotIndex = fileName.lastIndexOf('.');
        if (dotIndex > 0) {
            fileExtension = fileName.substring(dotIndex);
            fileName = fileName.substring(0, dotIndex);
        }

        // Regular expression to detect the pattern "(number)"
        Pattern pattern = Pattern.compile("\\((\\d+)\\)$");
        Matcher matcher = pattern.matcher(fileName);

        int counter = 0;

        if (matcher.find()) {
            // If the pattern (number) exists, extract the number and increment it
            counter = Integer.parseInt(matcher.group(1)) + 1;
            fileName = fileName.substring(0, matcher.start()); // Remove the old "(number)"
        } else {
            // If no (number) pattern, we start with (0)
            counter = 0;
        }

        // Create the new file name with the incremented number
        String newFileName = fileName + "(" + counter + ")" + fileExtension;
        File renamedFile = new File(originalFile.getParent(), newFileName);

        // Rename the file
        boolean success = originalFile.renameTo(renamedFile);

        if (success) {
            return renamedFile; // Return the renamed file
        } else {
            System.out.println("Failed to rename the file.");
            return null;
        }
    }

    //全局对话框  为service打造
    public void define_cancel_service(final Context context, String ipAddr,String s, String name, long l) {
        // 加载布局文件
        View view = View.inflate(context, R.layout.dialog_deivce, null);

        TextView subtitleView = (TextView) view.findViewById(R.id.subtitle);
        Button conf = (Button) view.findViewById(R.id.img_comf);
        Button canl = (Button) view.findViewById(R.id.img_canl);

        filename = name;
        filesize = l;

        String subtitle = getResources().getString(R.string.pc_to_phone_file_dialog_subtitle);
        subtitleView.setText(String.format(subtitle, s,name, bytekb(l)));

        //Log.d("lszz", "lszz String s===s=" + s);
        //Log.d("lszz", "lszz long   l===l=" + l);


        // 创建Dialog
        dialog = new AlertDialog.Builder(context).create();
        dialog.setCancelable(false);// 设置点击dialog以外区域不取消Dialog

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            dialog.getWindow().setType(WindowManager.LayoutParams.TYPE_APPLICATION_OVERLAY);
        } else {
            dialog.getWindow().setType(WindowManager.LayoutParams.TYPE_SYSTEM_ALERT);
        }

        dialog.show();
        dialog.setContentView(view);
        //todo 设置上下位置
        dialog.getWindow().setGravity(Gravity.CENTER);
        WindowManager.LayoutParams params = dialog.getWindow().getAttributes();

        conf.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {
                Log.d(TAG, "is confirm");
                Toast.makeText(context, "Confirm", Toast.LENGTH_SHORT).show();
                Libp2p_clipboard.ifClipboardPasteFile( name,  ipAddr,true);
                Intent intent = new Intent(MyApplication.getContext(), TestActivity.class);
                intent.putExtra("booleanKey", true);
                intent.putExtra("filename", name);
                intent.putExtra("filesize", l);
                Log.d(TAG, "get callbackMethodFileDone filename=" + filename);
                Log.d(TAG, "get callbackMethodFileDone l=" + l);
                Log.d(TAG, "get callbackMethodFileDone s=" + s);
                //intent.putExtra("bitmappath", s);
                intent.setFlags(Intent.FLAG_ACTIVITY_CLEAR_TASK);
                intent.setFlags(Intent.FLAG_ACTIVITY_NEW_TASK);
                startActivity(intent);
                dialog.dismiss();
            }
        });
        canl.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {
                Toast.makeText(context, "Cancel", Toast.LENGTH_SHORT).show();
                Libp2p_clipboard.ifClipboardPasteFile(name,ipAddr,false);
                dialog.dismiss();
            }
        });


    }

    public static int findFreePort() {
        int port = 0;
        try (ServerSocket serverSocket = new ServerSocket(0)) {
            port = serverSocket.getLocalPort();
        } catch (IOException e) {
            e.printStackTrace();
        }
        return port;
    }

    public static String getWifiIpAddress(Context context) {
        WifiManager wifiManager = (WifiManager) context.getSystemService(Context.WIFI_SERVICE);
        if (wifiManager != null && wifiManager.getConnectionInfo() != null) {
            int ipAddress = wifiManager.getConnectionInfo().getIpAddress();
            return Formatter.formatIpAddress(ipAddress);
        }
        return null;
    }

    public void getClientList() {
        //从libp2p获取列表数据
        String getlist = Libp2p_clipboard.getClientList();
        if (!getlist.isEmpty()) {
            Log.d("lszz", "getlist=======+++=" + getlist);
            //String[] strArray = getlist.split("#");

            Intent intent = new Intent("com.example.MY_CUSTOM_EVENT");
            // 可以添加额外的数据到intent
            //intent.putExtra("data_key", getlist);
            LocalBroadcastManager.getInstance(this).sendBroadcast(intent);


        }
    }

    private Point getDefaultDisplay() {
        Point realSize = new Point();
        Display display = windowManager.getDefaultDisplay();
        try {
            Display.class.getMethod("getRealSize", Point.class).invoke(display, realSize);
        } catch (IllegalAccessException e) {
            throw new RuntimeException(e);
        } catch (InvocationTargetException e) {
            throw new RuntimeException(e);
        } catch (NoSuchMethodException e) {
            throw new RuntimeException(e);
        }
        return realSize;
    }

    Handler mCheckBoundshandler = new Handler() {
        @Override
        public void handleMessage(Message msg) {
            super.handleMessage(msg);
            if (mIsDebugfloatWindow) {
                Log.d(TAG, "mIsDebugfloatWindow true, skip adjust position of floatview");
                return;
            }

            if (floatView != null && floatViewForAdjust != null) {
                int[] l = new int[2];
                floatViewForAdjust.getLocationOnScreen(l);
                int x = l[0];
                int y = l[1];
                if ((x == mAdjustViewX) && (y == mAdjustViewY)) {
                    Log.d(TAG, "skip this update");
                } else {
                    mAdjustViewX = x;
                    mAdjustViewY = y;

                    Log.d(TAG, "mCheckBoundshandler() floatViewForAdjust x: " + x + " y:" + y);
                    Point realSize = new Point();
                    Display display = windowManager.getDefaultDisplay();
                    try {
                        Display.class.getMethod("getRealSize", Point.class).invoke(display, realSize);
                    } catch (IllegalAccessException e) {
                        throw new RuntimeException(e);
                    } catch (InvocationTargetException e) {
                        throw new RuntimeException(e);
                    } catch (NoSuchMethodException e) {
                        throw new RuntimeException(e);
                    }
                    Log.d(TAG, "mCheckBoundshandler() realSize.x " + realSize.x + " realSize.y" + realSize.y);
                    int expected_x = realSize.x / 2 - 1;
                    int expected_y = realSize.y / 2 - 1;
                    Log.d(TAG, "mCheckBoundshandler() expected_x: " + expected_x + " expected_y" + expected_y);
                    if (expected_x != x || expected_y != y) {
                        params = new WindowManager.LayoutParams();
                        params.type = WindowManager.LayoutParams.TYPE_APPLICATION_OVERLAY;
                        params.flags = WindowManager.LayoutParams.FLAG_NOT_TOUCH_MODAL
                                | WindowManager.LayoutParams.FLAG_NOT_FOCUSABLE
                                | WindowManager.LayoutParams.FLAG_LAYOUT_NO_LIMITS;
                        params.format = PixelFormat.TRANSLUCENT;
                        params.gravity = Gravity.CENTER;
                        params.x = expected_x - x;
                        params.y = expected_y - y;
                        Log.d(TAG, "mCheckBoundshandler() params.x: " + params.x + ", params.y:" + params.y);
                        params.height = 6;
                        params.width = 6;
                        windowManager.updateViewLayout(floatView, params);
                    } else {
                        Log.d(TAG, "mCheckBoundshandler() expected_x and x is same, do nothing");
                    }
                }
            }
            mCheckBoundshandler.removeMessages(0);
            mCheckBoundshandler.sendEmptyMessageDelayed(0, 3000);
        }
    };


    public void getNetInfoFromLocalIp() throws SocketException {

        Enumeration<NetworkInterface> networkInterfaces = NetworkInterface.getNetworkInterfaces();
        while (networkInterfaces.hasMoreElements()) {
            NetworkInterface networkInterface = networkInterfaces.nextElement();
            // 在这里处理每个网络接口
            String netname = networkInterface.getName();
            if (netname.equals("wlan0")) {
                getnetname = "wlan0";
                getindex = networkInterface.getIndex();
            }

        }

    }
}
