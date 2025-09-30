package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/charmbracelet/huh"
	"github.com/dustin/go-humanize"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/runner/internal/model_hub"
	"github.com/NexaAI/nexa-sdk/runner/internal/render"
	"github.com/NexaAI/nexa-sdk/runner/internal/store"
	"github.com/NexaAI/nexa-sdk/runner/internal/types"
)

var (
	modelHub  string
	localPath string
)

// pull creates a command to download and cache a model by name.
// Usage: nexa pull <model-name>
func pull() *cobra.Command {
	pullCmd := &cobra.Command{
		GroupID: "model",
		Use:     "pull <model-name>",

		Short: "Pull model from HuggingFace",
		Long:  "Download and cache a model by name.",
	}

	pullCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)

	pullCmd.Flags().SortFlags = false
	pullCmd.Flags().StringVarP(&modelHub, "model-hub", "", "", "specify model hub to use: volces|s3|hf|localfs")
	pullCmd.Flags().StringVarP(&localPath, "local-path", "", "", "[localfs] path to local directory")

	pullCmd.Run = func(cmd *cobra.Command, args []string) {
		pullModel(args[0])
	}

	return pullCmd
}

// remove creates a command to delete a cached model by name.
// Usage: nexa remove <model-name>
func remove() *cobra.Command {
	removeCmd := &cobra.Command{
		GroupID: "model",
		Use:     "remove <model-name>",
		Aliases: []string{"rm"},
		Short:   "Remove cached model",
		Long:    "Delete a cached model by name. This will remove the model files from the local cache.",
	}

	removeCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)

	removeCmd.Run = func(cmd *cobra.Command, args []string) {
		name := normalizeModelName(args[0])

		s := store.Get()
		e := s.Remove(name)
		if e != nil {
			fmt.Println(render.GetTheme().Error.Sprintf("✘  Failed to remove model: %s", name))
		} else {
			fmt.Printf("✔  Removed %s\n", name)
		}
	}

	return removeCmd
}

// clean creates a command to remove all cached models and free up storage.
// Usage: nexa clean
func clean() *cobra.Command {
	cleanCmd := &cobra.Command{
		GroupID: "model",
		Use:     "clean",
		Short:   "remove all cached models",
		Long:    "Remove all cached models and free up storage. This will delete all model files from the local cache.",
	}

	cleanCmd.Run = func(cmd *cobra.Command, args []string) {
		s := store.Get()
		c := s.Clean()
		fmt.Printf("✔  Removed %d models\n", c)
	}

	return cleanCmd
}

// list creates a command to display all cached models in a formatted table.
// Shows model names and their storage sizes.
// Usage: nexa list
func list() *cobra.Command {
	listCmd := &cobra.Command{
		GroupID: "model",
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all cached models",
		Long:    "Display all cached models in a formatted table, showing model names, types, and sizes.",
	}
	verbose := listCmd.Flags().BoolP("verbose", "v", false, "show detailed model info")

	listCmd.Run = func(cmd *cobra.Command, args []string) {
		s := store.Get()
		models, e := s.List()
		if e != nil {
			fmt.Println(e)
			return
		}

		// Create formatted table output
		tw := table.NewWriter()
		tw.SetOutputMirror(os.Stdout)
		tw.SetStyle(table.StyleLight)
		if *verbose {
			tw.AppendHeader(table.Row{"NAME", "SIZE", "PLUGIN", "TYPE", "QUANTS"})
			for _, model := range models {
				tw.AppendRow(table.Row{
					model.Name,
					humanize.IBytes(uint64(model.GetSize())),
					model.PluginId,
					model.ModelType,
					strings.Join(func() []string {
						quants := make([]string, 0)
						for q := range model.ModelFile {
							if model.ModelFile[q].Downloaded {
								quants = append(quants, q)
							}
						}
						slices.Sort(quants)
						return quants
					}(), ","),
				})
			}
		} else {
			tw.AppendHeader(table.Row{"NAME", "SIZE", "QUANTS"})
			for _, model := range models {
				tw.AppendRow(table.Row{model.Name, humanize.IBytes(uint64(model.GetSize())), strings.Join(func() []string {
					quants := make([]string, 0)
					for q := range model.ModelFile {
						if model.ModelFile[q].Downloaded && q != "N/A" {
							quants = append(quants, q)
						}
					}
					return quants
				}(), ",")})
			}
		}
		tw.Render()
	}

	return listCmd
}

