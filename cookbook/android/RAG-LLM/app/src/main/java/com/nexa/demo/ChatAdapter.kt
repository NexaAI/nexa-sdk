package com.nexa.demo

import android.content.Intent
import android.net.Uri
import android.text.TextUtils
import android.text.method.LinkMovementMethod
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.Button
import android.widget.ImageView
import android.widget.LinearLayout
import android.widget.TextView
import androidx.recyclerview.widget.RecyclerView
import com.nexa.demo.activity.FileContentActivity
import com.nexa.demo.bean.EmbedResultBean
import io.noties.markwon.Markwon
import io.noties.markwon.ext.strikethrough.StrikethroughPlugin
import io.noties.markwon.ext.tables.TablePlugin
import io.noties.markwon.linkify.LinkifyPlugin
import java.io.File


data class Message(
    val content: String,
    val type: MessageType,
    val images: List<File> = emptyList(),
    val audio: List<File> = emptyList(),
    val embedResultBean: EmbedResultBean? = null,  // Deprecated, keeping for compatibility
    val retrievedChunks: List<EmbedResultBean> = emptyList(),  // Store multiple retrieved chunks
    val formattedPrompt: String? = null  // Store the full LLM prompt
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

interface OnCitationsClick {
    fun onClick(position: Int, retrievedChunks: List<EmbedResultBean>)
}

class ChatAdapter(private val messages: List<Message>, private val onCitationsClick: OnCitationsClick) :
    RecyclerView.Adapter<RecyclerView.ViewHolder>() {

    override fun getItemViewType(position: Int): Int {
        val message = messages[position]
        return message.type.value
    }

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): RecyclerView.ViewHolder {
        val inflater = LayoutInflater.from(parent.context)
        val type = MessageType.from(viewType)
        return if (type == MessageType.USER) {
            UserViewHolder(inflater.inflate(R.layout.item_user_message, parent, false))
        } else if (type == MessageType.ASSISTANT) {
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
        if (holder is ProfileViewHolder) holder.bind(position, message, onCitationsClick)
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
        private val markwon: Markwon = Markwon.builder(itemView.context)
            .usePlugin(StrikethroughPlugin.create())
            .usePlugin(TablePlugin.create(itemView.context))
            .usePlugin(LinkifyPlugin.create())
            .build()

        fun bind(message: Message) {
            markwon.setMarkdown(tvMessage, message.content)
            tvMessage.movementMethod = LinkMovementMethod.getInstance()
        }
    }

    class ProfileViewHolder(itemView: View) : RecyclerView.ViewHolder(itemView) {
        private val tvMessage: TextView = itemView.findViewById(R.id.tv_message)
        private val btnCitations: Button = itemView.findViewById(R.id.btn_citations)
        private val btnViewPrompt: Button = itemView.findViewById(R.id.btn_view_prompt)
        
        fun bind(position: Int, message: Message, onCitationsClick: OnCitationsClick) {
            tvMessage.text = message.content
            
            // Handle multiple citations
            if (message.retrievedChunks.isNotEmpty()) {
                btnCitations.text = "${message.retrievedChunks.size} Citations"
                btnCitations.visibility = View.VISIBLE
                btnCitations.setOnClickListener {
                    onCitationsClick.onClick(position, message.retrievedChunks)
                }
            } else {
                btnCitations.visibility = View.GONE
            }
            
            // Handle view prompt button
            if (!TextUtils.isEmpty(message.formattedPrompt)) {
                btnViewPrompt.setOnClickListener {
                    it.context.startActivity(
                        Intent(
                            it.context,
                            FileContentActivity::class.java
                        ).apply {
                            this.putExtra(
                                FileContentActivity.KEY_PROMPT_CONTENT,
                                message.formattedPrompt
                            )
                        })
                }
                btnViewPrompt.visibility = View.VISIBLE
            } else {
                btnViewPrompt.visibility = View.GONE
            }
        }
        
        private fun dpToPx(dp: Int, context: android.content.Context): Int {
            return (dp * context.resources.displayMetrics.density).toInt()
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