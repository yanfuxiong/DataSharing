package com.realtek.crossshare;

import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.ImageView;
import android.widget.TextView;
import androidx.annotation.NonNull;
import androidx.recyclerview.widget.RecyclerView;

import java.util.List;

public class LicenseAdapter extends RecyclerView.Adapter<LicenseAdapter.ViewHolder> {

    public interface OnItemClickListener {
        void onClick(LicenseItem item);
    }

    private List<LicenseItem> data;
    private OnItemClickListener listener;

    public LicenseAdapter(List<LicenseItem> data, OnItemClickListener listener) {
        this.data = data;
        this.listener = listener;
    }

    @NonNull
    @Override
    public ViewHolder onCreateViewHolder(@NonNull ViewGroup parent, int viewType) {
        View view = LayoutInflater.from(parent.getContext())
                .inflate(R.layout.item_license, parent, false);
        return new ViewHolder(view);
    }

    @Override
    public void onBindViewHolder(@NonNull ViewHolder holder, int position) {
        LicenseItem item = data.get(position);
        holder.tvName.setText(item.getName());
        holder.itemView.setOnClickListener(v -> {
            if (listener != null) listener.onClick(item);
        });
    }

    @Override
    public int getItemCount() {
        return data.size();
    }

    static class ViewHolder extends RecyclerView.ViewHolder {
        TextView tvName;
        ImageView ivArrow;
        ViewHolder(View itemView) {
            super(itemView);
            tvName = itemView.findViewById(R.id.tv_license_name);
            ivArrow = itemView.findViewById(R.id.iv_arrow);
        }
    }
}
