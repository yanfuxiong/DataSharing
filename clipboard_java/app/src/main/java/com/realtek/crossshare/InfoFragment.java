package com.realtek.crossshare;

import android.os.Bundle;
import android.util.Log;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.TextView;

import androidx.fragment.app.Fragment;
import androidx.fragment.app.FragmentManager;

public class InfoFragment extends Fragment {
    private static final String TAG = "InfoFragment";

    @Override
    public View onCreateView(LayoutInflater inflater, ViewGroup container,
                             Bundle savedInstanceState) {
        View view = inflater.inflate(R.layout.layout_fm_info, container, false);

        View licenseRow = view.findViewById(R.id.layout_license);
        TextView tv_ip_value = view.findViewById(R.id.tv_ip_value);
        TextView tv_device_value = view.findViewById(R.id.tv_device_value);
        TextView tv_version_value = view.findViewById(R.id.tv_version_value);
        licenseRow.setOnClickListener(v -> {
            FragmentManager fm = requireActivity().getSupportFragmentManager();
            fm.beginTransaction()
                    .hide(this)
                    .add(R.id.fragment_container, new LicenseListFragment(), "license_list_fragment")
                    .addToBackStack(null)
                    .commit();
        });

        String ipdevice = null;
        String devicename = null;
        String softwareinfo = null;
        if (getActivity() instanceof TestActivity) {
            ipdevice = ((TestActivity) getActivity()).getMyIp();
            devicename = ((TestActivity) getActivity()).DeviceName();
            softwareinfo = ((TestActivity) getActivity()).getSoftwareInfo();
            Log.i(TAG, "ipdevice=" + ipdevice + " devicename=" + devicename + " softwareinfo=" + softwareinfo);
            tv_ip_value.setText(ipdevice);
            tv_device_value.setText(devicename);
            tv_version_value.setText(softwareinfo);
        }
        return view;
    }


}