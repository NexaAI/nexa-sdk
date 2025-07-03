package nexa_sdk

/*
#include <stdlib.h>
#include "ml.h"

extern bool go_generate_stream_on_token(char*, void*);
*/
import "C"

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg" // Register JPEG format
	_ "image/png"  // Register PNG format
	"os"
	"slices"
	"unsafe"
)

// ImageSamplerConfig represents sampling parameters for image generation
type ImageSamplerConfig struct {
	Method        string  // Sampling method: "ddim", "ddpm", etc.
	Steps         int32   // Number of denoising steps
	GuidanceScale float32 // Classifier-free guidance scale
	Eta           float32 // DDIM eta parameter
	Seed          int32   // Random seed (-1 for random)
}

// ImageGenerationConfig represents configuration for image generation
type ImageGenerationConfig struct {
	Prompts         []string           // Required positive prompts
	NegativePrompts []string           // Optional negative prompts
	Height          int32              // Output image height
	Width           int32              // Output image width
	SamplerConfig   ImageSamplerConfig // Sampling parameters
	LoraID          int32              // LoRA ID (-1 for none)
	InitImage       *Image             // Initial image (nil for txt2img)
	Strength        float32            // Denoising strength for img2img
}

// SchedulerConfig represents diffusion scheduler configuration
type SchedulerConfig struct {
	Type              string  // Scheduler type: "ddim", etc.
	NumTrainTimesteps int32   // Training timesteps
	BetaStart         float32 // Beta schedule start
	BetaEnd           float32 // Beta schedule end
	BetaSchedule      string  // Beta schedule: "scaled_linear"
	PredictionType    string  // Prediction type: "epsilon", "v_prediction"
	TimestepType      string  // Timestep type: "discrete", "continuous"
	TimestepSpacing   string  // Timestep spacing: "linspace", "leading", "trailing"
	InterpolationType string  // Interpolation type: "linear", "exponential"
	ConfigPath        string  // Optional config file path
}

// ImageGen wraps the C library ImageGen structure and provides Go interface
type ImageGen struct {
	ptr *C.ml_ImageGen // Pointer to the underlying C ImageGen structure
}

// NewImageGen creates a new ImageGen instance with the specified model and configuration
func NewImageGen(modelPath string, schedulerConfigPath string, device string) *ImageGen {
	cModelPath := C.CString(modelPath)
	defer C.free(unsafe.Pointer(cModelPath))

	var cSchedulerConfigPath *C.char
	if schedulerConfigPath != "" {
		cSchedulerConfigPath = C.CString(schedulerConfigPath)
		defer C.free(unsafe.Pointer(cSchedulerConfigPath))
	}

	var cDevice *C.char
	if device != "" {
		cDevice = C.CString(device)
		defer C.free(unsafe.Pointer(cDevice))
	}

	return &ImageGen{
		ptr: C.ml_imagegen_create(cModelPath, cSchedulerConfigPath, cDevice),
	}
}

// Destroy frees the memory allocated for the ImageGen instance
func (ig *ImageGen) Destroy() {
	C.ml_imagegen_destroy(ig.ptr)
	ig.ptr = nil
}

// LoadModel loads a model from the specified path
func (ig *ImageGen) LoadModel(modelPath string, extraData unsafe.Pointer) bool {
	cModelPath := C.CString(modelPath)
	defer C.free(unsafe.Pointer(cModelPath))

	return bool(C.ml_imagegen_load_model(ig.ptr, cModelPath, extraData))
}

// Close closes the ImageGen instance
func (ig *ImageGen) Close() {
	C.ml_imagegen_close(ig.ptr)
}

// SetScheduler sets the scheduler configuration
func (ig *ImageGen) SetScheduler(config SchedulerConfig) {
	cConfig := C.ml_SchedulerConfig{}

	if config.Type != "" {
		cConfig._type = C.CString(config.Type)
		defer C.free(unsafe.Pointer(cConfig._type))
	}

	cConfig.num_train_timesteps = C.int32_t(config.NumTrainTimesteps)
	cConfig.beta_start = C.float(config.BetaStart)
	cConfig.beta_end = C.float(config.BetaEnd)

	if config.BetaSchedule != "" {
		cConfig.beta_schedule = C.CString(config.BetaSchedule)
		defer C.free(unsafe.Pointer(cConfig.beta_schedule))
	}

	if config.PredictionType != "" {
		cConfig.prediction_type = C.CString(config.PredictionType)
		defer C.free(unsafe.Pointer(cConfig.prediction_type))
	}

	if config.TimestepType != "" {
		cConfig.timestep_type = C.CString(config.TimestepType)
		defer C.free(unsafe.Pointer(cConfig.timestep_type))
	}

	if config.TimestepSpacing != "" {
		cConfig.timestep_spacing = C.CString(config.TimestepSpacing)
		defer C.free(unsafe.Pointer(cConfig.timestep_spacing))
	}

	if config.InterpolationType != "" {
		cConfig.interpolation_type = C.CString(config.InterpolationType)
		defer C.free(unsafe.Pointer(cConfig.interpolation_type))
	}

	if config.ConfigPath != "" {
		cConfig.config_path = C.CString(config.ConfigPath)
		defer C.free(unsafe.Pointer(cConfig.config_path))
	}

	C.ml_imagegen_set_scheduler(ig.ptr, &cConfig)
}

