package com.rtk.myapplication;


public class Device {


    private String name;
    private int iconResId; // 假设图标资源ID是整数

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
