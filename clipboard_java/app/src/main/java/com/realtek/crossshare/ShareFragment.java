package com.realtek.crossshare;

import android.annotation.SuppressLint;
import android.content.Context;
import android.os.Bundle;
import android.util.Log;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.LinearLayout;
import android.widget.TextView;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;
import androidx.fragment.app.Fragment;
import androidx.recyclerview.widget.LinearLayoutManager;
import androidx.recyclerview.widget.RecyclerView;

import com.tencent.mmkv.MMKV;

import java.util.ArrayList;
import java.util.List;

import libp2p_clipboard.Libp2p_clipboard;

public class ShareFragment extends Fragment {

    private static final String TAG = "ShareFragment";
    private static final String SOURCE_HDMI1 = "HDMI1";
    private static final String SOURCE_HDMI2 = "HDMI2";
    private static final String SOURCE_USBC = "USBC";
    private static final String SOURCE_MIRACAST = "Miracast";
    private static final String KEY_CLIENT_LIST = "client_list_data";
    private String clientlist;
    private List<Device> deviceList = new ArrayList<>();
    private RecyclerView recyclerViewdevice;
    private Context context;
    private LinearLayout share_layout, share_layout_info;
    private TextView diasserver;
    private View viewline;
    private String paramValue;
    private MMKV mmkv;

    private FmShareAdapter fmShareAdapter;

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
        Log.i(TAG, "onViewCreated getP2pList=" + getP2pList );
        Log.i(TAG, "onViewCreated mkkv clientlist=" + clientlist);
        recyclerViewdevice.setLayoutManager(new LinearLayoutManager(context, LinearLayoutManager.VERTICAL, false));
        recyclerViewdevice.addItemDecoration(new SpaceItemDecoration(3));
        fmShareAdapter = new FmShareAdapter(context, deviceList);
        recyclerViewdevice.setAdapter(fmShareAdapter);

        if (!getP2pList.isEmpty()) {
            updateClientlist(clientlist);
            share_layout.setVisibility(View.GONE);
            share_layout_info.setVisibility(View.VISIBLE);
            recyclerViewdevice.setVisibility(View.VISIBLE);
            viewline.setVisibility(View.VISIBLE);
        }else{
            share_layout.setVisibility(View.VISIBLE);
            share_layout_info.setVisibility(View.GONE);
            recyclerViewdevice.setVisibility(View.GONE);
            viewline.setVisibility(View.GONE);
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

        Log.i(TAG, "updateClientlist clientlist=" + clientlist);
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
            Log.i(TAG,"updateClientlist name="+name + " sourcePortType="+sourcePortType);
            if (sourcePortType.contains(SOURCE_HDMI1)) {
                deviceList.add(new Device(name, ip, R.drawable.hdmi));
            } else if (sourcePortType.contains(SOURCE_HDMI2)) {
                deviceList.add(new Device(name, ip, R.drawable.hdmi2));
            } else if (sourcePortType.contains(SOURCE_MIRACAST)) {
                deviceList.add(new Device(name, ip, R.drawable.miracast));
            } else if (sourcePortType.contains(SOURCE_USBC)) {
                deviceList.add(new Device(name, ip, R.drawable.usb_p));
            } else {
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
    }

    @Override
    public void onHiddenChanged(boolean hidden) {
        super.onHiddenChanged(hidden);
    }

    @Override
    public void onResume() {
        super.onResume();
        TextView toolbarTitle = getActivity().findViewById(R.id.toolbar_title);
        if (toolbarTitle != null) {
            toolbarTitle.setText("Cross Share");
        }
    }

    public void setMonitorName(String monitorName) {
        Log.i(TAG,"monitorName="+monitorName);
        if(!monitorName.isEmpty()) {
            diasserver.setText(monitorName);
        }
    }

    public void setDiasStatus(long status) {
        Log.i(TAG,"status="+status);
        if(status == 6 || status == 7) {
            share_layout.setVisibility(View.GONE);
            share_layout_info.setVisibility(View.VISIBLE);
            recyclerViewdevice.setVisibility(View.VISIBLE);
            viewline.setVisibility(View.VISIBLE);
        }else{
            share_layout.setVisibility(View.VISIBLE);
            share_layout_info.setVisibility(View.GONE);
            recyclerViewdevice.setVisibility(View.GONE);
            viewline.setVisibility(View.GONE);

        }
        fmShareAdapter.notifyDataSetChanged();
    }
}
