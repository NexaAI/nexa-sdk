// Copyright 2024-2026 Nexa AI, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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