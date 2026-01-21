package com.nexa.demo

/**
 * Configuration for RAG (Retrieval-Augmented Generation) system
 */
object RAGConfig {
    const val DEFAULT_MAX_CHUNK_SIZE = 512
    const val DEFAULT_MIN_CHUNK_SIZE = 16
    // Chunk size in words
    var CHUNK_SIZE = 128
    
    // Number of chunks to retrieve for each query
    var N_CHUNKS = 6
}