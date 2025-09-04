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

// pull creates a command to download and cache a model by name.
// Usage: nexa pull <model-name>
func pull() *cobra.Command {
	pullCmd := &cobra.Command{}
	pullCmd.Use = "pull <model-name>"

	pullCmd.Short = "Pull model from HuggingFace"
	pullCmd.Long = "Download and cache a model by name."

	pullCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)

	pullCmd.Run = func(cmd *cobra.Command, args []string) {
		name := normalizeModelName(args[0])

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
				return
			}
		}

		spin := render.NewSpinner("download manifest from: " + name)
		spin.Start()
		files, hmf, err := model_hub.ModelInfo(context.TODO(), name)
		spin.Stop()
		if err != nil {
			fmt.Println(render.GetTheme().Error.Sprintf("Download manifest error: %s", err))
			return
		}

		if mf != nil {
			newManifest, err := chooseQuantFiles(*mf)
			if err != nil {
				return
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
				manifest.PluginId = hmf.PluginId
				manifest.ModelType = hmf.ModelType
			}

			if manifest.PluginId == "" {
				manifest.PluginId = choosePluginId(name)
			}
			if manifest.ModelType == "" {
				if ctype, err := chooseModelType(); err != nil {
					return
				} else {
					manifest.ModelType = ctype
				}
			}

			err := chooseFiles(name, files, &manifest)
			if err != nil {
				return
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
	}

	return pullCmd
}

// remove creates a command to delete a cached model by name.
// Usage: nexa remove <model-name>
func remove() *cobra.Command {
	removeCmd := &cobra.Command{
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
			fmt.Println(e)
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
		Use:   "clean",
		Short: "remove all cached models",
		Long:  "Remove all cached models and free up storage. This will delete all model files from the local cache.",
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
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all cached models",
		Long:    "Display all cached models in a formatted table, showing model names, types, and sizes.",
	}

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
		tw.AppendHeader(table.Row{"NAME", "SIZE"})
		for _, model := range models {
			tw.AppendRow(table.Row{model.Name, humanize.IBytes(uint64(model.GetSize()))})
		}
		tw.Render()
	}

	return listCmd
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

// order big to small
func quantGreaterThan(a, b string, order []string) bool {
	// empty
	if a == "" || b == "" {
		return a != ""
	}

	a = strings.ToUpper(a)
	b = strings.ToUpper(b)

	// same
	if a == b {
		return false
	}

	// order
	ca := slices.Index(order, a)
	cb := slices.Index(order, b)
	if ca >= 0 && cb >= 0 {
		return ca < cb
	} else if ca >= 0 || cb >= 0 {
		return ca >= 0
	}

	// normal
	if a[0] == b[0] {
		return a > b
	} else {
		return a[0] == 'F'
	}
}

