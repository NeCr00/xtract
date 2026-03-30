package extract

import "github.com/NeCr00/xtract/internal/model"

// RunAllLayers runs extraction layers 1 through 7 on the given context.
func RunAllLayers(ctx *model.ExtractionContext) []model.Result {
	// Pre-allocate with reasonable estimate based on content size.
	// Typical: ~1 URL per 500 bytes of content.
	est := min(max(len(ctx.Content)/500, 64), 4096)
	results := make([]model.Result, 0, est)
	results = append(results, ExtractLayer1(ctx)...)
	results = append(results, ExtractLayer2(ctx)...)
	results = append(results, ExtractLayer3(ctx)...)
	results = append(results, ExtractLayer4(ctx)...)
	results = append(results, ExtractLayer5(ctx)...)
	results = append(results, ExtractLayer6(ctx)...)
	results = append(results, ExtractLayer7(ctx)...)
	return results
}
