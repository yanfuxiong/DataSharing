package com.realtek.crossshare;

import android.annotation.SuppressLint;
import android.graphics.Bitmap;
import android.util.Log;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.ImageView;
import android.widget.LinearLayout;
import android.widget.ProgressBar;
import android.widget.TextView;

import androidx.annotation.NonNull;
import androidx.core.content.ContextCompat;
import androidx.recyclerview.widget.DiffUtil;
import androidx.recyclerview.widget.RecyclerView;

import com.realtek.crossshare.R;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.List;
import java.util.ArrayList;

public class FileTransferAdapter extends RecyclerView.Adapter<FileTransferAdapter.ViewHolder> {

    private OnItemClickListener listener;

    public interface OnItemClickListener {
        void onDeleteClick(int position);
        void onCancelClick(boolean isallfile,String filename);
        void onOpenFileClick(boolean isallfile,String path);
    }


    private List<FileTransferItem> fileTransferList= new ArrayList<>();

    private List<FileTransferItem> files;

    public FileTransferAdapter(List<FileTransferItem> fileTransferList,OnItemClickListener listener) {
        this.fileTransferList = fileTransferList;
        this.files = fileTransferList;
        this.listener = listener;
    }


    @NonNull
    @Override
    public ViewHolder onCreateViewHolder(@NonNull ViewGroup parent, int viewType) {
        View view = LayoutInflater.from(parent.getContext()).inflate(R.layout.item_file_transfer, parent, false);
        return new ViewHolder(view);
    }

    @SuppressLint("ResourceAsColor")
    @Override
    public void onBindViewHolder(@NonNull ViewHolder holder, int position) {
        FileTransferItem item = fileTransferList.get(position);

        String shownName = item.getDisplayName() != null && !item.getDisplayName().isEmpty()
                ? item.getDisplayName()
                : item.getFileName();
        holder.fileName.setText(shownName);
        holder.fileSize.setText(bytekb(item.getFileSize()));
        holder.fileTime.setText(item.getDateInfo());
        holder.mImageView.setImageBitmap(item.getBitmap());
        holder.progressBar.setMax(100);

        holder.progressBar.setProgress((int) item.getCurrentProgress());
        holder.percentage.setText(String.valueOf(item.getCurrentProgress()) + "%");

        holder.file_revice.setText(item.getFile_revice());
        holder.file_count.setText(item.getFile_count());

        holder.file_revice_size.setText(item.getFile_size());
        holder.file_count_size.setText(item.getFile_sizecount());
        holder.file_devicename.setText("From "+item.getFile_devicename());

        holder.mImageView.setVisibility(View.VISIBLE);
        holder.mImageView.setImageResource(R.drawable.device);

        holder.close.setVisibility(View.VISIBLE);
        holder.close.setOnClickListener(v -> {
            if (listener != null) {
                if ((int) item.getCurrentProgress() == 100) {
                    listener.onDeleteClick(position);
                } else {
                    if(item.getStatus() == FileTransferItem.Status.CANCEL || item.getStatus() == FileTransferItem.Status.ERROR){
                        listener.onDeleteClick(position);
                    }else {
                        if (item.getFile_tpye() == FileTransferItem.FILE_TYPE_SINGLE ||
                                item.getFile_tpye() == FileTransferItem.FILE_TYPE_DEFAULT) {
                            listener.onCancelClick(false, holder.fileName.getText().toString());
                        } else {
                            listener.onCancelClick(true, null);
                        }
                    }
                }
            }
        });

        holder.file_opne.setOnClickListener(v -> {
            //to do openfile
            if (listener != null) {
                if (item.getFile_tpye() == FileTransferItem.FILE_TYPE_SINGLE ||
                        item.getFile_tpye() == FileTransferItem.FILE_TYPE_DEFAULT) {
                    listener.onOpenFileClick(false,item.getFilePath());
                } else {
                    listener.onOpenFileClick(true,item.getFilePath());
                }
            }
        });

        if (item.getFile_tpye() == FileTransferItem.FILE_TYPE_SINGLE ||
                item.getFile_tpye() == FileTransferItem.FILE_TYPE_DEFAULT) {
            holder.layout_file_info.setVisibility(LinearLayout.GONE);
        } else {
            holder.layout_file_info.setVisibility(LinearLayout.VISIBLE);
        }

        holder.file_devicename.setVisibility(View.VISIBLE);
        switch (item.getStatus()) {
            case IN_PROGRESS:
                holder.progressBar.setProgressDrawable(ContextCompat.getDrawable(MyApplication.getContext(), R.drawable.progress_receiving));
                int color = ContextCompat.getColor(MyApplication.getContext(), R.color.purple_800);
                holder.result.setTextColor(color);
                holder.percentage.setTextColor(color);
                holder.result.setVisibility(View.GONE);
                break;
            case COMPLETED:
                holder.progressBar.setProgressDrawable(ContextCompat.getDrawable(MyApplication.getContext(), R.drawable.progress_success));
                int color2 = ContextCompat.getColor(MyApplication.getContext(), R.color.teal_200);
                holder.result.setTextColor(color2);
                holder.percentage.setTextColor(color2);
                holder.result.setVisibility(View.VISIBLE);
                holder.result.setText("Complete");
                break;
            case ERROR:
                holder.progressBar.setProgressDrawable(ContextCompat.getDrawable(MyApplication.getContext(), R.drawable.progress_failed));
                int color3 = ContextCompat.getColor(MyApplication.getContext(), R.color.red);
                holder.result.setTextColor(color3);
                holder.percentage.setTextColor(color3);
                holder.result.setVisibility(View.VISIBLE);
                holder.result.setText("Error");
                break;
            case CANCEL:
                holder.progressBar.setProgressDrawable(ContextCompat.getDrawable(MyApplication.getContext(), R.drawable.progress_failed));
                int color4 = ContextCompat.getColor(MyApplication.getContext(), R.color.red);
                holder.result.setTextColor(color4);
                holder.percentage.setTextColor(color4);
                holder.result.setVisibility(View.VISIBLE);
                holder.result.setText("Cancel transfers");
                holder.close.setVisibility(View.GONE);

        }

        if((int) item.getCurrentProgress() == 100){
            holder.close.setVisibility(View.VISIBLE);
            holder.close.setImageResource(R.drawable.garbage);
            holder.result.setVisibility(View.VISIBLE);
            holder.result.setText("Complete");
            holder.mImageView.setVisibility(View.VISIBLE);
            holder.mImageView.setImageResource(R.drawable.device);
            holder.file_opne.setVisibility(View.VISIBLE);
        }else{
            holder.close.setVisibility(View.VISIBLE);
            holder.close.setImageResource(R.drawable.cancel);
            holder.file_opne.setVisibility(View.INVISIBLE);
        }

        if(item.getStatus() == FileTransferItem.Status.CANCEL || item.getStatus() == FileTransferItem.Status.ERROR){
            holder.close.setVisibility(View.VISIBLE);
            holder.close.setImageResource(R.drawable.garbage);
        }

    }


