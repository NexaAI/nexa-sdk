package com.nexa.demo.adapter

import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.recyclerview.widget.RecyclerView
import com.bumptech.glide.Glide
import com.nexa.demo.databinding.ItemSelectImageBinding
import java.io.File

class SelectImagesAdapter(allImages: ArrayList<String>) :
    RecyclerView.Adapter<SelectImagesAdapter.MyViewHolder>() {

    private val data = arrayListOf<String>()
    private val allPercent = arrayListOf<Int>()
    private var showPercent = false

    init {
        data.addAll(allImages)
    }

    override fun onCreateViewHolder(
        parent: ViewGroup,
        viewType: Int
    ): MyViewHolder {
        return MyViewHolder(
            ItemSelectImageBinding.inflate(
                LayoutInflater.from(parent.context), parent,
                false
            )
        )
    }

    override fun onBindViewHolder(
        holder: MyViewHolder,
        position: Int
    ) {
        holder.binding.cvPercent.visibility = if (showPercent) {
            View.VISIBLE
        } else {
            View.GONE
        }
        if (position < allPercent.size) {
            holder.binding.tvPercent.text = "${allPercent[position]}%"
        }
        Glide.with(holder.binding.root).load(File(data[position]))
            .into(holder.binding.ivImage)
    }

    override fun getItemCount(): Int {
        return data.size
    }

    fun updateImages(allImages: ArrayList<String>) {
        showPercent = false
        data.clear()
        data.addAll(allImages)
        notifyDataSetChanged()
    }

    fun updatePercent(allPercent: ArrayList<Int>) {
        showPercent = true
        this.allPercent.clear()
        this.allPercent.addAll(allPercent)
        notifyDataSetChanged()
    }

    class MyViewHolder(val binding: ItemSelectImageBinding) :
        RecyclerView.ViewHolder(binding.root) {
    }
}