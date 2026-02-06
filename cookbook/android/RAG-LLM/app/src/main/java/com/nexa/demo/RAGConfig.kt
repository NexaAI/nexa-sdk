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