// SetSampler sets the sampler configuration
func (ig *ImageGen) SetSampler(config ImageSamplerConfig) {
	cConfig := C.ml_ImageSamplerConfig{}

	if config.Method != "" {
		cConfig.method = C.CString(config.Method)
		defer C.free(unsafe.Pointer(cConfig.method))
	}

	cConfig.steps = C.int32_t(config.Steps)
	cConfig.guidance_scale = C.float(config.GuidanceScale)
	cConfig.eta = C.float(config.Eta)
	cConfig.seed = C.int32_t(config.Seed)

	C.ml_imagegen_set_sampler(ig.ptr, &cConfig)
}

// ResetSampler resets the sampler to default settings
func (ig *ImageGen) ResetSampler() {
	C.ml_imagegen_reset_sampler(ig.ptr)
}

// Txt2Img generates an image from text prompt
func (ig *ImageGen) Txt2Img(prompt string, config ImageGenerationConfig) (*Image, error) {
	cPrompt := C.CString(prompt)
	defer C.free(unsafe.Pointer(cPrompt))

	cConfig, cleanup := ig.convertImageGenerationConfig(config)
	defer cleanup()

	cImage := C.ml_imagegen_txt2img(ig.ptr, cPrompt, cConfig)
	if cImage.data == nil {
		return nil, ErrSDK
	}
	return &Image{ptr: &cImage}, nil
}

// Img2Img generates an image from an initial image and text prompt
func (ig *ImageGen) Img2Img(initImage *Image, prompt string, config ImageGenerationConfig) (*Image, error) {
	if initImage == nil || initImage.ptr == nil {
		return nil, ErrSDK
	}

	cPrompt := C.CString(prompt)
	defer C.free(unsafe.Pointer(cPrompt))

	cConfig, cleanup := ig.convertImageGenerationConfig(config)
	defer cleanup()

	cImage := C.ml_imagegen_img2img(ig.ptr, initImage.ptr, cPrompt, cConfig)
	if cImage.data == nil {
		return nil, ErrSDK
	}

	return &Image{ptr: &cImage}, nil
}

// Generate generates an image using the provided configuration
func (ig *ImageGen) Generate(config ImageGenerationConfig) (*Image, error) {
	cConfig, cleanup := ig.convertImageGenerationConfig(config)
	defer cleanup()

	cImage := C.ml_imagegen_generate(ig.ptr, cConfig)
	if cImage.data == nil {
		return nil, ErrSDK
	}

	return &Image{ptr: &cImage}, nil
}

// convertImageGenerationConfig converts Go config to C config and returns cleanup function
func (ig *ImageGen) convertImageGenerationConfig(config ImageGenerationConfig) (*C.ml_ImageGenerationConfig, func()) {
	cConfig := &C.ml_ImageGenerationConfig{}
	var cleanupFuncs []func()

	if len(config.Prompts) > 0 {
		cprompts, clean, err := stringSliceToC(config.Prompts)
		if err != nil {
			panic(fmt.Sprintf("failed to convert prompts: %v", err))
		}
		cConfig.prompts = cprompts
		cleanupFuncs = append(cleanupFuncs, clean)
	}

	// Convert negative prompts
	if len(config.NegativePrompts) > 0 {
		cprompts, clean, err := stringSliceToC(config.NegativePrompts)
		if err != nil {
			panic(fmt.Sprintf("failed to convert prompts: %v", err))
		}
		cConfig.prompts = cprompts
		cleanupFuncs = append(cleanupFuncs, clean)
	}

	// Set basic parameters
	cConfig.height = C.int32_t(config.Height)
	cConfig.width = C.int32_t(config.Width)
	cConfig.lora_id = C.int32_t(config.LoraID)
	cConfig.strength = C.float(config.Strength)

	// Set sampler config
	if config.SamplerConfig.Method != "" {
		cConfig.sampler_config.method = C.CString(config.SamplerConfig.Method)
		cleanupFuncs = append(cleanupFuncs, func() {
			C.free(unsafe.Pointer(cConfig.sampler_config.method))
		})
	}
	cConfig.sampler_config.steps = C.int32_t(config.SamplerConfig.Steps)
	cConfig.sampler_config.guidance_scale = C.float(config.SamplerConfig.GuidanceScale)
	cConfig.sampler_config.eta = C.float(config.SamplerConfig.Eta)
	cConfig.sampler_config.seed = C.int32_t(config.SamplerConfig.Seed)

	// Set init image if provided
	if config.InitImage != nil && config.InitImage.ptr != nil {
		cConfig.init_image = config.InitImage.ptr
	}

	// Return cleanup function
	cleanup := func() {
		for _, fn := range cleanupFuncs {
			fn()
		}
	}

	return cConfig, cleanup
}

