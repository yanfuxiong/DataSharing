package com.realtek.crossshare;

import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.TextView;

import androidx.annotation.NonNull;
import androidx.recyclerview.widget.RecyclerView;

import java.util.List;

public class FileListAdapter extends RecyclerView.Adapter<FileListAdapter.ViewHolder> {

    private List<String> fileNames;

    public FileListAdapter(List<String> fileNames) {
        this.fileNames = fileNames;
    }
    @NonNull
    @Override
    public ViewHolder onCreateViewHolder(@NonNull ViewGroup parent, int viewType) {
        View view = LayoutInflater.from(parent.getContext()).inflate(R.layout.item_file, parent, false);
        return new ViewHolder(view);
    }
    @Override
    public void onBindViewHolder(@NonNull ViewHolder holder, int position) {
        holder.tvFileName.setText(fileNames.get(position));
    }
    @Override
    public int getItemCount() {
        return fileNames.size();
    }
    static class ViewHolder extends RecyclerView.ViewHolder {
        TextView tvFileName;
        ViewHolder(View itemView) {
            super(itemView);
            tvFileName = itemView.findViewById(R.id.tv_file_name);
        }
    }
}
