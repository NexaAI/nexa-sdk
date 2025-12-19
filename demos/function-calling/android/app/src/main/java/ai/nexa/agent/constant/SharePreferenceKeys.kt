package ai.nexa.agent.constant

class SharePreferenceKeys {
    companion object {
        const val KEY_LAST_LOAD_MODEL_ID = "last_load_model_id"
        /**
         * 当 navigation pop 时，由于会加载初始的 modelId，所以不能直接加载 model，
         *  应该让 MainChatScreen 去加载
         */
        const val KEY_PREPARE_LOAD_MODEL_ID = "prepare_load_model_id"
        const val KEY_SHOW_OPERATION_POPUP = "show_operation_popup"
        const val KEY_S3_URL = "key_s3_url"
    }
    enum class FileName(val fileName: String) {
        ModelDownloadState("mode_download_state"),
        CommonConfig("common_config"),
        ModelS3Url("model_s3_url")
    }
}