func stringSliceToC(goStrings []string) (**C.char, func(), error) {
	l := len(goStrings)
	if l == 0 {
		return nil, func() {}, nil
	}

	cArr := C.malloc(C.size_t(l+1) * C.size_t(unsafe.Sizeof(uintptr(0))))
	if cArr == nil {
		return nil, nil, errors.New("failed to allocate memory")
	}

	cleanupFuncs := make([]func(), 0, l+1)
	cleanupFuncs = append(cleanupFuncs, func() { C.free(unsafe.Pointer(cArr)) })
	cArrSlice := (*[1 << 30]*C.char)(cArr)[:l+1]
	for i, s := range goStrings {
		cs := C.CString(s)
		cleanupFuncs = append(cleanupFuncs, func() { C.free(unsafe.Pointer(cs)) })
		cArrSlice[i] = cs
	}
	cArrSlice[l] = nil

	clean := func() {
		for _, fn := range slices.Backward(cleanupFuncs) {
			fn()
		}
	}
	return (**C.char)(unsafe.Pointer(cArr)), clean, nil
}

// Image represents an image structure
type Image struct {
	ptr *C.ml_Image // Pointer to the underlying C Image structure
}

// NewImage creates a new Image instance from the specified file path
func NewImage(path string) (*Image, error) {
	// Open the image file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image[%s]: %w", path, err)
	}
	defer file.Close()

	// Decode the image
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %s (format: %s): %w", path, format, err)
	}

	// Get image dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Validate image dimensions
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid image dimensions: %dx%d", width, height)
	}

	// Determine optimal format and channels
	var channels int
	var pixelData []uint8

	// Try to preserve original format when possible
	switch srcImg := img.(type) {
	case *image.RGBA:
		channels = 4
		pixelData = srcImg.Pix
	case *image.NRGBA:
		channels = 4
		// Convert NRGBA to RGBA
		rgbaImg := image.NewRGBA(bounds)
		draw.Draw(rgbaImg, bounds, srcImg, bounds.Min, draw.Src)
		pixelData = rgbaImg.Pix
	case *image.Gray:
		channels = 1
		pixelData = srcImg.Pix
	default:
		// Convert other formats to RGB (3 channels)
		channels = 3
		rgbImg := image.NewRGBA(bounds)
		draw.Draw(rgbImg, bounds, img, bounds.Min, draw.Src)

		// Extract RGB channels (skip alpha)
		pixelData = make([]uint8, width*height*3)
		for i := 0; i < len(rgbImg.Pix); i += 4 {
			rgbIdx := (i / 4) * 3
			pixelData[rgbIdx] = rgbImg.Pix[i]     // R
			pixelData[rgbIdx+1] = rgbImg.Pix[i+1] // G
			pixelData[rgbIdx+2] = rgbImg.Pix[i+2] // B
		}
	}

	// Calculate data size
	dataSize := width * height * channels
	if len(pixelData) != dataSize {
		return nil, fmt.Errorf("pixel data size mismatch: expected %d, got %d", dataSize, len(pixelData))
	}

	// Allocate C memory for image data
	cData := (*C.float)(C.malloc(C.size_t(dataSize * 4))) // 4 bytes per float32
	if cData == nil {
		return nil, fmt.Errorf("failed to allocate memory for image data")
	}

	// Create cleanup function for error cases
	cleanupData := func() {
		C.free(unsafe.Pointer(cData))
	}

	// Convert pixel data to float32 array (optimized)
	floatData := (*[1 << 30]C.float)(unsafe.Pointer(cData))[:dataSize:dataSize]
	for i, pixel := range pixelData {
		// Normalize pixel values from [0, 255] to [0.0, 1.0]
		floatData[i] = C.float(float32(pixel) / 255.0)
	}

	// Create C.ml_Image structure
	cImage := (*C.ml_Image)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_Image{}))))
	if cImage == nil {
		cleanupData()
		return nil, fmt.Errorf("failed to allocate memory for ml_Image")
	}

	// Initialize the C structure
	cImage.data = cData
	cImage.width = C.int32_t(width)
	cImage.height = C.int32_t(height)
	cImage.channels = C.int32_t(channels)

	return &Image{ptr: cImage}, nil
}

func (img *Image) Save(path string) error {
	if img.ptr == nil {
		return fmt.Errorf("image pointer is nil")
	}

	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	C.ml_image_save(img.ptr, cPath)
	img.ptr = nil
	return nil
}

func (img *Image) Free() {
	if img.ptr != nil {
		C.ml_image_free(img.ptr)
		img.ptr = nil
	}
}

// GetData returns the raw image data
func (img *Image) GetRaw() []float32 {
	if img.ptr == nil {
		return nil
	}
	l := int(img.ptr.width * img.ptr.height * img.ptr.channels)
	return (*[1 << 30]float32)(unsafe.Pointer(img.ptr.data))[:l:l]
}
