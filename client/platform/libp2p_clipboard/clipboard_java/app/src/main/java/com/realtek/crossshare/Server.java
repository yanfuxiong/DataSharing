package com.realtek.crossshare;


public class Server {

    public String monitorName;
    public String instance;
    public String ipAddr;

    public Server(String monitorName, String instance, String ipAddr) {
        this.monitorName = monitorName;
        this.instance = instance;
        this.ipAddr = ipAddr;
    }
    public String getmonitorName() { return monitorName; }
    public String getinstance() { return instance; }
    public String getipAddr() { return ipAddr; }
}
