package cmd

import (
	"fmt"
	"os"

	"github.com/johncoder/jot/internal/eval"
	"github.com/spf13/cobra"
)

var evalList bool
var evalAll bool

var evalCmd = &cobra.Command{
	Use:   "eval [file] [name]",
	Short: "Evaluate code blocks in markdown files",
	Long:  `Evaluate code blocks in markdown files using standards-compliant metadata links.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if evalList {
			if len(args) == 0 {
				fmt.Fprintln(os.Stderr, "No files specified.")
				return nil
			}
			for _, file := range args {
				fmt.Printf("%s:\n", file)
				err := eval.ListEvalBlocks(file)
				if err != nil {
					fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
				}
			}
			return nil
		}
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "No files specified.")
			return nil
		}
		
		file := args[0]
		var targetName string
		if len(args) > 1 {
			targetName = args[1]
		}
		
		var results []*eval.EvalResult
		var err error
		
		if targetName != "" {
			// Execute specific block by name
			results, err = eval.ExecuteEvaluableBlockByName(file, targetName)
		} else if evalAll {
			// Execute all blocks
			results, err = eval.ExecuteEvaluableBlocks(file)
		} else {
			return fmt.Errorf("please specify a block name or use --all to execute all blocks")
		}
		
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error executing blocks in %s: %v\n", file, err)
			return err
		}
		
		err = eval.UpdateMarkdownWithResults(file, results)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating results in %s: %v\n", file, err)
			return err
		}
		
		if targetName != "" {
			fmt.Printf("Updated result for block '%s' in %s\n", targetName, file)
		} else if evalAll {
			fmt.Printf("Updated results in %s\n", file)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(evalCmd)
	evalCmd.Flags().BoolVarP(&evalList, "list", "l", false, "List evaluable code blocks")
	evalCmd.Flags().BoolVarP(&evalAll, "all", "a", false, "Execute all evaluable code blocks")
}
