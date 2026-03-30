package main

import (
	"fmt"
	"os"
	"time"

	"github.com/NeCr00/xtract/internal/extract"
	"github.com/NeCr00/xtract/internal/model"
)

func main() {
	data, _ := os.ReadFile(os.Args[1])
	content := string(data)
	ctx := &model.ExtractionContext{
		Content:  content,
		FileName: os.Args[1],
		FileType: "js",
		Lines:    model.NewLineIndex(content),
	}

	layers := []struct {
		name string
		fn   func(*model.ExtractionContext) []model.Result
	}{
		{"Layer1 (Regex)", extract.ExtractLayer1},
		{"Layer2 (AST)", extract.ExtractLayer2},
		{"Layer3 (Framework)", extract.ExtractLayer3},
		{"Layer4 (Config)", extract.ExtractLayer4},
		{"Layer5 (Infra)", extract.ExtractLayer5},
		{"Layer6 (Comments)", extract.ExtractLayer6},
		{"Layer7 (Decode)", extract.ExtractLayer7},
	}

	fmt.Printf("File: %s (%d bytes)\n\n", os.Args[1], len(content))

	totalURLs := 0
	totalTime := time.Duration(0)
	for _, l := range layers {
		start := time.Now()
		results := l.fn(ctx)
		elapsed := time.Since(start)
		totalTime += elapsed
		totalURLs += len(results)
		fmt.Printf("  %-25s %8.2fs  %5d URLs\n", l.name, elapsed.Seconds(), len(results))
	}
	fmt.Printf("\n  %-25s %8.2fs  %5d URLs\n", "TOTAL", totalTime.Seconds(), totalURLs)
}
