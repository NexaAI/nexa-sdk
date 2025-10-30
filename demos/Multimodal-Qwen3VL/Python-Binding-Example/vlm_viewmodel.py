#!/usr/bin/env python3

from dataclasses import dataclass
from typing import Optional, List

from vlm_service import VLMService

@dataclass
class ModelInfo:
    repo_id: str
    plugin_ids: List[str]
    model_path: str
    mmproj_path: str
    devices: List[str]

class VLMViewModel:
    """ViewModel wrapper around VLMService for UI layers (e.g. Gradio).
    """

    def __init__(self) -> None:
        self._vlm: Optional[VLMService] = None

    def reset(self) -> None:
        """Reset and dispose the current VLMService instance."""
        if self._vlm is not None:
            self._vlm.reset()
            self._vlm = None

    def create(
        self,
        repo_id: str,
        plugin_id: str,
        device: str,
        system_prompt: str,
    ):
        """Lazily create (or return existing) VLMService instance.

        Returns the active VLMService instance.
        """
        model_info = self._model_of_repo(repo_id=repo_id)
        
        if self._vlm is None:
            self._vlm = VLMService(
                model_name=model_info.model_path,
                mmproj_name=model_info.mmproj_path,
                plugin_id=plugin_id,
                device=device,
                system_prompt=system_prompt,
            )

    def stream_response(
        self,
        prompt: str,
        images: Optional[List[str]] = None,
        audios: Optional[List[str]] = None,
        max_tokens: int = 200,
        temperature: float = 0.7,
        top_p: float = 0.9
    ):
        yield from self._vlm.stream_response(
            prompt=prompt, 
            images=images, 
            audios=audios, 
            max_tokens=max_tokens, 
            temperature=temperature,top_p=top_p
        )
        
    def _model_of_repo(self, repo_id: str) -> ModelInfo:
        return next((model for model in self.models if model.repo_id == repo_id), None)
        
    def update_system_prompt(self, system_prompt: str) -> None:
        """Update the system prompt of the active VLM and reset conversation."""
        if self._vlm is not None:
            self._vlm.update_system_prompt(system_prompt)
            
    @property
    def models(self) -> List[ModelInfo]:
        """All models"""
        return [
            ModelInfo(
                repo_id="NexaAI/Qwen3-VL-4B-Instruct-GGUF",
                plugin_ids=["nexaml", "cpu_gpu"],
                model_path="NexaAI/Qwen3-VL-4B-Instruct-GGUF/Qwen3-VL-4B-Instruct.Q4_0.gguf",
                mmproj_path="NexaAI/Qwen3-VL-4B-Instruct-GGUF/mmproj.F32.gguf",
                devices=["gpu", "cpu"]
            )
        ]

