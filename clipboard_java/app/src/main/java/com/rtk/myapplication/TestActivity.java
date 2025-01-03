package com.rtk.myapplication;

import android.annotation.SuppressLint;
import android.app.Activity;
import android.app.AlertDialog;
import android.app.AppOpsManager;
import android.app.Service;
import android.content.BroadcastReceiver;
import android.content.ClipData;
import android.content.ClipboardManager;
import android.content.ComponentName;
import android.content.ContentResolver;
import android.content.ContentValues;
import android.content.Context;
import android.content.Intent;
import android.content.ServiceConnection;
import android.content.pm.PackageManager;
import android.database.Cursor;
import android.graphics.BitmapFactory;
import android.graphics.Canvas;
import android.graphics.Color;
import android.graphics.Paint;
import android.graphics.drawable.ColorDrawable;
import android.media.MediaPlayer;
import android.net.ConnectivityManager;
import android.net.LinkAddress;
import android.net.LinkProperties;
import android.net.NetworkInfo;
import android.net.Uri;
import android.net.wifi.WifiManager;
import android.os.Binder;
import android.os.Build;
import android.os.Bundle;
import android.os.Environment;
import android.os.Handler;
import android.os.IBinder;
import android.provider.MediaStore;
import android.provider.OpenableColumns;
import android.provider.Settings;
import android.text.TextUtils;
import android.text.format.Formatter;
import android.util.Base64;
import android.util.DisplayMetrics;
import android.util.Log;
import android.view.Gravity;
import android.view.LayoutInflater;
import android.view.MotionEvent;
import android.view.View;
import android.view.ViewGroup;
import android.view.Window;
import android.view.WindowManager;
import android.widget.Button;
import android.widget.CheckBox;
import android.widget.CompoundButton;
import android.widget.EditText;
import android.widget.ImageView;
import android.widget.LinearLayout;
import android.widget.PopupWindow;
import android.widget.ProgressBar;
import android.widget.TextView;
import android.widget.Toast;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;
import androidx.core.app.ActivityCompat;
import androidx.core.content.ContextCompat;
import androidx.core.content.FileProvider;

import androidx.localbroadcastmanager.content.LocalBroadcastManager;
import androidx.recyclerview.widget.LinearLayoutManager;
import androidx.recyclerview.widget.RecyclerView;

import java.io.ByteArrayOutputStream;
import java.io.DataOutputStream;
import java.io.File;
import java.io.FileNotFoundException;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.io.UnsupportedEncodingException;
import java.lang.reflect.Field;
import java.lang.reflect.Method;
import java.net.NetworkInterface;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.nio.charset.StandardCharsets;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.atomic.AtomicReference;

import libp2p_clipboard.Callback;
import libp2p_clipboard.Libp2p_clipboard;

import android.graphics.Bitmap;
import android.widget.VideoView;

import java.net.ServerSocket;

import android.graphics.Matrix;

import com.tencent.mmkv.MMKV;

import android.content.IntentFilter;

public class TestActivity extends Activity {

    private static final int REQUEST_CODE = 1024;
    private static final String TAG = "TestActivity";
    private ClipboardManager clipboardManager;
    private int testCount = 0;
    private TextView mServerStatus;
    private TextView mClientStatus;
    private TextView mPeerMessage;
    private EditText mServerId;
    private EditText mServerIpInfo;
    private boolean mIsConnected = false;
    private String mLastString = "";
    private ImageView imageView;

    private Button btnGetImage;
    private Button btnSetImage;
    private ImageView imageView2, imageView3;
    private TextView textView;
    Bitmap bitmap;
    VideoView videview;
    Button mbutton, mbuttonpaste, buttom_w, buttom_r;
    String value;
    String mimetype;
    String sharedText;
    byte[] getbyteArray;
    String sizeInMB;
    Intent intent;
    String action;
    String base64String, clearbase64String;
    Context mContext;
    Bitmap bitmapShare;
    byte[] imageData;
    String text;
    Uri uri;
    ContentResolver resolver;
    boolean check = false;
    MMKV kv;
    boolean boxischeck;
    TextView textView_name, textView_size, mConnCountView, mFileConnCountView, mConnectionsView,
            mMyConnectionView, mSwVersionView;
    String fileRealpath, saveFilePath,share_file_name;
    RecyclerView recyclerView, recyclerView2;
    long countSize;
    long countSizebuf;
    double countbuf;
    ProgressBar progress_bar;
    int progress;
    String filename;
    long filesize;
    String bitmappath;
    String filenamea = "";
    private FloatClipboardService myService;
    private boolean isBound = false;

    private FileTransferAdapter adapter;
    private List<FileTransferItem> fileTransferList = new ArrayList<>();

    RecyclerView recyclerViewdevice;
    private DeviceAdapter deviceAdapter;
    private List<Device> deviceList;
    private Map<String, String> deviceNameIdMap;
    LinearLayout linearLayout;
    ImageView share_image, back_icon;
    Button img_button;
    boolean isimage =false;
    TextView share_file;
    LinearLayout layout;
    private static final String SOURCE_HDMI1 = "HDMI1";
    private static final String SOURCE_HDMI2 = "HDMI2";
    private static final String SOURCE_USBC = "USBC";
    private static final String SOURCE_MIRACAST = "Miracast";
    private boolean mIsAndroidShareFile = false;
    TextView file_page_back_icon;

    private ServiceConnection connection = new ServiceConnection() {

        @Override
        public void onServiceConnected(ComponentName className, IBinder service) {
            FloatClipboardService.LocalBinder binder = (FloatClipboardService.LocalBinder) service;
            myService = binder.getService();
            isBound = true;
            myService.setCallback(new FloatClipboardService.DataCallback() {
                @Override
                public void onDataReceived(String name, double data) {
                    runOnUiThread(() -> {
                        Log.i("lsz", "ServiceConnection get datadatadata==" + data);
                        progress = (int) data;
                        MyApplication.getMyAdapter().updateProgress(name, progress);
                    });
                }

                @Override
                public void onMsgReceived(String name, String msg) {
                    Log.i("lsz", "onMsgReceived msg==" + msg);
                    //updateFileListWithError(name, 2);
                    MyApplication.getMyAdapter().updateFileListWithError(name, 2);
                }

                @Override
                public void onBitmapReceived(Bitmap bitmap, String path) {

                    //getFileList(filename, filesize, bitmap,progress);
                    if (path != null && !path.isEmpty()) {
                        filename = path.substring(path.lastIndexOf("/") + 1);
                    }
                    MyApplication.getMyAdapter().setBitmap(filename, bitmap);
                }

                @Override
                public void onCallbackMethodFileDone(String path) {
                    String filename = " ";
                    if (path != null && !path.isEmpty()) {
                        filename = path.substring(path.lastIndexOf("/") + 1);
                    }
                    MyApplication.getMyAdapter().getFileTime(filename);
                }
            });
        }

        @Override
        public void onServiceDisconnected(ComponentName arg0) {
            isBound = false;
        }
    };


