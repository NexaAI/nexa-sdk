from nexa.gguf.llama.llama_cpp import (
    LLAMA_FTYPE_ALL_F32,
    LLAMA_FTYPE_MOSTLY_F16,
    LLAMA_FTYPE_MOSTLY_Q4_0,
    LLAMA_FTYPE_MOSTLY_Q4_1,
    LLAMA_FTYPE_MOSTLY_Q8_0,
    LLAMA_FTYPE_MOSTLY_Q5_0,
    LLAMA_FTYPE_MOSTLY_Q5_1,
    LLAMA_FTYPE_MOSTLY_Q2_K,
    LLAMA_FTYPE_MOSTLY_Q3_K_S,
    LLAMA_FTYPE_MOSTLY_Q3_K_M,
    LLAMA_FTYPE_MOSTLY_Q3_K_L,
    LLAMA_FTYPE_MOSTLY_Q4_K_S,
    LLAMA_FTYPE_MOSTLY_Q4_K_M,
    LLAMA_FTYPE_MOSTLY_Q5_K_S,
    LLAMA_FTYPE_MOSTLY_Q5_K_M,
    LLAMA_FTYPE_MOSTLY_Q6_K,
    LLAMA_FTYPE_MOSTLY_IQ2_XXS,
    LLAMA_FTYPE_MOSTLY_IQ2_XS,
    LLAMA_FTYPE_MOSTLY_Q2_K_S,
    LLAMA_FTYPE_MOSTLY_IQ3_XS,
    LLAMA_FTYPE_MOSTLY_IQ3_XXS,
    LLAMA_FTYPE_MOSTLY_IQ1_S,
    LLAMA_FTYPE_MOSTLY_IQ4_NL,
    LLAMA_FTYPE_MOSTLY_IQ3_S,
    LLAMA_FTYPE_MOSTLY_IQ3_M,
    LLAMA_FTYPE_MOSTLY_IQ2_S,
    LLAMA_FTYPE_MOSTLY_IQ2_M,
    LLAMA_FTYPE_MOSTLY_IQ4_XS,
    LLAMA_FTYPE_MOSTLY_IQ1_M,
    LLAMA_FTYPE_MOSTLY_BF16,
    LLAMA_FTYPE_MOSTLY_Q4_0_4_4,
    LLAMA_FTYPE_MOSTLY_Q4_0_4_8,
    LLAMA_FTYPE_MOSTLY_Q4_0_8_8,
    LLAMA_FTYPE_MOSTLY_TQ1_0,
    LLAMA_FTYPE_MOSTLY_TQ2_0,
)
from nexa.gguf.llama.llama_cpp import (
    GGML_TYPE_F32,
    GGML_TYPE_F16,
    GGML_TYPE_Q4_0,
    GGML_TYPE_Q4_1,
    GGML_TYPE_Q5_0,
    GGML_TYPE_Q5_1,
    GGML_TYPE_Q8_0,
    GGML_TYPE_Q8_1,
    GGML_TYPE_Q2_K,
    GGML_TYPE_Q3_K,
    GGML_TYPE_Q4_K,
    GGML_TYPE_Q5_K,
    GGML_TYPE_Q6_K,
    GGML_TYPE_Q8_K,
    GGML_TYPE_IQ2_XXS,
    GGML_TYPE_IQ2_XS,
    GGML_TYPE_IQ3_XXS,
    GGML_TYPE_IQ1_S,
    GGML_TYPE_IQ4_NL,
    GGML_TYPE_IQ3_S,
    GGML_TYPE_IQ2_S,
    GGML_TYPE_IQ4_XS,
    GGML_TYPE_I8,
    GGML_TYPE_I16,
    GGML_TYPE_I32,
    GGML_TYPE_I64,
    GGML_TYPE_F64,
    GGML_TYPE_IQ1_M,
    GGML_TYPE_BF16,
    GGML_TYPE_Q4_0_4_4,
    GGML_TYPE_Q4_0_4_8,
    GGML_TYPE_Q4_0_8_8,
    GGML_TYPE_COUNT,
)
from nexa.gguf.llama.llama_cpp import llama_model_quantize_params, llama_model_quantize
# from nexa.gguf.llama._utils_transformers import suppress_stdout_stderr
import os
import logging
import argparse
from typing import Optional
from pathlib import Path
import json

