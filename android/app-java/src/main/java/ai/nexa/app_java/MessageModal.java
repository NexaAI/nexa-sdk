package ai.nexa.app_java;

public class MessageModal {

    private String message;
    private String sender;
    private String imageUri;
    private long ttft;
    private double tps;
    private double decodingSpeed;
    private int totalTokens;

    public MessageModal(String message, String sender, String imageUri) {
        this.message = message;
        this.sender = sender;
        this.imageUri = imageUri;
        this.ttft = 0;
        this.tps = 0.0;
        this.decodingSpeed = 0.0;
        this.totalTokens = 0;
    }

    public String getMessage() {
        return message;
    }

    public void setMessage(String message) {
        this.message = message;
    }

    public String getSender() {
        return sender;
    }

    public void setSender(String sender) {
        this.sender = sender;
    }

    public String getImageUri() {
        return imageUri;
    }

    public void setImageUri(String imageUri) {
        this.imageUri = imageUri;
    }

    public long getTtft() {
        return ttft;
    }

    public void setTtft(long ttft) {
        this.ttft = ttft;
    }

    public double getTps() {
        return tps;
    }

    public void setTps(double tps) {
        this.tps = tps;
    }

    public double getDecodingSpeed() {
        return decodingSpeed;
    }

    public void setDecodingSpeed(double decodingSpeed) {
        this.decodingSpeed = decodingSpeed;
    }

    public int getTotalTokens() {
        return totalTokens;
    }

    public void setTotalTokens(int totalTokens) {
        this.totalTokens = totalTokens;
    }
}
