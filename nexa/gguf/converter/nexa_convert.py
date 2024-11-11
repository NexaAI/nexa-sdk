import os
import logging
import argparse
from typing import Optional
from pathlib import Path
import json

from nexa.gguf.llama.llama_cpp import GGML_TYPE_COUNT, LLAMA_FTYPE_MOSTLY_Q4_0
from nexa.gguf.converter.constants import LLAMA_QUANTIZATION_TYPES, GGML_TYPES
from nexa.gguf.llama.llama_cpp import llama_model_quantize_params, llama_model_quantize

# Set up logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


def quantize_model(
    input_file: str,
    output_file: Optional[str] = None,
    ftype: str = "q4_0",
    nthread: int = 4,
    **kwargs
) -> None:
    """
    Quantize a GGUF model file.

    Args:
        input_file (str): Path to the input GGUF file.
        output_file (Optional[str]): Path to the output quantized file. If None, a default path will be used.
        ftype (str): Quantization type (default: "q4_0").
        nthread (int): Number of threads to use for quantization (default: 4).
        **kwargs: Additional parameters for quantization:
            output_tensor_type (str): Output tensor type.
            token_embedding_type (str): Token embeddings tensor type.
            allow_requantize (bool): Allow quantizing non-f32/f16 tensors.
            quantize_output_tensor (bool): Quantize output.weight.
            only_copy (bool): Only copy tensors - ftype, allow_requantize and quantize_output_tensor are ignored.
            pure (bool): Quantize all tensors to the default type.
            keep_split (bool): Quantize to the same number of shards.
            imatrix (ctypes.c_void_p): Pointer to importance matrix data.
            kv_overrides (ctypes.c_void_p): Pointer to vector containing overrides.

    Raises:
        FileNotFoundError: If the input file doesn't exist.
        ValueError: If an invalid quantization type is provided.
    """
    # Check if input file exists
    if not os.path.isfile(input_file):
        raise FileNotFoundError(f"Input file does not exist: {input_file}")

    # Set up output file path
    if output_file is None:
        output_dir = os.path.join(os.path.dirname(input_file), "quantized_models")
        os.makedirs(output_dir, exist_ok=True)
        output_file = os.path.join(output_dir, f"{os.path.basename(input_file).split('.')[0]}_{ftype}.gguf")
    else:
        output_dir = os.path.dirname(output_file)
        os.makedirs(output_dir, exist_ok=True)

    # Set up quantization parameters
    params = llama_model_quantize_params()
    params.nthread = nthread

    # Handle ftype
    if ftype in LLAMA_QUANTIZATION_TYPES:
        params.ftype = LLAMA_QUANTIZATION_TYPES[ftype]
    else:
        logger.warning(f"Provided ftype '{ftype}' not found in LLAMA_QUANTIZATION_TYPES. Using default Q4_0.")
        params.ftype = LLAMA_FTYPE_MOSTLY_Q4_0

    # Handle output_tensor_type
    output_tensor_type = kwargs.get('output_tensor_type', '')
    if output_tensor_type:
        if output_tensor_type in GGML_TYPES:
            params.output_tensor_type = GGML_TYPES[output_tensor_type]
        else:
            logger.warning(f"Provided output_tensor_type '{output_tensor_type}' not found in GGML_TYPES. Using default COUNT.")
            params.output_tensor_type = GGML_TYPE_COUNT
    else:
        params.output_tensor_type = GGML_TYPE_COUNT

    # Handle token_embedding_type
    token_embedding_type = kwargs.get('token_embedding_type', '')
    if token_embedding_type:
        if token_embedding_type in GGML_TYPES:
            params.token_embedding_type = GGML_TYPES[token_embedding_type]
        else:
            logger.warning(f"Provided token_embedding_type '{token_embedding_type}' not found in GGML_TYPES. Using default COUNT.")
            params.token_embedding_type = GGML_TYPE_COUNT
    else:
        params.token_embedding_type = GGML_TYPE_COUNT

    logger.info(f"Starting quantization of {input_file}")
    logger.info(f"Output file: {output_file}")

    try:
        llama_model_quantize(
            input_file.encode("utf-8"),
            output_file.encode("utf-8"),
            params,
        )
    except Exception as e:
        logger.error(f"Quantization failed: {str(e)}")
        raise
    

