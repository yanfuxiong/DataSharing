package com.realtek.crossshare;
import android.content.res.AssetManager;
import android.os.Bundle;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.TextView;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;
import androidx.fragment.app.Fragment;
import androidx.recyclerview.widget.LinearLayoutManager;
import androidx.recyclerview.widget.RecyclerView;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.util.ArrayList;
import java.util.List;

public class LicenseDetailFragment extends Fragment {

    private static final String ARG_LIB_NAME = "lib_name";
    private static final String ARG_ASSET_FILE = "asset_file";

    public static LicenseDetailFragment newInstance(String name, String assetFile) {
        Bundle args = new Bundle();
        args.putString(ARG_LIB_NAME, name);
        args.putString(ARG_ASSET_FILE, assetFile);
        LicenseDetailFragment fragment = new LicenseDetailFragment();
        fragment.setArguments(args);
        return fragment;
    }

    @Override
    public View onCreateView(LayoutInflater inflater, ViewGroup container,
                             Bundle savedInstanceState) {
        return inflater.inflate(R.layout.fragment_license_detail, container, false);

    }


    @Override
    public void onViewCreated(@NonNull View view, @Nullable Bundle savedInstanceState) {
        super.onViewCreated(view, savedInstanceState);
        TextView tvLibName = view.findViewById(R.id.tv_license_lib_name);
        TextView tvLicense = view.findViewById(R.id.tv_license_detail);

        Bundle args = getArguments();
        String libName = args != null ? args.getString(ARG_LIB_NAME, "") : "";
        /*
        license file path:
        /app/src/main/assets//mbprogresshud_license.txt
         */
        String assetFile = args != null ? args.getString(ARG_ASSET_FILE, "") : "";

        //tvLibName.setText(libName);
        TextView toolbarTitle = getActivity().findViewById(R.id.toolbar_title);
        if (toolbarTitle != null) {
            toolbarTitle.setText(libName);
        }
        String content = readAssetFile(assetFile);
        tvLicense.setText(content != null ? content : "License file not found.");
    }

    private String readAssetFile(String assetFile) {
        if (getContext() == null) return null;
        try {
            AssetManager assetManager = getContext().getAssets();
            InputStream is = assetManager.open(assetFile);
            BufferedReader reader = new BufferedReader(new InputStreamReader(is));
            StringBuilder builder = new StringBuilder();
            String line;
            while ((line = reader.readLine()) != null) {
                builder.append(line).append("\n");
            }
            reader.close();
            return builder.toString();
        } catch (IOException e) {
            e.printStackTrace();
            return null;
        }
    }
}