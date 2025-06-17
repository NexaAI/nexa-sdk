package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	nexa_sdk "github.com/NexaAI/nexa-sdk/nexa-sdk"
)

// TODO: remove test
func infer() *cobra.Command {
	inferCmd := &cobra.Command{}
	inferCmd.Use = "infer"

	model := inferCmd.Flags().StringP("model", "m", "", "path to gguf model file")
	stream := inferCmd.Flags().BoolP("stream", "s", false, "enable stream mode")
	inferCmd.Run = func(cmd *cobra.Command, args []string) {
		p := nexa_sdk.NewLLMPipeline()
		p.LoadModel(*model)

		r := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("\nEnter text: ")

			txt, e := r.ReadString('\n')
			if e != nil {
				fmt.Printf("ReadString Error: %s\n", e)
				break
			}

			start := time.Now()
			var token_count int
			if *stream {
				e := p.GenerateStream(txt)
				if e != nil {
					fmt.Printf("GenerateStream Error: %s\n", e)
					break
				}

				token_count = 0
				fmt.Print("\033[33m")
				for {
					token, e := p.GenerateNextToken()
					if e != nil {
						fmt.Printf("GenerateNextToken Error: %s\n", e)
					}
					if token == "" {
						break
					}
					fmt.Print(token)
					token_count += 1
				}

			} else {
				fmt.Print("\033[33m")
				_, token_count, e = p.Generate(txt)
				if e != nil {
					fmt.Printf("Generat Error: %s\n", e)
					break
				}
			}

			duration := time.Since(start).Seconds()
			fmt.Print("\033[34m")
			fmt.Printf("\nGenerate %d token in %f s, speed is %f token/s\n",
				token_count,
				duration,
				float64(token_count)/duration)

			fmt.Print("\033[0m")
		}

		p.Close()
		p.Destroy()
	}
	return inferCmd
}