func choosePluginId(name string) string {
	switch {
	case strings.Contains(name, "mlx"):
		return "mlx"
	case strings.Contains(name, "gemma-3n-cuda"):
		return "nexa_cuda_ort_llama_cpp"
	case strings.Contains(name, "gemma-3n"):
		return "nexa_dml_llama_cpp"
	case strings.Contains(name, "prefect-illustrious") || strings.Contains(name, "sdxl-base"):
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

type ggufEntireInfo struct {
	Quant string
	Size  int64
}

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
	ggufFragments := make(map[string][]model_hub.ModelFileInfo)
	// qwen2.5-7b-instruct-q8_0-00003-of-00003.gguf original name is qwen2.5-7b-instruct-q8_0 *d-of-*d like this
	for _, file := range files {
		lower := strings.ToLower(file.Name)
		if strings.HasSuffix(lower, ".gguf") {
			if strings.HasPrefix(lower, "mmproj") {
				mmprojs = append(mmprojs, file)
			} else {
				name := partRegex.ReplaceAllString(file.Name, "")
				ggufFragments[name] = append(ggufFragments[name], file)
			}
		} else if strings.HasSuffix(lower, "tokenizer.json") {
			tokenizers = append(tokenizers, file)
		} else if strings.HasSuffix(lower, ".onnx") || strings.HasSuffix(lower, ".nexa") {
			onnxFiles = append(onnxFiles, file)
		}
	}

	ggufEntires := make(map[string]ggufEntireInfo, len(ggufFragments))
	for gguf := range ggufFragments {
		var size int64
		for _, f := range ggufFragments[gguf] {
			size += f.Size
		}
		quant := strings.ToUpper(quantRegix.FindString(gguf))
		if quant == "" {
			quant = "N/A"
		}
		ggufEntires[gguf] = ggufEntireInfo{Quant: quant, Size: size}
	}

	// choose model file
	if len(ggufEntires) > 0 {
		// detect gguf
		if len(ggufEntires) == 1 {
			// single quant
			fileInfo := types.ModelFileInfo{}
			for name, ggufEntire := range ggufEntires {
				fileInfo.Name = name
				fileInfo.Size = ggufEntire.Size
				fileInfo.Downloaded = true
				res.ModelFile[ggufEntire.Quant] = fileInfo
			}

		} else {
			// interactive choose

			// sort key by quant
			ggufEntireNames := make([]string, 0, len(ggufEntires))
			for k := range ggufEntires {
				ggufEntireNames = append(ggufEntireNames, k)
			}
			sort.Slice(ggufEntireNames, func(i, j int) bool {
				return quantGreaterThan(
					ggufEntires[ggufEntireNames[i]].Quant,
					ggufEntires[ggufEntireNames[j]].Quant,
					[]string{"Q4_K_M", "Q4_0", "Q8_0"})
			})

			var options []huh.Option[string]
			for i, ggufEntireName := range ggufEntireNames {
				fmtStr := "%-10s [%7s]"
				if i == 0 {
					fmtStr += " (default)"
				}
				options = append(options, huh.NewOption(
					fmt.Sprintf("%-10s [%7s]",
						ggufEntires[ggufEntireName].Quant,
						humanize.IBytes(uint64(ggufEntires[ggufEntireName].Size))),
					ggufEntireName,
				))
			}

			var file string
			if err = huh.NewSelect[string]().
				Title("Choose a quant version to download").
				Options(options...).
				Value(&file).
				Run(); err != nil {
				return err
			}

			for k, ggufFragment := range ggufFragments {
				downloaded := k == file
				quant := strings.ToUpper(quantRegix.FindString(k))
				// sort files by name
				sort.Slice(ggufFragment, func(i, j int) bool {
					return ggufFragment[i].Name < ggufFragment[j].Name
				})
				res.ModelFile[quant] = types.ModelFileInfo{
					Name:       ggufFragment[0].Name,
					Downloaded: downloaded,
					Size:       ggufEntires[k].Size,
				}
				for _, file := range ggufFragments[k][1:] {
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
		case 1:
			res.MMProjFile.Name = mmprojs[0].Name
			res.MMProjFile.Size = mmprojs[0].Size
			res.MMProjFile.Downloaded = true

		default:
			// Get mmproj file sizes for display
			if err != nil {
				fmt.Println(render.GetTheme().Error.Sprintf("get filesize error: %s", err))
				return err
			}

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

		// fallback to onnx file as mmproj if no regular mmproj found and exactly one onnx file exists
		if res.MMProjFile.Name == "" && len(onnxFiles) == 1 {
			res.MMProjFile.Name = onnxFiles[0].Name
			res.MMProjFile.Size = onnxFiles[0].Size
			res.MMProjFile.Downloaded = true
		}

		// detect tokenizer - only if both gguf and onnx files are found - specifically for gemma 3n in ort-llama-cpp case
		if len(onnxFiles) > 0 {
			switch len(tokenizers) {
			case 0:
				// No tokenizer file found - skip
			case 1:
				res.TokenizerFile.Name = tokenizers[0].Name
				if err != nil {
					fmt.Println(render.GetTheme().Error.Sprintf("get filesize error: [%s] %s", tokenizers[0], err))
					return err
				}
				res.TokenizerFile.Size = tokenizers[0].Size
				res.TokenizerFile.Downloaded = true

			default:
				return fmt.Errorf("multiple tokenizer files found: %v. Expected exactly one tokenizer file", tokenizers)
			}
		}
	} else {
		// other format

		// quant
		var quant string
		if q := strings.ToUpper(quantRegix.FindString(name)); q != "" {
			quant = q
		} else if q, err := model_hub.GetFileContent(context.TODO(), name, "config.json"); err != nil {
			quant = "N/A"
		} else if b, err := sonic.Get(q, "quantization_config", "bits"); err != nil {
			quant = "N/A"
		} else if q, err := b.Float64(); err != nil {
			quant = "N/A"
		} else {
			quant = fmt.Sprintf("%dBIT", uint32(q))
		}

		// detect main model file
		// add other files
		for _, file := range files {
			if res.ModelFile[quant].Name == "" {
				lower := strings.ToLower(file.Name)
				if strings.HasSuffix(lower, "safetensors") || strings.HasSuffix(lower, "npz") {
					res.ModelFile[quant] = types.ModelFileInfo{Name: file.Name, Size: file.Size}
					continue
				}
			}
			res.ExtraFiles = append(res.ExtraFiles, types.ModelFileInfo{Name: file.Name, Size: file.Size})
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
	// Find the longest quant name for alignment
	options := make([]huh.Option[string], 0, len(mf.ModelFile))
	for q, m := range mf.ModelFile {
		if !m.Downloaded {
			options = append(options, huh.NewOption(
				fmt.Sprintf("%-10s [%7s]", q, humanize.IBytes(uint64(m.Size))), q,
			))
		}
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

	file := mf.ModelFile[quant].Name
	ggufName := partRegex.ReplaceAllString(file, "")
	for i, f := range mf.ExtraFiles {
		if ggufName == partRegex.ReplaceAllString(file, "") {
			mf.ExtraFiles[i] = types.ModelFileInfo{
				Name:       f.Name,
				Downloaded: true,
			}
		}

	}

	return &mf, nil
}