    @Override
    public int getItemCount() {
        return fileTransferList.size();
    }

    public static class ViewHolder extends RecyclerView.ViewHolder {
        TextView fileName, fileSize, fileTime, percentage, result, file_revice,file_count,file_revice_size,file_count_size,file_devicename;
        ProgressBar progressBar;
        ImageView mImageView ,close,file_opne;
        LinearLayout layout_file_info;

        public ViewHolder(@NonNull View itemView) {
            super(itemView);

            fileName = itemView.findViewById(R.id.file_name);
            fileSize = itemView.findViewById(R.id.file_size);
            fileTime = itemView.findViewById(R.id.file_time);
            progressBar = itemView.findViewById(R.id.progress_bar);
            mImageView = itemView.findViewById(R.id.imageView);
            percentage = itemView.findViewById(R.id.percentage);
            result = itemView.findViewById(R.id.result);
            file_revice =itemView.findViewById(R.id.file_revice);
            file_count =itemView.findViewById(R.id.file_count);
            file_revice_size =itemView.findViewById(R.id.file_revice_size);
            file_count_size  =itemView.findViewById(R.id.file_count_size);
            file_devicename=itemView.findViewById(R.id.file_device);
            close=itemView.findViewById(R.id.close);
            layout_file_info = itemView.findViewById(R.id.layout_file_info);
            file_opne=itemView.findViewById(R.id.file_opne);
        }
    }

    public void updateProgress(String fileName, long progress) {
        //Log.i("lsz", "init filename fileTransferList.progress()=" + progress);
        if(fileTransferList != null) {
            for (int i = 0; i < fileTransferList.size(); i++) {
                FileTransferItem item = fileTransferList.get(i);
                if (item.getFileName().equals(fileName)) {
                    item.setCurrentProgress(progress);
                    //if (progress == 100) {
                    //    item.setStatus(FileTransferItem.Status.COMPLETED);
                    //} else {
                    //    item.setStatus(FileTransferItem.Status.IN_PROGRESS);
                    //}
                    item.setFile_tpye(FileTransferItem.FILE_TYPE_SINGLE);
                    notifyItemChanged(i);
                    //notifyDataSetChanged();
                    break;
                }
            }
        }

    }


    public void setBitmap(String fileName, Bitmap mBitmap) {
        if(fileTransferList != null) {
            for (int i = 0; i < fileTransferList.size(); i++) {
                FileTransferItem item = fileTransferList.get(i);
                if (item.getFileName().equals(fileName)) {
                    item.setBitmap(mBitmap);
                    item.setStatus(FileTransferItem.Status.COMPLETED);
                    item.setCurrentProgress(100);
                    notifyItemChanged(i);
                    break;
                }

            }
        }

    }

