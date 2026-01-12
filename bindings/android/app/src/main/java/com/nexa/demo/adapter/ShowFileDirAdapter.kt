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
