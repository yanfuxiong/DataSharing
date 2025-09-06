package com.realtek.crossshare;

import android.annotation.SuppressLint;
import android.app.AlertDialog;
import android.content.Context;
import android.content.Intent;
import android.graphics.drawable.AnimationDrawable;
import android.net.Uri;
import android.os.Build;
import android.os.Bundle;
import android.text.TextUtils;
import android.util.Log;
import android.view.Gravity;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.view.WindowManager;
import android.widget.Button;
import android.widget.ImageView;
import android.widget.LinearLayout;
import android.widget.TextView;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;
import androidx.fragment.app.Fragment;
import androidx.recyclerview.widget.LinearLayoutManager;
import androidx.recyclerview.widget.RecyclerView;

import com.google.gson.Gson;
import com.google.gson.reflect.TypeToken;
import com.tencent.mmkv.MMKV;

import java.lang.reflect.Type;
import java.util.ArrayList;
import java.util.List;
import android.os.Handler;

import libp2p_clipboard.Libp2p_clipboard;

public class ShareFragment extends Fragment {

    private static final String TAG = "ShareFragment";
    private static final String SOURCE_HDMI1 = "HDMI1";
    private static final String SOURCE_HDMI2 = "HDMI2";
    private static final String SOURCE_USBC1 = "USBC1";
    private static final String SOURCE_USBC2 = "USBC2";
    private static final String SOURCE_DP1 = "DP1";
    private static final String SOURCE_DP2 = "DP2";
    private static final String SOURCE_COMPUTER = "Computer";
    private static final String SOURCE_MIRACAST = "Miracast";
    private static final String KEY_CLIENT_LIST = "client_list_data";
    private String clientlist;
    private List<Device> deviceList = new ArrayList<>();
    private RecyclerView recyclerViewdevice;
    private Context context;
    private LinearLayout share_layout, share_layout_info, server_layout;
    private TextView diasserver;
    private View viewline;
    private String paramValue;
    private MMKV mmkv;
    private RecyclerView serverListView;
    private ServerAdapter serverAdapter;
    private List<Server> serverList = new ArrayList<>();
    private AlertDialog dialog,connectingDialog,warningDialog;
    private Handler handler = new Handler();
    private boolean isGoCallBack = false;
    private Runnable timeoutRunnable;
    private  String mmonitorName ="";
    private  String minstance ="";
    private  String mipAddr ="";
    private  int diasstatus =-1;
    private ImageView lanserversearsh;
    private FmShareAdapter fmShareAdapter;
    private boolean isbrowseLanServer = false;
    private boolean isMiarcastIn = false;
    private String pendingInstance = null;
    private boolean isFragmentActive = false;

    public ShareFragment() {}

    @Override
    public void onCreate(@Nullable Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        mmkv = MMKV.defaultMMKV();
    }

    @SuppressLint("WrongViewCast")
    @Override
    public View onCreateView(LayoutInflater inflater, ViewGroup container,
                             Bundle savedInstanceState) {
        return inflater.inflate(R.layout.layout_fm_share, container, false);
    }

