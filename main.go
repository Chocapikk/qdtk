package main

import (
	"io"
	"log"
	"os"

	"github.com/alecthomas/kong"
	"github.com/chocapikk/qdtk/lib"
)

var App struct {
	List   lib.ListCommand   `cmd:"" help:"List collections"`
	Dump   lib.DumpCommand   `cmd:"" help:"Dump collection data"`
	Stats  lib.StatsCommand  `cmd:"" help:"Get database statistics"`
	Search lib.SearchCommand `cmd:"" help:"Search in collection"`
	Url    string            `required:"" name:"url" help:"Qdrant URL (e.g., https://qdrant.example.com or http://host:6333)"`
	Debug  bool              `short:"d" help:"Debug mode" default:"false"`
}

func main() {
	ctx := kong.Parse(&App,
		kong.Name("qdtk"),
		kong.Description("Qdrant ToolKit - Navigate and dump data from Qdrant vector databases"),
		kong.UsageOnError(),
	)

	client := lib.NewQdrantClient(App.Url)

	if App.Debug {
		client.LogOutput = os.Stderr
		log.SetOutput(os.Stderr)
	} else {
		client.LogOutput = io.Discard
		log.SetOutput(io.Discard)
	}

	err := ctx.Run(client)
	ctx.FatalIfErrorf(err)
}