    private BroadcastReceiver broadcastReceiver = new BroadcastReceiver() {
        @Override
        public void onReceive(Context context, Intent intent) {
            long data = intent.getLongExtra("data", -1L);
            Log.i(TAG, "init broadcastReceiver" + data);
            getClientList();
        }
    };


    private BroadcastReceiver broadcastReceivera = new BroadcastReceiver() {
        @Override
        public void onReceive(Context context, Intent intent) {
            //long data = intent.getLongExtra("data", -1L);
            progress = intent.getIntExtra("countbuf", 0);
            filename = intent.getStringExtra("filename");
            filesize = intent.getLongExtra("filesize", 0);
            Log.i(TAG, "init filename" + filename);
            Log.i(TAG, "init filesize" + filesize);
            Log.i(TAG, "init broadcastReceiver progress" + progress);
            adapter.updateProgress(filename, progress);


        }
    };

    @Override
    protected void onCreate(@Nullable Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        kv = MMKV.defaultMMKV();
        mContext = this;

        intent = getIntent();
        action = intent.getAction();
        mimetype = intent.getType();

        boolean booleanValue = getIntent().getBooleanExtra("booleanKey", false);
        Log.i(TAG, "booleanValue: booleanValue====" + booleanValue);
        //adapter = new FileTransferAdapter(fileTransferList);

        Log.i(TAG, "onCreate: intent====" + intent);
        Log.i(TAG, "onCreate: action====" + action);
        Log.i(TAG, "onCreate: mimetype====" + mimetype);

        if (Intent.ACTION_SEND.equals(action) || booleanValue) {
            setTheme(R.style.TransparentTheme);
            setContentView(R.layout.layout_main);
            mIsAndroidShareFile = true;

            getWindow().setBackgroundDrawableResource(android.R.color.transparent);
            findViewByIdDevice();
            linearLayout.setVisibility(View.VISIBLE);

            try {
               // testCrossShare();
                getShare(intent, action, mimetype);
            } catch (IOException e) {
                throw new RuntimeException(e);
            } catch (Exception e) {
                throw new RuntimeException(e);
            }

            doShareMain();
            getClientList();

        } else {
            setTheme(R.style.Theme_MyApplication);
            getWindow().setBackgroundDrawableResource(android.R.color.white);
            Log.i(TAG, "onCreate: checkPermission");
            checkPermission(mContext);
            setContentView(R.layout.layout_testactivity_title);
            mMyConnectionView = findViewById(R.id.my_connections_view);
            if (mMyConnectionView != null) {
                mMyConnectionView.setText(getMyIpDeviceName());
            }
            mSwVersionView = findViewById(R.id.sw_version_view);
            if (mSwVersionView != null) {
                mSwVersionView.setText(getSoftwareInfo());
            }
        }
        Log.d(TAG, "mIsAndroidShareFile: "+ mIsAndroidShareFile);


        if(action == null && mimetype == null){
            Log.i("lszz", "onCreate: CheckBox boxischeck=== is null");
            setTheme(R.style.Theme_MyApplication);
            setContentView(R.layout.myactivity);
            RecyclerView recyclerView2 = findViewById(R.id.recycler_view);
            recyclerView2.setLayoutManager(new LinearLayoutManager(this));
            //adapter = new FileTransferAdapter(fileTransferList);
            recyclerView2.setAdapter(adapter);
            mConnCountView = findViewById(R.id.connection_count);

            setIntent(intent);

            boolean booleanValue2 = getIntent().getBooleanExtra("booleanKey", false); // 第二个参数是默认值，如果没找到键则使用默认值
            filename = getIntent().getStringExtra("filename");
            filesize = getIntent().getLongExtra("filesize", -1L);
            //bitmappath = getIntent().getStringExtra("bitmappath");
            countSize = filesize;
            Log.d(TAG, "booleanValue booleanValue=" + booleanValue2);
            Log.d(TAG, "filename filenamea=" + filenamea);
            Log.d(TAG, "filename filename=" + filename);
            Log.d(TAG, "filesize=" + filesize);
            //Log.d(TAG, "String bitmappath=" + bitmappath);

            if (booleanValue2) {

                if (!filenamea.equals(filename)) {
                    filenamea = filename;
                    FileTransferItem item = new FileTransferItem(filename, filesize, BitmapHolder.getBitmap());
                    fileTransferList.add(item);
                    adapter.notifyItemInserted(fileTransferList.size() - 1);
                }
            }

            file_page_back_icon = findViewById(R.id.file_page_back_icon);
            if (file_page_back_icon != null) {
                file_page_back_icon.setOnClickListener(new View.OnClickListener() {
                    @Override
                    public void onClick(View view) {
                        Log.d(TAG, "from file page to main page");
                        setContentView(R.layout.layout_testactivity_title);
                        getClientList();
                        mMyConnectionView = findViewById(R.id.my_connections_view);
                        if (mMyConnectionView != null) {
                            mMyConnectionView.setText(getMyIpDeviceName());
                        }
                        mSwVersionView = findViewById(R.id.sw_version_view);
                        if (mSwVersionView != null) {
                            mSwVersionView.setText(getSoftwareInfo());
                        }
                    }
                });
            }
        }

        //alertDialog("aa",1918522);

        boxischeck = kv.decodeBool("ischeck", false);
        Log.i("lszz", "onCreate: CheckBox boxischeck===" + boxischeck);




        IntentFilter intentFilter = new IntentFilter("com.example.MY_CUSTOM_EVENT");
        LocalBroadcastManager.getInstance(this).registerReceiver(broadcastReceiver, intentFilter);


        Log.i(TAG, "lsz path = " + getExternalFilesDir(null));
        //requestPermission();
        //   getIpAddres();

    }

    private String getMyIpDeviceName() {
        return getWifiIpAddress(MyApplication.getContext()) + " " +
                Settings.Global.getString(MyApplication.getContext().getContentResolver(), "device_name");
    }

    private String getSoftwareInfo() {
        return Libp2p_clipboard.getVersion()+" ("+ Libp2p_clipboard.getBuildDate()+")";
    }