    @Override
    public void onViewCreated(@NonNull View view, @Nullable Bundle savedInstanceState) {
        super.onViewCreated(view, savedInstanceState);
        context = getContext();
        share_layout = view.findViewById(R.id.share_layout);
        share_layout_info = view.findViewById(R.id.share_layout_info);
        recyclerViewdevice = view.findViewById(R.id.share_listview);
        diasserver = view.findViewById(R.id.diasserver);
        viewline = view.findViewById(R.id.viewline);

        //paramValue = mmkv.decodeString("paramValue");
        clientlist = mmkv.decodeString(KEY_CLIENT_LIST);
        String getP2pList = Libp2p_clipboard.getClientList();
        LogUtils.i(TAG, "onViewCreated getP2pList=" + getP2pList );
        LogUtils.i(TAG, "onViewCreated mkkv clientlist=" + clientlist);
        recyclerViewdevice.setLayoutManager(new LinearLayoutManager(context, LinearLayoutManager.VERTICAL, false));
        recyclerViewdevice.addItemDecoration(new SpaceItemDecoration(3));
        fmShareAdapter = new FmShareAdapter(context, deviceList);
        recyclerViewdevice.setAdapter(fmShareAdapter);

        lanserversearsh = getActivity().findViewById(R.id.searsh);
        lanserversearsh.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                startSearchDialog(MyApplication.getContext());
            }
        });
        ImageView searchAnimView = view.findViewById(R.id.search_anim_view);
        AnimationDrawable animationDrawable = (AnimationDrawable) searchAnimView.getDrawable();
        animationDrawable.start();
        server_layout = view.findViewById(R.id.server_layout);
        serverListView = view.findViewById(R.id.server_listview);
        serverListView.setLayoutManager(new LinearLayoutManager(context, LinearLayoutManager.VERTICAL, false));
        serverAdapter = new ServerAdapter(context, serverList, new OnClickListener() {
            @Override
            public void onServerClick(String monitorName, String instance, String ipAddr) {
                LogUtils.d(TAG, "serverAdapter onServerClick " + monitorName + "， "+ instance+"， " +ipAddr);
                showConnectingDialog(getContext(),"Connecting to the montitor...");
                mmonitorName=monitorName;
                minstance=instance;
                mipAddr=ipAddr;
                new Thread(() -> {
                    LogUtils.i(TAG, "onServerClick: now connectLanServer with instance="+instance);
                    Libp2p_clipboard.workerConnectLanServer(instance);
                }).start();

            }
        });

        serverListView.setAdapter(serverAdapter);

        if (!getP2pList.isEmpty()) {
            updateClientlist(clientlist);
            share_layout.setVisibility(View.GONE);
            share_layout_info.setVisibility(View.VISIBLE);
            recyclerViewdevice.setVisibility(View.VISIBLE);
            viewline.setVisibility(View.VISIBLE);
        }else{
            diasstatus = mmkv.decodeInt("diasstatus", -1);
            LogUtils.d(TAG, "onViewCreated diasstatus " + diasstatus);
            if(diasstatus ==1){
                share_layout.setVisibility(View.VISIBLE);
                share_layout_info.setVisibility(View.GONE);
                recyclerViewdevice.setVisibility(View.GONE);
                viewline.setVisibility(View.GONE);
            }else if(diasstatus ==2 || diasstatus ==4){
                share_layout.setVisibility(View.GONE);
                share_layout_info.setVisibility(View.GONE);
                server_layout.setVisibility(View.VISIBLE);
                serverListView.setVisibility(View.VISIBLE);
                lanserversearsh.setVisibility(View.GONE);
                viewline.setVisibility(View.GONE);
            }else if (diasstatus == 6 || diasstatus == 7) {
                share_layout.setVisibility(View.GONE);
                share_layout_info.setVisibility(View.VISIBLE);
                recyclerViewdevice.setVisibility(View.VISIBLE);
                viewline.setVisibility(View.VISIBLE);
                server_layout.setVisibility(View.GONE);
                lanserversearsh.setVisibility(View.VISIBLE);
            }else{
                share_layout.setVisibility(View.VISIBLE);
                share_layout_info.setVisibility(View.GONE);
                recyclerViewdevice.setVisibility(View.GONE);
                viewline.setVisibility(View.GONE);
            }
        }
    }

    /**
     * update adapter
     */
    public void updateClientlist(String listdata) {
        clientlist = listdata;

        if (listdata != null && !listdata.isEmpty()) {
            mmkv.encode(KEY_CLIENT_LIST, clientlist);
        } else {
            mmkv.removeValueForKey(KEY_CLIENT_LIST);
        }

        LogUtils.i(TAG, "updateClientlist clientlist=" + clientlist);
        recyclerViewdevice.setVisibility(View.VISIBLE);
        if (deviceList == null) return;
        deviceList.clear();
        if (listdata == null || listdata.isEmpty()) {
            fmShareAdapter.notifyDataSetChanged();
            return;
        }
        String[] strArray = listdata.split(",");
        for (String getlistvalue : strArray) {
            String[] info = getlistvalue.split("#");
            String ip = info[0];
            String id = info.length > 1 ? info[1] : info[0];
            String name = info.length > 2 ? info[2] : info[0];
            String sourcePortType = info.length > 3 ? info[3] : info[0];
            LogUtils.i(TAG,"updateClientlist name="+name + " sourcePortType="+sourcePortType);
            if (sourcePortType.contains(SOURCE_HDMI1)) {
                deviceList.add(new Device(name, ip, R.drawable.hdmi));
            } else if (sourcePortType.contains(SOURCE_HDMI2)) {
                deviceList.add(new Device(name, ip, R.drawable.hdmi2));
            } else if (sourcePortType.contains(SOURCE_MIRACAST)) {
                deviceList.add(new Device(name, ip, R.drawable.miracast));
            } else if (sourcePortType.contains(SOURCE_USBC1)) {
                deviceList.add(new Device(name, ip, R.drawable.usb_c1));
            } else if (sourcePortType.contains(SOURCE_USBC2)) {
                deviceList.add(new Device(name, ip, R.drawable.usb_c2));
            }else if (sourcePortType.contains(SOURCE_DP1)) {
                deviceList.add(new Device(name, ip, R.drawable.dp1));
            }else if (sourcePortType.contains(SOURCE_DP2)) {
                deviceList.add(new Device(name, ip, R.drawable.dp2));
            }else if (sourcePortType.contains(SOURCE_COMPUTER)) {
                deviceList.add(new Device(name, ip, R.drawable.computer));
            }else {
                deviceList.add(new Device(name, ip, R.drawable.src_default));
            }
        }
        fmShareAdapter.notifyDataSetChanged();
    }

    @Override
    public void onPause() {
        super.onPause();
        if (clientlist != null && !clientlist.isEmpty()) {
            mmkv.encode(KEY_CLIENT_LIST, clientlist);
        }
        isFragmentActive = false;
        if (connectingDialog != null && connectingDialog.isShowing()) connectingDialog.dismiss();
        if (warningDialog != null && warningDialog.isShowing()) warningDialog.dismiss();
        if (dialog != null && dialog.isShowing()) dialog.dismiss();
    }

    @Override
    public void onHiddenChanged(boolean hidden) {
        super.onHiddenChanged(hidden);
        LogUtils.d(TAG, "onHiddenChanged hidden=" + hidden + ", diasstatus=" + diasstatus);
        if (!hidden) {
            if (lanserversearsh != null) {
                diasstatus = mmkv.decodeInt("diasstatus", -1);
                if (diasstatus == 6 || diasstatus == 7) {
                    LogUtils.d(TAG, "onHiddenChanged: making search icon VISIBLE for status=" + diasstatus);
                    lanserversearsh.setVisibility(View.VISIBLE);
                } else {
                    LogUtils.d(TAG, "onHiddenChanged: making search icon GONE for status=" + diasstatus);
                    lanserversearsh.setVisibility(View.GONE);
                }
            }
        } else {
            if (lanserversearsh != null) {
                LogUtils.d(TAG, "onHiddenChanged: hiding search icon as fragment is hidden");
                lanserversearsh.setVisibility(View.GONE);
            }
        }
    }

    @Override
    public void onResume() {
        super.onResume();
        TextView toolbarTitle = getActivity().findViewById(R.id.toolbar_title);
        if (toolbarTitle != null) {
            toolbarTitle.setText("Cross Share");
        }
        isFragmentActive = true;
        String monitorName = MMKV.defaultMMKV().decodeString("LANSERVER_MONITOR_NAME", "");
        String instance = MMKV.defaultMMKV().decodeString("LANSERVER_MONITOR_INSTANCE", "");
        String ipAddr = MMKV.defaultMMKV().decodeString("LANSERVER_MONITOR_IPADDE", "");
        diasstatus = mmkv.decodeInt("diasstatus", -1);
        LogUtils.i(TAG,"onResume monitorName="+monitorName +   ", instance="+instance+ ", ipAddr="+ipAddr + ", diasstatus="+diasstatus+ ", isMiarcastIn="+isMiarcastIn);
        if(!instance.isEmpty() && (diasstatus != 6 &&  diasstatus !=7) && isMiarcastIn){
            isGoCallBack = false;
            if (timeoutRunnable != null) {
                handler.removeCallbacks(timeoutRunnable);
                timeoutRunnable = null;
            }
            timeoutRunnable = new Runnable() {
                @Override
                public void run() {
                    if (!isGoCallBack && timeoutRunnable != null) {
                        // 5s，remove favorite monitor
                        LogUtils.i(TAG,"removeValueForKey 5s remove favorite monitor");
                        MMKV.defaultMMKV().removeValueForKey("LANSERVER_MONITOR_NAME");
                        MMKV.defaultMMKV().removeValueForKey("LANSERVER_MONITOR_INSTANCE");
                        MMKV.defaultMMKV().removeValueForKey("LANSERVER_MONITOR_IPADDE");
                        timeoutRunnable = null;
                    }
                }
            };
            handler.postDelayed(timeoutRunnable, 5000);
        }
    }

    public void setMonitorName(String monitorName) {
        LogUtils.i(TAG,"monitorName="+monitorName);
        if(!monitorName.isEmpty()) {
            diasserver.setText(monitorName);
        }
    }

    public void setDiasStatus(long status) {
        LogUtils.i(TAG,"getstatus="+status + ", mmonitorName="+mmonitorName+ ", minstance="+minstance);
        diasstatus=(int)status;
        mmkv.encode("diasstatus", diasstatus);
        if (status == 6 || status == 7) {
            share_layout.setVisibility(View.GONE);
            share_layout_info.setVisibility(View.VISIBLE);
            recyclerViewdevice.setVisibility(View.VISIBLE);
            viewline.setVisibility(View.VISIBLE);
            server_layout.setVisibility(View.GONE);
            lanserversearsh.setVisibility(View.VISIBLE);
            if(!TextUtils.isEmpty(minstance)) {
                MMKV.defaultMMKV().encode("LANSERVER_MONITOR_NAME", mmonitorName);
                MMKV.defaultMMKV().encode("LANSERVER_MONITOR_INSTANCE", minstance);
                MMKV.defaultMMKV().encode("LANSERVER_MONITOR_IPADDE", mipAddr);
            }
            if (connectingDialog != null) connectingDialog.dismiss();
        } else if (status >= 2) {
            share_layout.setVisibility(View.GONE);
            share_layout_info.setVisibility(View.GONE);
            server_layout.setVisibility(View.VISIBLE);
            serverListView.setVisibility(View.VISIBLE);
            lanserversearsh.setVisibility(View.GONE);
            viewline.setVisibility(View.GONE);
            showListserver(MyApplication.getServerList());
        } else {
            share_layout.setVisibility(View.VISIBLE);
            share_layout_info.setVisibility(View.GONE);
            server_layout.setVisibility(View.GONE);
            serverListView.setVisibility(View.GONE);
            lanserversearsh.setVisibility(View.GONE);
            viewline.setVisibility(View.GONE);
        }
        if (status == 5) {
            if (warningDialog != null) warningDialog.dismiss();
            if (connectingDialog != null && connectingDialog.isShowing()) connectingDialog.dismiss();
            showCommonWarningDialog(getContext(), "Verification failed",status);
        } else if (status == 8) {
            if (warningDialog != null) warningDialog.dismiss();
            if (connectingDialog != null && connectingDialog.isShowing()) connectingDialog.dismiss();
            showCommonWarningDialog(getContext(), "Connection failed",status);
        } else if (status == 4) {
            if (connectingDialog != null) connectingDialog.dismiss();
            showConnectingDialog(getContext(), "Verifying...");
        }
        fmShareAdapter.notifyDataSetChanged();
    }

    public void addServer(String monitorName, String instance, String ipAddr, String version, long timestamp) {
        LogUtils.i(TAG,"addServer monitorName="+monitorName +   ", instance="+instance+ ", ipAddr="+ipAddr);

        String savedMonitorName = MMKV.defaultMMKV().decodeString("LANSERVER_MONITOR_NAME", "");
        String savedInstance = MMKV.defaultMMKV().decodeString("LANSERVER_MONITOR_INSTANCE", "");
        String savedIpAddr = MMKV.defaultMMKV().decodeString("LANSERVER_MONITOR_IPADDE", "");
        LogUtils.i(TAG,"addServer savedMonitorName="+savedMonitorName +   ", savedInstance="+savedInstance+ ", savedIpAddr="+savedIpAddr+ ", isbrowseLanServer="+isbrowseLanServer+ ", isMiarcastIn="+isMiarcastIn);

        if (!savedInstance.isEmpty() && !isbrowseLanServer) {
            if (savedInstance.equals(instance)) {
                isGoCallBack = true;
                if (timeoutRunnable != null) {
                    handler.removeCallbacks(timeoutRunnable);
                    timeoutRunnable = null;
                }
                if(isMiarcastIn) {
                    new Thread(new Runnable() {
                        @Override
                        public void run() {
                            try {
                                Thread.sleep(500);
                            } catch (InterruptedException e) {
                                e.printStackTrace();
                            }
                            LogUtils.i(TAG, "addServer: now connectLanServer with instance="+instance);
                            Libp2p_clipboard.workerConnectLanServer(instance);
                        }
                    }).start();
                }else{
                    pendingInstance = instance;
                    LogUtils.i(TAG, "addServer pending connectLanServer, wait for Miracast");
                }
            }
        }

        if(diasstatus >=2){
            showListserver(MyApplication.getServerList());
        }
    }



    public void startSearchDialog(Context context) {
        View view = View.inflate(context, R.layout.dialog_lanserver, null);
        Button button_cancel = view.findViewById(R.id.btnCancel);
        Button button_confirm = view.findViewById(R.id.btnConfirm);
        dialog = new AlertDialog.Builder(context).create();
        dialog.setCancelable(false);

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            dialog.getWindow().setType(WindowManager.LayoutParams.TYPE_APPLICATION_OVERLAY);
        } else {
            dialog.getWindow().setType(WindowManager.LayoutParams.TYPE_SYSTEM_ALERT);
        }

        dialog.show();
        dialog.setContentView(view);
        dialog.getWindow().setGravity(Gravity.CENTER);
        button_cancel.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {
                dialog.dismiss();
            }
        });

        button_confirm.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {
                isbrowseLanServer = true;
                LogUtils.i(TAG,"startSearchDialog, start browseLanServer, remove favorite moniter");
                MMKV.defaultMMKV().removeValueForKey("LANSERVER_MONITOR_NAME");
                MMKV.defaultMMKV().removeValueForKey("LANSERVER_MONITOR_INSTANCE");
                MMKV.defaultMMKV().removeValueForKey("LANSERVER_MONITOR_IPADDE");
                MyApplication.clearServerList();
                Libp2p_clipboard.browseLanServer();
                dialog.dismiss();
            }
        });

    }


    @Override
    public void onDestroyView() {
        LogUtils.i(TAG,"onDestroyView");
        super.onDestroyView();
        isFragmentActive = false;
        if(timeoutRunnable != null) handler.removeCallbacks(timeoutRunnable);
        if (connectingDialog != null && connectingDialog.isShowing()) connectingDialog.dismiss();
        if (warningDialog != null && warningDialog.isShowing()) warningDialog.dismiss();
        if (dialog != null && dialog.isShowing()) dialog.dismiss();
    }


    public void showCommonWarningDialog(Context context, String message,long status) {
        if (!isFragmentActive || getActivity() == null || getActivity().isFinishing()) {
            LogUtils.i(TAG, "showCommonWarningDialog: fragment not active, skip show");
            return;
        }
        View view = View.inflate(context, R.layout.dialog_common_warning, null);
        TextView dialogTitle = view.findViewById(R.id.dialogTitle);
        TextView dialogMessage = view.findViewById(R.id.dialogMessage);
        Button btnContinue = view.findViewById(R.id.btnContinue);
        if(status == 5){
            dialogMessage.setText("Please ensure the phone screenis displayed on the monitor");
        }else if(status == 8){
            dialogMessage.setText("Please ensure the monitor is on the local network");
        }

        dialogTitle.setText(message);
        warningDialog = new AlertDialog.Builder(context).create();
        warningDialog.setCancelable(false);

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            warningDialog.getWindow().setType(WindowManager.LayoutParams.TYPE_APPLICATION_OVERLAY);
        } else {
            warningDialog.getWindow().setType(WindowManager.LayoutParams.TYPE_SYSTEM_ALERT);
        }

        warningDialog.show();
        warningDialog.setContentView(view);
        warningDialog.getWindow().setGravity(Gravity.CENTER);
        btnContinue.setOnClickListener(v -> warningDialog.dismiss());
    }

    public void showConnectingDialog(Context context, String message) {
        if (!isFragmentActive || getActivity() == null || getActivity().isFinishing()) {
            LogUtils.i(TAG, "showConnectingDialog: fragment not active, skip show");
            return;
        }
        View view = View.inflate(context, R.layout.dialog_connecting, null);
        TextView dialogMessage = view.findViewById(R.id.dialogMessage);
        dialogMessage.setText(message);

        connectingDialog = new AlertDialog.Builder(context).create();
        connectingDialog.setCancelable(false);

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            connectingDialog.getWindow().setType(WindowManager.LayoutParams.TYPE_APPLICATION_OVERLAY);
        } else {
            connectingDialog.getWindow().setType(WindowManager.LayoutParams.TYPE_SYSTEM_ALERT);
        }

        connectingDialog.show();
        connectingDialog.setContentView(view);
        connectingDialog.getWindow().setGravity(Gravity.CENTER);

    }


    public void showListserver(List<Server> servers){
        serverList.clear();
        if(servers != null){
            serverList.addAll(servers);
        }
        LogUtils.i(TAG,"showListserver serverList.size="+serverList.size());
        for (Server s : serverList) {
            LogUtils.i(TAG,"showListserver s.getmonitorName="+s.getmonitorName());
        }
        serverAdapter.notifyDataSetChanged();
    }

    public void setMiarcastState(boolean isIn) {
        isMiarcastIn=isIn;
        LogUtils.i(TAG, "setMiarcastState isIn=" + isIn);
        if(!isMiarcastIn){
            isbrowseLanServer = false;
            handler.removeCallbacks(timeoutRunnable);
            timeoutRunnable = null;
            MyApplication.clearServerList();
        }
        if (isMiarcastIn && pendingInstance != null) {
            new Thread(new Runnable() {
                @Override
                public void run() {
                    try {
                        Thread.sleep(500);
                    } catch (InterruptedException e) {
                        e.printStackTrace();
                    }
                    LogUtils.i(TAG, "setMiarcastState: now connectLanServer with cached instance="+pendingInstance);
                    Libp2p_clipboard.workerConnectLanServer(pendingInstance);
                    pendingInstance = null;
                }
            }).start();
        }
    }

}
