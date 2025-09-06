package com.realtek.crossshare;


public class Device {


    private String name,ip;
    private int iconResId; // assume image resid is integer

    public Device(String name, String ip,int iconResId) {
        this.name = name;
        this.ip=ip;
        this.iconResId = iconResId;
    }

    public String getName() {
        return name;
    }

    public String getIp() {
        return ip;
    }

    public int getIconResId() {
        return iconResId;
    }

}
