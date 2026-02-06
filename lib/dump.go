package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
)

type DumpCommand struct {
	Collection  string `required:"" short:"c" help:"Collection name (use '*' or 'all' for all collections)"`
	OutputFile  string `short:"o" help:"Output file (default: stdout)"`
	Limit       int    `short:"l" help:"Max documents to dump (0 = unlimited)" default:"0"`
	BatchSize   int    `short:"b" help:"Batch size for scrolling" default:"100"`
	WithVectors bool   `short:"v" help:"Include vectors in output" default:"false"`
	PayloadOnly bool   `short:"p" help:"Output only payload (no metadata)" default:"false"`
	NoProgress  bool   `help:"Disable progress bar" default:"false"`
	Quiet       bool   `short:"q" help:"Quiet mode (minimal output)" default:"false"`
}

func (cmd *DumpCommand) Run(client *QdrantClient) error {
	var outputWriter io.Writer
	var err error

	if cmd.OutputFile != "" {
		outputWriter, err = os.Create(cmd.OutputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outputWriter.(*os.File).Close()
		if !cmd.Quiet {
			PrintInfo(fmt.Sprintf("Output: %s", boldYellow(cmd.OutputFile)))
		}
	} else {
		outputWriter = os.Stdout
		cmd.NoProgress = true
		cmd.Quiet = true
	}

	if cmd.Collection == "*" || strings.ToLower(cmd.Collection) == "all" {
		return cmd.dumpAllCollections(client, outputWriter)
	}

	return cmd.dumpCollection(client, cmd.Collection, outputWriter)
}

func (cmd *DumpCommand) dumpAllCollections(client *QdrantClient, outputWriter io.Writer) error {
	var resp QdrantResponse[CollectionsList]
	err := client.Request("GET", "/collections", nil, &resp)
	if err != nil {
		return err
	}

	if !cmd.Quiet {
		PrintSection(fmt.Sprintf("Dumping %d collections", len(resp.Result.Collections)))
	}

	var totalDumped int64
	for _, col := range resp.Result.Collections {
		if !cmd.Quiet {
			fmt.Printf("\n  %s %s\n", boldCyan("▸"), boldYellow(col.Name))
		}
		err := cmd.dumpCollection(client, col.Name, outputWriter)
		if err != nil {
			if !cmd.Quiet {
				PrintError(fmt.Sprintf("Error: %v", err))
			}
		}
	}

	if !cmd.Quiet {
		fmt.Println()
		PrintSuccess(fmt.Sprintf("Dump completed: %d total points", totalDumped))
	}

	return nil
}

func (cmd *DumpCommand) dumpCollection(client *QdrantClient, collection string, outputWriter io.Writer) error {
	var infoResp QdrantResponse[CollectionInfo]
	err := client.Request("GET", "/collections/"+collection, nil, &infoResp)
	if err != nil {
		return fmt.Errorf("failed to get collection info: %w", err)
	}

	totalPoints := infoResp.Result.PointsCount
	if totalPoints == 0 {
		if !cmd.Quiet {
			fmt.Printf("    %s\n", dim("Empty collection, skipping"))
		}
		return nil
	}

	maxDump := totalPoints
	if cmd.Limit > 0 && int64(cmd.Limit) < totalPoints {
		maxDump = int64(cmd.Limit)
	}

	if !cmd.Quiet && !cmd.NoProgress {
		fmt.Printf("    %s %s  %s %s\n",
			dim("Points:"), cyan(totalPoints),
			dim("Dumping:"), cyan(maxDump))
	}

	jsonEncoder := json.NewEncoder(outputWriter)

	var bar *progressbar.ProgressBar
	if !cmd.NoProgress && !cmd.Quiet {
		bar = progressbar.NewOptions64(maxDump,
			progressbar.OptionSetDescription("    "),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetWidth(30),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "█",
				SaucerHead:    "▓",
				SaucerPadding: "░",
				BarStart:      "│",
				BarEnd:        "│",
			}),
			progressbar.OptionOnCompletion(func() { fmt.Fprint(os.Stderr, "\n") }),
		)
	}

	var offset *string
	var dumped int64

	for {
		scrollReq := ScrollRequest{
			Limit:       cmd.BatchSize,
			WithPayload: true,
			WithVector:  cmd.WithVectors,
			Offset:      offset,
		}

		var scrollResp QdrantResponse[ScrollResponse]
		err := client.Request("POST", "/collections/"+collection+"/points/scroll", scrollReq, &scrollResp)
		if err != nil {
			if strings.Contains(err.Error(), "timeout") {
				if !cmd.Quiet {
					PrintWarning("Timeout, retrying in 5s...")
				}
				time.Sleep(5 * time.Second)
				continue
			}
			return fmt.Errorf("scroll failed: %w", err)
		}

		if len(scrollResp.Result.Points) == 0 {
			break
		}

		for _, point := range scrollResp.Result.Points {
			if cmd.Limit > 0 && dumped >= int64(cmd.Limit) {
				break
			}

			var output interface{}
			if cmd.PayloadOnly {
				output = point.Payload
			} else {
				output = DumpPoint{
					Collection: collection,
					Id:         point.Id,
					Payload:    point.Payload,
					Vector:     point.Vector,
				}
			}

			if err := jsonEncoder.Encode(output); err != nil {
				return fmt.Errorf("failed to encode point: %w", err)
			}

			dumped++
			if bar != nil {
				bar.Add(1)
			}
		}

		if cmd.Limit > 0 && dumped >= int64(cmd.Limit) {
			break
		}

		offset = scrollResp.Result.NextPageOffset
		if offset == nil {
			break
		}
	}

	if bar != nil {
		bar.Finish()
	}

	if !cmd.Quiet && cmd.NoProgress {
		PrintSuccess(fmt.Sprintf("Dumped %d points", dumped))
	}

	return nil
}

type DumpPoint struct {
	Collection string                 `json:"_collection"`
	Id         interface{}            `json:"_id"`
	Payload    map[string]interface{} `json:"payload"`
	Vector     interface{}            `json:"vector,omitempty"`
}
