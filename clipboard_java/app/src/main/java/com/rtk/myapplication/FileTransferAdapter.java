package com.rtk.myapplication;

import android.annotation.SuppressLint;
import android.graphics.Bitmap;
import android.util.Log;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.ImageView;
import android.widget.ProgressBar;
import android.widget.TextView;

import androidx.annotation.NonNull;
import androidx.core.content.ContextCompat;
import androidx.recyclerview.widget.RecyclerView;

import java.text.SimpleDateFormat;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.Date;
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

    @SuppressLint("ResourceAsColor")
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
        holder.percentage.setText(String.valueOf(item.getCurrentProgress()) + "%");

        Log.i("lsz", "init filename item.getStatus()" + item.getStatus());
        switch (item.getStatus()) {
            case IN_PROGRESS:
                holder.progressBar.setProgressDrawable(ContextCompat.getDrawable(MyApplication.getContext(), R.drawable.progress_receiving));
                int color = ContextCompat.getColor(MyApplication.getContext(), R.color.purple_800);
                holder.result.setTextColor(color);
                holder.percentage.setTextColor(color);
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
        }



    }


    @Override
    public int getItemCount() {
        return fileTransferList.size();
    }

    public static class ViewHolder extends RecyclerView.ViewHolder {
        TextView fileName, fileSize, fileTime, percentage, result;
        ProgressBar progressBar;
        ImageView mImageView;

        public ViewHolder(@NonNull View itemView) {
            super(itemView);

            fileName = itemView.findViewById(R.id.file_name);
            fileSize = itemView.findViewById(R.id.file_size);
            fileTime = itemView.findViewById(R.id.file_time);
            progressBar = itemView.findViewById(R.id.progress_bar);
            mImageView = itemView.findViewById(R.id.imageView);
            percentage = itemView.findViewById(R.id.percentage);
            result = itemView.findViewById(R.id.result);
        }
    }

    public void updateProgress(String fileName, long progress) {
        //Log.i("lsz", "init filename fileTransferList.progress()" + progress);
        for (int i = 0; i < fileTransferList.size(); i++) {
            FileTransferItem item = fileTransferList.get(i);
            if (item.getFileName().equals(fileName)) {
                item.setCurrentProgress(progress);
                notifyItemChanged(i);
                item.setStatus(FileTransferItem.Status.IN_PROGRESS);
                notifyDataSetChanged();
                break;
            }
        }

    }


    public void setBitmap(String fileName, Bitmap mBitmap) {
        for (int i = 0; i < fileTransferList.size(); i++) {
            FileTransferItem item = fileTransferList.get(i);
            if (item.getFileName().equals(fileName)) {
                item.setBitmap(mBitmap);
                //item.setSuccess(true);
                item.setStatus(FileTransferItem.Status.COMPLETED);
                notifyItemChanged(i);
                break;
            }

        }

    }

    public void getFileTime(String fileName) {
        for (int i = 0; i < fileTransferList.size(); i++) {
            FileTransferItem item = fileTransferList.get(i);
            if (item.getFileName().equals(fileName)) {
                item.setDateInfo(formatReceiveTime());
                notifyItemChanged(i);
                break;
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
        for (int i = 0; i < fileTransferList.size(); i++) {
            FileTransferItem fileInfo = fileTransferList.get(i);
            if (fileInfo.getFileName().equals(fileName)) {
                fileInfo.setStatus(FileTransferItem.Status.ERROR);
                notifyItemChanged(i);
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
}
