package lib

import (
	"fmt"
	"os"
)

type ListCommand struct {
	Verbose bool `short:"v" help:"Show detailed collection info"`
}

func (cmd *ListCommand) Run(client *QdrantClient) error {
	var resp QdrantResponse[CollectionsList]
	err := client.Request("GET", "/collections", nil, &resp)
	if err != nil {
		return err
	}

	if len(resp.Result.Collections) == 0 {
		PrintWarning("No collections found")
		return nil
	}

	PrintSection(fmt.Sprintf("Collections (%d)", len(resp.Result.Collections)))

	for _, col := range resp.Result.Collections {
		var infoResp QdrantResponse[CollectionInfo]
		err := client.Request("GET", "/collections/"+col.Name, nil, &infoResp)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s %s %s\n", red("●"), col.Name, dim("(error)"))
			continue
		}
		info := infoResp.Result

		if cmd.Verbose {
			PrintCollectionHeader(col.Name, info.PointsCount)
			PrintItem("    Vectors", info.VectorsCount)
			PrintItem("    Status", info.Status)
			PrintItem("    Segments", info.SegmentsCount)
			fmt.Println()
		} else {
			PrintCollectionHeader(col.Name, info.PointsCount)
		}
	}

	return nil
}

type StatsCommand struct{}

func (cmd *StatsCommand) Run(client *QdrantClient) error {
	// Get telemetry info
	var telemetry QdrantResponse[TelemetryInfo]
	err := client.Request("GET", "/telemetry", nil, &telemetry)

	PrintSection("Database Info")

	if err == nil && telemetry.Result.App.Version != "" {
		PrintItem("Version", fmt.Sprintf("Qdrant %s", telemetry.Result.App.Version))
	}

	// Get collections and sum up stats
	var resp QdrantResponse[CollectionsList]
	err = client.Request("GET", "/collections", nil, &resp)
	if err != nil {
		return err
	}

	PrintItem("Collections", len(resp.Result.Collections))
	fmt.Println()

	PrintSection("Collection Details")
	PrintTableHeader("Collection", "Points", "Vectors", "Status")

	var totalPoints, totalVectors int64
	for _, col := range resp.Result.Collections {
		var infoResp QdrantResponse[CollectionInfo]
		err := client.Request("GET", "/collections/"+col.Name, nil, &infoResp)
		if err != nil {
			PrintTableRow(col.Name, "error", "-", "-")
			continue
		}
		info := infoResp.Result
		totalPoints += info.PointsCount
		totalVectors += info.VectorsCount

		PrintTableRow(col.Name, info.PointsCount, info.VectorsCount, info.Status)
	}

	fmt.Println()
	PrintDivider()
	PrintTotal("Total Points", totalPoints)
	PrintTotal("Total Vectors", totalVectors)

	return nil
}
