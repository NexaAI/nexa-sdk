#include "ml.h"

/**
 * @brief Get error message string for error code
 * 
 * This function maps error codes to human-readable error messages.
 * Error codes follow hierarchical naming: ML_ERROR_[CATEGORY]_[SUBCATEGORY]_[ERROR_TYPE]
 * 
 * @param error_code Error code from ml_ErrorCode enumeration
 * @return Error message string (const char*)
 */
const char* ml_get_error_message(ml_ErrorCode error_code) {
    switch (error_code) {
        case ML_SUCCESS:
            return "Success";
            
        /* ====================================================================== */
        /*                              COMMON ERRORS (100xxx)                     */
        /* ====================================================================== */
        
        /* General errors */
        case ML_ERROR_COMMON_UNKNOWN:
            return "Unknown error";
        case ML_ERROR_COMMON_INVALID_INPUT:
            return "Invalid input parameters or handle";
        case ML_ERROR_COMMON_MEMORY_ALLOCATION:
            return "Memory allocation failed";
        case ML_ERROR_COMMON_FILE_NOT_FOUND:
            return "File not found or inaccessible";
        case ML_ERROR_COMMON_NOT_INITIALIZED:
            return "Library not initialized";
        case ML_ERROR_COMMON_NOT_SUPPORTED:
            return "Operation not supported";
        case ML_ERROR_COMMON_MODEL_LOAD:
            return "Model loading failed";
        case ML_ERROR_COMMON_EMBEDDING_GENERATION:
            return "Embedding generation failed";
        case ML_ERROR_COMMON_EMBEDDING_DIMENSION:
            return "Invalid embedding dimension";
        case ML_ERROR_COMMON_RERANK_FAILED:
            return "Reranking failed";
        case ML_ERROR_COMMON_IMG_GENERATION:
            return "Image generation failed";
        case ML_ERROR_COMMON_IMG_PROMPT:
            return "Invalid image prompt";
            
        /* ====================================================================== */
        /*                              LLM ERRORS (200xxx)                        */
        /* ====================================================================== */
        
        /* Tokenization errors */
        case ML_ERROR_LLM_TOKENIZATION_FAILED:
            return "Tokenization failed";
        case ML_ERROR_LLM_TOKENIZATION_CONTEXT_LENGTH:
            return "Context length exceeded";
            
        /* Generation errors */
        case ML_ERROR_LLM_GENERATION_FAILED:
            return "Text generation failed";
        case ML_ERROR_LLM_GENERATION_PROMPT_TOO_LONG:
            return "Input prompt too long";
            
        /* ====================================================================== */
        /*                              VLM ERRORS (300xxx)                        */
        /* ====================================================================== */
        
        /* Image processing errors */
        case ML_ERROR_VLM_IMAGE_LOAD:
            return "Image loading failed";
        case ML_ERROR_VLM_IMAGE_FORMAT:
            return "Unsupported image format";
            
        /* Audio processing errors */
        case ML_ERROR_VLM_AUDIO_LOAD:
            return "Audio loading failed";
        case ML_ERROR_VLM_AUDIO_FORMAT:
            return "Unsupported audio format";
            
        /* Generation errors */
        case ML_ERROR_VLM_GENERATION_FAILED:
            return "Multimodal generation failed";
            
        /* ====================================================================== */
        /*                              OCR ERRORS (400xxx)                        */
        /* ====================================================================== */
        
        case ML_ERROR_OCR_DETECTION:
            return "OCR text detection failed";
        case ML_ERROR_OCR_RECOGNITION:
            return "OCR text recognition failed";
        case ML_ERROR_OCR_MODEL:
            return "OCR model error";
            
        /* ====================================================================== */
        /*                              ASR ERRORS (500xxx)                        */
        /* ====================================================================== */
        
        case ML_ERROR_ASR_TRANSCRIPTION:
            return "ASR transcription failed";
        case ML_ERROR_ASR_AUDIO_FORMAT:
            return "Unsupported ASR audio format";
        case ML_ERROR_ASR_LANGUAGE:
            return "Unsupported ASR language";
            
        /* ====================================================================== */
        /*                              TTS ERRORS (600xxx)                        */
        /* ====================================================================== */
        
        case ML_ERROR_TTS_SYNTHESIS:
            return "TTS synthesis failed";
        case ML_ERROR_TTS_VOICE:
            return "TTS voice not found";
        case ML_ERROR_TTS_AUDIO_FORMAT:
            return "TTS audio format error";
              
        default:
            return "Unknown error code";
    }
} 