# Set up logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

LLAMA_QUANTIZATION_TYPES = {
    "q4_0": LLAMA_FTYPE_MOSTLY_Q4_0,
    "q4_1": LLAMA_FTYPE_MOSTLY_Q4_1,
    "q5_0": LLAMA_FTYPE_MOSTLY_Q5_0,
    "q5_1": LLAMA_FTYPE_MOSTLY_Q5_1,
    "q8_0": LLAMA_FTYPE_MOSTLY_Q8_0,
    "q2_k": LLAMA_FTYPE_MOSTLY_Q2_K,
    "q3_k_s": LLAMA_FTYPE_MOSTLY_Q3_K_S,
    "q3_k_m": LLAMA_FTYPE_MOSTLY_Q3_K_M,
    "q3_k_l": LLAMA_FTYPE_MOSTLY_Q3_K_L,
    "q4_k_s": LLAMA_FTYPE_MOSTLY_Q4_K_S,
    "q4_k_m": LLAMA_FTYPE_MOSTLY_Q4_K_M,
    "q5_k_s": LLAMA_FTYPE_MOSTLY_Q5_K_S,
    "q5_k_m": LLAMA_FTYPE_MOSTLY_Q5_K_M,
    "q6_k": LLAMA_FTYPE_MOSTLY_Q6_K,
    "iq2_xxs": LLAMA_FTYPE_MOSTLY_IQ2_XXS,
    "iq2_xs": LLAMA_FTYPE_MOSTLY_IQ2_XS,
    "q2_k_s": LLAMA_FTYPE_MOSTLY_Q2_K_S,
    "iq3_xs": LLAMA_FTYPE_MOSTLY_IQ3_XS,
    "iq3_xxs": LLAMA_FTYPE_MOSTLY_IQ3_XXS,
    "iq1_s": LLAMA_FTYPE_MOSTLY_IQ1_S,
    "iq4_nl": LLAMA_FTYPE_MOSTLY_IQ4_NL,
    "iq3_s": LLAMA_FTYPE_MOSTLY_IQ3_S,
    "iq3_m": LLAMA_FTYPE_MOSTLY_IQ3_M,
    "iq2_s": LLAMA_FTYPE_MOSTLY_IQ2_S,
    "iq2_m": LLAMA_FTYPE_MOSTLY_IQ2_M,
    "iq4_xs": LLAMA_FTYPE_MOSTLY_IQ4_XS,
    "iq1_m": LLAMA_FTYPE_MOSTLY_IQ1_M,
    "f16": LLAMA_FTYPE_MOSTLY_F16,
    "f32": LLAMA_FTYPE_ALL_F32,
    "bf16": LLAMA_FTYPE_MOSTLY_BF16,
    "q4_0_4_4": LLAMA_FTYPE_MOSTLY_Q4_0_4_4,
    "q4_0_4_8": LLAMA_FTYPE_MOSTLY_Q4_0_4_8,
    "q4_0_8_8": LLAMA_FTYPE_MOSTLY_Q4_0_8_8,
    "tq1_0": LLAMA_FTYPE_MOSTLY_TQ1_0,
    "tq2_0": LLAMA_FTYPE_MOSTLY_TQ2_0,
}

