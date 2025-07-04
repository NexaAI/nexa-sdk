#include "ml.h"

extern "C"
{
    /** Free image data structure and its pixel data */
    void ml_image_free(ml_Image *image) {}
    void ml_image_save(const ml_Image *image, const char *filename) {}

    /* ====================  Lifecycle Management  ============================== */

    /** Create and initialize an image generator instance */
    ml_ImageGen *ml_imagegen_create(ml_Path model_path, ml_Path scheduler_config_path, const char *device) { return nullptr; };

    /** Destroy image generator instance and free associated resources */
    void ml_imagegen_destroy(ml_ImageGen *handle) {};

    /** Load model from path with optional extra configuration data */
    bool ml_imagegen_load_model(ml_ImageGen *handle, ml_Path model_path, const void *extra_data) { return false; };

    /** Close and cleanup image generator resources */
    void ml_imagegen_close(ml_ImageGen *handle) {};

    /* ====================  Configuration  ===================================== */

    /** Configure diffusion scheduler parameters */
    void ml_imagegen_set_scheduler(ml_ImageGen *handle, const ml_SchedulerConfig *config) {};

    /** Configure image generation sampling parameters */
    void ml_imagegen_set_sampler(ml_ImageGen *handle, const ml_ImageSamplerConfig *config) {};

    /** Reset sampling parameters to defaults */
    void ml_imagegen_reset_sampler(ml_ImageGen *handle) {};

    /* ====================  Image Generation  ================================== */

    /** Generate image from text prompt */
    ml_Image ml_imagegen_txt2img(
        ml_ImageGen *handle, const char *prompt_utf8, const ml_ImageGenerationConfig *config)
    {
        return ml_Image{nullptr, 0, 0, 0};
    };

    /** Generate image from initial image and prompt */
    ml_Image ml_imagegen_img2img(ml_ImageGen *handle, const ml_Image *init_image, const char *prompt_utf8,
                                 const ml_ImageGenerationConfig *config)
    {
        return ml_Image{nullptr, 0, 0, 0};
    };

    /** Generate image using full configuration */
    ml_Image ml_imagegen_generate(ml_ImageGen *handle, const ml_ImageGenerationConfig *config)
    {
        return ml_Image{nullptr, 0, 0, 0};
    };

    /* ====================  LoRA Management  ================================== */

    /** Set active LoRA adapter by ID */
    void ml_imagegen_set_lora(ml_ImageGen *handle, int32_t lora_id) {};

    /** Add LoRA adapter from file. Returns LoRA ID on success, negative on error */
    int32_t ml_imagegen_add_lora(ml_ImageGen *handle, ml_Path lora_path) { return -255; };

    /** Remove LoRA adapter by ID */
    void ml_imagegen_remove_lora(ml_ImageGen *handle, int32_t lora_id) {};

    /** List all loaded LoRA adapter IDs. Returns count, negative on error */
    int32_t ml_imagegen_list_loras(ml_ImageGen *handle, int32_t **out) { return -255; };
}