package main

import (
	"fmt"
	"strconv"
	"time"

	nexa_sdk "github.com/NexaAI/nexa-sdk/nexa-sdk"
	"github.com/spf13/cobra"
)

func genImage() *cobra.Command {
	imgCmd := &cobra.Command{
		Use: "image generate",
	}
	var model, out string
	var prompts []string
	var genType string
	var scheduler string
	// 创建图像生成器实例
	imgCmd.Flags().StringVarP(&model, "model", "m", "stabilityai/sdxl-turbo", "Model name for image generation")
	imgCmd.Flags().StringSliceVarP(&prompts, "prompt", "p", nil, "Prompt for image generation")
	imgCmd.Flags().StringVar(&genType, "type", "txt2img", "Type of image generation: txt2img, img2img")
	imgCmd.Flags().StringVarP(&out, "output", "o", "out.png", "Output file name for the generated image")
	imgCmd.Flags().StringVarP(&scheduler, "scheduler", "s", "", "Scheduler type for image generation")

	imgCmd.Run = func(cmd *cobra.Command, args []string) {
		nexa_sdk.Init()
		defer nexa_sdk.DeInit()

		imageGen := nexa_sdk.NewImageGen(model, scheduler, "")
		defer imageGen.Destroy()

		switch genType {
		case "txt2img":
			MLXText2Img(imageGen, prompts)
		case "img2img":
			MLXImg2Img(imageGen, prompts)
		default:
			fmt.Println("Unknown image generation type. Please use txt2img, img2img.")
			return
		}
	}

	return imgCmd
}

// MLXText2Img 文本到图像生成功能
func MLXText2Img(imageGen *nexa_sdk.ImageGen, prompts []string) {
	fmt.Println("\n===> MLX Text-to-Image Generation")

	if len(prompts) == 0 {
		fmt.Println("Error: Empty prompt provided")
		return
	}

	fmt.Printf("Prompt: %s\n", prompts)

	// 创建配置 - 使用SDXL-Turbo默认设置
	config := nexa_sdk.ImageGenerationConfig{
		Prompts:         prompts,
		NegativePrompts: nil,
		Height:          512, // 提高分辨率
		Width:           512,
		SamplerConfig: nexa_sdk.ImageSamplerConfig{
			Method:        "ddim",
			Steps:         4,   // 稍微增加步数
			GuidanceScale: 1.0, // 轻微引导
			Eta:           0.0,
			Seed:          -1, // 随机种子
		},
		LoraID:   -1,
		Strength: 1.0,
	}

	// 生成图像
	fmt.Println("Generating image...")
	result, err := imageGen.Txt2Img(prompts[0], config)
	if err != nil {
		fmt.Printf("Failed to generate image: %v\n", err)
		return
	}
	defer result.Free()

	outputPath := strconv.Itoa(int(time.Now().Unix())) + ".jpeg"
	err = result.Save(outputPath)
	if err != nil {
		fmt.Printf("Failed to save image: %v\n", err)
	} else {
		fmt.Printf("Image-to-image generation completed! Image saved to: %s\n", outputPath)
	}
}

// MLXImg2Img 图像到图像生成功能
func MLXImg2Img(imageGen *nexa_sdk.ImageGen, prompts []string) {
	fmt.Println("\n===> MLX Image-to-Image Generation")

	if len(prompts) == 0 {
		fmt.Println("Error: Empty prompt provided")
		return
	}

	// 创建测试初始图像
	initImg, err := createTestImage()
	if err != nil {
		fmt.Printf("Failed to create test image: %v\n", err)
		return
	}
	defer initImg.Free()

	fmt.Println("Loaded initial image")

	// 保存初始图像作为参考
	err = initImg.Save("init_image.png")
	if err != nil {
		fmt.Printf("Warning: Failed to save init image: %v\n", err)
	}

	// 创建配置
	config := nexa_sdk.ImageGenerationConfig{
		Prompts:         prompts,
		NegativePrompts: nil,
		Height:          512, // 标准尺寸
		Width:           512,
		SamplerConfig: nexa_sdk.ImageSamplerConfig{
			Method:        "ddim",
			Steps:         20,  // 增加步数以获得更好质量
			GuidanceScale: 7.5, // 标准引导比例
			Eta:           0.0,
			Seed:          -1, // 随机种子
		},
		LoraID:    -1,
		InitImage: initImg,
		Strength:  0.8, // 80%强度用于img2img
	}

	// 生成图像
	fmt.Println("Generating image...")
	result, err := imageGen.Img2Img(initImg, prompts[0], config)
	if err != nil {
		fmt.Printf("Failed to generate image: %v\n", err)
		return
	}
	defer result.Free()

	outputPath := strconv.Itoa(int(time.Now().Unix())) + ".jpeg"
	err = result.Save(outputPath)
	if err != nil {
		fmt.Printf("Failed to save image: %v\n", err)
	} else {
		fmt.Printf("Image-to-image generation completed! Image saved to: %s\n", outputPath)
	}
}

// createTestImage 创建一个简单的测试图像
func createTestImage() (*nexa_sdk.Image, error) {
	// 首先尝试从命令行指定的图像文件
	if image != "" {
		img, err := nexa_sdk.NewImage(image)
		if err == nil {
			return &img, nil
		}
		fmt.Printf("Warning: Failed to load specified image %s: %v\n", image, err)
	}

	// 尝试从默认测试文件创建图像
	testImagePath := "test_input.png"
	img, err := nexa_sdk.NewImage(testImagePath)
	if err != nil {
		return nil, fmt.Errorf("test image file not found: %s. Please provide a test image file using -i flag or place a test_input.png file in the current directory", testImagePath)
	}
	return &img, nil
}
