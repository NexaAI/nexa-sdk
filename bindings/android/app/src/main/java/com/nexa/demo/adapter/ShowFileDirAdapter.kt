package com.nexa.demo.adapter

import android.util.Log
import android.view.LayoutInflater
import android.view.ViewGroup
import androidx.recyclerview.widget.RecyclerView
import com.nexa.demo.bean.IndexSelectedDirBean
import com.nexa.demo.databinding.ItemSelectFileBinding
import java.io.File

class ShowFileDirAdapter : RecyclerView.Adapter<ShowFileDirAdapter.MyViewHolder>() {

    private val rootDir = File("/sdcard")
    private val showDirs =
        rootDir.listFiles()!!.filter { file -> file.isDirectory && !file.name.startsWith(".") }
            .sorted()
            .map {
                IndexSelectedDirBean(it)
            }

    override fun onCreateViewHolder(
        parent: ViewGroup,
        viewType: Int
    ): MyViewHolder {
        return MyViewHolder(
            ItemSelectFileBinding.inflate(
                LayoutInflater.from(parent.context), parent,
                false
            )
        )
    }

    override fun onBindViewHolder(
        holder: MyViewHolder,
        position: Int
    ) {
        if (showDirs[position].subImageFiles == null || showDirs[position].subVideoFiles == null) {
            showDirs[position].subImageFiles = arrayListOf()
            showDirs[position].subVideoFiles = arrayListOf()
            showDirs[position].dir.listFiles()?.forEach {subFile->
                if (subFile.name.endsWith("jpg", true) ||
                    subFile.name.endsWith("jpeg", true) ||
                    subFile.name.endsWith("png", true)
                ) {
                    showDirs[position].subImageFiles!!.add(subFile.absolutePath)
                } else if (subFile.name.endsWith("mp4", true)) {
                    showDirs[position].subVideoFiles!!.add(subFile.absolutePath)
                }
            }
        }

        holder.binding.cbSelected.isChecked = showDirs[position].isSelected
        holder.binding.tvDirName.text = showDirs[position].dir.name.toString()
        holder.binding.tvFileCount.text = "${showDirs[position].subImageFiles?.size ?: "0"} items"
        holder.binding.cbSelected.setOnClickListener {
            showDirs[position].isSelected = holder.binding.cbSelected.isChecked
        }
    }

    override fun getItemCount(): Int {
        return showDirs.size
    }

    fun getSelectedImages(): ArrayList<String> {
        val allSelectedImages = arrayListOf<String>()
        showDirs.filter { it.isSelected }.forEach {
            allSelectedImages.addAll(it.subImageFiles!!.toTypedArray())
        }
        return allSelectedImages
    }

    class MyViewHolder(val binding: ItemSelectFileBinding) :
        RecyclerView.ViewHolder(binding.root) {}
}
