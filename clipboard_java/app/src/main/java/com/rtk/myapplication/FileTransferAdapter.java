package com.rtk.myapplication;

import android.graphics.Bitmap;
import android.util.Log;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.ImageView;
import android.widget.ProgressBar;
import android.widget.TextView;

import androidx.annotation.NonNull;
import androidx.recyclerview.widget.RecyclerView;

import java.util.List;

public class FileTransferAdapter extends RecyclerView.Adapter<FileTransferAdapter.ViewHolder> {

    private List<FileTransferItem> fileTransferList;

    public FileTransferAdapter(List<FileTransferItem> fileTransferList) {
        this.fileTransferList = fileTransferList;
    }

    @NonNull
    @Override
    public ViewHolder onCreateViewHolder(@NonNull ViewGroup parent, int viewType) {
        View view = LayoutInflater.from(parent.getContext()).inflate(R.layout.item_file_transfer, parent, false);
        return new ViewHolder(view);
    }

    @Override
    public void onBindViewHolder(@NonNull ViewHolder holder, int position) {
        FileTransferItem item = fileTransferList.get(position);
        holder.fileName.setText(item.getFileName());
        holder.fileSize.setText(bytekb(item.getFileSize()));
        holder.fileTime.setText(item.getDateInfo());
        holder.mImageView.setImageBitmap(item.getBitmap());
        holder.progressBar.setMax(100);
        //Log.i("lsz", "init filename progressprogressprogress"+(int) item.getCurrentProgress());
        holder.progressBar.setProgress((int) item.getCurrentProgress());
    }

    @Override
    public int getItemCount() {
        return fileTransferList.size();
    }

    public static class ViewHolder extends RecyclerView.ViewHolder {
        TextView fileName, fileSize, fileTime;
        ProgressBar progressBar;
        ImageView mImageView;

        public ViewHolder(@NonNull View itemView) {
            super(itemView);

            fileName = itemView.findViewById(R.id.file_name);
            fileSize = itemView.findViewById(R.id.file_size);
            fileTime = itemView.findViewById(R.id.file_time);
            progressBar = itemView.findViewById(R.id.progress_bar);
            mImageView = itemView.findViewById(R.id.imageView);
        }
    }

    // 用于更新进度的辅助方法
    public void updateProgress(String fileName, long progress) {

        for (FileTransferItem item : fileTransferList) {
            if (item.getFileName().equals(fileName)) {
                item.setCurrentProgress(progress);
                notifyItemChanged(fileTransferList.indexOf(item));
                break;
            }
        }

    }

    public void setBitmap(String fileName, Bitmap mBitmap) {

        for (FileTransferItem item : fileTransferList) {
            if (item.getFileName().equals(fileName)) {
                item.setBitmap(mBitmap);
                notifyItemChanged(fileTransferList.indexOf(item));
                break;
            }

        }

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
            return String.format("%.2f", mb) + " MB";
        } else if (bytes / KB >= 1) {
            double kb = Math.round((double) bytes / 1024.0 * 100.0) / 100.0;
            return String.format("%.2f", kb) + " KB";
        } else {
            return bytes + "B";
        }
    }

}