// pull

func pullModel(name string) error {
	name = normalizeModelName(name)

	s := store.Get()

	mf, err := s.GetManifest(name)
	if err == nil {
		downloaded := true
		for _, f := range mf.ModelFile {
			if !f.Downloaded {
				downloaded = false
				break
			}
		}

		if downloaded {
			fmt.Println(render.GetTheme().Info.Sprint("Already downloaded all quant"))
			return nil
		}
	}

	// specify model hub
	if localPath != "" && modelHub == "" {
		modelHub = "localfs"
	}
	if modelHub != "" {
		switch strings.ToLower(modelHub) {
		case "volces":
			model_hub.SetHub(model_hub.NewVolces())
		case "s3":
			model_hub.SetHub(model_hub.NewS3())
		case "hf", "huggingface":
			model_hub.SetHub(model_hub.NewHuggingFace())
		case "local", "localfs":
			if localPath == "" {
				return fmt.Errorf("local path is required for localfs model hub")
			}
			model_hub.SetHub(model_hub.NewLocalFS(localPath))
		default:
			return fmt.Errorf("unknown model hub: %s", modelHub)
		}
	}

	spin := render.NewSpinner("download manifest from: " + name)
	spin.Start()
	files, hmf, err := model_hub.ModelInfo(context.TODO(), name)
	spin.Stop()
	if err != nil {
		fmt.Println(render.GetTheme().Error.Sprintf("Get ModelInfo error: %s", err))
		return err
	}

	if hmf != nil && !isValidVersion(hmf.MinSDKVersion) {
		fmt.Println(render.GetTheme().Error.Sprintf("Model requires NexaSDK version %s or higher. Please upgrade your NexaSDK CLI.", hmf.MinSDKVersion))
		return fmt.Errorf("model requires higher version")
	}

	if mf != nil {
		newManifest, err := chooseQuantFiles(*mf)
		if err != nil {
			return err
		}
		pgCh, errCh := s.PullExtraQuant(context.TODO(), *mf, *newManifest)
		bar := render.NewProgressBar(newManifest.GetSize()-mf.GetSize(), "downloading")

		for pg := range pgCh {
			bar.Set(pg.TotalDownloaded)
		}
		bar.Exit()

		for err := range errCh {
			bar.Clear()
			fmt.Println(render.GetTheme().Error.Sprintf("Error: %s", err))
		}
	} else {
		var manifest types.ModelManifest

		if hmf != nil {
			manifest.ModelName = hmf.ModelName
			manifest.PluginId = hmf.PluginId
			manifest.DeviceId = hmf.DeviceId
			manifest.ModelType = hmf.ModelType
			manifest.MinSDKVersion = hmf.MinSDKVersion
		}

		if manifest.ModelName == "" {
			manifest.ModelName = name
		}
		if manifest.PluginId == "" {
			manifest.PluginId = choosePluginId(name)
		}
		if manifest.ModelType == "" {
			if ctype, err := chooseModelType(); err != nil {
				fmt.Println(render.GetTheme().Error.Sprintf("Error: %s", err))
				return err
			} else {
				manifest.ModelType = ctype
			}
		}

		err := chooseFiles(name, files, &manifest)
		if err != nil {
			fmt.Println(render.GetTheme().Error.Sprintf("Error: %s", err))
			return err
		}

		// TODO: replace with go-pretty
		pgCh, errCh := s.Pull(context.TODO(), manifest)
		bar := render.NewProgressBar(manifest.GetSize(), "downloading")

		for pg := range pgCh {
			bar.Set(pg.TotalDownloaded)
		}
		bar.Exit()

		for err := range errCh {
			bar.Clear()
			fmt.Println(render.GetTheme().Error.Sprintf("Error: %s", err))
		}
	}

	return nil
}

