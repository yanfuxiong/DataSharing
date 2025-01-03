package com.rtk.myapplication;

import android.app.Activity;
import android.graphics.Bitmap;
import android.graphics.BitmapFactory;
import android.os.Bundle;
import android.util.Log;
import android.widget.TextView;
import android.widget.Toast;

import androidx.annotation.Nullable;
import androidx.recyclerview.widget.LinearLayoutManager;
import androidx.recyclerview.widget.RecyclerView;

import java.io.File;
import java.util.Arrays;
import java.util.List;

public class getFileActivity extends Activity {


    @Override
    protected void onCreate(@Nullable Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.layout_getfileactivity);

        boolean booleanValue = getIntent().getBooleanExtra("booleanKey", false);
        String filename = getIntent().getStringExtra("filename");
        long filesize = getIntent().getLongExtra("filesize",-1L);
        String bitmappath = getIntent().getStringExtra("bitmappath");

        Log.d("lszz", "lsz filename filename=" + filename);
        Log.d("lszz", "long filesize=" + filesize);
        Log.d("lszz", "String bitmappath=" + bitmappath);
        Log.d("lszz", "getBitmap=" + getBitmap(bitmappath));

        if (booleanValue) {
            getFileList(filename,filesize,getBitmap(bitmappath));

            File file = new File(bitmappath);
            if(file.exists()){
                Log.d("lszz", "storage private app file is exists,now remove");
                file.delete();
            }
        }

    }


    public void getFileList(String name, long siez, Bitmap bitmap) {
        RecyclerView recyclerView = findViewById(R.id.my_recycler_view2);
        recyclerView.setLayoutManager(new LinearLayoutManager(this, LinearLayoutManager.VERTICAL, false));

        List<GetFile> users = Arrays.asList(
                new GetFile(name, siez,bitmap,11)
        );

        Toast.makeText(MyApplication.getContext(), "文件已存入storage/emulated/0/Download/", Toast.LENGTH_SHORT).show();
        MyFileAdapter myadapter = new MyFileAdapter(users);
        recyclerView.setAdapter(myadapter);

    }



    public  Bitmap getBitmap(String privateFilePath) {
        File file = new File(privateFilePath);
        Log.i("lsz","ss name "+file.getName().substring(file.getName().length() -3));
        if(file.getName().substring(file.getName().length() -3).equals("png") ||
                file.getName().substring(file.getName().length() -3).equals("jpg")) {
            Log.i("lsz","ss name  init");

            return BitmapFactory.decodeFile(file.getAbsolutePath());
        }
        return null;
    }
}
