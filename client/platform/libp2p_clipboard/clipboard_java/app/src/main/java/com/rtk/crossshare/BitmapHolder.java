package com.rtk.crossshare;

import android.graphics.Bitmap;

public class BitmapHolder {

    private static Bitmap bitmap;

    public static void setBitmap(Bitmap bitmap) {
        BitmapHolder.bitmap = bitmap;
    }

    public static Bitmap getBitmap() {
        return bitmap;
    }
}
