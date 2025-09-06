package com.realtek.crossshare;

import android.content.Context;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.TextView;
import androidx.recyclerview.widget.RecyclerView;
import java.util.List;

public class ServerAdapter extends RecyclerView.Adapter<ServerAdapter.ViewHolder> {

    private Context context;
    private List<Server> serverList;
    private OnClickListener listener;

    public ServerAdapter(Context context, List<Server> serverList,OnClickListener listener) {
        this.context = context;
        this.serverList = serverList;
        this.listener = listener;
    }

    @Override
    public ViewHolder onCreateViewHolder(ViewGroup parent, int viewType) {
        View view = LayoutInflater.from(context).inflate(R.layout.item_server, parent, false);
        return new ViewHolder(view);
    }

    @Override
    public void onBindViewHolder(ViewHolder holder, int position) {
        Server server = serverList.get(position);
        holder.name.setText(server.getmonitorName());
        holder.ip.setText(server.getipAddr());

        holder.itemView.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                if (listener != null) {
                    listener.onServerClick(server.getmonitorName(),server.getinstance(),server.getipAddr());
                }
            }
        });
    }

    @Override
    public int getItemCount() {
        return serverList.size();
    }

    static class ViewHolder extends RecyclerView.ViewHolder {
        TextView name, ip;
        ViewHolder(View itemView) {
            super(itemView);
            name = itemView.findViewById(R.id.server_name);
            ip = itemView.findViewById(R.id.server_ip);
        }
    }
}
