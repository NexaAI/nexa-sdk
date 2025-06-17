package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/internal/store"
	nexa_sdk "github.com/NexaAI/nexa-sdk/nexa-sdk"
)

// TODO: remove test
func infer() *cobra.Command {
	inferCmd := &cobra.Command{}
	inferCmd.Use = "infer"

	inferCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)
	inferCmd.Run = func(cmd *cobra.Command, args []string) {
		s := store.NewStore()
		p := nexa_sdk.NewLLMPipeline()
		p.LoadModel(s.ModelfilePath(args[0]))

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

			e = p.GenerateStream(txt)
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
