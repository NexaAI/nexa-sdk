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
        verbose (bool): Enable verbose output.
        **kwargs: Additional parameters for quantization:
            output_tensor_type (int): Output tensor type.
            token_embedding_type (int): Token embeddings tensor type.
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
    params.ftype = LLAMA_QUANTIZATION_TYPES.get(ftype, LLAMA_FTYPE_MOSTLY_Q4_0)
    params.output_tensor_type = GGML_TYPE_COUNT
    params.token_embedding_type = GGML_TYPE_COUNT

    # Set additional parameters from kwargs
    for key, value in kwargs.items():
        if hasattr(params, key):
            setattr(params, key, value)

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

def convert_hf_to_quantized_gguf(input_hf_directory: str, output_file: str, ftype: str = "q4_0", **kwargs) -> None:
    # Convert input paths to absolute paths
    input_hf_directory = os.path.abspath(input_hf_directory)
    output_file = os.path.abspath(output_file)

    # Create tmp file path
    tmp_dir = Path.home().absolute() / ".cache" / "nexa" / "tmp_models"
    tmp_dir.mkdir(parents=True, exist_ok=True)
    tmp_file_name = f"{Path(input_hf_directory).name}-f16.gguf"
    tmp_file_path = tmp_dir / tmp_file_name

    # Convert HF model to GGUF
    nexa_convert_hf_to_gguf(model=input_hf_directory, outfile=str(tmp_file_path.absolute()), **kwargs)

    # Quantize GGUF model
    quantize_model(str(tmp_file_path.absolute()), output_file, ftype, **kwargs)
    




def main():
    parser = argparse.ArgumentParser(description="Quantize a GGUF model file.")
    parser.add_argument("input_file", type=str, help="Path to the input GGUF file")
    parser.add_argument("-o", "--output_file", type=str, help="Path to the output quantized file")
    parser.add_argument("-f", "--ftype", type=str, default="q4_0", help="Quantization type (default: q4_0)")
    parser.add_argument("-t", "--nthread", type=int, default=4, help="Number of threads to use (default: 4)")
    
    # Additional parameters matching llama_model_quantize_params
    parser.add_argument("--output_tensor_type", type=str, help="Output tensor type")
    parser.add_argument("--token_embedding_type", type=str, help="Token embedding type")
    parser.add_argument("--allow_requantize", action="store_true", help="Allow quantizing non-f32/f16 tensors")
    parser.add_argument("--quantize_output_tensor", action="store_true", help="Quantize output.weight")
    parser.add_argument("--only_copy", action="store_true", help="Only copy tensors (ignores ftype, allow_requantize, and quantize_output_tensor)")
    parser.add_argument("--pure", action="store_true", help="Quantize all tensors to the default type")
    parser.add_argument("--keep_split", action="store_true", help="Quantize to the same number of shards")
    parser.add_argument("--verbose", action="store_true", help="Enable verbose output")

    args = parser.parse_args()

    # Prepare kwargs for additional parameters
    kwargs = {
        k: v for k, v in vars(args).items()
        if k not in ["input_file", "output_file", "ftype", "nthread", "verbose"] and v is not None
    }

    # Convert string types to GGML types if specified
    if args.output_tensor_type:
        kwargs["output_tensor_type"] = GGML_TYPES.get(args.output_tensor_type, GGML_TYPE_COUNT)
    if args.token_embedding_type:
        kwargs["token_embedding_type"] = GGML_TYPES.get(args.token_embedding_type, GGML_TYPE_COUNT)

    try:
        quantize_model(args.input_file, args.output_file, args.ftype, args.nthread, verbose=args.verbose, **kwargs)
    except Exception as e:
        logger.error(f"Error during quantization: {str(e)}")
        exit(1)

if __name__ == "__main__":
    # main()
    convert_hf_to_quantized_gguf("../models/octopus-v2", "../models/octopus-v2-q4_0.gguf", "q4_0")