// =============== quant name parse ===============
var quantRegix = regexp.MustCompile(`(` + strings.Join([]string{
	"[fF][pP][0-9]+",                 // FP32, FP16, FP64
	"[fF][0-9]+",                     // F64, F32, F16
	"[iI][0-9]+",                     // I64, I32, I16, I8
	"[qQ][0-9]+(_[A-Za-z0-9]+)*",     // Q8_0, Q8_1, Q8_K, Q6_K, Q5_0, Q5_1, Q5_K, Q4_0, Q4_1, Q4_K, Q3_K, Q2_K
	"[iI][qQ][0-9]+(_[A-Za-z0-9]+)*", // IQ4_NL, IQ4_XS, IQ3_S, IQ3_XXS, IQ2_XXS, IQ2_S, IQ2_XS, IQ1_S, IQ1_M
	"[bB][fF][0-9]+",                 // BF16
	"[0-9]+[bB][iI][tT]",             // 1bit, 2bit, 3bit, 4bit, 16bit, 1BIT, 16Bit, etc.
}, "|") + `)`)

func getQuant(name string) string {
	quant := strings.ToUpper(quantRegix.FindString(name))
	if quant == "" {
		quant = "N/A"
	}
	return quant
}

func choosePluginId(name string) string {
	switch {
	case strings.Contains(name, "mlx") || strings.Contains(name, "sdxl-turbo"):
		return "mlx"
	case strings.Contains(name, "gemma-3n-cuda"):
		return "nexa_cuda_ort_llama_cpp"
	case strings.Contains(name, "gemma-3n"):
		return "nexa_dml_llama_cpp"
	case strings.Contains(name, "Prefect-illustrious") || strings.Contains(name, "sdxl-base"):
		return "nexa_dml"
		// return "nexa_cuda"
	default:
		return "llama_cpp"
	}

}

func chooseModelType() (types.ModelType, error) {
	var modelType types.ModelType
	if err := huh.NewSelect[types.ModelType]().
		Title("Choose Model Type").
		Options(huh.NewOptions(
			types.ModelTypeLLM, types.ModelTypeVLM, types.ModelTypeEmbedder, types.ModelTypeReranker,
			types.ModelTypeASR, types.ModelTypeTTS, types.ModelTypeCV, types.ModelTypeImageGen)...).
		Value(&modelType).
		Run(); err != nil {
		return "", err
	}
	return modelType, nil
}

var partRegex = regexp.MustCompile(`-\d+-of-\d+\.gguf$`)

