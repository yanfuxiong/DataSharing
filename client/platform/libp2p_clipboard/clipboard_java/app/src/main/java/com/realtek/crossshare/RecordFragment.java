package com.realtek.crossshare;

import android.app.AlertDialog;
import android.content.Context;
import android.content.DialogInterface;
import android.os.Build;
import android.os.Bundle;
import android.util.Log;
import android.view.Gravity;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.view.WindowManager;
import android.widget.Button;
import android.widget.TextView;

import androidx.annotation.NonNull;
import androidx.fragment.app.Fragment;
import androidx.recyclerview.widget.LinearLayoutManager;
import androidx.recyclerview.widget.RecyclerView;

import com.google.android.material.tabs.TabLayout;

import java.util.ArrayList;
import java.util.Iterator;
import java.util.List;

import libp2p_clipboard.Libp2p_clipboard;

public class RecordFragment extends Fragment {
    private static final String TAG = "RecordFragment";
    private TabLayout tabLayout;
    private FileTransferAdapter adapter;
    private RecyclerView recyclerView;
    private int currentTabPosition = 0;
    private AlertDialog dialog;
    private static final int TAB_IN_PROGRESS = 0;
    private static final int TAB_RECEIVED = 1;
    private static final int TAB_FAILED = 2;

    public interface OnFileActionListener {
        void onOpenFileClick(boolean isallfile, String path);
    }

    private static String deviceip;
    private static String deviceid;
    private static long currenttimestamp;

    private OnFileActionListener actionListener;

    @Override
    public void onAttach(@NonNull Context context) {
        super.onAttach(context);
        if (context instanceof OnFileActionListener) {
            actionListener = (OnFileActionListener) context;
        }
    }

    @Override
    public View onCreateView(LayoutInflater inflater, ViewGroup container,
                             Bundle savedInstanceState) {
        View view = inflater.inflate(R.layout.layout_fm_record, container, false);

        tabLayout = view.findViewById(R.id.tab_layout);
        recyclerView = view.findViewById(R.id.recycler_view);

        MyApplication app = (MyApplication) requireActivity().getApplication();
        List<FileTransferItem> allFileItems = app.getAllFileItems();


        adapter = new FileTransferAdapter(new ArrayList<>(), new FileTransferAdapter.OnItemClickListener() {
            @Override
            public void onDeleteClick(int position) {
                FileTransferItem toDelete = adapter.getItem(position);
                Iterator<FileTransferItem> iterator = allFileItems.iterator();
                while (iterator.hasNext()) {
                    FileTransferItem item = iterator.next();
                    if (item.getTimestamp() == toDelete.getTimestamp()) {
                        iterator.remove();
                        break;
                    }
                }
                filterItemsByTab(currentTabPosition, allFileItems);
                adapter.notifyDataSetChangedAll();
            }

            @Override
            public void onCancelClick(boolean isallfile, String filename) {
                if (isallfile) {
                    cancel_transfers(MyApplication.getContext(), null, true, allFileItems);
                } else {
                    cancel_transfers(MyApplication.getContext(), filename, false, allFileItems);
                }
            }

            @Override
            public void onOpenFileClick(boolean isallfile, String path) {
                if (actionListener != null) {
                    actionListener.onOpenFileClick(isallfile, path);
                }
            }
        });
        recyclerView.setLayoutManager(new LinearLayoutManager(getContext()));
        recyclerView.addItemDecoration(new SpaceItemDecoration(5));
        recyclerView.setAdapter(adapter);

        // 设置Tab
        tabLayout.removeAllTabs();
        tabLayout.addTab(tabLayout.newTab().setText("In Progress"));
        tabLayout.addTab(tabLayout.newTab().setText("Received"));
        tabLayout.addTab(tabLayout.newTab().setText("Fail to send"));

        tabLayout.addOnTabSelectedListener(new TabLayout.OnTabSelectedListener() {
            @Override
            public void onTabSelected(TabLayout.Tab tab) {
                currentTabPosition = tab.getPosition();
                filterItemsByTab(currentTabPosition, allFileItems);
                adapter.notifyDataSetChangedAll();
            }

            @Override
            public void onTabUnselected(TabLayout.Tab tab) {
            }

            @Override
            public void onTabReselected(TabLayout.Tab tab) {
            }
        });


        filterItemsByTab(currentTabPosition, allFileItems);
        adapter.notifyDataSetChangedAll();
        return view;
    }


    private void filterItemsByTab(int tabPosition, List<FileTransferItem> allFileItems) {
        List<FileTransferItem> filteredItems = new ArrayList<>();
        for (FileTransferItem item : allFileItems) {
            Log.e(TAG,"item.getStatus()="+item.getStatus());
            if (tabPosition == TAB_IN_PROGRESS &&
                    (item.getStatus() == FileTransferItem.Status.IN_PROGRESS || item.getStatus() == FileTransferItem.Status.PENDING)) {
                filteredItems.add(item);
            } else if (tabPosition == TAB_RECEIVED && item.getStatus() == FileTransferItem.Status.COMPLETED) {
                filteredItems.add(item);
            } else if (tabPosition == TAB_FAILED && (item.getStatus() == FileTransferItem.Status.CANCEL || item.getStatus() == FileTransferItem.Status.ERROR)) {
                filteredItems.add(item);
            }
        }
        adapter.updateDataItems(filteredItems);
    }

