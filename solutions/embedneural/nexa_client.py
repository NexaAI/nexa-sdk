# Copyright 2024-2026 Nexa AI, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# !/usr/bin/env python3

from typing import Union, List
from openai import OpenAI


class NexaClient:
    """Client for Nexa serve OpenAI-compatible API."""
    
    def __init__(self, base_url: str = "http://localhost:18181", model: str = "NexaAI/EmbedNeural"):
        """
        Initialize Nexa client.
        
        Args:
            base_url: Base URL of nexa serve API (should include /v1 if needed)
            model: Model name to use for embeddings
        """
        # Ensure base_url ends with /v1 for OpenAI-compatible API
        if not base_url.endswith('/v1'):
            if base_url.endswith('/'):
                base_url = base_url + 'v1'
            else:
                base_url = base_url + '/v1'
        self.client = OpenAI(base_url=base_url, api_key="not-needed")
        self.model = model
    
    def get_embedding(self, input: Union[str, List[str]]):
        """
        Get embedding for text or image path(s).
        
        Args:
            input: Single text string, image path, or list of texts/image paths
            
        Returns:
            list of embeddings. If single input, returns 1D list.
            If list input, returns 2D list with shape (len(input), embedding_dim).
        """
        if isinstance(input, str):
            input_list = [input]
            single_input = True
        else:
            input_list = input
            single_input = False
        
        try:
            response = self.client.embeddings.create(
                model=self.model,
                input=input_list,
                encoding_format="float"
            )
            
            embeddings = [item.embedding for item in response.data]
            
            if single_input:
                return embeddings[0]
            return embeddings
            
        except Exception as e:
            raise RuntimeError(f"Failed to get embedding from Nexa API: {e}")

