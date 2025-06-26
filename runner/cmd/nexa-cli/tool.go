package main

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/internal/store"
	"github.com/NexaAI/nexa-sdk/nexa-sdk"
)

func embedding() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Use = "embedding <model-name> <prompt>"

	cmd.Args = cobra.MatchAll(cobra.ExactArgs(2), cobra.OnlyValidArgs)
	cmd.Run = func(cmd *cobra.Command, args []string) {
		s := store.NewStore()

		file, err := s.ModelfilePath(args[0])
		if err != nil {
			fmt.Println(text.FgRed.Sprintf("Error: %s", err))
			return
		}
		p := nexa_sdk.NewEmbedder(file, nil, nil)
		defer p.Destroy()

		res, err := p.Embed([]string{args[1]})
		if err != nil {
			fmt.Println(text.FgRed.Sprintf("Error: %s", err))
			return
		} else {
			for i := range res {
				fmt.Println(text.FgYellow.Sprintf("%f", res[i]))
			}
		}

	}

	return cmd
}

func reranking() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Use = "reranking <model-name> <tokenizer-path> <query> [text...]"

	cmd.Args = cobra.MatchAll(cobra.MinimumNArgs(4), cobra.OnlyValidArgs)
	cmd.Run = func(cmd *cobra.Command, args []string) {
		s := store.NewStore()

		file, err := s.ModelfilePath(args[0])
		if err != nil {
			fmt.Println(text.FgRed.Sprintf("Error: %s", err))
			return
		}
		p := nexa_sdk.NewReranker(file, args[1], nil)
		defer p.Destroy()

		res, err := p.Rerank(args[2], args[3:])
		if err != nil {
			fmt.Println(text.FgRed.Sprintf("Error: %s", err))
			return
		} else {
			for i := range res {
				fmt.Println(text.FgYellow.Sprintf("%f", res[i]))
			}
		}

	}

	return cmd
}

func tool() *cobra.Command {
	toolCmd := &cobra.Command{}
	toolCmd.Use = "tool <command>"

	toolCmd.AddCommand(embedding())
	toolCmd.AddCommand(reranking())
	return toolCmd
}
