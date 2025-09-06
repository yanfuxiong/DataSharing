package com.realtek.crossshare;

import android.app.AlertDialog;
import android.app.AppOpsManager;
import android.app.Service;
import android.content.ActivityNotFoundException;
import android.content.BroadcastReceiver;
import android.content.ClipData;
import android.content.ClipboardManager;
import android.content.ComponentName;
import android.content.ContentResolver;
import android.content.ContentUris;
import android.content.ContentValues;
import android.content.Context;
import android.content.DialogInterface;
import android.content.Intent;
import android.content.ServiceConnection;
import android.content.pm.PackageManager;
import android.database.Cursor;
import android.graphics.BitmapFactory;
import android.media.MediaPlayer;
import android.net.ConnectivityManager;
import android.net.LinkAddress;
import android.net.LinkProperties;
import android.net.Uri;
import android.net.wifi.WifiManager;
import android.os.Binder;
import android.os.Build;
import android.os.Bundle;
import android.os.Environment;
import android.os.Handler;
import android.os.IBinder;
import android.os.Looper;
import android.provider.DocumentsContract;
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
import android.view.View;
import android.view.ViewGroup;
import android.view.WindowManager;
import android.webkit.MimeTypeMap;
import android.widget.Button;
import android.widget.EditText;
import android.widget.ImageView;
import android.widget.LinearLayout;
import android.widget.ProgressBar;
import android.widget.TextView;
import android.widget.Toast;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;
import androidx.appcompat.app.AppCompatActivity;
import androidx.appcompat.widget.Toolbar;
import androidx.core.app.ActivityCompat;
import androidx.core.content.ContextCompat;
import androidx.core.content.FileProvider;

import androidx.fragment.app.Fragment;
import androidx.fragment.app.FragmentManager;
import androidx.fragment.app.FragmentTransaction;
import androidx.localbroadcastmanager.content.LocalBroadcastManager;
import androidx.recyclerview.widget.LinearLayoutManager;
import androidx.recyclerview.widget.RecyclerView;
import androidx.viewpager2.widget.ViewPager2;

import java.io.ByteArrayOutputStream;
import java.io.File;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.io.UnsupportedEncodingException;
import java.lang.reflect.Field;
import java.lang.reflect.Method;
import java.nio.ByteBuffer;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.concurrent.atomic.AtomicReference;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

import libp2p_clipboard.Libp2p_clipboard;

import android.graphics.Bitmap;
import android.widget.VideoView;

import com.google.android.material.bottomnavigation.BottomNavigationView;
import com.google.android.material.bottomsheet.BottomSheetDialog;
import com.google.android.material.tabs.TabLayout;
import com.google.zxing.client.android.Intents;
import com.tencent.mmkv.MMKV;

import android.content.IntentFilter;
import org.json.JSONArray;
import org.json.JSONObject;

import androidx.core.view.ViewCompat;
import androidx.core.view.WindowInsetsCompat;

import android.util.TypedValue;

public class TestActivity extends AppCompatActivity implements View.OnClickListener, RecordFragment.OnFileActionListener {

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
    String value, valueipid, valueip, valueid;
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
            mMyConnectionView, mSwVersionView,mDiasIDView;
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
    private List<Device> deviceList = new ArrayList<>();
    private Map<String, String> deviceNameIdMap;
    LinearLayout linearLayout;
    ImageView share_image, back_icon;
    Button img_button;
    boolean isimage =false;
    TextView share_file;
    LinearLayout layout;
    private static final String SOURCE_HDMI1 = "HDMI1";
    private static final String SOURCE_HDMI2 = "HDMI2";
    private static final String SOURCE_USBC1 = "USBC1";
    private static final String SOURCE_USBC2 = "USBC2";
    private static final String SOURCE_DP1 = "DP1";
    private static final String SOURCE_DP2 = "DP2";
    private static final String SOURCE_COMPUTER = "Computer";
    private static final String SOURCE_MIRACAST = "Miracast";
    private boolean mIsAndroidShareFile = false;
    ImageView file_page_back_icon,file_recored_back_icon;
    private static final String SETTINGS_DEBUG_QRCODE = "debug_qr_code";
    private final int REQUEST_CODE_SCAN = 100;
    private static final int CAMERA_PERMISSION_REQUEST_CODE = 100;
    private String paramValue;
    private static final int REQUEST_STORAGE_PERMISSION = 101;
    List<String> listfolder = new ArrayList<>();
    List<String> listfile = new ArrayList<>();
    private WifiManager.WifiLock wifiLock;
    private int mClickCount = 0;
    private Handler mCountHandler = new Handler(Looper.getMainLooper());
    private Runnable mResetRunnable = new Runnable() {
        @Override
        public void run() {
            mClickCount = 0;
        }
    };
    private  static String deviceip;
    private  static String deviceid;
    private  static long currenttimestamp;
    private AlertDialog dialog;
    private  static String openfilepath;
    private List<String> filePaths = new ArrayList<>();
    private String jsonString;

    private TabLayout tabLayout;
    private ViewPager2 viewPager;
    private BottomNavigationView bottomNavigationView;
    private Fragment currentFragment;
    private ShareFragment shareFragment;
    private RecordFragment recordFragment;
    private InfoFragment infoFragment;
    private static FileTransferItem items;
    public static boolean isForeground = false;
    private ImageView btnShare, btnRecord, btnInfo,imagecamera;
    private TextView textShare, textRecord, textInfo;
    private LinearLayout layoutShare,layoutRecord,layoutInfo;
    private List<String> pathList = new ArrayList<>();
    private Map<Long, List<String>> deviceFileMap = new HashMap<>();
    List<String> fileNamesList = new ArrayList<>();
    private List<Server> serverList = new ArrayList<>();

