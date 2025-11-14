# !/usr/bin/env python3

import numpy as np
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
    
    def get_embedding(self, input: Union[str, List[str]]) -> np.ndarray:
        """
        Get embedding for text or image path(s).
        
        Args:
            input: Single text string, image path, or list of texts/image paths
            
        Returns:
            numpy array of embeddings. If single input, returns 1D array.
            If list input, returns 2D array with shape (len(input), embedding_dim).
        """
        # Ensure input is a list
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
            
            # Extract embeddings from response
            embeddings = [item.embedding for item in response.data]
            embeddings_array = np.array(embeddings)
            
            # Return 1D array for single input, 2D array for multiple inputs
            if single_input:
                return embeddings_array[0]
            return embeddings_array
            
        except Exception as e:
            raise RuntimeError(f"Failed to get embedding from Nexa API: {e}")

