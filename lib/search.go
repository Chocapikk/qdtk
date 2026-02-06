package lib

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type SearchCommand struct {
	Collection string `required:"" short:"c" help:"Collection name"`
	Query      string `short:"q" help:"Text to search in payloads (searches all text fields)"`
	Field      string `short:"f" help:"Specific field to search in"`
	Limit      int    `short:"l" help:"Max results" default:"10"`
	Raw        bool   `short:"r" help:"Output raw JSON" default:"false"`
}

func (cmd *SearchCommand) Run(client *QdrantClient) error {
	if cmd.Query == "" {
		return cmd.scrollDocuments(client)
	}
	return cmd.searchPayload(client)
}

func (cmd *SearchCommand) scrollDocuments(client *QdrantClient) error {
	scrollReq := ScrollRequest{
		Limit:       cmd.Limit,
		WithPayload: true,
		WithVector:  false,
	}

	var scrollResp QdrantResponse[ScrollResponse]
	err := client.Request("POST", "/collections/"+cmd.Collection+"/points/scroll", scrollReq, &scrollResp)
	if err != nil {
		return err
	}

	return cmd.outputResults(scrollResp.Result.Points)
}

func (cmd *SearchCommand) searchPayload(client *QdrantClient) error {
	var allMatches []Point
	var offset *string
	batchSize := 100
	maxScroll := 10000

	searched := 0
	for searched < maxScroll && len(allMatches) < cmd.Limit {
		scrollReq := ScrollRequest{
			Limit:       batchSize,
			WithPayload: true,
			WithVector:  false,
			Offset:      offset,
		}

		var scrollResp QdrantResponse[ScrollResponse]
		err := client.Request("POST", "/collections/"+cmd.Collection+"/points/scroll", scrollReq, &scrollResp)
		if err != nil {
			return err
		}

		if len(scrollResp.Result.Points) == 0 {
			break
		}

		for _, point := range scrollResp.Result.Points {
			if cmd.matchesQuery(point) {
				allMatches = append(allMatches, point)
				if len(allMatches) >= cmd.Limit {
					break
				}
			}
		}

		searched += len(scrollResp.Result.Points)
		if scrollResp.Result.NextPageOffset == nil {
			break
		}
		offsetStr := fmt.Sprintf("%v", scrollResp.Result.NextPageOffset)
		offset = &offsetStr
		if offsetStr == "" || offsetStr == "<nil>" {
			break
		}
	}

	if !cmd.Raw {
		PrintInfo(fmt.Sprintf("Searched %d documents, found %d matches", searched, len(allMatches)))
		fmt.Println()
	}
	return cmd.outputResults(allMatches)
}

func (cmd *SearchCommand) matchesQuery(point Point) bool {
	query := strings.ToLower(cmd.Query)

	if cmd.Field != "" {
		if val, ok := point.Payload[cmd.Field]; ok {
			return containsQuery(val, query)
		}
		return false
	}

	for _, val := range point.Payload {
		if containsQuery(val, query) {
			return true
		}
	}
	return false
}

func containsQuery(val interface{}, query string) bool {
	switch v := val.(type) {
	case string:
		return strings.Contains(strings.ToLower(v), query)
	case map[string]interface{}:
		for _, subVal := range v {
			if containsQuery(subVal, query) {
				return true
			}
		}
	case []interface{}:
		for _, item := range v {
			if containsQuery(item, query) {
				return true
			}
		}
	}
	return false
}

func (cmd *SearchCommand) outputResults(points []Point) error {
	if cmd.Raw {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(points)
	}

	for i, point := range points {
		fmt.Printf("%s %s %v\n",
			boldCyan(fmt.Sprintf("[%d]", i+1)),
			dim("ID:"),
			yellow(point.Id))

		payloadJson, _ := json.MarshalIndent(point.Payload, "    ", "  ")
		fmt.Printf("    %s\n\n", string(payloadJson))
	}

	return nil
}