    public void notifyAllUI() {
        adapter.notifyDataSetChangedAll();
    }


    public void updateData(String ip, String id, long timestamp, FileTransferItem item) {
        deviceip = ip;
        deviceid = id;
        currenttimestamp = timestamp;
        if (getActivity() == null || item == null) return;
        getActivity().runOnUiThread(() -> {
            MyApplication app = (MyApplication) requireActivity().getApplication();
            List<FileTransferItem> allFileItems = app.getAllFileItems();

            boolean found = false;
            for (int i = 0; i < allFileItems.size(); i++) {
                FileTransferItem oldItem = allFileItems.get(i);
                if (oldItem.getTimestamp() == item.getTimestamp()) {
                    if (oldItem.getStatus() != FileTransferItem.Status.CANCEL
                            && oldItem.getStatus() != FileTransferItem.Status.ERROR) {
                        oldItem.setCurrentProgress(item.getCurrentProgress());
                        oldItem.setFileSize(item.getFileSize());
                        if (item.getCurrentProgress() == 100) {
                            oldItem.setStatus(FileTransferItem.Status.IN_PROGRESS);
                            // delay1.5s
                            recyclerView.postDelayed(() -> {
                                oldItem.setStatus(FileTransferItem.Status.COMPLETED);
                                filterItemsByTab(currentTabPosition, allFileItems);
                                adapter.notifyDataSetChangedAll();
                            }, 1500);
                        } else {
                            oldItem.setStatus(FileTransferItem.Status.IN_PROGRESS);
                        }
                    }
                    found = true;
                    break;
                }
            }
            if (!found) {
                if (item.getStatus() != FileTransferItem.Status.CANCEL
                        && item.getStatus() != FileTransferItem.Status.ERROR) {
                    if (item.getCurrentProgress() == 100) {
                        item.setStatus(FileTransferItem.Status.IN_PROGRESS);
                        allFileItems.add(0, item);
                        recyclerView.postDelayed(() -> {
                            item.setStatus(FileTransferItem.Status.COMPLETED);
                            filterItemsByTab(currentTabPosition, allFileItems);
                            adapter.notifyDataSetChangedAll();
                        }, 1500);
                    } else {
                        allFileItems.add(0, item);
                    }
                }
            }

            filterItemsByTab(currentTabPosition, allFileItems);
        });
    }



    // dialog for file cancel transfers
    public void cancel_transfers(final Context context, String filename, boolean isAllFile, List<FileTransferItem> allFileItems) {
        View view = View.inflate(context, R.layout.dialog_cancelfile, null);

        TextView titleView = view.findViewById(R.id.title);
        TextView subtitleView = view.findViewById(R.id.subtitle);
        Button conf = view.findViewById(R.id.img_comf);
        Button canl = view.findViewById(R.id.img_canl);

        if (isAllFile) {
            titleView.setText("Cancel all transfers in progress");
            subtitleView.setText("All you sure want to cancel all transfers ?");
        } else {
            titleView.setText("Cancel this transfers in progress ");
            subtitleView.setText("All you sure want to cancel this transfers of " + filename + " ?");
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
                LogUtils.i(TAG, "cancelFileTrans onClick isAllFile="+isAllFile);
                if (isAllFile) {
                    for (FileTransferItem item : allFileItems) {
                        if (item.getStatus() == FileTransferItem.Status.IN_PROGRESS ||
                                item.getStatus() == FileTransferItem.Status.PENDING) {
                            item.setStatus(FileTransferItem.Status.CANCEL);
                        }
                    }
                } else {
                    for (FileTransferItem item : allFileItems) {
                        if (item.getFileName().equals(filename)) {
                            item.setStatus(FileTransferItem.Status.CANCEL);
                        }
                    }
                }
                filterItemsByTab(currentTabPosition, allFileItems);
                adapter.notifyDataSetChangedAll();
                Libp2p_clipboard.cancelFileTrans(deviceip, deviceid, currenttimestamp);
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

    public void selectTab(int tabIndex) {
        if (tabLayout != null && tabLayout.getTabCount() > tabIndex) {
            TabLayout.Tab tab = tabLayout.getTabAt(tabIndex);
            if (tab != null) tab.select();
        }
    }

    @Override
    public void onHiddenChanged(boolean hidden) {
        super.onHiddenChanged(hidden);
    }

    public void onFileTransferError(String errorMsg, long timestamp ) {
        MyApplication app = (MyApplication) requireActivity().getApplication();
        List<FileTransferItem> allFileItems = app.getAllFileItems();
        for (FileTransferItem item : allFileItems) {
            LogUtils.i(TAG, "onFileTransferError item.getFileName()=" + item.getFileName() + ", timestamp="+timestamp);
            if (item.getTimestamp() ==timestamp) {
                if (item.getStatus() != FileTransferItem.Status.CANCEL) {
                    item.setStatus(FileTransferItem.Status.ERROR);
                    break;
                }
            }
        }
        filterItemsByTab(currentTabPosition, allFileItems);
        adapter.notifyDataSetChangedAll();
    }
}