    public void getFileTime(String fileName) {
        if(fileTransferList != null) {
            for (int i = 0; i < fileTransferList.size(); i++) {
                FileTransferItem item = fileTransferList.get(i);
                if (item.getFileName().equals(fileName)) {
                    item.setDateInfo(formatReceiveTime());
                    notifyItemChanged(i);
                    break;
                }

            }
        }

    }

    private String formatReceiveTime() {
        // Get the current date and time
        LocalDateTime now = LocalDateTime.now();
        // Define the desired date format
        DateTimeFormatter formatter = DateTimeFormatter.ofPattern("yyyy.MM.dd HH:mm:ss");
        // Format the current date and time
        return now.format(formatter);
    }


    public static String bytekb(long bytes) {
        int GB = 1024 * 1024 * 1024;
        int MB = 1024 * 1024;
        int KB = 1024;

        if (bytes / GB >= 1) {
            double gb = Math.round((double) bytes / 1024.0 / 1024.0 / 1024.0 * 100.0) / 100.0;
            return String.format("%.2f", gb) + " GB";
        } else if (bytes / MB >= 1) {
            double mb = Math.round((double) bytes / 1024.0 / 1024.0 * 100.0) / 100.0;
            return String.format("%.2f", mb) + " MB";
        } else if (bytes / KB >= 1) {
            double kb = Math.round((double) bytes / 1024.0 * 100.0) / 100.0;
            return String.format("%.2f", kb) + " KB";
        } else {
            return bytes + "B";
        }
    }

    public void updateFileListWithError(String fileName, int msg) {
        if(fileTransferList != null) {
            for (int i = 0; i < fileTransferList.size(); i++) {
                FileTransferItem fileInfo = fileTransferList.get(i);
                if (fileInfo.getFileName().equals(fileName)) {
                    fileInfo.setStatus(FileTransferItem.Status.ERROR);
                    notifyItemChanged(i);
                }
            }
        }
    }

    public void removeSameFile(String fileName, long filesize) {
        for (int i = 0; i < fileTransferList.size(); i++) {
            FileTransferItem fileInfo = fileTransferList.get(i);
            if (fileInfo.getFileName().equals(fileName) && fileInfo.getFileSize()==filesize ) {
                fileTransferList.remove(i);
                break;
            }
        }
    }


    public void  updateFileList(String devicename, String currentFileName, long sentFileCnt, long totalFileCnt, long currentFileSize, long totalSize, long sentSize,long percentage ,long timestamp) {

        if(fileTransferList != null) {

            for (int i = 0; i < fileTransferList.size(); i++) {

                FileTransferItem item = fileTransferList.get(i);
                if(item.getTimestamp() == timestamp ) {
                    item.setCurrentProgress(percentage);
                    item.setFile_revice(String.valueOf(sentFileCnt) + "/");
                    item.setFile_count(String.valueOf(totalFileCnt));
                    item.setFileSize(currentFileSize);
                    item.setFileName(currentFileName);
                    item.setFile_size(bytekb(sentSize));
                    item.setFile_sizecount(bytekb(totalSize));
                    if((int)percentage == 100){
                        item.setDateInfo(formatReceiveTime());
                    }
                    item.setFile_devicename(devicename);
                    if(totalFileCnt == 1){
                        item.setFile_tpye(FileTransferItem.FILE_TYPE_SINGLE);
                    }else {
                        item.setFile_tpye(FileTransferItem.FILE_TYPE_MULTIPLE);
                    }
                    notifyItemChanged(i);
                    //notifyDataSetChanged();
                    break;
                }
            }
        }
    }



    public void removeItem(int position) {
        fileTransferList.remove(position);
        notifyItemRemoved(position);
        notifyItemRangeChanged(position, fileTransferList.size());
    }

    public void removeAllItem() {
        fileTransferList.clear();
        notifyDataSetChanged();
    }

    public void cancelTransfers(){
        for (int i = 0; i < fileTransferList.size(); i++) {
            FileTransferItem item = fileTransferList.get(i);
            item.setStatus(FileTransferItem.Status.CANCEL);
            notifyItemChanged(i);
            break;

        }
    }

    public List<FileTransferItem> getItemList() {
        return new ArrayList<>(fileTransferList);
    }



    public void updateDataItems(List<FileTransferItem> newData) {
        fileTransferList.clear();
        if (newData != null) {
            fileTransferList.addAll(newData);
        }
        for (int i = 0; i < fileTransferList.size(); i++) {
            notifyItemChanged(i);
        }
    }

    public FileTransferItem getItem(int position) {
        return fileTransferList.get(position);
    }

    public void notifyDataSetChangedAll() {
        notifyDataSetChanged();
    }

    public void updateFileTransferDisplayName(long timestamp, String realName, String downLoadPath) {
        for (FileTransferItem item : fileTransferList) {
            if (item.getTimestamp() == timestamp) {
                item.setDisplayName(realName);
                item.setFilePath(downLoadPath);
                break;
            }
        }
        notifyDataSetChangedAll();
    }
}
