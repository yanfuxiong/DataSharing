package com.realtek.crossshare;

import android.graphics.Bitmap;

public class FileTransferItem {

    private String fileName;
    private long fileSize;

    public void setFileSize(long fileSize) {
        this.fileSize = fileSize;
    }

    private long currentProgress;
    private Bitmap bitmap;

    public void setFile_revice(String file_revice) {
        this.file_revice = file_revice;
    }

    public String getFile_count() {
        return file_count;
    }

    public void setFile_count(String file_count) {
        this.file_count = file_count;
    }

    public void setFile_sizecount(String file_sizecount) {
        this.file_sizecount = file_sizecount;
    }

    public void setFile_size(String file_size) {
        this.file_size = file_size;
    }

    public String getFile_revice() {
        return file_revice;
    }

    public String getFile_sizecount() {
        return file_sizecount;
    }

    public String getFile_size() {
        return file_size;
    }

    private String dateInfo;
    private Status status;
    private String file_revice;
    private String file_count;
    private String file_size;
    private String file_sizecount;
    public static final int FILE_TYPE_SINGLE = 1;
    public static final int FILE_TYPE_MULTIPLE = 2;
    public static final int FILE_TYPE_DEFAULT = -1;

    public void setFile_tpye(int file_tpye) {
        this.file_tpye = file_tpye;
    }

    public int getFile_tpye() {
        return file_tpye;
    }

    private String  file_devicename;
    private int file_tpye;

    public String getFile_devicename() {
        return file_devicename;
    }

    public void setFile_devicename(String file_devicename) {
        this.file_devicename = file_devicename;
    }

    public FileTransferItem(String fileName, long fileSize, Bitmap bitmap) {
        this.fileName = fileName;
        this.fileSize = fileSize;
        this.bitmap = bitmap;
        this.currentProgress = 0;
        this.status=Status.PENDING;
        this.file_revice="";
        this.file_count="";
        this.file_size="";
        this.file_sizecount="";
        this.file_devicename="";
        this.file_tpye = FILE_TYPE_DEFAULT;
    }

    public Status getStatus() {
        return status;
    }

    public void setStatus(Status status) {
        this.status = status;
    }

    public String getFileName() {
        return fileName;
    }

    public void setFileName(String fileName) {
        this.fileName = fileName;
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


    public enum Status {
        PENDING,
        IN_PROGRESS,
        COMPLETED,
        ERROR,
        CANCEL
    }
}
