package ai.nexa.app_java;

import android.content.Context;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.ImageView;
import android.widget.TextView;

import androidx.annotation.NonNull;
import androidx.recyclerview.widget.RecyclerView;

import com.bumptech.glide.Glide;

import java.util.ArrayList;

public class MessageRVAdapter extends RecyclerView.Adapter {

    private ArrayList<MessageModal> messageModalArrayList;
    private Context context;

    public MessageRVAdapter(ArrayList<MessageModal> messageModalArrayList, Context context) {
        this.messageModalArrayList = messageModalArrayList;
        this.context = context;
    }

    @NonNull
    @Override
    public RecyclerView.ViewHolder onCreateViewHolder(@NonNull ViewGroup parent, int viewType) {
        View view;
        switch (viewType) {
            case 0:
                view = LayoutInflater.from(parent.getContext()).inflate(R.layout.user_msg, parent, false);
                return new UserViewHolder(view);
            case 1:
                view = LayoutInflater.from(parent.getContext()).inflate(R.layout.bot_msg, parent, false);
                return new BotViewHolder(view);
        }
        return null;
    }

    @Override
    public void onBindViewHolder(@NonNull RecyclerView.ViewHolder holder, int position) {
        MessageModal modal = messageModalArrayList.get(position);
        switch (modal.getSender()) {
            case "user":
                UserViewHolder userHolder = (UserViewHolder) holder;
                if (modal.getImageUri() != null && !modal.getImageUri().isEmpty()) {
                    userHolder.userImage.setVisibility(View.VISIBLE);
                    userHolder.userTV.setVisibility(View.GONE);
                    Glide.with(userHolder.itemView.getContext())
                            .load(modal.getImageUri())
                            .into(userHolder.userImage);
                } else {
                    userHolder.userImage.setVisibility(View.GONE);
                    userHolder.userTV.setVisibility(View.VISIBLE);
                    userHolder.userTV.setText(modal.getMessage());
                }
                break;
            case "bot":
                BotViewHolder botViewHolder = (BotViewHolder) holder;
                botViewHolder.botTV.setText(modal.getMessage());
                if (modal.getTtft() > 0) {
                    double ttftInSeconds = modal.getTtft() / 1000.0;
                    String metrics = String.format(
                            "Total Tokens: %d\n" +
                                    "TTFT: %.2fs\n" +
                                    "TPS: %.2f tokens/s\n" +
                                    "Decoding: %.2f tokens/s",
                            modal.getTotalTokens(),
                            ttftInSeconds,
                            modal.getTps(),
                            modal.getDecodingSpeed());
                    botViewHolder.metricsTV.setText(metrics);
                    botViewHolder.metricsTV.setVisibility(View.VISIBLE);
                } else {
                    botViewHolder.metricsTV.setVisibility(View.GONE);
                }
                break;
        }
    }

    @Override
    public int getItemCount() {
        return messageModalArrayList.size();
    }

    @Override
    public int getItemViewType(int position) {
        switch (messageModalArrayList.get(position).getSender()) {
            case "user":
                return 0;
            case "bot":
                return 1;
            default:
                return -1;
        }
    }

    public static class UserViewHolder extends RecyclerView.ViewHolder {
        TextView userTV;
        ImageView userImage;

        public UserViewHolder(@NonNull View itemView) {
            super(itemView);
            userTV = itemView.findViewById(R.id.idTVUser);
            userImage = itemView.findViewById(R.id.idIVUserImage);
        }
    }

    public static class BotViewHolder extends RecyclerView.ViewHolder {
        TextView botTV;
        TextView metricsTV;

        public BotViewHolder(@NonNull View itemView) {
            super(itemView);
            botTV = itemView.findViewById(R.id.idTVBot);
            metricsTV = itemView.findViewById(R.id.idTVMetrics);
        }
    }
}
