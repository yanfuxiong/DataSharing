package com.realtek.crossshare;

import android.os.Bundle;
import android.util.DisplayMetrics;
import android.util.Log;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.Button;
import android.widget.LinearLayout;
import android.widget.TextView;
import android.widget.Toast;

import androidx.annotation.Nullable;
import androidx.appcompat.app.AppCompatActivity;
import androidx.recyclerview.widget.LinearLayoutManager;
import androidx.recyclerview.widget.RecyclerView;
import com.google.android.material.bottomsheet.BottomSheetDialog;
import org.json.JSONArray;
import org.json.JSONObject;

import java.io.File;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;

import libp2p_clipboard.Libp2p_clipboard;

public class ShareActivity extends AppCompatActivity {

    private static final String TAG = "ShareActivity";
    private ArrayList<String> fileNamesList;
    private ArrayList<String> filePaths;
    private List<Device> deviceList;
    private Map<String, String> deviceNameIdMap;
    private DeviceAdapter deviceAdapter;
    private String value, valueipid, valueip, valueid;
    private String jsonString;
    private BottomSheetDialog dialog;

    @Override
    protected void onCreate(@Nullable Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        Log.i(TAG, "onCreate");

        fileNamesList = getIntent().getStringArrayListExtra("fileNamesList");
        filePaths = getIntent().getStringArrayListExtra("filePaths");
        deviceList = MyApplication.getDeviceList();
        deviceNameIdMap = MyApplication.getDeviceNameIdMap();
        Log.i(TAG,"deviceList"+deviceList.size());
        LogUtils.i(TAG,"deviceNameIdMap"+deviceNameIdMap);

        showBottomSheetDialog();
    }

    public void showBottomSheetDialog(){
        dialog = new BottomSheetDialog(this);
        View sheetView = LayoutInflater.from(this).inflate(R.layout.layout_main_share, null);

        DisplayMetrics metrics = new DisplayMetrics();
        getWindowManager().getDefaultDisplay().getMetrics(metrics);

        int desiredHeight;
        if (fileNamesList.size() <= 15) {
            desiredHeight = ViewGroup.LayoutParams.WRAP_CONTENT;
        } else {
            desiredHeight = (int) (getResources().getDisplayMetrics().heightPixels * 0.7);
        }
        LinearLayout root = sheetView.findViewById(R.id.bottom_sheet_root);
        root.setLayoutParams(new ViewGroup.LayoutParams(
                ViewGroup.LayoutParams.MATCH_PARENT,
                desiredHeight
        ));

        RecyclerView fileRecycler = sheetView.findViewById(R.id.file_list);
        fileRecycler.setLayoutManager(new LinearLayoutManager(this, LinearLayoutManager.VERTICAL, false));
        FileListAdapter fileListAdapter = new FileListAdapter(fileNamesList);
        fileRecycler.setAdapter(fileListAdapter);

        RecyclerView deviceRecycler = sheetView.findViewById(R.id.recycler_devicelist);
        deviceRecycler.setLayoutManager(new LinearLayoutManager(this, LinearLayoutManager.HORIZONTAL, false));

        if (deviceAdapter == null) {
            deviceAdapter = new DeviceAdapter(this, deviceList);
        }
        deviceRecycler.setAdapter(deviceAdapter);
        dialog.setContentView(sheetView);
        dialog.setCanceledOnTouchOutside(true);
        dialog.setCancelable(true);
        dialog.show();

        dialog.setOnCancelListener(dialogInterface -> {
            for (String path : filePaths) {
                File f = new File(path);
                if(f.exists()){
                    boolean deleted = f.delete();
                    LogUtils.i(TAG, "Dialog canceled, delete original file " + path + ", deleted=" + deleted);
                }
            }
            finish();
        });

        TextView fileCount = sheetView.findViewById(R.id.filecount);
        fileRecycler.post(new Runnable() {
            @Override
            public void run() {
                if (fileRecycler.canScrollVertically(1)) {
                    fileCount.setVisibility(View.VISIBLE);
                    fileCount.setText("Total" + " "+fileNamesList.size() +" files" );
                } else {
                    fileCount.setVisibility(View.GONE);
                }
            }
        });

        deviceAdapter.setOnItemClickListener(new MyAdapter.OnItemClickListener() {
            @Override
            public void onItemClick(View view, int position) {
                String name = ((TextView) view).getText().toString();
                Log.d(TAG, "select device valueid name="+name);
                valueipid = deviceNameIdMap.get(name);
                String[] parts = valueipid.split(":");
                valueid = parts[0];
                valueip = parts[1];
                Log.d(TAG, "select device valueid=" + valueid + " valueip="+valueip);
                Log.d(TAG, "select device name:" + name + ", id:" + valueipid);
                value = valueid;
                sendjson(valueid,valueip);
            }
        });

        Button transfer_button = sheetView.findViewById(R.id.transfer_button);
        transfer_button.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View view) {
                LogUtils.d(TAG, "sendMultiFilesDropRequest valueip=" + valueip + " valueid="+ valueid + " jsonString="+jsonString);
                if (valueid == null | valueip == null) {
                    Toast.makeText(ShareActivity.this, "Please select a connection", Toast.LENGTH_SHORT).show();
                } else {
                    Libp2p_clipboard.sendMultiFilesDropRequest(jsonString);
                    finish();
                }
            }
        });
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
    protected void onResume() {
        super.onResume();
        Log.i(TAG, "onResume()");
    }

    @Override
    protected void onPause() {
        super.onPause();
        Log.i(TAG, "onPause()");
    }

    @Override
    protected void onStop() {
        super.onStop();
        Log.i(TAG, "onStop()");
    }

    @Override
    protected void onDestroy() {
        Log.i(TAG, "onDestroy");
        if (dialog != null && dialog.isShowing()) {
            dialog.dismiss();
        }
        super.onDestroy();
    }
}

