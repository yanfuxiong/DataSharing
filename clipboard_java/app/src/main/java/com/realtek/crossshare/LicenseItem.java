package com.realtek.crossshare;


public class LicenseItem {

    private String name;
    private String assetFile; // e.g. "mbprogresshud_license.txt"

    public LicenseItem(String name, String assetFile) {
        this.name = name;
        this.assetFile = assetFile;
    }
    public String getName() { return name; }
    public String getAssetFile() { return assetFile; }

}
