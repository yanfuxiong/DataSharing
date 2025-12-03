package com.rtk.crossshare;

import android.graphics.Bitmap;

public class GetFile {

        private String filename;
        private long filesize;
        private Bitmap bitmap;
        private int progress;

        public GetFile(String filename, long filesize,Bitmap bitmap,int progress) {
            this.filename = filename;
            this.filesize = filesize;
            this.bitmap = bitmap;
            this.progress = progress;
        }

        public String getFilename() {
            return filename;
        }

        public long getFilesize() {
            return filesize;
        }
        public Bitmap getBitmap() {
            return bitmap;
        }

        public int getProgress() {
            return progress;
        }


}
