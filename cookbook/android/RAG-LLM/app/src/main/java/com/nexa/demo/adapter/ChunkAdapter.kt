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

import android.content.Intent
import android.view.LayoutInflater
import android.view.ViewGroup
import androidx.recyclerview.widget.RecyclerView
import com.nexa.demo.activity.FileContentActivity
import com.nexa.demo.bean.EmbedResultBean
import com.nexa.demo.databinding.ItemCitationBinding
import java.io.File

class ChunkAdapter : RecyclerView.Adapter<ChunkAdapter.MyViewHolder>() {

    private val data = mutableListOf<EmbedResultBean>()

    fun updateData(retrievedChunks: List<EmbedResultBean>) {
        data.clear()
        data.addAll(retrievedChunks)
        notifyDataSetChanged()
    }

    override fun onCreateViewHolder(
        parent: ViewGroup,
        viewType: Int
    ): MyViewHolder {
        return MyViewHolder(
            ItemCitationBinding.inflate(
                LayoutInflater.from(parent.context),
                parent,
                false
            )
        )
    }

    override fun onBindViewHolder(
        holder: MyViewHolder,
        position: Int
    ) {
        holder.binding.tvPosition.text = "${position + 1}"
        val fileName = File(data[position].path).name
        holder.binding.tvFileName.text = fileName
        holder.binding.btnChunk.text = "Chunk ${data[position].chunkIndex + 1}"
        val chunkTxt = data[position].txt ?: ""
        holder.binding.btnChunk.setOnClickListener {
            holder.binding.btnChunk.context.startActivity(
                Intent(
                    it.context,
                    FileContentActivity::class.java
                ).apply {
                    putExtra(
                        FileContentActivity.KEY_PROMPT_CONTENT,
                        chunkTxt
                    )
                    putExtra(
                        FileContentActivity.KEY_TITLE,
                        fileName
                    )
                })
        }
    }

    override fun getItemCount(): Int {
        return data.size
    }

    class MyViewHolder(val binding: ItemCitationBinding) :
        RecyclerView.ViewHolder(binding.root)
}