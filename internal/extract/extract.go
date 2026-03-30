package extract

import "github.com/Necr00/xtract/internal/model"

// RunAllLayers runs extraction layers 1 through 7 on the given context.
func RunAllLayers(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	results = append(results, ExtractLayer1(ctx)...)
	results = append(results, ExtractLayer2(ctx)...)
	results = append(results, ExtractLayer3(ctx)...)
	results = append(results, ExtractLayer4(ctx)...)
	results = append(results, ExtractLayer5(ctx)...)
	results = append(results, ExtractLayer6(ctx)...)
	results = append(results, ExtractLayer7(ctx)...)
	return results
}
