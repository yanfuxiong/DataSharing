package com.realtek.crossshare;

import android.os.Bundle;
import android.util.Log;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;
import androidx.fragment.app.Fragment;
import androidx.fragment.app.FragmentManager;
import androidx.recyclerview.widget.DividerItemDecoration;
import androidx.recyclerview.widget.LinearLayoutManager;
import androidx.recyclerview.widget.RecyclerView;

import java.util.ArrayList;
import java.util.List;

public class LicenseListFragment extends Fragment {

    private RecyclerView recyclerView;
    @Override
    public View onCreateView(LayoutInflater inflater, ViewGroup container,
                             Bundle savedInstanceState) {
        return inflater.inflate(R.layout.fragment_license_list, container, false);

    }


    @Override
    public void onViewCreated(@NonNull View view, @Nullable Bundle savedInstanceState) {
        super.onViewCreated(view, savedInstanceState);
        recyclerView = view.findViewById(R.id.recycler_license);
        recyclerView.setLayoutManager(new LinearLayoutManager(getContext()));
        recyclerView.addItemDecoration(new DividerItemDecoration(requireContext(), DividerItemDecoration.VERTICAL));
        List<LicenseItem> licenseItems = new ArrayList<>();
        licenseItems.add(new LicenseItem("Appcompat", "appcompat.txt"));
        licenseItems.add(new LicenseItem("Constraintlayout", "constraintlayout.txt"));
        licenseItems.add(new LicenseItem("Gson", "gson.txt"));
        licenseItems.add(new LicenseItem("Lifecycle", "lifecycle.txt"));
        licenseItems.add(new LicenseItem("Material", "material.txt"));
        licenseItems.add(new LicenseItem("Mmkv", "mmkv.txt"));
        licenseItems.add(new LicenseItem("Zxing", "zxing.txt"));
        
        LicenseAdapter adapter = new LicenseAdapter(licenseItems, item -> {

            getParentFragmentManager()
                    .beginTransaction()
                    .replace(R.id.fragment_container, LicenseDetailFragment.newInstance(item.getName(), item.getAssetFile()))
                    .addToBackStack(null)
                    .commit();
        });
        recyclerView.setAdapter(adapter);

    }
}