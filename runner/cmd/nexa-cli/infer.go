package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/internal/store"
	nexa_sdk "github.com/NexaAI/nexa-sdk/nexa-sdk"
)

func infer() *cobra.Command {
	inferCmd := &cobra.Command{}
	inferCmd.Use = "infer"

	inferCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)
	inferCmd.Run = func(cmd *cobra.Command, args []string) {
		s := store.NewStore()
		nexa_sdk.Init()

		p := nexa_sdk.NewLLM(s.ModelfilePath(args[0]), nil, 4096, nil)

		var history []nexa_sdk.ChatMessage
		var lastLen int

		r := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("\nEnter text: ")

			txt, e := r.ReadString('\n')
			if e != nil {
				if e == io.EOF {
					fmt.Println()
					break
				}
				fmt.Printf("ReadString Error: %s\n", e)
				break
			}
			txt = strings.TrimSpace(txt)
			if txt == "/exit" {
				break
			}
			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleUser, Content: txt})

			formatted, e := p.ApplyChatTemplate(history)
			if e != nil {
				fmt.Printf("ApplyChatTemplat Error: %s\n", e)
				break
			}

			start := time.Now()
			var count int
			var full strings.Builder

			fmt.Print("\033[33m")
			dataCh, errCh := p.GenerateStream(context.Background(), formatted[lastLen:])
			for r := range dataCh {
				full.WriteString(r)
				fmt.Print(r)
				count++
			}
			fmt.Print("\033[0m\n")

			e, ok := <-errCh
			if ok {
				fmt.Printf("GenerateStream Error: %s\n", e)
				return
			}

			duration := time.Since(start).Seconds()
			fmt.Print("\033[34m")
			fmt.Printf("\nGenerate %d token in %f s, speed is %f token/s\n",
				count,
				duration,
				float64(count)/duration)

			fmt.Print("\033[0m")

			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleAssistant, Content: full.String()})

			formatted, e = p.ApplyChatTemplate(history)
			if e != nil {
				fmt.Printf("ApplyChatTemplat Error: %s\n", e)
				break
			}
			lastLen = len(formatted)
		}

		p.Destroy()
		nexa_sdk.DeInit()
	}
	return inferCmd
}