    private void getClibMessageLoop() {
        new Handler().postDelayed(new Runnable() {
            @Override
            public void run() {
                getClibMessage();
                getClibMessageLoop();
            }
        }, 3000);
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

    private void getClibMessage() {
        if (clipboardManager.hasPrimaryClip()) {
            // clipboard has data
            ClipData clipData = clipboardManager.getPrimaryClip();
            if (clipData != null && clipData.getItemCount() > 0) {
                CharSequence itemText = clipData.getItemAt(0).getText();
                Log.i(TAG, "getClibMessage: itemText==" + itemText.toString());
            }
        }
    }


    private void setClibMessage() {
        ClipData clipData = ClipData.newPlainText(null, "编辑后的文本数据+" + testCount);
        clipboardManager.setPrimaryClip(clipData);
        testCount++;
    }

    @Override
    protected void onResume() {
        super.onResume();
        Log.i(TAG, "lsz onResume: mIsConnected: " + mIsConnected);

        if (checkFloatPermission(mContext) == true) {
            Intent serviceIntent = new Intent(MyApplication.getContext(), FloatClipboardService.class);
            Log.d(TAG, "startForegroundService in onResume");
            startForegroundService(serviceIntent);
            Intent intent = new Intent(MyApplication.getContext(), FloatClipboardService.class);
            bindService(intent, connection, Context.BIND_AUTO_CREATE);
        } else {
            Log.d(TAG, "startForegroundService in onResume: skip, due to no permission");
        }

    }

    private void requestPermission() {
        if (ActivityCompat.checkSelfPermission(this, android.Manifest.permission.READ_EXTERNAL_STORAGE) == PackageManager.PERMISSION_GRANTED &&
                ContextCompat.checkSelfPermission(this, android.Manifest.permission.WRITE_EXTERNAL_STORAGE) == PackageManager.PERMISSION_GRANTED) {
            Toast.makeText(this, "Get read external storage permission successfully", Toast.LENGTH_SHORT).show();
        } else {
            ActivityCompat.requestPermissions(this, new String[]{android.Manifest.permission.READ_EXTERNAL_STORAGE, android.Manifest.permission.WRITE_EXTERNAL_STORAGE}, REQUEST_CODE);
        }

    }

    @Override
    public void onRequestPermissionsResult(int requestCode, @NonNull String[] permissions, @NonNull int[] grantResults) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults);
        if (requestCode == REQUEST_CODE) {
            if (ActivityCompat.checkSelfPermission(this, android.Manifest.permission.READ_EXTERNAL_STORAGE) == PackageManager.PERMISSION_GRANTED &&
                    ContextCompat.checkSelfPermission(this, android.Manifest.permission.WRITE_EXTERNAL_STORAGE) == PackageManager.PERMISSION_GRANTED) {
                Toast.makeText(this, "Get read external storage permission successfully", Toast.LENGTH_SHORT).show();
            } else {
                Toast.makeText(this, "Get read external storage permission successfully", Toast.LENGTH_SHORT).show();
            }
        }
    }

    private String getIpAddres() {

        ConnectivityManager connectivityManager = (ConnectivityManager) getApplicationContext().getSystemService(Service.CONNECTIVITY_SERVICE);
        LinkProperties linkProperties = connectivityManager.getLinkProperties(connectivityManager.getActiveNetwork());
        List<LinkAddress> addressList = linkProperties.getLinkAddresses();
        StringBuffer sbf = new StringBuffer();
        for (LinkAddress linkAddress : addressList) {
            sbf.append(linkAddress.toString()).append("#");
            Log.d(TAG, "xyf getIpAddreslinkAddress.toString(): " + linkAddress.toString());
        }
        Log.d(TAG, "xyf getIpAddres: " + sbf.toString());
        return sbf.toString();
    }

    public static String getWifiIpAddress(Context context) {
        WifiManager wifiManager = (WifiManager) context.getSystemService(Context.WIFI_SERVICE);
        if (wifiManager != null && wifiManager.getConnectionInfo() != null) {
            int ipAddress = wifiManager.getConnectionInfo().getIpAddress();
            return Formatter.formatIpAddress(ipAddress);
        }
        return null;
    }


    public String getTextFromClipboard(Context context) {
        ClipboardManager clipboard = (ClipboardManager) context.getSystemService(Context.CLIPBOARD_SERVICE);
        if (clipboard.hasPrimaryClip()) {
            ClipData.Item item = clipboard.getPrimaryClip().getItemAt(0);
            return item.getText().toString();
        }
        return null;
    }

    public void getClipboard() {

        ClipboardManager clipboard = (ClipboardManager) getSystemService(Context.CLIPBOARD_SERVICE);
        ClipData clip = clipboard.getPrimaryClip();
        Log.i("lsz", "clip" + clip);
        if (clip != null && clip.getItemCount() > 0) {
            ClipData.Item item = clip.getItemAt(0);
            Uri uri = item.getUri();
            Log.i("lsz", "uri=" + uri);
            try (InputStream inputStream = getContentResolver().openInputStream(uri)) {
                Bitmap bitmap = BitmapFactory.decodeStream(inputStream);
                imageView2.setImageBitmap(bitmap);
            } catch (IOException e) {
                e.printStackTrace();
            }
        } else {
            Log.i("lsz", "no data");
        }


    }

    private void getClipFromClipboard() throws IOException {
        AtomicReference<ClipData> clipDataRef = new AtomicReference<>(null);
        ClipboardUtils clipboardUtils = ClipboardUtils.getInstance();
        clipboardUtils.getPrimaryClip(clipDataRef);
        Log.e("clip", "lsz len===hasClip " + clipboardUtils.hasClip());
        for (int i = 0; i < clipboardUtils.getItemCount(clipDataRef); i++) {
            Log.e("clip", "lsz len=getItemType" + clipboardUtils.getItemType(clipDataRef, i));
            if (clipboardUtils.getItemType(clipDataRef, i) == clipboardUtils.CLIPBOARD_DATA_TYPE_TEXT) {
                text = clipboardUtils.getTextItem(clipDataRef, i);
                textView.setText(text);
                //sendToPC(text);
            } else if (clipboardUtils.getItemType(clipDataRef, i) == clipboardUtils.CLIPBOARD_DATA_TYPE_IMAGE) {
                Bitmap bitmap1 = clipboardUtils.getImageItem(clipDataRef, i);
                Log.e("clip", "lsz len===hasClip bitmap1" + bitmap1);

                if (bitmap1 != null) {
                /*
                convert byte array to bitmap
                */
                    //Bitmap drawableicon = BitmapFactory.decodeResource(getResources(), R.drawable.liu2);
                    //byte[] imageData = bitmapToByteArray(drawableicon);
                    //Bitmap bitmap3 = BitmapFactory.decodeByteArray(imageData, 0, imageData.length);
                    imageView2.setImageBitmap(bitmap1);

                    //byte[] imageData = bitmapToByteArray(bitmap1);
                    imageData = bitmapToByteArray(bitmap1);

                    /*//bitmap to byteArray
                    int bytes = bitmap1.getByteCount();
                    ByteBuffer buf = ByteBuffer.allocate(bytes);
                    bitmap1.copyPixelsToBuffer(buf);
                    byte[] byteArray = buf.array();
                    Log.e("lszz","GoLog  dddbyteArray="+byteArray.length);
                    sendToPCIMG(byteArray);*/


                    //Bitmap bitmap = BitmapFactory.decodeByteArray(imageData, 0, imageData.length);
                    //imageView3.setImageBitmap(bitmap);

                } else {
                    Toast.makeText(TestActivity.this, " Clipboard img is empty", Toast.LENGTH_SHORT).show();
                }


            } else {
                Log.e("clip", "not support format");
                Toast.makeText(TestActivity.this, "lsz111 Clipboard is empty", Toast.LENGTH_SHORT).show();
            }
        }


    }


    public void copyTextToClipboard(String text) {
        ClipboardManager clipboard = (ClipboardManager) getSystemService(Context.CLIPBOARD_SERVICE);
        ClipData clip = ClipData.newPlainText("label", text);
        clipboard.setPrimaryClip(clip);
    }


    private void setClipToClipboard() {
        /*ClipboardUtils clipboardUtils = ClipboardUtils.getInstance();
        clipboardUtils.clearClip();

        AtomicReference<ClipData> clipDataRef = ClipboardUtils.createClipdataRef();
        //clipboardUtils.addTextItem(clipDataRef, "test text12a");
        clipboardUtils.addImageItem(clipDataRef, bitmap);
        // bitmapToByteArray(bitmap);
        clipboardUtils.setPrimaryClip(clipDataRef);*/


        Bitmap drawableicon = bitmap;//BitmapFactory.decodeResource(getResources(), R.drawable.liu2);
        ClipboardManager mClipboard = (ClipboardManager) getSystemService(Context.CLIPBOARD_SERVICE);
        Uri imageUri = getImageUri(this, drawableicon);
        ClipData theClip = ClipData.newUri(getContentResolver(), "Image", imageUri);
        mClipboard.setPrimaryClip(theClip);

    }

    public Uri getImageUri(Context inContext, Bitmap inImage) {
        ByteArrayOutputStream bytes = new ByteArrayOutputStream();
        inImage.compress(Bitmap.CompressFormat.JPEG, 100, bytes);
        String path = MediaStore.Images.Media.insertImage(inContext.getContentResolver(), inImage, "Title", null);
        return Uri.parse(path);
    }

    private void sendToPC(String text) {
        Log.i(TAG, "lsz GoLog sendToPC: text:" + text);
        new Thread(new Runnable() {
            @Override
            public void run() {
                Libp2p_clipboard.sendMessage(text);
            }
        }).start();
    }

    /*
    bitmap转数组
     */
    public static byte[] bitmapToByteArray(Bitmap bitmap) {
        ByteArrayOutputStream outputStream = new ByteArrayOutputStream();
        bitmap.compress(Bitmap.CompressFormat.PNG, 100, outputStream);
        Log.d("lsz", "outputStream. imag toByteArray()=" + outputStream.toByteArray().toString());
        return outputStream.toByteArray();

    }


    public void testCrossShare() throws IOException {

        // Get intent, action and MIME type
        intent = getIntent();
        action = intent.getAction();
        mimetype = intent.getType();
        Log.d("lszz", "action=" + action + "/type==" + mimetype);
        if (Intent.ACTION_SEND.equals(action) && mimetype != null) {
            if (mimetype.startsWith("image/")) {
                Uri imageUri = (Uri) intent.getParcelableExtra(Intent.EXTRA_STREAM);
                InputStream inputStream = this.getContentResolver().openInputStream(imageUri);
                long sizeInBytes = getImageSize(inputStream);
                sizeInMB = bytekb(sizeInBytes);
                Log.d("lszz", "sizeInMB=" + sizeInMB);
            } else if (mimetype.startsWith("video/mp4")) {
                Uri uri = (Uri) intent.getParcelableExtra(Intent.EXTRA_STREAM);
                InputStream inputStream = this.getContentResolver().openInputStream(uri);
                long sizeInBytes = getImageSize(inputStream);
                sizeInMB = bytekb(sizeInBytes);
            }
        }
    }

    private static final int REQUEST_EXTERNAL_STORAGE = 1;
    private static String[] PERMISSIONS_STORAGE = {
            android.Manifest.permission.WRITE_EXTERNAL_STORAGE
    };

    private void verifyStoragePermissions() {
        // Check if we have write permission
        int permission = ActivityCompat.checkSelfPermission(this, android.Manifest.permission.WRITE_EXTERNAL_STORAGE);

        if (permission != PackageManager.PERMISSION_GRANTED) {
            // We don't have permission so prompt the user
            ActivityCompat.requestPermissions(
                    this,
                    PERMISSIONS_STORAGE,
                    REQUEST_EXTERNAL_STORAGE
            );
        }
    }

    public String getRealPathFromURI(Context context, Uri contentUri) {
        String[] proj = {MediaStore.Images.Media.DATA};
        Cursor cursor = context.getContentResolver().query(contentUri, proj, null, null, null);
        int column_index = cursor.getColumnIndexOrThrow(MediaStore.Images.Media.DATA);
        cursor.moveToFirst();
        fileRealpath = cursor.getString(column_index);
        cursor.close();
        Log.d("lszz", "get file path=" + fileRealpath);
        return fileRealpath;
    }

    public String getPathFromMediaStoreUri(Context context, Uri uri) {
        String[] projection = {MediaStore.Images.Media.DISPLAY_NAME, MediaStore.Images.Media.DATA};
        try (Cursor cursor = context.getContentResolver().query(uri, projection, null, null, null)) {
            if (cursor != null && cursor.moveToFirst()) {
                int columnNameIndex = cursor.getColumnIndex(MediaStore.Images.Media.DISPLAY_NAME);
                int columnDataIndex = cursor.getColumnIndex(MediaStore.Images.Media.DATA);

                String displayName = cursor.getString(columnNameIndex);
                String filePath = cursor.getString(columnDataIndex);
                // If filePath is null, you might need to construct it manually based on the DISPLAY_NAME
                if (filePath == null || filePath.isEmpty()) {
                    final String[] split = context.getExternalFilesDir(null).toString().split("/");
                    final String[] directories = Arrays.copyOfRange(split, 0, split.length - 1);
                    final StringBuilder sb = new StringBuilder();
                    for (int i = 0; i < directories.length; i++) {
                        sb.append("/");
                        sb.append(directories[i]);
                    }
                    sb.append("/");
                    sb.append(displayName);
                    filePath = sb.toString();
                    saveFilePath = filePath;
                }
                Log.d("lszz", "get file filePath filePath=" + filePath);
                return filePath;
            }
        } catch (Exception e) {
            e.printStackTrace();
        }
        return null;
    }

    private String getFileNameFromUri(Uri zipUri) {
        Cursor returnCursor = getContentResolver().query(zipUri, null, null, null, null);
        /*
         * Get the column indexes of the data in the Cursor,
         * then get the data from the Cursor
         */
        int nameIndex = returnCursor.getColumnIndex(OpenableColumns.DISPLAY_NAME);
        returnCursor.moveToFirst();
        String fileName = returnCursor.getString(nameIndex);
        Log.d("lszz", "get fileName=" + fileName);
        returnCursor.close();
        return fileName;
    }

    private void playAudio(Uri audioUri) {
        MediaPlayer mediaPlayer = new MediaPlayer();
        try {
            mediaPlayer.setDataSource(this, audioUri);
            mediaPlayer.prepare();
            mediaPlayer.start();
        } catch (IOException e) {
            e.printStackTrace();
        }
    }


    private void downLoad(Uri uri) {
        try {
            InputStream inputStream = this.getContentResolver().openInputStream(uri);
            File saveDir = Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DOWNLOADS);//保存在内部存储的Download下
            File saveFile = new File(saveDir, getFileNameFromUri(uri));
            OutputStream outputStream = new FileOutputStream(saveFile);
            byte[] buffer = new byte[1024];
            int length = inputStream.read(buffer);
            while (length > 0) {
                outputStream.write(buffer, 0, length);
                length = inputStream.read(buffer);
            }
            inputStream.close();
            outputStream.close();
            Toast.makeText(TestActivity.this, "file has been save to Download of internal storage", Toast.LENGTH_SHORT).show();
        } catch (IOException e) {
            e.printStackTrace();
        }
    }

    //If a Log over 4000 string, AS can't print all of it
    public void printLongString(String longString) {
        int maxLength = 4000; // set max length of each substring
        int index = 0;
        int count = 0;

        while (index < longString.length()) {
            int endIndex = Math.min(index + maxLength, longString.length());
            String subString = longString.substring(index, endIndex);
            Log.d("lsz", "PartLog " + (count + 1) + ": " + subString);
            index += maxLength;
            count++;
        }
    }


    private void sendToPCIMG(byte[] value) {

        runOnUiThread(new Runnable() {
            @Override
            public void run() {
                //Log.i("lszzz", "GoLog sendToPC: img byte[] value:==" + value);
                Log.e("lszzz", "GoLog sendToPC: img byte[] value length:===" + value.length);

                base64String = Base64.encodeToString(value, Base64.DEFAULT);
                clearbase64String = removeInvalidCharacters(base64String);
                //printLongString(clearbase64String);
                Log.i("lszz", "GoLog clearbase64String===" + clearbase64String);
                Log.i("lszz", "GoLog clearbase64String.length()===" + clearbase64String.length());

                Libp2p_clipboard.sendImage(clearbase64String);

            }
        });

        /*byte[] valuea =Base64.decode(clearbase64String, Base64.DEFAULT);
        Log.i("lszzz", "GoLog sendToPC: img value[]==" + value.length);
        Bitmap bitmap = BitmapFactory.decodeByteArray(valuea, 0, valuea.length);
        imageView3.setImageBitmap(bitmap);
        imageView3.setVisibility(View.VISIBLE);*/

    }

    public static String encodeToStringWithoutSlash(byte[] data) {
        String base64String = Base64.encodeToString(data, Base64.DEFAULT);
        return base64String.replace("/", "");
    }

    public String base64StringToNormalString(String base64String) {
        byte[] decodedBytes = Base64.decode(base64String, Base64.DEFAULT);
        return new String(decodedBytes);
    }

    public String byteToString(byte[] data) {
        int index = data.length;
        for (int i = 0; i < data.length; i++) {
            if (data[i] == 0) {
                index = i;
                break;
            }
        }
        byte[] temp = new byte[index];
        Arrays.fill(temp, (byte) 0);
        System.arraycopy(data, 0, temp, 0, index);
        String str;
        try {
            str = new String(temp, "GBK");
        } catch (UnsupportedEncodingException e) {
            // TODO Auto-generated catch block
            e.printStackTrace();
            return "";
        }
        return str;
    }


    public byte[] getbyteArray(Bitmap bitmap) {
        int bytes = bitmap.getByteCount();
        ByteBuffer buf = ByteBuffer.allocate(bytes);
        bitmap.copyPixelsToBuffer(buf);
        byte[] byteArray = buf.array();
        return byteArray;
    }



    private long getImageSize(InputStream inputStream) throws IOException {
        return inputStream.available(); // 返回输入流中可用的字节数
    }


    public static String bytekb(long bytes) {
//格式化小数
        int GB = 1024 * 1024 * 1024;
        int MB = 1024 * 1024;
        int KB = 1024;

        if (bytes / GB >= 1) {
            double gb = Math.round((double) bytes / 1024.0 / 1024.0 / 1024.0 * 100.0) / 100.0;
            return String.format("%.2f", gb) + " GB";
        } else if (bytes / MB >= 1) {
            double mb = Math.round((double) bytes / 1024.0 / 1024.0 * 100.0) / 100.0;

            Log.i("lsz", "1111==" + String.format("%.2f", mb));
            return String.format("%.2f", mb) + " MB";
        } else if (bytes / KB >= 1) {
            double kb = Math.round((double) bytes / 1024.0 * 100.0) / 100.0;
            return String.format("%.2f", kb) + " KB";
        } else {
            return bytes + "B";
        }
    }

    public void getShare(Intent intent, String action, String mimetype) throws Exception {

        Log.d("lszz", "action=" + action + "/type=" + mimetype);
        if (Intent.ACTION_SEND.equals(action) && mimetype != null) {
            if ("text/plain".equals(mimetype)) {
                Uri uri = (Uri) intent.getParcelableExtra(Intent.EXTRA_STREAM);
                // getRealPathFromURI(mContext, uri);

                try {
                    //put share file to private directory of app, so libp2p can read it
                    InputStream inputStream = this.getContentResolver().openInputStream(uri);
                    File saveDir = getExternalFilesDir(null);//Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DOWNLOADS);//保存在内部存储的Download下
                    File saveFile = new File(saveDir, getFileNameFromUri(uri));
                    Log.i("lszz", "uri.getPath();:=saveDir " + saveDir.getPath());
                    Log.i("lszz", "uri.getPath();:=saveFile = " + saveFile.getPath());
                    saveFilePath = saveFile.getPath();
                    OutputStream outputStream = new FileOutputStream(saveFile);
                    byte[] buffer = new byte[1024];
                    int length = inputStream.read(buffer);
                    while (length > 0) {
                        outputStream.write(buffer, 0, length);
                        length = inputStream.read(buffer);
                    }
                    inputStream.close();
                    outputStream.close();
                } catch (IOException e) {
                    e.printStackTrace();
                }
            } else if (mimetype.startsWith("image/") || mimetype.startsWith("video/")) {
                Uri imageUri = (Uri) intent.getParcelableExtra(Intent.EXTRA_STREAM);
                Log.d("lszz", "imageUri=aa=" + imageUri);
                getPathFromMediaStoreUri(mContext, imageUri);

                //imageView2.setImageURI(imageUri);
                share_image.setImageURI(imageUri);

                try {
                    InputStream inputStream2 = this.getContentResolver().openInputStream(imageUri);
                    File saveDir = getExternalFilesDir(null);//Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DOWNLOADS);//保存在内部存储的Download下
                    File saveFile = new File(saveDir, getFileNameFromUri(imageUri));
                    Log.i("lszz", "uri.getPath();:=saveDir " + saveDir.getPath());
                    Log.i("lszz", "uri.getPath();:=saveFile = " + saveFile.getPath());
                    Log.i("lszz", "uri.getPath();:=saveFile = " + saveFile.getPath());
                    Log.i("lszz", "uri.getPath();:=saveFile = isimage" + saveFile.getName());
                    if(saveFile.getName().substring(saveFile.getName().length() -3).equals("png") ||
                            saveFile.getName().substring(saveFile.getName().length() -3).equals("jpg")) {
                        isimage = true;
                    }else{
                        isimage = false;
                    }
                    saveFilePath = saveFile.getPath();
                    //Toast.makeText(TestActivity.this, "文件已經保存在內部儲存空間的Download下", Toast.LENGTH_SHORT).show();
                    OutputStream outputStream = new FileOutputStream(saveFile);
                    byte[] buffer = new byte[1024];
                    int length = inputStream2.read(buffer);
                    while (length > 0) {
                        outputStream.write(buffer, 0, length);
                        length = inputStream2.read(buffer);
                    }
                    inputStream2.close();
                    outputStream.close();
                } catch (IOException e) {
                    e.printStackTrace();
                }

            } else if (mimetype.startsWith("application/")) {
                Uri uri = (Uri) intent.getParcelableExtra(Intent.EXTRA_STREAM);
                //getRealPathFromURI(mContext, uri);
                try {
                    InputStream inputStream = this.getContentResolver().openInputStream(uri);
                    File saveDir = getExternalFilesDir(null);
                    ;//Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DOWNLOADS);//保存在内部存储的Download下
                    File saveFile = new File(saveDir, getFileNameFromUri(uri));
                    Log.i("lszz", "uri.getPath();:=saveDir " + saveDir.getPath());
                    Log.i("lszz", "uri.getPath();:=saveFile = " + saveFile.getPath());
                    if(saveFile.getName().substring(saveFile.getName().length() -3).equals("png") ||
                            saveFile.getName().substring(saveFile.getName().length() -3).equals("jpg")) {
                        isimage = true;
                    }else{
                        isimage = false;
                    }
                    share_file_name= saveFile.getName();
                    saveFilePath = saveFile.getPath();
                    OutputStream outputStream = new FileOutputStream(saveFile);
                    byte[] buffer = new byte[1024];
                    int length = inputStream.read(buffer);
                    while (length > 0) {
                        outputStream.write(buffer, 0, length);
                        length = inputStream.read(buffer);
                    }
                    inputStream.close();
                    outputStream.close();
                } catch (IOException e) {
                    e.printStackTrace();
                }
            }


        }

    }

    public static String removeInvalidCharacters(String base64String) {
        String regex = "[^A-Za-z0-9+/=]";
        String cleanString = base64String.replaceAll(regex, "");
        return cleanString;
    }

    public Bitmap getBitmap(String privateFilePath) {
        Log.i(TAG, "getBitmap privateFilePath:" + privateFilePath);
        if (privateFilePath != null) {
            File file = new File(privateFilePath);
            if (file.exists()) {
                if (file.getName().substring(file.getName().length() - 3).equals("png") ||
                        file.getName().substring(file.getName().length() - 3).equals("jpg")) {

                    return BitmapFactory.decodeFile(file.getAbsolutePath());
                }
            } else {
                return null;
            }
        } else {
            Log.i(TAG, "getBitmap privateFilePath is null");
        }
        return null;
    }

    @Override
    protected void onStop() {
        Log.d(TAG, "lsz onStop");
        //finish();
        super.onStop();
    }

    @Override
    protected void onNewIntent(Intent intent) {
        Log.d(TAG, "lsz onNewIntent");
        super.onNewIntent(intent);

        setIntent(intent);


        boolean booleanValue = getIntent().getBooleanExtra("booleanKey", false);
        Log.d(TAG, "booleanValue booleanValue=" + booleanValue);
        if(booleanValue) {
            setContentView(R.layout.myactivity);
            RecyclerView recyclerView2 = findViewById(R.id.recycler_view);
            recyclerView2.setLayoutManager(new LinearLayoutManager(this));
            adapter = new FileTransferAdapter(fileTransferList);

            MyApplication.setMyAdapter(adapter);

            filename = getIntent().getStringExtra("filename");
            filesize = getIntent().getLongExtra("filesize", -1L);
            countSize = filesize;
            Log.d(TAG, "filename=" + filename);
            Log.d(TAG, "filesize=" + filesize);

            adapter.removeSameFile(filename, filesize);

            FileTransferItem item = new FileTransferItem(filename, filesize, null);
            fileTransferList.add(0, item);
            recyclerView2.setAdapter(adapter);

            file_page_back_icon = findViewById(R.id.file_page_back_icon);
            if (file_page_back_icon != null) {
                file_page_back_icon.setOnClickListener(new View.OnClickListener() {
                    @Override
                    public void onClick(View view) {
                        Log.d(TAG, "from file page to main page");
                        setContentView(R.layout.layout_testactivity_title);
                        getClientList();
                        mMyConnectionView = findViewById(R.id.my_connections_view);
                        if (mMyConnectionView != null) {
                            mMyConnectionView.setText(getMyIpDeviceName());
                        }
                        mSwVersionView = findViewById(R.id.sw_version_view);
                        if (mSwVersionView != null) {
                            mSwVersionView.setText(getSoftwareInfo());
                        }
                    }
                });
            }
        }
        Log.d(TAG, "mIsAndroidShareFile: "+ mIsAndroidShareFile);
        //update connection info on right-upper corner
        getClientList();
    }

    @Override
    protected void onDestroy() {
        Log.d(TAG, "lsz onDestroy activity");
        super.onDestroy();
        LocalBroadcastManager.getInstance(this).unregisterReceiver(broadcastReceiver);
        LocalBroadcastManager.getInstance(this).unregisterReceiver(broadcastReceivera);

//        if (isBound) {
//            // rempve callback
//            myService.setCallback(null);
//            unbindService(connection);
//            isBound = false;
//        }
        /*android.os.Process.killProcess(android.os.Process.myPid());*/
    }

    @Override
    protected void onRestart() {
        Log.d(TAG, "lsz onRestart");
        super.onRestart();
    }

    @Override
    protected void onPause() {
        Log.d(TAG, "lsz onPause");
        super.onPause();
        //recyclerView.setVisibility(View.GONE);
        // mbutton.setVisibility(View.GONE);
    }