def convert_hf_to_quantized_gguf(
    input_path: str, 
    output_file: str = None, 
    ftype: str = "q4_0", 
    convert_type: str = "f16", 
    **kwargs
) -> Optional[str]:
    """
    Convert a model in safetensors format to a quantized GGUF file.

    This function handles the conversion of Hugging Face models to GGUF format and subsequent quantization.
    It can process both directories containing .safetensors files and existing .gguf files.

    Args:
        input_path (str): Path to the input Hugging Face model directory or GGUF file.
        output_file (str, optional): Path to the output quantized GGUF file. If None, a default path will be used.
        ftype (str, optional): Quantization type (default: "q4_0").
        convert_type (str, optional): Conversion type for safetensors to GGUF (default: "f16").
        **kwargs: Additional keyword arguments for the conversion and quantization process.

    Returns:
        Optional[str]: Path to the output quantized GGUF file if successful, None otherwise.

    Raises:
        FileNotFoundError: If the input directory or file does not exist.
        ValueError: If the input path is invalid or no .safetensors files are found in the directory.

    Note:
        - For directory inputs, this function first converts the model to GGUF format, then quantizes it.
        - For .gguf file inputs, it directly applies quantization.
        - Temporary files are created and cleaned up during the process.
    """
    # Convert input path to absolute path
    input_path = os.path.abspath(input_path)
    
    # Set default output file if not provided
    if not output_file:
        input_name = os.path.basename(input_path)
        if input_path.endswith('.gguf'):
            input_name = os.path.splitext(input_name)[0]  # Remove .gguf extension
        output_file = os.path.abspath(f"./{input_name}-{ftype}.gguf")
    else:
        output_file = os.path.abspath(output_file)

    if os.path.isdir(input_path):
        if not os.path.exists(input_path):
            logger.error(f"Input directory does not exist: {input_path}")
            return None
        
        safetensors_files = [f for f in os.listdir(input_path) if f.endswith('.safetensors')]
        if safetensors_files:
            # Create tmp file path
            tmp_dir = Path.home().absolute() / ".cache" / "nexa" / "tmp_models"
            tmp_dir.mkdir(parents=True, exist_ok=True)
            tmp_file_name = f"{Path(input_path).name}-{convert_type}.gguf"
            tmp_file_path = tmp_dir / tmp_file_name

            try:
                # Convert HF model to GGUF
                from nexa_gguf.convert_hf_to_gguf import nexa_convert_hf_to_gguf
                nexa_convert_hf_to_gguf(model=input_path, outfile=str(tmp_file_path.absolute()), outtype=convert_type, **kwargs)

                # Quantize GGUF model
                quantize_model(str(tmp_file_path.absolute()), output_file, ftype, **kwargs)
                return output_file
            finally:
                # Delete the temporary file
                if tmp_file_path.exists():
                    tmp_file_path.unlink()
        else:
            logger.error(f"No .safetensors files found in directory: {input_path}")
            return None
    elif input_path.endswith('.gguf'):
        # Directly call quantize_model with input_path
        quantize_model(input_file=input_path, output_file=output_file, ftype=ftype, **kwargs)
        return output_file
    else:
        logger.error(f"Invalid input path: {input_path}. Must be a directory with .safetensors files or a .gguf file.")
        return None
    

def main():
    parser = argparse.ArgumentParser(description="Convert and quantize a Hugging Face model to GGUF format.")
    # nexa convert specific arguments
    parser.add_argument("input_path", type=str, help="Path to the input Hugging Face model directory or GGUF file")
    parser.add_argument("ftype", nargs='?', type=str, default="q4_0", help="Quantization type (default: q4_0)")
    parser.add_argument("output_file", nargs='?', type=str, help="Path to the output quantized GGUF file")
        
    # Arguments for convert_hf_to_gguf
    # Reference: https://github.com/ggerganov/llama.cpp/blob/c8c07d658a6cefc5a50cfdf6be7d726503612303/convert_hf_to_gguf.py#L4284-L4344
    parser.add_argument("--convert_type", type=str, default="f16", help="Conversion type for safetensors to GGUF (default: f16)")
    parser.add_argument("--bigendian", action="store_true", help="Use big endian format")
    parser.add_argument("--use_temp_file", action="store_true", help="Use a temporary file during conversion")
    parser.add_argument("--no_lazy", action="store_true", help="Disable lazy loading")
    parser.add_argument("--metadata", type=json.loads, help="Additional metadata as JSON string")
    parser.add_argument("--split_max_tensors", type=int, default=0, help="Maximum number of tensors per split")
    parser.add_argument("--split_max_size", type=str, default="0", help="Maximum size per split")
    parser.add_argument("--no_tensor_first_split", action="store_true", help="Disable tensor-first splitting")
    parser.add_argument("--vocab_only", action="store_true", help="Only process vocabulary")
    parser.add_argument("--dry_run", action="store_true", help="Perform a dry run without actual conversion")
    
    # Arguments for quantization
    # Reference: https://github.com/ggerganov/llama.cpp/blob/c8c07d658a6cefc5a50cfdf6be7d726503612303/examples/quantize/quantize.cpp#L109-L133
    parser.add_argument("--nthread", type=int, default=4, help="Number of threads to use (default: 4)")
    parser.add_argument("--output_tensor_type", type=str, help="Output tensor type")
    parser.add_argument("--token_embedding_type", type=str, help="Token embedding type")
    parser.add_argument("--allow_requantize", action="store_true", help="Allow quantizing non-f32/f16 tensors")
    parser.add_argument("--quantize_output_tensor", action="store_true", help="Quantize output.weight")
    parser.add_argument("--only_copy", action="store_true", help="Only copy tensors (ignores ftype, allow_requantize, and quantize_output_tensor)")
    parser.add_argument("--pure", action="store_true", help="Quantize all tensors to the default type")
    parser.add_argument("--keep_split", action="store_true", help="Quantize to the same number of shards")

    args = parser.parse_args()

    # Prepare kwargs for additional parameters
    kwargs = {
        k: v for k, v in vars(args).items()
        if k not in ["input_path", "output_file", "ftype", "convert_type"] and v is not None
    }

    # Convert string types to GGML types if specified
    if args.output_tensor_type:
        kwargs["output_tensor_type"] = GGML_TYPES.get(args.output_tensor_type, GGML_TYPE_COUNT)
    if args.token_embedding_type:
        kwargs["token_embedding_type"] = GGML_TYPES.get(args.token_embedding_type, GGML_TYPE_COUNT)

    try:
        convert_hf_to_quantized_gguf(args.input_path, args.output_file, args.ftype, args.convert_type, **kwargs)
    except Exception as e:
        logger.error(f"Error during conversion and quantization: {str(e)}")
        exit(1)

if __name__ == "__main__":
    main()
