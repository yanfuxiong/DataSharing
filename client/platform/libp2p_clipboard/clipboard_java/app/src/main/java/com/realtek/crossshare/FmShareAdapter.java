package com.realtek.crossshare;

import android.content.Context;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.ImageView;
import android.widget.TextView;

import androidx.annotation.NonNull;
import androidx.recyclerview.widget.RecyclerView;

import java.util.List;

public class FmShareAdapter extends RecyclerView.Adapter<FmShareAdapter.DeviceViewHolder> {

    private static MyAdapter.OnItemClickListener mOnItemClickListener;

    public void setOnItemClickListener(MyAdapter.OnItemClickListener mOnItemClickListener) {
        this.mOnItemClickListener = mOnItemClickListener;
    }

    private List<Device> deviceList;
    private Context context;

    public FmShareAdapter(Context context, List<Device> deviceList) {
        this.context = context;
        this.deviceList = deviceList;
    }

    @NonNull
    @Override
    public DeviceViewHolder onCreateViewHolder(@NonNull ViewGroup parent, int viewType) {
        View view = LayoutInflater.from(context).inflate(R.layout.list_fmshare_device, parent, false);
        return new DeviceViewHolder(view);
    }

    @Override
    public void onBindViewHolder(@NonNull DeviceViewHolder holder, int position) {
        Device device = deviceList.get(position);
        holder.deviceName.setText(device.getName());
        holder.deviceIp.setText(device.getIp());
        holder.deviceIcon.setImageResource(device.getIconResId());

        if (mOnItemClickListener != null) {

            holder.deviceIcon.setOnClickListener(new View.OnClickListener() {
                @Override
                public void onClick(View v) {
                    int position = holder.getLayoutPosition(); // 1
                    mOnItemClickListener.onItemClick(holder.deviceName, position); // 2
                }
            });
        }

    }

    @Override
    public int getItemCount() {
        return deviceList.size();
    }

    static class DeviceViewHolder extends RecyclerView.ViewHolder {
        ImageView deviceIcon;
        TextView deviceName,deviceIp;

        DeviceViewHolder(@NonNull View itemView) {
            super(itemView);
            deviceIcon = itemView.findViewById(R.id.image);
            deviceName = itemView.findViewById(R.id.devicename);
            deviceIp = itemView.findViewById(R.id.deviceip);
        }
    }

    public interface OnItemClickListener {
        void onItemClick(View view, int position);
    }
}
