package com.nexa.demo

import android.net.Uri
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.ImageView
import android.widget.LinearLayout
import android.widget.TextView
import androidx.recyclerview.widget.RecyclerView
import java.io.File


data class Message(
    val content: String,
    val type: MessageType,
    val images: List<File> = emptyList(),
    val audio: List<File> = emptyList()
)

enum class MessageType(val value: Int) {
    USER(0),
    ASSISTANT(1),
    PROFILE(2),
    IMAGES(3);

    companion object {
        fun from(value: Int): MessageType =
            entries.firstOrNull { it.value == value } ?: PROFILE
    }
}
class ChatAdapter(private val messages: List<Message>) :
    RecyclerView.Adapter<RecyclerView.ViewHolder>() {

    override fun getItemViewType(position: Int): Int  {
        val message = messages[position]
        return message.type.value
    }

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): RecyclerView.ViewHolder {
        val inflater = LayoutInflater.from(parent.context)
        val type = MessageType.from(viewType)
        return if (type == MessageType.USER) {
            UserViewHolder(inflater.inflate(R.layout.item_user_message, parent, false))
        } else if (type == MessageType.ASSISTANT){
            AiViewHolder(inflater.inflate(R.layout.item_ai_message, parent, false))
        } else if (type == MessageType.IMAGES) {
            ImagesViewHolder(inflater.inflate(R.layout.item_image_message, parent, false))
        } else {
            ProfileViewHolder(inflater.inflate(R.layout.item_profile_message, parent, false))
        }
    }

    override fun onBindViewHolder(holder: RecyclerView.ViewHolder, position: Int) {
        val message = messages[position]
        if (holder is UserViewHolder) holder.bind(message)
        if (holder is AiViewHolder) holder.bind(message)
        if (holder is ImagesViewHolder) holder.bind(message)
        if (holder is ProfileViewHolder) holder.bind(message)
    }

    override fun getItemCount() = messages.size

    class UserViewHolder(itemView: View) : RecyclerView.ViewHolder(itemView) {
        private val tvMessage: TextView = itemView.findViewById(R.id.tv_message)
        fun bind(message: Message) {
            tvMessage.text = message.content
        }
    }

    class AiViewHolder(itemView: View) : RecyclerView.ViewHolder(itemView) {
        private val tvMessage: TextView = itemView.findViewById(R.id.tv_message)
        fun bind(message: Message) {
            tvMessage.text = message.content
        }
    }

    class ProfileViewHolder(itemView: View) : RecyclerView.ViewHolder(itemView) {
        private val tvMessage: TextView = itemView.findViewById(R.id.tv_message)
        fun bind(message: Message) {
            tvMessage.text = message.content
        }
    }

    class ImagesViewHolder(itemView: View) : RecyclerView.ViewHolder(itemView) {
        private val imageContainer: LinearLayout = itemView.findViewById(R.id.image_container)
        fun bind(message: Message) {
            val savedImageFiles = message.images
            imageContainer.removeAllViews()
            val context = itemView.context

            for (file in savedImageFiles) {
                val itemView = LayoutInflater.from(context)
                    .inflate(R.layout.item_image_item_message, imageContainer, false)
                val ivImage = itemView.findViewById<ImageView>(R.id.iv_image)
                ivImage.setImageURI(Uri.fromFile(file))
                imageContainer.addView(itemView)
            }
        }
    }
}