    private ServiceConnection connection = new ServiceConnection() {

        @Override
        public void onServiceConnected(ComponentName className, IBinder service) {
            FloatClipboardService.LocalBinder binder = (FloatClipboardService.LocalBinder) service;
            myService = binder.getService();
            isBound = true;
            myService.setActivityActive(true);
            myService.setCallback(new FloatClipboardService.DataCallback() {
                @Override
                public void onDataReceived(String name, double data, String ip, String id, long timestamp) {
                    runOnUiThread(() -> {
                        Log.i("lsz", "ServiceConnection get datadatadata==" + data);
                        deviceip=ip;
                        deviceid=id;
                        currenttimestamp= timestamp;
                        progress = (int) data;
                        MyApplication.getMyAdapter().updateProgress(name, progress);
                    });
                }

                @Override
                public void onMsgReceived(String name, String msg,long timestamp) {
                    LogUtils.i(TAG, "onMsgReceived msg==" + msg + ", name="+name + ", timestamp="+timestamp);

                    File appExternalFilesDir = getExternalFilesDir(null);
                    final String filesDir = appExternalFilesDir.getAbsolutePath();
                    List<String> completedFiles = new ArrayList<>();
                    List<String> filesToProcess = deviceFileMap.get(timestamp);
                    LogUtils.i(TAG, "onMsgReceived filesToProcess list=" +filesToProcess);
                    if (filesToProcess != null) {
                        for (String filePath : filesToProcess) {
                            File f = new File(filePath);

                            if (f.exists()) {
                                String relPath = filePath.substring(filesDir.length() + 1);
                                File destDir = Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DOWNLOADS);
                                File destFile = getUniqueDestFile(destDir, relPath);
                                destFile.getParentFile().mkdirs();
                                try (FileInputStream in = new FileInputStream(f);
                                     FileOutputStream out = new FileOutputStream(destFile)) {
                                    byte[] buffer = new byte[8192];
                                    int len;
                                    while ((len = in.read(buffer)) != -1) {
                                        out.write(buffer, 0, len);
                                    }
                                    LogUtils.i(TAG, "onMsgReceived err copy file success: " + destFile.getAbsolutePath() + ", timestamp="+timestamp);
                                } catch (IOException e) {
                                    LogUtils.e(TAG, "onMsgReceived copy err: " + f.getAbsolutePath() + " -> " + destFile.getAbsolutePath() + ", error: " + e.getMessage() + ", timestamp="+timestamp);
                                }
                                boolean deleted = f.delete();
                                LogUtils.i(TAG, "onMsgReceived src file  " + f.getAbsolutePath() + ", delet: " + deleted);
                                completedFiles.add(filePath);
                            } else {
                                LogUtils.i(TAG, "onMsgReceived err file: " + filePath);
                            }
                        }
                        deviceFileMap.remove(timestamp);

                        if (deviceFileMap.isEmpty()) {
                            LogUtils.i(TAG, "onMsgReceived All transfers finished. Now safe to delete empty dirs.");
                            deleteEmptyDirs(appExternalFilesDir);
                            LogUtils.i(TAG, "onMsgReceived deleteEmptyDirs done: " + appExternalFilesDir.getAbsolutePath());
                        } else {
                            LogUtils.i(TAG, "onMsgReceived Still active transfers: " + deviceFileMap.keySet() + ". Not deleting empty dirs.");
                        }

                    }

                    runOnUiThread(() -> {
                        if (recordFragment != null) {
                            recordFragment.onFileTransferError(msg ,timestamp);
                        }

                    });
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
                    openfilepath=path;
                    String filename = " ";
                    if (path != null && !path.isEmpty()) {
                        filename = path.substring(path.lastIndexOf("/") + 1);
                    }
                    MyApplication.getMyAdapter().getFileTime(filename);
                }

                @Override
                public void sendFileListinfo(String ip,String id,String devicename, String currentFileName, long sentFileCnt, long totalFileCnt, long currentFileSize, long totalSize, long sentSize,long percentage,List<String> filePathList,long timestamp) {
                    deviceip=ip;
                    deviceid=id;
                    currenttimestamp= timestamp;
                    runOnUiThread(() -> {
                        recordFragment.updateData(deviceip,deviceid,currenttimestamp,items);
                        if(percentage == 100){
                            recordFragment.notifyAllUI();
                        }

                    });
                    openfilepath=currentFileName;
                    File file = new File(currentFileName);
                    String fileName = file.getName();
                    MyApplication.getMyAdapter().updateFileList(devicename, fileName,sentFileCnt,totalFileCnt,currentFileSize,totalSize,sentSize,percentage,timestamp);

                    //filesFolder is "/storage/emulated/0/Android/data/com.realtek.crossshare/files"
                    File appExternalFilesDir = getExternalFilesDir(null);
                    final String filesDir = appExternalFilesDir.getAbsolutePath();
                    pathList=filePathList;
                    deviceFileMap.put(timestamp, filePathList);
                    if ((int)percentage == 100) {
                        for (String filePath : filePathList) {
                            File f = new File(filePath);
                            if (f.exists()) {
                                String relPath = filePath.substring(filesDir.length() + 1);
                                File destDir = Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DOWNLOADS);
                                File destFile = getUniqueDestFile(destDir, relPath);
                                destFile.getParentFile().mkdirs();
                                try (FileInputStream in = new FileInputStream(f);
                                     FileOutputStream out = new FileOutputStream(destFile)) {
                                    byte[] buffer = new byte[8192];
                                    int len;
                                    while ((len = in.read(buffer)) != -1) {
                                        out.write(buffer, 0, len);
                                    }
                                    LogUtils.i(TAG, "sendFileListinfo copy file success: " + destFile.getAbsolutePath() + ", timestamp="+timestamp);
                                } catch (IOException e) {
                                    LogUtils.e(TAG, "sendFileListinfo copy file err: " + f.getAbsolutePath() + " -> " + destFile.getAbsolutePath() + ", error: " + e.getMessage()  + ", timestamp="+timestamp);
                                }

                                MyApplication.getMyAdapter().updateFileTransferDisplayName(timestamp , destFile.getName(),destFile.getAbsolutePath());

                                boolean deleted = f.delete();
                                LogUtils.i(TAG, "sendFileListinfo src file " + f.getAbsolutePath() + ", delet: " + deleted);
                            } else {
                                LogUtils.i(TAG, "sendFileListinfo file not exist: " + filePath);
                            }
                        }

                        deviceFileMap.remove(timestamp);
                        if (deviceFileMap.isEmpty()) {
                            LogUtils.i(TAG, "sendFileListinfo All transfers finished. Now safe to delete empty dirs.");
                            deleteEmptyDirs(appExternalFilesDir);
                            LogUtils.i(TAG, "sendFileListinfo deleteEmptyDirs done: " + appExternalFilesDir.getAbsolutePath());
                        } else {
                            LogUtils.i(TAG, "sendFileListinfo Still active transfers: " + deviceFileMap.keySet() + ". Not deleting empty dirs.");
                        }

                    }

                }

                @Override
                public void sendFileDone(String filesInfo, String platform, String deviceName, long timestamp) {
                    LogUtils.i(TAG, "sendFileDone getFilePaths ="+MyApplication.getFilePaths());
                    filePaths=MyApplication.getFilePaths();
                    for (String path : filePaths) {
                        File f = new File(path);
                        if(f.exists()){
                            boolean deleted = f.delete();
                            LogUtils.i(TAG, "sendFileDone src file " + f.getAbsolutePath() + ", delet: " + deleted);
                        }
                    }
                }

                @Override
                public void sendFileitem(FileTransferItem item) {
                    items = item;
                }

                @Override
                public void disPlayinfoCallback(boolean bool) {
                    //todo
                    runOnUiThread(() -> {
                        if (shareFragment != null) {
                            shareFragment.setMiarcastState(bool);
                        }
                    });
                }

                @Override
                public void onSetMonitorName(String monitorName) {
                    runOnUiThread(() -> {
                        if (shareFragment != null) {
                            Log.i(TAG,"callbackSetMonitorName="+monitorName);
                            shareFragment.setMonitorName(monitorName);
                        }
                    });
                }

                @Override
                public void onSetUpdateDiasStatus(long status) {

                    runOnUiThread(() -> {
                        if (shareFragment != null) {
                            Log.i(TAG,"onSetUpdateDiasStatus status="+status);
                            shareFragment.setDiasStatus(status);
                        }
                    });
                }

                @Override
                public void onGetClientlist() {
                    getClientList();
                    String clientdevice = Libp2p_clipboard.getClientList();
                    runOnUiThread(() -> {

                        if (shareFragment != null) {
                            if (!clientdevice.isEmpty()) {
                                shareFragment.updateClientlist(clientdevice);
                            }else {
                                shareFragment.updateClientlist(null);
                            }
                        }
                    });
                }

                @Override
                public void onGetLanserverInfo(String monitorName, String instance, String ipAddr, String version, long timestamp) {
                    
                    runOnUiThread(() -> {
                        Log.i(TAG, "onGetLanserverInfo monitorNam="+monitorName + ", instance="+instance +", ipAddr="+ipAddr + ", version="+version+ ", timestamp="+timestamp);

                        Server newServer = new Server(monitorName, instance, ipAddr);
                        serverList.add(newServer);
                        MyApplication.setServerList(serverList);

                        if (shareFragment != null) {
                            shareFragment.addServer(monitorName,instance,ipAddr,version,timestamp);
                        }
                    });
                }

                @Override
                public void onUpdateClientVersion(String clienVersion) {
                    runOnUiThread(() -> {
                        LogUtils.i(TAG, "onUpdateClientVersion="+clienVersion);
                        appUpgrade(MyApplication.getContext());
                    });
                }

                @Override
                public void onFileListDragNotify() {
                    runOnUiThread(() -> {
                        boolean isFinishing = isFinishing();
                        boolean isDestroyed = isDestroyed();
                        LogUtils.i(TAG, "onFileListDragNotify, isFinishing:" + isFinishing +
                                ", isDestroyed:" + isDestroyed);
                        if (!isFinishing && !isDestroyed) {
                            LogUtils.i(TAG, "onFileListDragNotify, switch to record fragment");
                            switchFragment(recordFragment, "record_fragment");
                            setSelectedTab(1);
                            if (recordFragment != null) {
                                recordFragment.selectTab(0);
                            }
                        }
                    });
                }

            });
            //user may swipe app from history, so we need to update info again
            Log.d(TAG, "onServiceConnected, try getClientList() once");
            getClientList();
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

