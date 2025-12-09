package com.nexa.demo.adapter

import android.util.Log
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.recyclerview.widget.RecyclerView
import com.bumptech.glide.Glide
import com.nexa.demo.bean.EmbedFileBean
import com.nexa.demo.databinding.ItemSelectImageBinding
import com.nexa.demo.fragments.IndexFragment
import java.io.File
import kotlin.math.min

class SelectImagesAdapter(allImages: ArrayList<String>) :
    RecyclerView.Adapter<SelectImagesAdapter.MyViewHolder>() {

    private val data = arrayListOf<EmbedFileBean>()
    private val topKData = arrayListOf<EmbedFileBean>()
    private val allPercent = arrayListOf<Int>()
    private val topK = arrayListOf<Int>()
    private var showPercent = false

    init {
        data.addAll(allImages.map {
            EmbedFileBean(it)
        })
        topKData.addAll(data)
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
            View.GONE
        } else {
            View.GONE
        }
        holder.binding.tvPercent.text = "${topKData[position].percent}%"
        Glide.with(holder.binding.root).load(File(topKData[position].filePath))
            .into(holder.binding.ivImage)
    }

    override fun getItemCount(): Int {
        Log.d(TAG, "show images ${topKData.size}")
        return topKData.size
    }

    fun updateImages(allImages: ArrayList<String>) {
        showPercent = false
        data.clear()
        topKData.clear()
        data.addAll(allImages.map { EmbedFileBean(it) })
        topKData.addAll(data)
        notifyDataSetChanged()
    }

    fun updatePercent(allPercent: ArrayList<Int>) {
        showPercent = true
        this.allPercent.clear()
        this.allPercent.addAll(allPercent)
        data.forEachIndexed { index, bean ->
            bean.percent = allPercent[index]
        }
        topKData.clear()
        topKData.addAll(data)
        topKData.sortWith(Comparator { bean, bean1 -> if (bean.percent!! > bean1.percent!!) -1 else 1 })

        ArrayList(topKData.subList(0, min(IndexFragment.embedTopK, topKData.size))).let {
            topKData.clear()
            topKData.addAll(it)
        }
        notifyDataSetChanged()
    }

    class MyViewHolder(val binding: ItemSelectImageBinding) :
        RecyclerView.ViewHolder(binding.root) {
    }

    companion object {
        private const val TAG = "SelectImagesAdapter"
    }
}