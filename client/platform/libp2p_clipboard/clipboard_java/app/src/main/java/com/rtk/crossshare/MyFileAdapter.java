package com.rtk.crossshare;

import android.util.Log;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.ImageView;
import android.widget.ProgressBar;
import android.widget.TextView;

import androidx.annotation.NonNull;
import androidx.recyclerview.widget.RecyclerView;

import com.rtk.crossshare.R;

import java.util.List;

public class MyFileAdapter extends RecyclerView.Adapter<MyFileAdapter.MyViewHolder> {

    private String[] mData;

    public MyFileAdapter(String[] data) {
        mData = data;
    }

    private List<GetFile> userList;

    public MyFileAdapter(List<GetFile> userList) {
        this.userList = userList;
    }

    private static OnItemClickListener mOnItemClickListener;
    private static OnItemLongClickListener mOnItemLongClickListener;

    public void setOnItemClickListener(OnItemClickListener mOnItemClickListener){
        this.mOnItemClickListener = mOnItemClickListener;
    }

    public void setOnItemLongClickListener(OnItemLongClickListener mOnItemLongClickListener) {
        this.mOnItemLongClickListener = mOnItemLongClickListener;
    }

    @NonNull
    @Override
    public MyViewHolder onCreateViewHolder(@NonNull ViewGroup parent, int viewType) {
        View view = LayoutInflater.from(parent.getContext())
                .inflate(R.layout.list_fileitem, parent, false);
        return new MyViewHolder(view);
    }

    @Override
    public void onBindViewHolder(@NonNull MyViewHolder holder, int position) {
        //holder.textView.setText(mData[position]);
        GetFile user = userList.get(position);
        holder.textView.setText(user.getFilename());
        holder.textView2.setText(bytekb(user.getFilesize()));
        //holder.mImageView.setImageBitmap(user.getBitmap());

        holder.mProgressBar.setProgress(user.getProgress());
        holder.mProgressBar.setMax(100);
        Log.i("lsz","holder.mProgressBar.getProgress()"+holder.mProgressBar.getProgress());
        if(holder.mProgressBar.getProgress() == 100){
            holder.mProgressBar.setVisibility(View.INVISIBLE);
        }

    }

    @Override
    public int getItemCount() {
        return userList.size();
    }

    public static class MyViewHolder extends RecyclerView.ViewHolder {

        public TextView textView,textView2;
        ImageView mImageView;
        ProgressBar mProgressBar;
        public MyViewHolder(@NonNull View itemView) {
            super(itemView);
            textView = itemView.findViewById(R.id.text_view);
            textView2 = itemView.findViewById(R.id.text_view2);
            mImageView= itemView.findViewById(R.id.imageView);
            mProgressBar= itemView.findViewById(R.id.progress_bar);
        }
    }

    public interface OnItemClickListener{
        void onItemClick(View view,int position);
    }

    public interface OnItemLongClickListener{
        void onItemLongClick(View view,int position);
    }


    public static String bytekb(long bytes) {
        int GB = 1024 * 1024 * 1024;
        int MB = 1024 * 1024;
        int KB = 1024;

        if (bytes / GB >= 1) {
            double gb =Math.round((double) bytes / 1024.0 / 1024.0 / 1024.0 * 100.0) / 100.0;
            return String.format("%.2f", gb) + " GB";
        } else if (bytes / MB >= 1) {
            double mb =Math.round((double) bytes / 1024.0 / 1024.0 * 100.0) / 100.0;
            return String.format("%.2f", mb) + " MB";
        } else if (bytes / KB >= 1) {
            double kb= Math.round((double) bytes / 1024.0 * 100.0) / 100.0;
            return String.format("%.2f", kb) + " KB";
        } else {
            return bytes + "B";
        }
    }
}
