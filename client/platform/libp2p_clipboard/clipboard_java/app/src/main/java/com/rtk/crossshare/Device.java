package com.rtk.crossshare;


public class Device {


    private String name;
    private int iconResId; // assume image resid is integer

    public Device(String name, int iconResId) {
        this.name = name;
        this.iconResId = iconResId;
    }

    public String getName() {
        return name;
    }

    public int getIconResId() {
        return iconResId;
    }

}
