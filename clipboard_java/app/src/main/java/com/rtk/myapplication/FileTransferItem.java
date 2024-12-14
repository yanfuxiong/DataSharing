package com.rtk.myapplication;

import android.graphics.Bitmap;

public class FileTransferItem {

    private String fileName;
    private long fileSize;
    private long currentProgress;
    private Bitmap bitmap;
    private String dateInfo;

    public FileTransferItem(String fileName, long fileSize, Bitmap bitmap) {
        this.fileName = fileName;
        this.fileSize = fileSize;
        this.bitmap = bitmap;
        this.currentProgress = 0;
    }

    public String getFileName() {
        return fileName;
    }

    public long getFileSize() {
        return fileSize;
    }

    public Bitmap getBitmap() {
        return bitmap;
    }

    public long getCurrentProgress() {
        return currentProgress;
    }

    public void setCurrentProgress(long currentProgress) {
        this.currentProgress = currentProgress;
    }

    public void setBitmap(Bitmap bitmap) {
        this.bitmap = bitmap;
    }

    public void setDateInfo(String dateInfo) { this.dateInfo = dateInfo;}
    public String getDateInfo() {return dateInfo; }
}