func chooseFiles(name string, files []model_hub.ModelFileInfo, res *types.ModelManifest) (err error) {
	if len(files) == 0 {
		err = fmt.Errorf("repo is empty")
		return
	}

	res.Name = name
	res.ModelFile = make(map[string]types.ModelFileInfo)

	// check gguf
	var mmprojs []model_hub.ModelFileInfo
	var tokenizers []model_hub.ModelFileInfo
	var onnxFiles []model_hub.ModelFileInfo
	var nexaFiles []model_hub.ModelFileInfo
	var npyFiles []model_hub.ModelFileInfo
	ggufs := make(map[string][]model_hub.ModelFileInfo) // key is gguf name without part
	// qwen2.5-7b-instruct-q8_0-00003-of-00003.gguf original name is qwen2.5-7b-instruct-q8_0 *d-of-*d like this

	for _, file := range files {
		name := strings.ToLower(file.Name)
		if strings.HasSuffix(name, ".gguf") {
			if strings.HasPrefix(name, "mmproj") {
				mmprojs = append(mmprojs, file)
			} else {
				name := partRegex.ReplaceAllString(file.Name, "")
				ggufs[name] = append(ggufs[name], file)
			}
		} else if strings.HasSuffix(name, "tokenizer.json") {
			tokenizers = append(tokenizers, file)
		} else if strings.HasSuffix(name, ".onnx") {
			onnxFiles = append(onnxFiles, file)
		} else if strings.HasSuffix(name, ".nexa") {
			nexaFiles = append(nexaFiles, file)
		} else if strings.HasSuffix(name, ".npy") {
			npyFiles = append(npyFiles, file)
		}
	}

	// choose model file
	if len(ggufs) > 0 {
		// detect gguf
		if len(ggufs) == 1 {
			// single quant
			fileInfo := types.ModelFileInfo{}
			for name, gguf := range ggufs {
				fileInfo.Name = gguf[0].Name
				fileInfo.Size = gguf[0].Size
				fileInfo.Downloaded = true
				res.ModelFile[getQuant(name)] = fileInfo
				// other fragments
				for _, file := range gguf[1:] {
					res.ExtraFiles = append(res.ExtraFiles, types.ModelFileInfo{
						Name:       file.Name,
						Downloaded: true,
						Size:       file.Size,
					})
				}
			}

		} else {
			// interactive choose
			var file string

			// sort key by quant
			ggufNames := make([]string, 0, len(ggufs))
			for k := range ggufs {
				ggufNames = append(ggufNames, k)

				if file == "" {
					file = k
					continue
				}

				// prefer Q4_K_M, Q4_0, Q8_0
				kq := getQuant(k)
				fq := getQuant(file)
				sortKey := []string{"Q8_0", "Q4_0", "Q4_K_M"}
				if slices.Index(sortKey, kq) > slices.Index(sortKey, fq) {
					file = k
				}
			}
			sort.Slice(ggufNames, func(i, j int) bool {
				return sumSize(ggufs[ggufNames[i]]) > sumSize(ggufs[ggufNames[j]])
			})

			var options []huh.Option[string]
			for _, ggufName := range ggufNames {
				fmtStr := "%-10s [%7s]"
				if ggufName == file {
					fmtStr += " (default)"
				}
				options = append(options, huh.NewOption(
					fmt.Sprintf(fmtStr, getQuant(ggufName), humanize.IBytes(uint64(sumSize(ggufs[ggufName])))),
					ggufName,
				))
			}

			if err = huh.NewSelect[string]().
				Title("Choose a quant version to download").
				Options(options...).
				Value(&file).
				Run(); err != nil {
				return err
			}

			for k, gguf := range ggufs {
				downloaded := k == file
				// sort files by name
				sort.Slice(gguf, func(i, j int) bool {
					return gguf[i].Name < gguf[j].Name
				})
				res.ModelFile[getQuant(k)] = types.ModelFileInfo{
					Name:       gguf[0].Name,
					Downloaded: downloaded,
					Size:       sumSize(ggufs[k]),
				}
				for _, file := range ggufs[k][1:] {
					res.ExtraFiles = append(res.ExtraFiles, types.ModelFileInfo{
						Name:       file.Name,
						Downloaded: downloaded,
						Size:       file.Size,
					})
				}
			}
		}

		// detect mmproj
		switch len(mmprojs) {
		case 0:
			// fallback to onnx file as mmproj if no regular mmproj found and exactly one onnx file exists
			if len(onnxFiles) == 1 {
				res.MMProjFile.Name = onnxFiles[0].Name
				res.MMProjFile.Size = onnxFiles[0].Size
				res.MMProjFile.Downloaded = true
			} else if len(nexaFiles) == 1 {
				// fallback to nexa file as mmproj if no onnx file and exactly one nexa file exists
				res.MMProjFile.Name = nexaFiles[0].Name
				res.MMProjFile.Size = nexaFiles[0].Size
				res.MMProjFile.Downloaded = true
			}
		case 1:
			res.MMProjFile.Name = mmprojs[0].Name
			res.MMProjFile.Size = mmprojs[0].Size
			res.MMProjFile.Downloaded = true

		default:
			// match biggest
			var file model_hub.ModelFileInfo
			for _, mmproj := range mmprojs {
				if mmproj.Size > file.Size {
					file = mmproj
				}
			}

			res.MMProjFile.Name = file.Name
			res.MMProjFile.Size = file.Size
			res.MMProjFile.Downloaded = true
		}

		// detect tokenizer for gguf models
		switch len(tokenizers) {
		case 0:
			// No tokenizer file found - skip
		case 1:
			res.TokenizerFile.Name = tokenizers[0].Name
			res.TokenizerFile.Size = tokenizers[0].Size
			res.TokenizerFile.Downloaded = true

		default:
			return fmt.Errorf("multiple tokenizer files found: %v. Expected exactly one tokenizer file", tokenizers)
		}

		// Always include .nexa files as extra files when gguf is the main model, except if used as mmproj
		for _, nexaFile := range nexaFiles {
			// Skip if this nexa file is being used as mmproj
			if res.MMProjFile.Name != nexaFile.Name {
				res.ExtraFiles = append(res.ExtraFiles, types.ModelFileInfo{
					Name:       nexaFile.Name,
					Downloaded: true,
					Size:       nexaFile.Size,
				})
			}
		}

		// Always include .npy files as extra files when gguf is the main model
		for _, npyFile := range npyFiles {
			res.ExtraFiles = append(res.ExtraFiles, types.ModelFileInfo{
				Name:       npyFile.Name,
				Downloaded: true,
				Size:       npyFile.Size,
			})
		}

	} else {
		// mlx

		// quant
		quant := getQuant(name)
		if quant == "N/A" {
			if q, err := model_hub.GetFileContent(context.TODO(), name, "config.json"); err != nil {
			} else if b, err := sonic.Get(q, "quantization_config", "bits"); err != nil {
			} else if q, err := b.Float64(); err != nil {
			} else {
				quant = fmt.Sprintf("%dBIT", uint32(q))
			}
		}

		// detect main model file
		isSupportedModelFile := func(filename string) bool {
			lower := strings.ToLower(filename)
			return strings.HasSuffix(lower, "safetensors") ||
				strings.HasSuffix(lower, "npz") ||
				strings.HasSuffix(lower, "nexa")
		}

		// First pass: prefer non-nested supported files (not in subdirectories)
		for _, file := range files {
			if isSupportedModelFile(file.Name) && !strings.Contains(file.Name, "/") {
				res.ModelFile[quant] = types.ModelFileInfo{Name: file.Name, Size: file.Size}
				break
			}
		}

		// Second pass: if no non-nested file found, fall back to any supported file
		if res.ModelFile[quant].Name == "" {
			for _, file := range files {
				if isSupportedModelFile(file.Name) {
					res.ModelFile[quant] = types.ModelFileInfo{Name: file.Name, Size: file.Size}
					break
				}
			}
		}

		// add other files to ExtraFiles
		for _, file := range files {
			if file.Name != res.ModelFile[quant].Name {
				res.ExtraFiles = append(res.ExtraFiles, types.ModelFileInfo{Name: file.Name, Size: file.Size})
			}
		}

		// fallback to first file
		if res.ModelFile[quant].Name == "" {
			res.ModelFile[quant] = types.ModelFileInfo{Name: files[0].Name, Size: files[0].Size}
			res.ExtraFiles = res.ExtraFiles[1:]
		}

		res.ModelFile[quant] = types.ModelFileInfo{
			Name:       res.ModelFile[quant].Name,
			Downloaded: true,
			Size:       res.ModelFile[quant].Size,
		}
		for i, v := range res.ExtraFiles {
			res.ExtraFiles[i] = types.ModelFileInfo{
				Name:       v.Name,
				Downloaded: true,
				Size:       v.Size,
			}
		}
	}

	return
}

