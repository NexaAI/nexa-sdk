package com.nexa.demo.adapter

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
        holder.binding.cbSelected.isChecked = showDirs[position].isSelected
        holder.binding.tvDirName.text = showDirs[position].dir.name.toString()
        holder.binding.tvFileCount.text = "${showDirs[position].dir.listFiles()?.size ?: "0"} items"
        holder.binding.cbSelected.setOnClickListener {
            showDirs[position].isSelected = holder.binding.cbSelected.isChecked
        }
    }

    override fun getItemCount(): Int {
        return showDirs.size
    }

    fun getSelectedDirs(): ArrayList<String> {
        return showDirs.filter { it.isSelected }.map {
            it.dir.absolutePath
        } as ArrayList<String>
    }

    class MyViewHolder(val binding: ItemSelectFileBinding) :
        RecyclerView.ViewHolder(binding.root) {}
}