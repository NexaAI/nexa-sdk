package com.nexa.demo.adapter

import android.view.LayoutInflater
import android.view.ViewGroup
import androidx.recyclerview.widget.RecyclerView
import com.nexa.demo.bean.IndexSelectedDirBean
import com.nexa.demo.databinding.ItemSelectFileBinding
import java.io.File

class ShowFileDirAdapter(val showDirs: List<IndexSelectedDirBean>) : RecyclerView.Adapter<ShowFileDirAdapter.MyViewHolder>() {

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