GGML_TYPES = {
    "f32": GGML_TYPE_F32,
    "f16": GGML_TYPE_F16,
    "q4_0": GGML_TYPE_Q4_0,
    "q4_1": GGML_TYPE_Q4_1,
    "q5_0": GGML_TYPE_Q5_0,
    "q5_1": GGML_TYPE_Q5_1,
    "q8_0": GGML_TYPE_Q8_0,
    "q8_1": GGML_TYPE_Q8_1,
    "q2_k": GGML_TYPE_Q2_K,
    "q3_k": GGML_TYPE_Q3_K,
    "q4_k": GGML_TYPE_Q4_K,
    "q5_k": GGML_TYPE_Q5_K,
    "q6_k": GGML_TYPE_Q6_K,
    "q8_k": GGML_TYPE_Q8_K,
    "iq2_xxs": GGML_TYPE_IQ2_XXS,
    "iq2_xs": GGML_TYPE_IQ2_XS,
    "iq3_xxs": GGML_TYPE_IQ3_XXS,
    "iq1_s": GGML_TYPE_IQ1_S,
    "iq4_nl": GGML_TYPE_IQ4_NL,
    "iq3_s": GGML_TYPE_IQ3_S,
    "iq2_s": GGML_TYPE_IQ2_S,
    "iq4_xs": GGML_TYPE_IQ4_XS,
    "i8": GGML_TYPE_I8,
    "i16": GGML_TYPE_I16,
    "i32": GGML_TYPE_I32,
    "i64": GGML_TYPE_I64,
    "f64": GGML_TYPE_F64,
    "iq1_m": GGML_TYPE_IQ1_M,
    "bf16": GGML_TYPE_BF16,
    "q4_0_4_4": GGML_TYPE_Q4_0_4_4,
    "q4_0_4_8": GGML_TYPE_Q4_0_4_8,
    "q4_0_8_8": GGML_TYPE_Q4_0_8_8,
}

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
    

from nexa_gguf.convert_hf_to_gguf import nexa_convert_hf_to_gguf

def convert_hf_to_quantized_gguf(input_path: str, output_file: str = None, ftype: str = "q4_0", **kwargs) -> None:
    # Convert input path to absolute path
    input_path = os.path.abspath(input_path)
    output_file = os.path.abspath(output_file)

    if os.path.isdir(input_path):
        if not os.path.exists(input_path):
            logger.error(f"Input directory does not exist: {input_path}")
            return
        
        safetensors_files = [f for f in os.listdir(input_path) if f.endswith('.safetensors')]
        if safetensors_files:
            # Create tmp file path
            tmp_dir = Path.home().absolute() / ".cache" / "nexa" / "tmp_models"
            tmp_dir.mkdir(parents=True, exist_ok=True)
            tmp_file_name = f"{Path(input_path).name}-f16.gguf"
            tmp_file_path = tmp_dir / tmp_file_name

            # Convert HF model to GGUF
            nexa_convert_hf_to_gguf(model=input_path, outfile=str(tmp_file_path.absolute()), **kwargs)

            # Quantize GGUF model
            quantize_model(str(tmp_file_path.absolute()), output_file, ftype, **kwargs)
        else:
            logger.error(f"No .safetensors files found in directory: {input_path}")
    elif input_path.endswith('.gguf'):
        # Directly call nexa_convert_hf_to_gguf with input_path
        quantize_model(input_file=input_path, output_file=output_file, ftype=ftype, **kwargs)
    else:
        logger.error(f"Invalid input path: {input_path}. Must be a directory with .safetensors files or a .gguf file.")
    




def main():
    parser = argparse.ArgumentParser(description="Convert and quantize a Hugging Face model to GGUF format.")
    parser.add_argument("input_path", type=str, help="Path to the input Hugging Face model directory or GGUF file")
    parser.add_argument("-o", "--output_file", type=str, help="Path to the output quantized GGUF file")
    parser.add_argument("-f", "--ftype", type=str, default="q4_0", help="Quantization type (default: q4_0)")
    parser.add_argument("-t", "--nthread", type=int, default=4, help="Number of threads to use (default: 4)")
    
    # Arguments for convert_hf_to_gguf
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
        if k not in ["input_path", "output_file", "ftype"] and v is not None
    }

    # Convert string types to GGML types if specified
    if args.output_tensor_type:
        kwargs["output_tensor_type"] = GGML_TYPES.get(args.output_tensor_type, GGML_TYPE_COUNT)
    if args.token_embedding_type:
        kwargs["token_embedding_type"] = GGML_TYPES.get(args.token_embedding_type, GGML_TYPE_COUNT)

    try:
        convert_hf_to_quantized_gguf(args.input_path, args.output_file, args.ftype, **kwargs)
    except Exception as e:
        logger.error(f"Error during conversion and quantization: {str(e)}")
        exit(1)

if __name__ == "__main__":
    main()
