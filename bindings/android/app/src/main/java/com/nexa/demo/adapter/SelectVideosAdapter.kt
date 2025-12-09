package com.nexa.demo.adapter

import android.view.LayoutInflater
import android.view.ViewGroup
import androidx.recyclerview.widget.RecyclerView
import com.nexa.demo.databinding.ItemSelectImageBinding

class SelectVideosAdapter : RecyclerView.Adapter<SelectVideosAdapter.MyViewHolder>() {
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

    }

    override fun getItemCount(): Int {
        return 20
    }

    class MyViewHolder(binding: ItemSelectImageBinding) : RecyclerView.ViewHolder(binding.root) {

    }
}