//    public String getNetwork() {
//
//        String name = "";
//        try {
//            List<NetworkInterface> interfaces = Collections.list(NetworkInterface.getNetworkInterfaces());
//            for (NetworkInterface intf : interfaces) {
//                if (intf.getName().startsWith("wlan")) {
//
//                    return "wlan0";
//                } else if (intf.getName().startsWith("eth")) {
//                    return "ethernet";
//                }
//            }
//        } catch (Exception e) {
//            e.printStackTrace();
//        }
//        return name;
//    }


    public static boolean checkFloatPermission(Context context) {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.KITKAT) {
            return true;
        }
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.M) {
            try {
                Class<?> cls = Class.forName("android.content.Context");
                Field declaredField = cls.getDeclaredField("APP_OPS_SERVICE");
                declaredField.setAccessible(true);
                Object obj = declaredField.get(cls);
                if (!(obj instanceof String)) {
                    return false;
                }
                String str2 = (String) obj;
                obj = cls.getMethod("getSystemService", String.class).invoke(context, str2);
                cls = Class.forName("android.app.AppOpsManager");
                Field declaredField2 = cls.getDeclaredField("MODE_ALLOWED");
                declaredField2.setAccessible(true);
                Method checkOp = cls.getMethod("checkOp", Integer.TYPE, Integer.TYPE, String.class);
                int result = (Integer) checkOp.invoke(obj, 24, Binder.getCallingUid(), context.getPackageName());
                return result == declaredField2.getInt(cls);
            } catch (Exception e) {
                return false;
            }
        } else {
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                AppOpsManager appOpsMgr = (AppOpsManager) context.getSystemService(Context.APP_OPS_SERVICE);
                if (appOpsMgr == null) return false;
                int mode = appOpsMgr.checkOpNoThrow("android:system_alert_window", android.os.Process.myUid(), context.getPackageName());
                return mode == AppOpsManager.MODE_ALLOWED || mode == AppOpsManager.MODE_IGNORED;
            } else {
                return Settings.canDrawOverlays(context);
            }
        }
    }


    public void getFileList(String name, long siez, Bitmap bitmap, int a) {
        recyclerView2 = findViewById(R.id.my_recycler_view2);
        recyclerView2.setLayoutManager(new LinearLayoutManager(this, LinearLayoutManager.VERTICAL, false));

        List<GetFile> users = Arrays.asList(
                new GetFile(name, siez, bitmap, a)
        );

        //Toast.makeText(TestActivity.this, "文件已存入storage/emulated/0/Download/", Toast.LENGTH_SHORT).show();
        MyFileAdapter myadapter = new MyFileAdapter(users);
        recyclerView2.setAdapter(myadapter);

    }

    public void getClientList() {
        recyclerViewdevice = findViewById(R.id.recycler_devicelist);
        deviceList = new ArrayList<>();
        deviceNameIdMap = new HashMap<String, String>();
        String connectionsViewText = "";

        // IP1#ID1#Name1,IP2#ID2#Name2,IP3#ID3#Name3
        String getlist = Libp2p_clipboard.getClientList();
        if (!getlist.isEmpty()) {
            String[] strArray = getlist.split(",");
            for (String getlistvalue : strArray) {
                Log.d("lszz","getlistvalue=hhh"+getlistvalue);
                String[] info = getlistvalue.split("#");
                String ip = info[0];
                String id = info.length >1?info[1]:info[0];
                String name = info.length >2?info[2]:info[0];
                Log.d("lszz","name: "+name);
                if (name.contains(SOURCE_HDMI1)) {
                    deviceList.add(new Device(name, R.drawable.hdmi));
                } else if (name.contains(SOURCE_HDMI2)){
                    deviceList.add(new Device(name, R.drawable.hdmi2));
                } else if (name.contains(SOURCE_MIRACAST)){
                    deviceList.add(new Device(name, R.drawable.miracast));
                } else if (name.contains(SOURCE_USBC)){
                    deviceList.add(new Device(name, R.drawable.usb_p));
                } else {
                    deviceList.add(new Device(name, R.drawable.src_default));
                }
                deviceNameIdMap.put(name, id);
                connectionsViewText = connectionsViewText + ip + " " + name + "\n";
            }
            mConnectionsView = findViewById(R.id.connections_view);
            if (mConnectionsView != null) {
                mConnectionsView.setText(connectionsViewText);
            }
            //update connection text
            mFileConnCountView = findViewById(R.id.file_connection_count);
            if (deviceList != null) {
                Log.d(TAG, "getClientList: " + deviceList.size());
                if (mFileConnCountView != null) {
                    mFileConnCountView.setText(String.valueOf(deviceList.size()));
                } else {
                    Log.d(TAG, "getClientList, mFileConnCountView is null");
                }
            }
            mConnCountView = findViewById(R.id.connection_count);
            if (deviceList != null) {
                Log.d(TAG, "getClientList: " + deviceList.size());
                if (mConnCountView != null) {
                    mConnCountView.setText(String.valueOf(deviceList.size()));
                } else {
                    Log.d(TAG, "getClientList, mConnCountView is null");
                }
            }


        } else {
            //update connection text
            mFileConnCountView = findViewById(R.id.file_connection_count);
            if (deviceList != null) {
                Log.d(TAG, "getClientList: 0");
                if (mFileConnCountView != null) {
                    mFileConnCountView.setText("0");
                } else {
                    Log.d(TAG, "yiwen: getClientList, mFileConnCountView is null");
                }
            }
            mConnCountView = findViewById(R.id.connection_count);
            if (deviceList != null) {
                Log.d(TAG, "getClientList: 0");
                if (mConnCountView != null) {
                    mConnCountView.setText("0");
                } else {
                    Log.d(TAG, "getClientList, mConnCountView is null");
                }
            }
            mConnectionsView = findViewById(R.id.connections_view);
            if (mConnectionsView != null) {
                mConnectionsView.setText("NA");
            }
        }

        deviceAdapter = new DeviceAdapter(this, deviceList);

        // 设置RecyclerView为横向布局
        RecyclerView.LayoutManager layoutManager = new LinearLayoutManager(this, LinearLayoutManager.HORIZONTAL, false);
        //Log.d("lszz","recyclerViewdevice=hhh"+recyclerViewdevice);
        if (recyclerViewdevice != null) {
            recyclerViewdevice.setLayoutManager(layoutManager);
            recyclerViewdevice.setAdapter(deviceAdapter);


            deviceAdapter.setOnItemClickListener(new MyAdapter.OnItemClickListener() {
                @Override
                public void onItemClick(View view, int position) {
                    String name = ((TextView) view).getText().toString();
                    value = deviceNameIdMap.get(name);
                    Log.d(TAG, "select device name:"+name+", id:"+value);
                    Toast.makeText(TestActivity.this, "You select：" + name, Toast.LENGTH_SHORT).show();
                }
            });
        }
    }


    public static Bitmap base64ToBitmapa(String base64String) {
        // remove prefix of base64 encoding string if exists
        if (base64String.contains(",")) {
            base64String = base64String.split(",")[1];
        }

        byte[] decodedBytes = Base64.decode(base64String, Base64.DEFAULT);

        //for (byte aa : decodedBytes) {
        //    Log.d(TAG, "lsz byte[].toString(): " + aa);
        //}
        Log.i(TAG, "lszz bitmap decodedBytes[] length" + decodedBytes.length);
        return BitmapFactory.decodeByteArray(decodedBytes, 0, decodedBytes.length);
    }


    public void setBitmapToClipboard(Context context, Bitmap bitmap) {
        Log.i(TAG, "lsz setBitmapToClipboard init");
        if (!Environment.getExternalStorageState().equals(Environment.MEDIA_MOUNTED)) {
            return;
        }

        File file = new File(context.getExternalFilesDir(null), "shared_image.png");
        Log.i(TAG, "lsz getExternalStorageState imageFile getPath=" + file.getPath());
        //Uri imageUri = FileProvider.getUriForFile(context, "com.rtk.myapplication", file);

        try (FileOutputStream out = new FileOutputStream(file)) {
            bitmap.compress(Bitmap.CompressFormat.PNG, 100, out);
        } catch (IOException e) {
            e.printStackTrace();
            return;
        }

        // get file's Uri
        Uri imageUri = FileProvider.getUriForFile(context, "com.rtk.myapplication", file);
        Log.i(TAG, "lsz getExternalStorageState imageFile imageUri=" + imageUri);
        ClipData clip = ClipData.newUri(context.getContentResolver(), "image/png", imageUri);

        // get ClipboardManager
        ClipboardManager clipboard = (ClipboardManager) context.getSystemService(Context.CLIPBOARD_SERVICE);

        // put ClipData into clipboard
        Log.i(TAG, "lsz setimg to clipboard ");
        clipboard.setPrimaryClip(clip);
    }


    public void checkPermission(Context mContext) {
        if (checkFloatPermission(mContext) == true) {
            Log.d(TAG, "checkPermission ok, do nothing");
        } else {
            if (!Settings.canDrawOverlays(TestActivity.this)) {
                Intent intent = new Intent(Settings.ACTION_MANAGE_OVERLAY_PERMISSION,
                        Uri.parse("package:" + TestActivity.this.getPackageName()));
                TestActivity.this.startActivity(intent);
            }
            Toast.makeText(TestActivity.this, "Please open float window permission", Toast.LENGTH_SHORT).show();
        }
    }

    public void getfindViewById() {

        btnGetImage = findViewById(R.id.btnGetImage);
        btnSetImage = findViewById(R.id.btnSetImage);
        imageView2 = findViewById(R.id.imageView2);
        imageView3 = findViewById(R.id.imageView3);
        textView = findViewById(R.id.textView);
        bitmap = BitmapFactory.decodeResource(getResources(), R.drawable.aa);
        videview = findViewById(R.id.videoview);

        //getClientList();
        mbutton = findViewById(R.id.buttom);
        //mbuttonpaste = findViewById(R.id.buttom_paste);
        buttom_w = findViewById(R.id.buttom_w);
        buttom_r = findViewById(R.id.buttom_r);
        textView_name = findViewById(R.id.textView_name);
        textView_size = findViewById(R.id.textView_size);

        progress_bar = findViewById(R.id.progress_bar);
        recyclerView = findViewById(R.id.my_recycler_view);
        //  recyclerView2 = findViewById(R.id.my_recycler_view2);

        recyclerView.setLayoutManager(new LinearLayoutManager(this, LinearLayoutManager.VERTICAL, false));
        String getlist = Libp2p_clipboard.getClientList();
        if (!getlist.isEmpty()) {
            Log.d(TAG, "recyclerView getlist==" + getlist);
            String[] strArray = getlist.split("#");

            //put data to myadapter
            MyAdapter myadapter = new MyAdapter(strArray);
            myadapter.setOnItemClickListener(new MyAdapter.OnItemClickListener() {
                @Override
                public void onItemClick(View view, int position) {
                    value = ((TextView) view).getText().toString();
                    Toast.makeText(TestActivity.this, "You select：" + value, Toast.LENGTH_SHORT).show();

                }
            });

            recyclerView.setAdapter(myadapter);
        }
        //Press button to send file to libp2p
        mbutton.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {

                //path:saveFilePath
                //cliendid:value
                String path = saveFilePath;//"/storage/emulated/0/_FileCopyBehavior_V3.pptx";
                String cliendid = value;//"192.168.2.151:6668";
                Log.d(TAG, "file_copy, path+==path saveFilePath=" + path);
                Log.d(TAG, "file_copy, cliendid+==cliendid=" + cliendid);
                if (cliendid == null | path == null) {
                    Toast.makeText(TestActivity.this, "Please select a connection", Toast.LENGTH_SHORT).show();
                } else {
                    File file = new File(path);
                    if (file.exists()) {
                        long fileSize = file.length();
                        Libp2p_clipboard.sendCopyFile(path, cliendid, fileSize);
                    }
                }


            }

            //}
        });
    }



    public void findViewByIdDevice() {
        linearLayout = findViewById(R.id.linearLayout);
        linearLayout.setVisibility(View.GONE);
        recyclerViewdevice = findViewById(R.id.recycler_devicelist);
        share_image = findViewById(R.id.share_image);

        layout = findViewById(R.id.frame_file);
        share_file = findViewById(R.id.share_file);

        img_button = findViewById(R.id.img_button);
        img_button.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {
                String path = saveFilePath;//"/storage/emulated/0/_FileCopyBehavior_V3.pptx";
                String cliendid = value;//"192.168.2.151:6668";
                Log.d(TAG, "file_copy, path+==path saveFilePath=" + path);
                Log.d(TAG, "file_copy, cliendid+==cliendid=" + cliendid);
                if (cliendid == null | path == null) {
                    Toast.makeText(TestActivity.this, "Please select a connection", Toast.LENGTH_SHORT).show();
                } else {
                    File file = new File(path);
                    if (file.exists()) {
                        long fileSize = file.length();
                        Libp2p_clipboard.sendCopyFile(path, cliendid, fileSize);
                    }
                }
            }
        });
    }

    public void doShareMain() {

        Log.d(TAG, "path+==isimage isimage=" + isimage);
        if(isimage){
            share_file.setVisibility(View.VISIBLE);
            layout.setVisibility(View.GONE);
        }else{
            share_image.setVisibility(View.GONE);
            share_file.setVisibility(View.VISIBLE);
            share_file.setText(share_file_name);
        }

        back_icon = findViewById(R.id.back_icon);
        back_icon.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {
                finish();
            }
        });

        img_button = findViewById(R.id.img_button);
        img_button.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {
                String path = saveFilePath;//"/storage/emulated/0/_FileCopyBehavior_V3.pptx";
                String cliendid = value;//"192.168.2.151:6668";
                Log.d(TAG, "file_copy, path+==path saveFilePath=" + path);
                Log.d(TAG, "file_copy, cliendid+==cliendid=" + cliendid);
                if (cliendid == null | path == null) {
                    Toast.makeText(TestActivity.this, "Please select a connection", Toast.LENGTH_SHORT).show();
                } else {
                    File file = new File(path);
                    if (file.exists()) {
                        long fileSize = file.length();
                        Libp2p_clipboard.sendCopyFile(path, cliendid, fileSize);
                        Log.d(TAG, "file_copy, finish after sendCopyFile");
                        finish();
                    }
                }
            }
        });
    }

    @Override
    protected void onUserLeaveHint() {
        super.onUserLeaveHint();
        Log.d(TAG, "onUserLeaveHint: User press home to leave app");
        if (mIsAndroidShareFile) {
            Log.d(TAG, "leave app and this is android file share window, finish()");
            finish();
        }
    }
}