func chooseQuantFiles(old types.ModelManifest) (*types.ModelManifest, error) {
	var mf types.ModelManifest
	d, _ := sonic.Marshal(old)
	sonic.Unmarshal(d, &mf)

	// sort key by quant
	ggufQuants := make([]string, 0, len(mf.ModelFile))
	for k := range mf.ModelFile {
		ggufQuants = append(ggufQuants, k)
	}
	sort.Slice(ggufQuants, func(i, j int) bool {
		return mf.ModelFile[ggufQuants[i]].Size > mf.ModelFile[ggufQuants[j]].Size
	})

	options := make([]huh.Option[string], 0, len(mf.ModelFile))
	for _, q := range ggufQuants {
		m := mf.ModelFile[q]
		if m.Downloaded {
			continue
		}
		options = append(options, huh.NewOption(
			fmt.Sprintf("%-10s [%7s]", q, humanize.IBytes(uint64(m.Size))), q,
		))
	}

	var quant string
	if err := huh.NewSelect[string]().
		Title("Choose a quant version to download").
		Options(options...).
		Value(&quant).
		Run(); err != nil {
		return nil, err
	}

	mf.ModelFile[quant] = types.ModelFileInfo{
		Name:       mf.ModelFile[quant].Name,
		Downloaded: true,
		Size:       mf.ModelFile[quant].Size,
	}

	// other fragments
	file := mf.ModelFile[quant].Name
	ggufName := partRegex.ReplaceAllString(file, "")
	for i, f := range mf.ExtraFiles {
		if ggufName == partRegex.ReplaceAllString(f.Name, "") {
			mf.ExtraFiles[i] = types.ModelFileInfo{
				Name:       f.Name,
				Downloaded: true,
			}
		}

	}

	return &mf, nil
}

func sumSize(files []model_hub.ModelFileInfo) int64 {
	var size int64
	for _, f := range files {
		size += f.Size
	}
	return size
}
