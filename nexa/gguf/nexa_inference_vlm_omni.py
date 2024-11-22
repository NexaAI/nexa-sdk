import ctypes
import logging
import os
import sys
from pathlib import Path
from streamlit.web import cli as stcli
from nexa.utils import nexa_prompt, SpinningCursorAnimation
from nexa.constants import (
    NEXA_RUN_OMNI_VLM_PROJECTOR_MAP,
    NEXA_RUN_OMNI_VLM_MAP
)
from nexa.gguf.lib_utils import is_gpu_available
from nexa.gguf.llama import omni_vlm_cpp
from nexa.gguf.llama._utils_transformers import suppress_stdout_stderr
from nexa.general import pull_model

class NexaOmniVlmInference:
    """
    A class used for vision language model inference.
    """

    def __init__(
        self,
        model_path=None,
        local_path=None,
        projector_local_path=None,
        device="auto",
        **kwargs,
    ):
        if model_path is None and local_path is None:
            raise ValueError("Either model_path or local_path must be provided.")

        self.model = None
        self.projector = None
        self.device = device
        self.context = None
        self.omni_vlm_version = "vlm-81-instruct"
        if self.device == "auto" or self.device == "gpu":
            self.n_gpu_layers = -1 if is_gpu_available() else 0
        else:
            self.n_gpu_layers = 0

        # Handle direct model file paths (e.g., omnivision:model-fp16)
        if model_path and ':model-' in model_path:
            base_name = model_path.split(':')[0]
            model_type = model_path.split('model-')[1]
            if base_name in NEXA_RUN_OMNI_VLM_PROJECTOR_MAP:
                self.model_path = model_path
                # Construct corresponding projector path
                self.projector_path = f"{base_name}:projector-{model_type}"
                self.downloaded_path, _ = pull_model(self.model_path, **kwargs)
                self.projector_downloaded_path, _ = pull_model(self.projector_path, **kwargs)
                self.omni_vlm_version = self._determine_vlm_version(model_path)

        else:
            # Handle other path formats and model loading scenarios
            self.projector_path = NEXA_RUN_OMNI_VLM_PROJECTOR_MAP.get(model_path, None)
            self.downloaded_path = local_path
            self.projector_downloaded_path = projector_local_path

            if self.downloaded_path is not None and self.projector_downloaded_path is not None:
                # when running from local, both path should be provided
                self.omni_vlm_version = self._determine_vlm_version(str(self.downloaded_path))
            elif self.downloaded_path is not None:
                if model_path in NEXA_RUN_OMNI_VLM_MAP:
                    self.projector_path = NEXA_RUN_OMNI_VLM_PROJECTOR_MAP[model_path]
                    self.projector_downloaded_path, _ = pull_model(self.projector_path, **kwargs)
                    self.omni_vlm_version = self._determine_vlm_version(model_path)
            elif model_path in NEXA_RUN_OMNI_VLM_MAP:
                self.model_path = NEXA_RUN_OMNI_VLM_MAP[model_path]
                self.projector_path = NEXA_RUN_OMNI_VLM_PROJECTOR_MAP[model_path]
                self.downloaded_path, _ = pull_model(self.model_path, **kwargs)
                self.projector_downloaded_path, _ = pull_model(self.projector_path, **kwargs)
                self.omni_vlm_version = self._determine_vlm_version(model_path)
            elif Path(model_path).parent.exists():
                local_dir = Path(model_path).parent
                model_name = Path(model_path).name
                tag_and_ext = model_name.split(":")[-1]
                self.downloaded_path = local_dir / f"model-{tag_and_ext}"
                self.projector_downloaded_path = local_dir / f"projector-{tag_and_ext}"
                if not (
                    self.downloaded_path.exists()
                    and self.projector_downloaded_path.exists()
                ):
                    logging.error(
                        f"Model or projector not found in {local_dir}. "
                        "Make sure to name them as 'model-<tag>.gguf' and 'projector-<tag>.gguf'."
                    )
                    exit(1)
                self.omni_vlm_version = self._determine_vlm_version(model_path)
            else:
                logging.error("VLM user model from hub is not supported yet.")
                exit(1)
        
        # Override version if specified in kwargs
        if 'omni_vlm_version' in kwargs:
            self.omni_vlm_version = kwargs.get('omni_vlm_version')
        print(f"Using omni-vlm-version: {self.omni_vlm_version}")
            
        with suppress_stdout_stderr():
            self._load_model()

    def _determine_vlm_version(self, path_str: str) -> str:
        """Helper function to determine VLM version from path string."""
        if 'ocr' in path_str:
            return "vlm-81-ocr"
        elif 'preview' in path_str:
            return "nano-vlm-instruct"
        return "vlm-81-instruct"
    
    @SpinningCursorAnimation()
    def _load_model(self):
        try:
            self.ctx_params_model = ctypes.c_char_p(self.downloaded_path.encode("utf-8"))
            self.ctx_params_mmproj = ctypes.c_char_p(self.projector_downloaded_path.encode("utf-8"))
            self.ctx_params_omni_vlm_version = ctypes.c_char_p(self.omni_vlm_version.encode("utf-8"))
            omni_vlm_cpp.omnivlm_init(self.ctx_params_model, self.ctx_params_mmproj, self.ctx_params_omni_vlm_version)
        except Exception as e:
            logging.error(f"Error loading model: {e}")
            raise
        
    def run(self):
        from nexa.gguf.llama._utils_spinner import start_spinner, stop_spinner
        
        while True:
            try:
                image_path = nexa_prompt("Image Path (required): ")
                if not os.path.exists(image_path):
                    print(f"Image path: {image_path} not found, running omni VLM without image input.")
                # Skip user input for OCR version
                user_input = "" if self.omni_vlm_version == "vlm-81-ocr" else nexa_prompt()

                stop_event, spinner_thread = start_spinner(
                style="default", 
                message=""  
                )

                response = self.inference(user_input, image_path)

                stop_spinner(stop_event, spinner_thread)

                print(f"\nResponse: {response}")
            except KeyboardInterrupt:
                print("\nExiting...")
                break
            except Exception as e:
                logging.error(f"\nError during audio generation: {e}", exc_info=True)
            print("\n")

    def inference(self, prompt: str, image_path: str):
        with suppress_stdout_stderr():
            if prompt and prompt[0].islower():
                prompt = prompt[0].upper() + prompt[1:]
                
            prompt = ctypes.c_char_p(prompt.encode("utf-8"))
            image_path = ctypes.c_char_p(image_path.encode("utf-8"))
            response = omni_vlm_cpp.omnivlm_inference(prompt, image_path)
            
            decoded_response = response.decode('utf-8')
            if '<|im_start|>assistant' in decoded_response:
                decoded_response = decoded_response.replace('<|im_start|>assistant', '').strip()
                
            return decoded_response

    def __del__(self):
        omni_vlm_cpp.omnivlm_free()

    def run_streamlit(self, model_path: str, is_local_path = False, hf = False, projector_local_path = None):
        """
        Run the Streamlit UI.
        """
        logging.info("Running Streamlit UI...")

        streamlit_script_path = (
            Path(os.path.abspath(__file__)).parent / "streamlit" / "streamlit_vlm_omni.py"
        )

        sys.argv = ["streamlit", "run", str(streamlit_script_path), model_path, str(is_local_path), str(hf), str(projector_local_path)]
        sys.exit(stcli.main())


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(
        description="Run audio-in text-out generation with nexa-omni-audio model"
    )
    parser.add_argument(
        "model_path",
        type=str,
        help="Path or identifier for the model in Nexa Model Hub",
    )
    parser.add_argument(
        "-d",
        "--device",
        type=str,
        choices=["auto", "cpu", "gpu"],
        default="auto",
        help="Device to use for inference (auto, cpu, or gpu)",
    )
    parser.add_argument(
        "-st",
        "--streamlit",
        action="store_true",
        help="Run the inference in Streamlit UI",
    )
    parser.add_argument(
        "--omni_vlm_version",
        type=str,
        choices=["vlm-81-ocr", "vlm-81-instruct", "nano-vlm-instruct"],
        default="vlm-81-instruct",
        help="omni-vlm-version to use",
    )

    args = parser.parse_args()
    kwargs = {k: v for k, v in vars(args).items() if v is not None}
    model_path = kwargs.pop("model_path")
    device = kwargs.pop("device", "auto")

    inference = NexaOmniVlmInference(model_path, device=device, **kwargs)
    if args.streamlit:
        inference.run_streamlit(model_path)
    else:
        inference.run()