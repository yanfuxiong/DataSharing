package com.rtk.myapplication;

import android.content.Context;
import android.content.ClipData;
import android.content.ClipboardManager;

import java.io.ByteArrayOutputStream;
import java.io.InputStream;
import java.util.Calendar;
import java.util.concurrent.atomic.AtomicReference;

import android.graphics.BitmapFactory;
import android.graphics.Bitmap;
import android.provider.MediaStore;
import android.net.Uri;
import android.util.Log;


public class ClipboardUtils {

    private static final String TAG = "ClipboardUtils";

    public static final int CLIPBOARD_DATA_TYPE_TEXT = 0;
    public static final int CLIPBOARD_DATA_TYPE_IMAGE = 1;
    public static final int CLIPBOARD_DATA_TYPE_UNSUPPORT = -1;

    private ClipboardManager mClipboardManager = null;

    public static Context context;

    public ClipboardUtils() {
        if (context == null) {
            Log.e(TAG, "set Content first");
            return;
        }
        mClipboardManager = (ClipboardManager) context.getSystemService(Context.CLIPBOARD_SERVICE);
        if (mClipboardManager == null) {
            Log.e(TAG, "get ClipboardManager error");
        }
    }

    public static void setContext(Context inContext) {
        context = inContext;
    }

    /**
     * 获取 ClipboardUtils 单例
     */
    public static ClipboardUtils getInstance() {
        return Holder.sInstance;
    }

    private static class Holder {
        private static ClipboardUtils sInstance = new ClipboardUtils();
    }

    public static AtomicReference<ClipData> createClipdataRef() {
        return new AtomicReference<>(null);
    }

    /**
     * 剪切板是否有数据
     */
    public boolean hasClip() {
        Log.d(TAG, "java call:hasclip");
        return mClipboardManager.hasPrimaryClip();
    }

    /**
     * 清除剪切板数据
     */
    public void clearClip() {
        Log.d(TAG, "java call:clearClip");
        mClipboardManager.clearPrimaryClip();
    }

    /**
     * 添加文本类型Item数据
     */
    public void addTextItem(AtomicReference<ClipData> clipDataRef, String text) {
        try {
            if (clipDataRef.get() == null) {
                ClipData clipData = ClipData.newPlainText("text_label", text);
                clipDataRef.set(clipData);
            } else {
                ClipData.Item item = ClipData.newPlainText("text_label", text).getItemAt(0);
                clipDataRef.get().addItem(item);
            }

            Log.e(TAG, "lenTextItem1=" + clipDataRef.get().getItemCount());

        } catch (Exception e) {
            Log.e(TAG, "Error adding text item to ClipData");
        }
    }

    /**
     * 添加图片类型Item数据
     */
    public void addImageItem(AtomicReference<ClipData> clipDataRef, Bitmap image) {
        try {
            ByteArrayOutputStream stream = new ByteArrayOutputStream();
            //image.compress(Bitmap.CompressFormat.PNG, 100, stream);
            image.compress(Bitmap.CompressFormat.JPEG, 100, stream);
            //String path = MediaStore.Images.Media.insertImage(context.getContentResolver(),image,"ImageX",null);
            String path = MediaStore.Images.Media.insertImage(context.getContentResolver(), image, "ImageX" + Calendar.getInstance().getTime(), null);
            Log.e(TAG, "image item to ClipData path" + path);
            if (clipDataRef.get() == null) {
                ClipData clipData = ClipData.newRawUri("image_label", Uri.parse(path));
                clipDataRef.set(clipData);
            } else {
                ClipData.Item item = ClipData.newRawUri("image_label", Uri.parse(path)).getItemAt(0);
                clipDataRef.get().addItem(item);
            }
        } catch (Exception e) {
            Log.e(TAG, "11Error adding image item to ClipData");
        }
    }

    /**
     * 根据索引获取剪贴板中的文本项
     */
    public String getTextItem(AtomicReference<ClipData> clipDataRef, int index) {
        try {
            if (clipDataRef.get() != null && index >= 0 && index < clipDataRef.get().getItemCount()) {
                ClipData.Item item = clipDataRef.get().getItemAt(index);
                int type = getItemType(clipDataRef, index);
                if (type != CLIPBOARD_DATA_TYPE_TEXT) {
                    return null;
                }
                // 直接返回文本内容，如果获取成功
                return item.getText().toString();
            } else {
                Log.d(TAG, "index override");
            }
        } catch (Exception e) {
            Log.e(TAG, "Error getting text item from ClipData");
        }
        // 如果索引无效或出现异常，返回null表示获取失败
        return null;
    }

    /**
     * 根据索引获取剪贴板中的图片项
     */
    public Bitmap getImageItem(AtomicReference<ClipData> clipDataRef, int index) {
        try {
            if (clipDataRef.get() != null && index >= 0 && index < clipDataRef.get().getItemCount()) {
                ClipData.Item item = clipDataRef.get().getItemAt(index);
                int type = getItemType(clipDataRef, index);
                if (type != CLIPBOARD_DATA_TYPE_IMAGE) {
                    return null;
                }
                if (item.getUri() != null) {
                    InputStream inputStream = context.getContentResolver().openInputStream(item.getUri());
                    Bitmap bitmap = BitmapFactory.decodeStream(inputStream);
                    if (bitmap != null) {
                        return bitmap;
                    }
                }
            } else {
                Log.d(TAG, "index override");
            }
        } catch (Exception e) {
            Log.e(TAG, "Error getting image item from ClipData");
        }
        return null; // 索引无效或数据类型不匹配
    }

    /**
     * 获取剪切板中Item的数量
     */
    public int getItemCount(AtomicReference<ClipData> clipDataRef) {

        return clipDataRef.get().getItemCount();
    }

    /**
     * 将当前的mGetClipData设置为剪贴板的主内容
     */
    public void setPrimaryClip(AtomicReference<ClipData> clipDataRef) {
        if (mClipboardManager != null && clipDataRef.get() != null) {
            mClipboardManager.setPrimaryClip(clipDataRef.get());
        }
    }

    /**
     * 获取剪贴板中主剪贴板的内容
     */
    public void getPrimaryClip(AtomicReference<ClipData> clipDataRef) {
        if (mClipboardManager != null && mClipboardManager.hasPrimaryClip()) {
            ClipData clipdata = mClipboardManager.getPrimaryClip();
            clipDataRef.set(clipdata);
        }
    }

    /**
     * 获取Item类型
     */
    public int getItemType(AtomicReference<ClipData> clipDataRef, int index) {
        if (clipDataRef.get() != null && index >= 0 && index < clipDataRef.get().getItemCount()) {
            ClipData.Item item = clipDataRef.get().getItemAt(index);
            Uri uri = item.getUri();
            if (uri == null) {
                return CLIPBOARD_DATA_TYPE_TEXT;
            } else {
                String mimeType = context.getContentResolver().getType(item.getUri());
                if (mimeType != null) {
                    if (mimeType.startsWith("image/")) {
                        return CLIPBOARD_DATA_TYPE_IMAGE;
                    } else {
                        return CLIPBOARD_DATA_TYPE_UNSUPPORT;
                    }
                }
            }
        }
        return -2;
    }

}