        if (Intent.ACTION_SEND_MULTIPLE.equals(action)) {
            List<Uri> sharedUris = intent.getParcelableArrayListExtra(Intent.EXTRA_STREAM);
            Log.i(TAG, "onCreate: sharedUris====" + sharedUris);
            if (sharedUris != null) {
                for (Uri uri : sharedUris) {
                    try {
                        InputStream inputStream = this.getContentResolver().openInputStream(uri);
                        File saveDir = getExternalFilesDir(null);
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
                        saveFilePath = saveFile.getPath();
                        filePaths.add(saveFilePath);
                        MyApplication.setFilePaths(filePaths);
                        fileNamesList.add(saveFile.getName());
                    } catch (Exception e) {
                        e.printStackTrace();
                    }

                }
            }

            getClientList();
            Intent intent = new Intent(this, ShareActivity.class);
            intent.putStringArrayListExtra("filePaths", new ArrayList<>(filePaths));
            intent.putStringArrayListExtra("fileNamesList", new ArrayList<>(fileNamesList));
            startActivity(intent);
            finish();

        } else if (Intent.ACTION_SEND.equals(action)) {
            setTheme(R.style.TransparentTheme);

            try {
                getShare(intent, action, mimetype);
            } catch (IOException e) {
                throw new RuntimeException(e);
            } catch (Exception e) {
                throw new RuntimeException(e);
            }

            getClientList();
            Intent intent = new Intent(this, ShareActivity.class);
            intent.putStringArrayListExtra("filePaths", new ArrayList<>(filePaths));
            intent.putStringArrayListExtra("fileNamesList", new ArrayList<>(fileNamesList));
            startActivity(intent);
            finish();

        } else {
            setTheme(R.style.Theme_MyApplication);
            //getWindow().setBackgroundDrawableResource(android.R.color.white);
            //Log.i(TAG, "onCreate: checkPermission");

            //setContentView(R.layout.myactivity);
            //file_page_back_icon = findViewById(R.id.file_page_back_icon);
            //file_page_back_icon.setOnClickListener(this);

            setContentView(R.layout.layout_main_activity);
            checkPermission(mContext);
            //Toolbar toolbar = findViewById(R.id.toolbar);
            //setSupportActionBar(toolbar);

            if (savedInstanceState == null) {
                shareFragment = new ShareFragment();
                recordFragment = new RecordFragment();
                infoFragment = new InfoFragment();
                getSupportFragmentManager().beginTransaction()
                        .add(R.id.fragment_container, shareFragment, "share_fragment")
                        .add(R.id.fragment_container, recordFragment, "record_fragment")
                        .add(R.id.fragment_container, infoFragment, "info_fragment")
                        .hide(recordFragment)
                        .hide(infoFragment)
                        .commitNow();
                currentFragment = shareFragment;
            } else {
                shareFragment = (ShareFragment) getSupportFragmentManager().findFragmentByTag("share_fragment");
                recordFragment = (RecordFragment) getSupportFragmentManager().findFragmentByTag("record_fragment");
                infoFragment = (InfoFragment) getSupportFragmentManager().findFragmentByTag("info_fragment");

                if (shareFragment != null && !shareFragment.isHidden()) {
                    currentFragment = shareFragment;
                } else if (recordFragment != null && !recordFragment.isHidden()) {
                    currentFragment = recordFragment;
                } else if (infoFragment != null && !infoFragment.isHidden()) {
                    currentFragment = infoFragment;
                }
            }

            View bottomBar = findViewById(R.id.bottom_bar);
            if(bottomBar != null) {
                ViewCompat.setOnApplyWindowInsetsListener(bottomBar, (v, insets) -> {
                    int bottomInset = insets.getInsets(WindowInsetsCompat.Type.systemBars()).bottom;
                    v.setPadding(0, 0, 0, bottomInset);
                    return insets;
                });
            }

            btnShare = findViewById(R.id.btn_share);
            btnRecord = findViewById(R.id.btn_record);
            btnInfo = findViewById(R.id.btn_info);
            textShare = findViewById(R.id.textshare);
            textRecord = findViewById(R.id.textrecord);
            textInfo = findViewById(R.id.textinfo);


            layoutShare = findViewById(R.id.layout_share);
            layoutRecord = findViewById(R.id.layout_record);
            layoutInfo = findViewById(R.id.layout_info);

            imagecamera = findViewById(R.id.toolbar_camera);
            imagecamera.setOnClickListener(v -> {
                checkCameraPermission();
            });

            layoutShare.setOnClickListener(v -> {
                switchFragment(shareFragment, "share_fragment");
                setSelectedTab(0);
                //imagecamera.setVisibility(View.VISIBLE);
            });
            layoutRecord.setOnClickListener(v -> {
                switchFragment(recordFragment, "record_fragment");
                setSelectedTab(1);
                imagecamera.setVisibility(View.GONE);
            });
            layoutInfo.setOnClickListener(v -> {
                switchFragment(infoFragment, "info_fragment");
                setSelectedTab(2);
                imagecamera.setVisibility(View.GONE);
            });

            if (booleanValue) {
                switchFragment(recordFragment, "record_fragment");
                setSelectedTab(1);
            } else {
                switchFragment(shareFragment, "share_fragment");
                setSelectedTab(0);
                //imagecamera.setVisibility(View.VISIBLE);
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
        handleIntent(getIntent());
        //createExternalFile(this);
        checkPermissionsAndCopy();
        creatWifiLock();
    }

    public void creatWifiLock(){
        Log.d(TAG, "WifiLock acquired");
        // get WifiLock
        WifiManager wm = (WifiManager) getApplicationContext().getSystemService(WIFI_SERVICE);
        wifiLock = wm.createWifiLock(
                WifiManager.WIFI_MODE_FULL_HIGH_PERF,
                "MyApp:WifiLock"
        );
        wifiLock.acquire();
    }

    public static String getFirstPart(String path) {
        int index = path.indexOf('/');
        if (index == -1) {
            return path;
        } else {
            return path.substring(0, index);
        }
    }

    private void copyFolderToDownloads(String Dirname) {
        File sourceDir = new File(getExternalFilesDir(null), Dirname);
        if (!sourceDir.exists() || !sourceDir.isDirectory()) {
            //Toast.makeText(this, "not exit", Toast.LENGTH_SHORT).show();
            return;
        }

        new Thread(() -> {
            try {
                copyFolderRecursively(sourceDir, Environment.DIRECTORY_DOWNLOADS);
                runOnUiThread(() -> Toast.makeText(this, "copy ok", Toast.LENGTH_SHORT).show());

                deleteRecursive(sourceDir);
            } catch (IOException e) {
                //runOnUiThread(() -> Toast.makeText(this, "copy ng: " + e.getMessage(), Toast.LENGTH_SHORT).show());
            }
        }).start();
    }


    private void copyFolderRecursively(File sourceDir, String targetParentPath) throws IOException {
        File[] files = sourceDir.listFiles();
        if (files == null) return;

        for (File file : files) {
            String targetRelativePath = targetParentPath + File.separator + sourceDir.getName();
            if (file.isDirectory()) {
                copyFolderRecursively(file, targetRelativePath);
            } else {
                copySingleFile(file, targetRelativePath);
            }
        }
    }


    private void copySingleFile(File sourceFile, String targetRelativePath) throws IOException {
        ContentResolver resolver = getContentResolver();
        ContentValues values = new ContentValues();
        values.put(MediaStore.MediaColumns.DISPLAY_NAME, sourceFile.getName());
        values.put(MediaStore.MediaColumns.MIME_TYPE, getMimeType(sourceFile.getName()));
        values.put(MediaStore.MediaColumns.RELATIVE_PATH, targetRelativePath);

        Uri targetUri = null;
        if (android.os.Build.VERSION.SDK_INT >= android.os.Build.VERSION_CODES.Q) {
            targetUri = resolver.insert(MediaStore.Downloads.EXTERNAL_CONTENT_URI, values);
        }
        if (targetUri == null) throw new IOException("not mkdir: " + sourceFile.getName());

        try (InputStream in = new FileInputStream(sourceFile);
             OutputStream out = resolver.openOutputStream(targetUri)) {
            if (out == null) throw new IOException("not write");
            byte[] buffer = new byte[1024];
            int len;
            while ((len = in.read(buffer)) > 0) {
                out.write(buffer, 0, len);
            }
        }
    }

    private String getMimeType(String fileName) {
        String extension = MimeTypeMap.getFileExtensionFromUrl(fileName);
        return extension != null
                ? MimeTypeMap.getSingleton().getMimeTypeFromExtension(extension)
                : "application/octet-stream";
    }

    private static boolean deleteRecursive(File fileOrDirectory) {
        if (fileOrDirectory == null || !fileOrDirectory.exists()) {
            return false;
        }
        if (fileOrDirectory.isDirectory()) {
            File[] children = fileOrDirectory.listFiles();
            if (children != null) {
                for (File child : children) {
                    deleteRecursive(child);
                }
            }
        }
        return fileOrDirectory.delete();
    }

    private void checkPermissionsAndCopy() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            if (checkSelfPermission(android.Manifest.permission.WRITE_EXTERNAL_STORAGE) != PackageManager.PERMISSION_GRANTED) {
                requestPermissions(
                        new String[]{android.Manifest.permission.WRITE_EXTERNAL_STORAGE},
                        REQUEST_STORAGE_PERMISSION
                );
            } else {
                //copyFolderToDownloads();
            }
        } else {
            //copyFolderToDownloads();
        }
    }

    public String copyFileToPublicDir(String privateFilePath) {
        boolean isRetry = false;
        FileInputStream fis = null;
        FileOutputStream fos = null;
        File srcFile = new File(privateFilePath);
        File saveDir = Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DOWNLOADS);
        File destFile = new File(saveDir, srcFile.getName());
        openfilepath= destFile.getPath();
        //if (destFile.exists()) {

        //}
        try {
            fis = new FileInputStream(srcFile);
            fos = new FileOutputStream(destFile);

            byte[] buffer = new byte[1024];
            int bytesRead;
            long fileSize = srcFile.length();
            long totalBytesRead = 0;

            while ((bytesRead = fis.read(buffer)) != -1) {
                fos.write(buffer, 0, bytesRead);
                totalBytesRead += bytesRead;

                //Log.i("lszz", "get totalBytesRead =" + totalBytesRead);
                if (totalBytesRead >= fileSize) {
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

    //add file to pass lan server certification:
    //"/storage/emulated/0/Android/data/com.realtek.crossshare/files/ID.SrcAndPort"
    public static void createExternalFile(Context context) {
        String fileName = "ID.SrcAndPort";
        String content = "13,0";

        File externalDir = context.getExternalFilesDir(null);
        if (externalDir == null) {
            return;
        }

        File targetFile = new File(externalDir, fileName);
        try {
            FileOutputStream fos = new FileOutputStream(targetFile);
            fos.write(content.getBytes());
            fos.close();
        } catch (IOException e) {
            e.printStackTrace();
        }
    }

    private void handleIntent(Intent intent) {
        Uri data = intent.getData();
        if (data != null) {
            String scheme = data.getScheme();
            if ("crossshare".equals(scheme)) {
                String param = data.getQueryParameter("param");
                //mDiasIDView.setText(param);
                kv.encode("paramValue", param);
                Log.i(TAG, "handleIntent get qr code param=="+param);
                if (!TextUtils.isEmpty(param)) {
                    new Handler().postDelayed(new Runnable() {
                        @Override
                        public void run() {
                            Intent broadcastIntent = new Intent("com.corssshare.qrparam");
                            broadcastIntent.putExtra("param", param);
                            Log.i(TAG, "GoLog handleIntent sendBroadcast qrparam");
                            broadcastIntent.setPackage(getPackageName());
                            LocalBroadcastManager.getInstance(mContext).sendBroadcast(broadcastIntent);
                        }
                    }, 1000);
                }
            }
        }
    }

    private String getMyIpDeviceName() {
        return getWifiIpAddress(MyApplication.getContext()) + " " +
                Settings.Global.getString(MyApplication.getContext().getContentResolver(), "device_name");
    }

    public String getMyIp() {
        return getWifiIpAddress(MyApplication.getContext()) ;
    }

    public String DeviceName() {
        return Settings.Global.getString(MyApplication.getContext().getContentResolver(), "device_name");
    }

    public String getSoftwareInfo() {
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
        isForeground = true;
        mMyConnectionView = findViewById(R.id.my_connections_view);
        if (mMyConnectionView != null) {
            mMyConnectionView.setText(getMyIpDeviceName());
        }

        if (checkFloatPermission(mContext) == true) {
            Intent serviceIntent = new Intent(MyApplication.getContext(), FloatClipboardService.class);
            Log.d(TAG, "startForegroundService in onResume");
            startForegroundService(serviceIntent);
            Intent intent = new Intent(MyApplication.getContext(), FloatClipboardService.class);
            bindService(intent, connection, Context.BIND_AUTO_CREATE);
            String latestVersion = kv.decodeString("LATEST_VERSION", " ");
            String currentVersion = Libp2p_clipboard.getVersion();
            boolean needUpgrade = kv.decodeBool("NEED_UPGRADE", false);
            LogUtils.i(TAG, "onResume isBound ="+isBound + ", myService="+myService + ", latestVersion=" +latestVersion + ", currentVersion=" +currentVersion+", needUpgrade=" +needUpgrade);
            if (!latestVersion.isEmpty() && currentVersion.compareTo(latestVersion) < 0) {
                needUpgrade = true;
                kv.encode("NEED_UPGRADE", true);
            } else {
                needUpgrade = false;
                kv.encode("NEED_UPGRADE", false);
            }
            LogUtils.i(TAG, "onResume needUpgrade ="+needUpgrade);
            if (needUpgrade) {
                if (dialog == null || !dialog.isShowing()) {
                    appUpgrade(MyApplication.getContext());
                    kv.encode("NEED_UPGRADE", false);
                }
            }
            if (isBound && myService != null) {
                Log.d(TAG, "startForegroundService in onResume setActivityActive");
                myService.setActivityActive(true);
            }

        } else {
            Log.d(TAG, "startForegroundService in onResume: skip, due to no permission");
        }
        showScanQrCodesButtonIfNeeded();
    }

    private void showScanQrCodesButtonIfNeeded() {
        LinearLayout ly = (LinearLayout)findViewById(R.id.linearlayout_qrcode);
        TextView diasIdTitle = (TextView)findViewById(R.id.dias_view_title);
        TextView diasId = (TextView)findViewById(R.id.dias_view);
        if (ly != null && diasIdTitle != null && diasId != null) {
            if (Settings.Global.getInt(MyApplication.getContext().getContentResolver(), SETTINGS_DEBUG_QRCODE, 0) == 1) {
                Log.d(TAG, "showScanQrCodesButtonIfNeeded, get 1, show it");
                ly.setVisibility(View.VISIBLE);
                diasIdTitle.setVisibility(View.VISIBLE);
                diasId.setVisibility(View.VISIBLE);
                paramValue = kv.decodeString("paramValue");
                diasId.setText(paramValue);
            } else {
                Log.d(TAG, "showScanQrCodesButtonIfNeeded, get 0, don't show it");
                ly.setVisibility(View.INVISIBLE);
                diasIdTitle.setVisibility(View.INVISIBLE);
                diasId.setVisibility(View.INVISIBLE);
            }
        }
        // todo: remove this after fixing "new diasid can connect old diasid's devices"
        TextView crossShareTitleTv = findViewById(R.id.crossshare_title);
        if (crossShareTitleTv != null) {
            crossShareTitleTv.setOnClickListener(new View.OnClickListener() {
                @Override
                public void onClick(View v) {
                    mClickCount++;
                    Log.i(TAG, "crossShareTitleTv onclick: "+mClickCount);

                    // Cancel any existing reset timer
                    mCountHandler.removeCallbacks(mResetRunnable);

                    if (mClickCount >= 7) {
                        Toast.makeText(v.getContext(), "clear diasid in storage", Toast.LENGTH_SHORT).show();
                        kv.removeValueForKey("paramValue");
                        mClickCount = 0;
                    } else {
                        // Restart timer to reset after 5 seconds of last click
                        mCountHandler.postDelayed(mResetRunnable, 5000);
                    }
                }
            });
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
        if (requestCode == CAMERA_PERMISSION_REQUEST_CODE) {
            if (grantResults.length > 0 && grantResults[0] == PackageManager.PERMISSION_GRANTED) {
                startBarcodeScan();
            } else {
                Toast.makeText(this, "Need camera permission,please try", Toast.LENGTH_SHORT).show();
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
                    imageView2.setImageBitmap(bitmap1);
                    imageData = bitmapToByteArray(bitmap1);
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
                    filePaths.add(saveFilePath);
                    MyApplication.setFilePaths(filePaths);
                    share_file_name= saveFile.getName();
                    fileNamesList.add(share_file_name);
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
                //share_image.setImageURI(imageUri);

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
                    filePaths.add(saveFilePath);
                    MyApplication.setFilePaths(filePaths);
                    fileNamesList.add(saveFile.getName());
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
                    fileNamesList.add(share_file_name);
                    saveFilePath = saveFile.getPath();
                    filePaths.add(saveFilePath);
                    MyApplication.setFilePaths(filePaths);
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
            }else{
                Uri uri = (Uri) intent.getParcelableExtra(Intent.EXTRA_STREAM);
                try {
                    InputStream inputStream = this.getContentResolver().openInputStream(uri);
                    File saveDir = getExternalFilesDir(null);
                    File saveFile = new File(saveDir, getFileNameFromUri(uri));
                    share_file_name= saveFile.getName();
                    fileNamesList.add(share_file_name);
                    saveFilePath = saveFile.getPath();
                    filePaths.add(saveFilePath);
                    MyApplication.setFilePaths(filePaths);
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
        if (isBound && myService != null) {
            myService.setActivityActive(false);
        }
        if (dialog != null && dialog.isShowing()) {
            dialog.dismiss();
        }
    }

    @Override
    protected void onNewIntent(Intent intent) {
        Log.d(TAG, "onNewIntent");
        super.onNewIntent(intent);

        handleIntent(intent);
        setIntent(intent);

        boolean booleanValue = getIntent().getBooleanExtra("booleanKey", false);
        Log.d(TAG, "booleanValue booleanValue=" + booleanValue);
    }

    @Override
    protected void onDestroy() {
        Log.d(TAG, "lsz onDestroy activity");
        if (wifiLock != null && wifiLock.isHeld()) {
            wifiLock.release();
            Log.d(TAG, "onDestroy WifiLock released");
        }
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
        isForeground = false;
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
        List<Device> latestList = new ArrayList<>();
        deviceNameIdMap = new HashMap<String, String>();
        String getlist = Libp2p_clipboard.getClientList();
        if (!getlist.isEmpty()) {
            String[] strArray = getlist.split(",");
            for (String getlistvalue : strArray) {
                String[] info = getlistvalue.split("#");
                String ip = info[0];
                String id = info.length > 1 ? info[1] : info[0];
                String name = info.length > 2 ? info[2] : info[0];
                String sourcePortType = info.length > 3 ? info[3] : info[0];
                int iconRes;
                if (sourcePortType.contains(SOURCE_HDMI1)) {
                    iconRes = R.drawable.hdmi;
                } else if (sourcePortType.contains(SOURCE_HDMI2)) {
                    iconRes = R.drawable.hdmi2;
                } else if (sourcePortType.contains(SOURCE_MIRACAST)) {
                    iconRes = R.drawable.miracast;
                } else if (sourcePortType.contains(SOURCE_USBC1)) {
                    iconRes = R.drawable.usb_c1;
                }else if (sourcePortType.contains(SOURCE_USBC2)) {
                    iconRes = R.drawable.usb_c2;
                }else if (sourcePortType.contains(SOURCE_DP1)) {
                    iconRes = R.drawable.dp1;
                }else if (sourcePortType.contains(SOURCE_DP2)) {
                    iconRes = R.drawable.dp2;
                }else if (sourcePortType.contains(SOURCE_COMPUTER)) {
                    iconRes = R.drawable.computer;
                } else {
                    iconRes = R.drawable.src_default;
                }
                latestList.add(new Device(name, ip, iconRes));
                deviceNameIdMap.put(name, id + ":" + ip);
                MyApplication.setDeviceNameIdMap(deviceNameIdMap);
            }
        }

        deviceList.clear();
        deviceList.addAll(latestList);
        if(deviceList != null) {
            MyApplication.setDeviceList(deviceList);
        }
        if (deviceAdapter != null) deviceAdapter.notifyDataSetChanged();
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
        //Uri imageUri = FileProvider.getUriForFile(context, "com.realtek.crossshare", file);

        try (FileOutputStream out = new FileOutputStream(file)) {
            bitmap.compress(Bitmap.CompressFormat.PNG, 100, out);
        } catch (IOException e) {
            e.printStackTrace();
            return;
        }

        // get file's Uri
        Uri imageUri = FileProvider.getUriForFile(context, "com.realtek.crossshare", file);
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

    public void sendjson(String id, String ip) {

        try {
            JSONObject jsonObject = new JSONObject();
            jsonObject.put("Id", id);
            jsonObject.put("Ip", ip);

            JSONArray pathArray = new JSONArray();
            for (String path : filePaths) {
                pathArray.put(path);
            }
            jsonObject.put("PathList", pathArray);

            jsonString = jsonObject.toString();
            Log.i(TAG, "sendjson jsonString=" + jsonString);
        } catch (
                Exception e) {
            e.printStackTrace();
        }

    }

    @Override
    protected void onUserLeaveHint() {
        super.onUserLeaveHint();
        Log.d(TAG, "onUserLeaveHint: User press home to leave app 0611");
//        if (mIsAndroidShareFile) {
//            Log.d(TAG, "leave app and this is android file share window, finish()");
//            finish();
//        }
    }

    @Override
    public void onClick(View v) {
        if(v.getId() == R.id.linearlayout_qrcode){
            Log.i(TAG,"onClick qr butt");
            checkCameraPermission();
        }

        if(v.getId() == R.id.linearlayout_transport_history){
            Log.i(TAG,"onClick layout linearlayout_transport_history");
            file_recored_back_icon = (ImageView)findViewById(R.id.file_recored_back_icon);
            Log.i(TAG,"file_recored_back_icon");
            setContentView(R.layout.myactivity);
            findviewfilemain();
        }

        if(v.getId() == R.id.file_recored_back_icon){
            file_recored_back_icon = (ImageView)findViewById(R.id.file_recored_back_icon);
            Log.i(TAG,"file_recored_back_icon");
            setContentView(R.layout.myactivity);
            findviewfilemain();
        }

        if(v.getId() == R.id.file_page_back_icon){
            file_page_back_icon = (ImageView)findViewById(R.id.file_page_back_icon);
            Log.i(TAG,"file_page_back_icon");
            setContentView(R.layout.layout_testactivity_title);
            findviewfilerecored();
            showScanQrCodesButtonIfNeeded();
        }

        if(v.getId() == R.id.file_allclose){
            Log.i(TAG,"file_allclose onClick");
            if(adapter != null ) {
                adapter.removeAllItem();
            }
        }
    }

    public void findviewfilerecored(){
        file_recored_back_icon = (ImageView)findViewById(R.id.file_recored_back_icon);
        file_recored_back_icon.setOnClickListener(this);

        getClientList();
        mMyConnectionView = findViewById(R.id.my_connections_view);
        if (mMyConnectionView != null) {
            mMyConnectionView.setText(getMyIpDeviceName());
        }
        mSwVersionView = findViewById(R.id.sw_version_view);
        if (mSwVersionView != null) {
            mSwVersionView.setText(getSoftwareInfo());
        }
        mDiasIDView = (TextView)findViewById(R.id.dias_view);
        paramValue = kv.decodeString("paramValue");
        mDiasIDView.setText(paramValue);

        LinearLayout layout = (LinearLayout)findViewById(R.id.linearlayout_qrcode);
        layout.setOnClickListener(this);

        LinearLayout ly_transport_history = (LinearLayout)findViewById(R.id.linearlayout_transport_history);
        ly_transport_history.setOnClickListener(this);

    }

    public void findviewfilemain(){
        file_recored_back_icon = (ImageView)findViewById(R.id.file_page_back_icon);
        file_recored_back_icon.setOnClickListener(this);

        recyclerView2 = findViewById(R.id.recycler_view);
        recyclerView2.setLayoutManager(new LinearLayoutManager(this, LinearLayoutManager.VERTICAL, false));
        recyclerView2.setItemAnimator(null);
        recyclerView2.setAdapter(adapter);

        //file_allclose = findViewById(R.id.file_allclose);
        ImageView file_allclose = findViewById(R.id.file_allclose);
        file_allclose.setOnClickListener(this);


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

    private void checkCameraPermission() {
        if (ContextCompat.checkSelfPermission(this, android.Manifest.permission.CAMERA)
                != PackageManager.PERMISSION_GRANTED) {
            ActivityCompat.requestPermissions(
                    this,
                    new String[]{android.Manifest.permission.CAMERA},
                    CAMERA_PERMISSION_REQUEST_CODE
            );
        } else {
            startBarcodeScan();
        }
    }

    private void startBarcodeScan(){
        Log.i(TAG,"startBarcodeScan start ScannerActivity");
        Intent intent = new Intent(this, ScannerActivity.class);
        startActivityForResult(intent, REQUEST_CODE_SCAN);
    }

    @Override
    protected void onActivityResult(int requestCode, int resultCode, @Nullable Intent data) {
        super.onActivityResult(requestCode, resultCode, data);

        if (requestCode == REQUEST_CODE_SCAN && resultCode == RESULT_OK) {
            if (data != null) {
                String result = data.getStringExtra(Intents.Scan.RESULT);
                Log.i(TAG, "onActivityResult qr result =" + result);
                //crossshare://scan?param=00E04C09A0C5
                Uri uri = Uri.parse(result);
                paramValue = uri.getQueryParameter("param");
                Log.i(TAG, "onActivityResult qr param=" + paramValue);

                ShareFragment shareFragment = (ShareFragment) getSupportFragmentManager().findFragmentByTag("share_fragment");
                if (shareFragment != null) {
                    switchFragment(shareFragment, "share_fragment");
                    setSelectedTab(0);
                }

                if (paramValue != null && !paramValue.isEmpty()) {
                    kv.encode("paramValue", paramValue);
                    new Handler().postDelayed(new Runnable() {
                        @Override
                        public void run() {
                            Intent broadcastIntent = new Intent("com.corssshare.qrparam");
                            broadcastIntent.putExtra("param", paramValue);
                            Log.i(TAG, "GoLog handleIntent sendBroadcast qrparam");
                            broadcastIntent.setPackage(getPackageName());
                            LocalBroadcastManager.getInstance(mContext).sendBroadcast(broadcastIntent);
                        }
                    }, 1000);
                }
            }
        }

    }

    private static void addUnique(List<String> list, String newStr) {
        if (!list.contains(newStr)) {
            list.add(newStr);
        }
    }

    // dialog for file cancel transfers
    public void cancel_transfers(final Context context, String filename,boolean isAllFile) {
        View view = View.inflate(context, R.layout.dialog_cancelfile, null);

        TextView titleView = (TextView) view.findViewById(R.id.title);
        TextView subtitleView = (TextView) view.findViewById(R.id.subtitle);
        Button conf = (Button) view.findViewById(R.id.img_comf);
        Button canl = (Button) view.findViewById(R.id.img_canl);

        if(isAllFile){
            titleView.setText("Cancel all transfers in progress");
            subtitleView.setText("All you sure want to cancel all transfers ?");
        }else{
            titleView.setText("Cancel this transfers in progress ");
            subtitleView.setText("All you sure want to cancel this transfers of "+filename + " ?");
        }
        dialog = new AlertDialog.Builder(context).create();
        dialog.setCancelable(false);

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            dialog.getWindow().setType(WindowManager.LayoutParams.TYPE_APPLICATION_OVERLAY);
        } else {
            dialog.getWindow().setType(WindowManager.LayoutParams.TYPE_SYSTEM_ALERT);
        }

        dialog.show();
        MyApplication.setDialogShown(true);
        dialog.setContentView(view);
        dialog.getWindow().setGravity(Gravity.CENTER);

        dialog.setOnDismissListener(new DialogInterface.OnDismissListener() {
            @Override
            public void onDismiss(DialogInterface dialog) {
                MyApplication.setDialogShown(false);
            }
        });

        conf.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {

                Libp2p_clipboard.cancelFileTrans(deviceip,deviceid,currenttimestamp);
                if(adapter != null ) {
                    adapter.cancelTransfers();
                }
                dialog.dismiss();
            }
        });
        canl.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {
                dialog.dismiss();
            }
        });


    }

    public void openPublicFileWithMediaStore(Context context , File file) {

        if (!file.exists()) {
            Toast.makeText(context, "file not exist", Toast.LENGTH_SHORT).show();
            return;
        }

        Log.i(TAG,"setReadable="+file.setReadable(true));
        Log.i(TAG, "setWritable=" + file.setWritable(true));
        Log.i(TAG,"getFileMimeType(file)="+getFileMimeType(file));
        Log.i(TAG, "file URI=" + getMediaUriFromPath(context, file.getAbsolutePath()));

        Intent intent2 = new Intent(Intent.ACTION_VIEW);
        intent2.setDataAndType(getMediaUriFromPath(context,file.getAbsolutePath()), getFileMimeType(file));
        intent2.addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION);
        try {
            startActivity(intent2);
        } catch (ActivityNotFoundException e) {
            Toast.makeText(context, "not app open", Toast.LENGTH_SHORT).show();
        } catch (SecurityException e) {
            Log.e(TAG, "security fail: " + e.getMessage());
        } catch (Exception e) {
            Log.e(TAG, "start fail: " + e.getClass().getSimpleName(), e);
        }

    }

    public void copyFileToDownloads(Context context, String privateFilePath, String displayName, String mimeType) {
        Log.i(TAG, "copyFileToDownloads privateFilePath=" + privateFilePath + " displayName="+displayName + " mimeType="+mimeType);
        File srcFile = new File(privateFilePath);
        ContentResolver resolver = context.getContentResolver();

        ContentValues values = new ContentValues();
        values.put(MediaStore.Downloads.DISPLAY_NAME, displayName);
        values.put(MediaStore.Downloads.MIME_TYPE, mimeType);
        values.put(MediaStore.Downloads.IS_PENDING, 1);

        Uri collection = null;
        if (android.os.Build.VERSION.SDK_INT >= android.os.Build.VERSION_CODES.Q) {
            collection = MediaStore.Downloads.getContentUri(MediaStore.VOLUME_EXTERNAL_PRIMARY);
        }

        Uri fileUri = resolver.insert(collection, values);
        if (fileUri == null) return;

        try (FileInputStream in = new FileInputStream(srcFile);
             OutputStream out = resolver.openOutputStream(fileUri)) {
            byte[] buffer = new byte[8192];
            int len;
            while ((len = in.read(buffer)) != -1) {
                out.write(buffer, 0, len);
            }
        } catch (IOException e) {
            e.printStackTrace();
            resolver.delete(fileUri, null, null);
            return;
        }


        values.clear();
        values.put(MediaStore.Downloads.IS_PENDING, 0);
        resolver.update(fileUri, values, null, null);

        boolean fileDel = srcFile.delete();
        Log.i(TAG, "copyFileToDownloads srcFile.delete()=" + fileDel);
        Toast.makeText(context, "file has been save to Download of internal storage", Toast.LENGTH_SHORT).show();
        fileUri.toString();
    }

    public void openDownloadsFolder() {

        File downloadsDir = Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DOWNLOADS);
        Uri downloadsUri = Uri.parse(downloadsDir.getAbsolutePath());

        Intent intent = new Intent(Intent.ACTION_VIEW);
        intent.setDataAndType(downloadsUri, "resource/folder");

        try {
            startActivity(intent);
        } catch (ActivityNotFoundException e) {
            Intent intentfile = new Intent(Intent.ACTION_OPEN_DOCUMENT_TREE);
            Uri uri = Uri.parse(Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DOWNLOADS).getAbsolutePath());
            intentfile.putExtra(DocumentsContract.EXTRA_INITIAL_URI, uri);
            startActivityForResult(intentfile, REQUEST_CODE);
        }
    }

    private static Uri getMediaUriFromPath(Context context, String filePath) {
        ContentResolver resolver = context.getContentResolver();
        String[] projection = new String[]{MediaStore.MediaColumns._ID};
        String selection = MediaStore.MediaColumns.DATA + " = ?";
        String[] selectionArgs = new String[]{filePath};

        List<Uri> contentUris = new ArrayList<>();

        contentUris.add(MediaStore.Images.Media.EXTERNAL_CONTENT_URI);
        contentUris.add(MediaStore.Video.Media.EXTERNAL_CONTENT_URI);
        contentUris.add(MediaStore.Audio.Media.EXTERNAL_CONTENT_URI);

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            contentUris.add(MediaStore.Downloads.EXTERNAL_CONTENT_URI);
        }

        contentUris.add(MediaStore.Files.getContentUri("external"));

        for (Uri uri : contentUris) {
            try (Cursor cursor = resolver.query(uri, projection, selection, selectionArgs, null)) {
                if (cursor != null && cursor.moveToFirst()) {
                    int idColumn = cursor.getColumnIndexOrThrow(MediaStore.MediaColumns._ID);
                    long id = cursor.getLong(idColumn);
                    return ContentUris.withAppendedId(uri, id);
                }
            } catch (Exception e) {
                Log.w(TAG, "Query failed for URI: " + uri, e);
            }
        }

        try {
            File file = new File(filePath);
            return FileProvider.getUriForFile(
                    context,
                    context.getPackageName(),
                    file
            );
        } catch (IllegalArgumentException e) {
            Log.e(TAG, "FileProvider error: " + e.getMessage());
            return null;
        }
    }

    private  String getFileMimeType(File file) {
        String name = file.getName();
        String extension = name.substring(name.lastIndexOf('.') + 1).toLowerCase();

        if ("jpg".equals(extension) || "jpeg".equals(extension)) {
            return "image/jpeg";
        } else if ("txt".equals(extension)) {
            return "text/plain";
        } else if ("png".equals(extension)) {
            return "image/png";
        } else if ("gif".equals(extension)) {
            return "image/gif";
        } else if ("mp4".equals(extension)) {
            return "video/mp4";
        } else if ("mp3".equals(extension)) {
            return "audio/mpeg";
        } else if ("doc".equals(extension)) {
            return "application/msword";
        } else if ("docx".equals(extension)) {
            return "application/vnd.openxmlformats-officedocument.wordprocessingml.document";
        } else if ("pdf".equals(extension)) {
            return "application/pdf";
        }  else if ("pptx".equals(extension)) {
            return "application/vnd.openxmlformats-officedocument.presentationml.presentation";
        }  else if ("pdf".equals(extension)) {
            return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet";
        }
        return "application/octet-stream";
    }






    @Override
    public void onOpenFileClick(boolean isallfile, String path) {
        //File srcFile = new File(openfilepath);
        //File saveDir = Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DOWNLOADS);
        LogUtils.e(TAG, "onOpenFileClick isallfile=" + isallfile + ", path"+path);
        File destFile = new File(path);
        if(isallfile){
            openDownloadsFolder();
        }else {
            LogUtils.e(TAG, "onOpenFileClick destFile=" + destFile);
            LogUtils.e(TAG, "onOpenFileClick destFile.getAbsolutePath()=" + destFile.getAbsolutePath());
            if (!isDownloadFileExists(MyApplication.getContext(), path)) {
                Toast.makeText(MyApplication.getContext(), "Download file not exist", Toast.LENGTH_SHORT).show();
                return;
            }
            openPublicFileWithMediaStore(MyApplication.getContext(), destFile);
        }
    }

    private boolean isDownloadFileExists(Context context, String filePath) {
        Uri uri = getMediaUriFromPath(context, filePath);
        return uri != null;
    }


    private void switchFragment(Fragment target, String tag) {
        FragmentManager frm = getSupportFragmentManager();

        frm.popBackStack(null, FragmentManager.POP_BACK_STACK_INCLUSIVE);
        frm.executePendingTransactions();

        if (target != null && target != currentFragment) {
            FragmentManager fm = getSupportFragmentManager();
            FragmentTransaction ft = fm.beginTransaction();
            if (!target.isAdded()) {
                ft.hide(currentFragment).add(R.id.fragment_container, target, tag).commitNow();
            } else {
                ft.hide(currentFragment).show(target).commitNow();
            }
            currentFragment = target;
            //Log.i(TAG, "After switch -> share: " + shareFragment.isAdded() + "," + shareFragment.isHidden());
            //Log.i(TAG, "After switch -> record: " + recordFragment.isAdded() + "," + recordFragment.isHidden());
            //Log.i(TAG, "After switch -> info: " + infoFragment.isAdded() + "," + infoFragment.isHidden());
        }
    }


    private void setSelectedTab(int index) {
        btnShare.setSelected(index == 0);
        btnRecord.setSelected(index == 1);
        btnInfo.setSelected(index == 2);
        textShare.setSelected(index == 0);
        textRecord.setSelected(index == 1);
        textInfo.setSelected(index == 2);
    }

    private File getUniqueDestFile(File destDir, String relPath) {
        File destFile = new File(destDir, relPath);
        if (!destFile.exists()) return destFile;
        String name = destFile.getName();
        String parent = destFile.getParent();
        String baseName, ext;
        int dotIdx = name.lastIndexOf(".");
        if (dotIdx != -1) {
            baseName = name.substring(0, dotIdx);
            ext = name.substring(dotIdx);
        } else {
            baseName = name;
            ext = "";
        }
        int idx = 1;
        File candidate;
        do {
            candidate = new File(parent, baseName + " (" + idx + ")" + ext);
            if (!candidate.exists()) {
                LogUtils.i(TAG, "file already exists,so rename: " + candidate.getAbsolutePath());
            }
            idx++;
        } while (candidate.exists());
        return candidate;
    }

    private void deleteEmptyDirs(File dir) {
        if (dir.isDirectory()) {
            File[] files = dir.listFiles();
            if (files != null) {
                for (File child : files) {
                    deleteEmptyDirs(child);
                }
            }

            if (dir.listFiles() == null || dir.listFiles().length == 0) {
                dir.delete();
            }
        }
    }


    public void appUpgrade(final Context context) {
        View view = View.inflate(context, R.layout.dialog_upgrade, null);
        Button button_upgrade = view.findViewById(R.id.button_upgrade);
        TextView subtitle = view.findViewById(R.id.subtitle);
        subtitle.setText(getResources().getString(R.string.version_warning, Libp2p_clipboard.getVersion()));
        dialog = new AlertDialog.Builder(context).create();
        dialog.setCancelable(false);

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            dialog.getWindow().setType(WindowManager.LayoutParams.TYPE_APPLICATION_OVERLAY);
        } else {
            dialog.getWindow().setType(WindowManager.LayoutParams.TYPE_SYSTEM_ALERT);
        }
        LogUtils.i(TAG, "appUpgrade dialog="+dialog + ", dialog.isShowing()="+dialog.isShowing());
        if (dialog != null && dialog.isShowing()) {
            return;
        }
        dialog.show();
        dialog.setContentView(view);
        dialog.getWindow().setGravity(Gravity.CENTER);
        button_upgrade.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {
                String packageName = getPackageName();
                Intent intent;
                if (isGooglePlayInstalled(context)) {
                    LogUtils.i(TAG, "appUpgrade startActivity to googleplay");
                    intent = new Intent(Intent.ACTION_VIEW, Uri.parse("market://details?id=" + packageName));
                } else {
                    LogUtils.i(TAG, "appUpgrade not appstore");
                    intent = new Intent(Intent.ACTION_VIEW, Uri.parse("https://play.google.com/store/apps/details?id=" + packageName));
                }
                intent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK);
                startActivity(intent);
                dialog.dismiss();
            }
        });

    }

    public static boolean isGooglePlayInstalled(Context context) {
        try {
            context.getPackageManager().getPackageInfo("com.android.vending", 0);
            return true;
        } catch (PackageManager.NameNotFoundException e) {
            return false;
        }
    }
}
