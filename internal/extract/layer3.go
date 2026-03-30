package extract

import (
	"regexp"
	"strings"

	"github.com/Necr00/xtract/internal/model"
)

// ExtractLayer3 runs all Layer 3 (Framework-Aware) techniques and returns results.
func ExtractLayer3(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	results = append(results, extractReactRouter(ctx)...)
	results = append(results, extractVueRouter(ctx)...)
	results = append(results, extractAngularRouter(ctx)...)
	results = append(results, extractNextJS(ctx)...)
	results = append(results, extractExpressRoutes(ctx)...)
	results = append(results, extractGraphQLOps(ctx)...)
	results = append(results, extractRESTInference(ctx)...)
	return results
}

// extractReactRouter detects React Router paths from <Route>, <Link>, <NavLink>,
// useNavigate, and route config arrays.
func extractReactRouter(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	patterns := []*regexp.Regexp{
		// <Route path="..." or <Route exact path="..."
		model.GetRegex(`<Route\s+(?:exact\s+)?path\s*=\s*["']([^"']+)["']`),
		// <Link to="..."
		model.GetRegex(`<Link\s+[^>]*to\s*=\s*["']([^"']+)["']`),
		// <NavLink to="..."
		model.GetRegex(`<NavLink\s+[^>]*to\s*=\s*["']([^"']+)["']`),
		// Generic to="..." in JSX context (adjacent to component-like tags)
		model.GetRegex(`\s+to\s*=\s*["'](/[^"']+)["']`),
		// useNavigate() with path argument: navigate('/path')
		model.GetRegex(`(?:useNavigate|navigate)\s*\(\s*["']([^"']+)["']`),
		// Route config array: {path: '/...'}
		model.GetRegex(`\{\s*path\s*:\s*["']([^"']+)["']`),
	}

	for _, re := range patterns {
		matches := re.FindAllStringSubmatchIndex(content, -1)
		for _, m := range matches {
			if len(m) >= 4 {
				path := content[m[2]:m[3]]
				results = append(results, model.Result{
					URL:             path,
					SourceFile:      ctx.FileName,
					SourceLine:      model.LineNumber(content, m[0]),
					DetectionMethod: "react_router",
					Category:        model.CatPageRoute,
					Confidence:      model.ConfHigh,
					TechniqueID:     34,
				})
			}
		}
	}

	return results
}

// extractVueRouter detects Vue Router paths from route configs, router-link,
// and $router.push calls.
func extractVueRouter(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	patterns := []*regexp.Regexp{
		// routes: [{path: '...'}] or {path: '...'} inside route arrays
		model.GetRegex(`(?:routes\s*:\s*\[[\s\S]*?)?\{\s*path\s*:\s*["']([^"']+)["']`),
		// <router-link to="..."
		model.GetRegex(`<router-link\s+[^>]*to\s*=\s*["']([^"']+)["']`),
		// <router-link :to="..."
		model.GetRegex(`<router-link\s+[^>]*:to\s*=\s*["']([^"']+)["']`),
		// this.$router.push('...')
		model.GetRegex(`this\.\$router\.push\s*\(\s*["']([^"']+)["']`),
		// router.push({path: '...'})
		model.GetRegex(`router\.push\s*\(\s*\{\s*path\s*:\s*["']([^"']+)["']`),
	}

	for _, re := range patterns {
		matches := re.FindAllStringSubmatchIndex(content, -1)
		for _, m := range matches {
			if len(m) >= 4 {
				path := content[m[2]:m[3]]
				results = append(results, model.Result{
					URL:             path,
					SourceFile:      ctx.FileName,
					SourceLine:      model.LineNumber(content, m[0]),
					DetectionMethod: "vue_router",
					Category:        model.CatPageRoute,
					Confidence:      model.ConfHigh,
					TechniqueID:     35,
				})
			}
		}
	}

	return results
}

// extractAngularRouter detects Angular Router paths from RouterModule,
// routerLink, and router.navigate calls.
func extractAngularRouter(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	patterns := []*regexp.Regexp{
		// RouterModule.forRoot([{path: '...'}]) or RouterModule.forChild
		model.GetRegex(`RouterModule\.for(?:Root|Child)\s*\(\s*\[[\s\S]*?\{\s*path\s*:\s*["']([^"']+)["']`),
		// routerLink="..."
		model.GetRegex(`routerLink\s*=\s*["']([^"']+)["']`),
		// [routerLink]="['/path']" or [routerLink]="['/path', param]"
		model.GetRegex(`\[routerLink\]\s*=\s*["']\s*\[\s*["']([^"']+)["']`),
		// this.router.navigate(['/path'])
		model.GetRegex(`this\.router\.navigate\s*\(\s*\[\s*["']([^"']+)["']`),
		// {path: '...', component: ...} in route config
		model.GetRegex(`\{\s*path\s*:\s*["']([^"']+)["']\s*,\s*component\s*:`),
	}

	for _, re := range patterns {
		matches := re.FindAllStringSubmatchIndex(content, -1)
		for _, m := range matches {
			if len(m) >= 4 {
				path := content[m[2]:m[3]]
				results = append(results, model.Result{
					URL:             path,
					SourceFile:      ctx.FileName,
					SourceLine:      model.LineNumber(content, m[0]),
					DetectionMethod: "angular_router",
					Category:        model.CatPageRoute,
					Confidence:      model.ConfHigh,
					TechniqueID:     36,
				})
			}
		}
	}

	return results
}

// extractNextJS detects Next.js routing patterns including <Link href>,
// useRouter().push, router.push, router.replace, and API route patterns.
func extractNextJS(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	patterns := []*regexp.Regexp{
		// <Link href="..."
		model.GetRegex(`<Link\s+[^>]*href\s*=\s*["']([^"']+)["']`),
		// useRouter().push('...') or router.push('...')
		model.GetRegex(`(?:useRouter\(\)\.push|router\.push)\s*\(\s*["']([^"']+)["']`),
		// router.replace('...')
		model.GetRegex(`router\.replace\s*\(\s*["']([^"']+)["']`),
		// Next.js API route patterns: /api/... in string contexts
		model.GetRegex(`["'](/api/[a-zA-Z0-9/_\-\[\]]+)["']`),
	}

	for _, re := range patterns {
		matches := re.FindAllStringSubmatchIndex(content, -1)
		for _, m := range matches {
			if len(m) >= 4 {
				path := content[m[2]:m[3]]
				cat := model.CatPageRoute
				if strings.HasPrefix(path, "/api/") {
					cat = model.CatAPIEndpoint
				}
				results = append(results, model.Result{
					URL:             path,
					SourceFile:      ctx.FileName,
					SourceLine:      model.LineNumber(content, m[0]),
					DetectionMethod: "nextjs",
					Category:        cat,
					Confidence:      model.ConfHigh,
					TechniqueID:     37,
				})
			}
		}
	}

	return results
}

// extractExpressRoutes detects Express.js route definitions including app and
// router method calls, extracting both the HTTP method and the path.
func extractExpressRoutes(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	re := model.GetRegex(`(?:app|router)\.(get|post|put|delete|patch|options|head|use)\s*\(\s*["']([^"']+)["']`)
	matches := re.FindAllStringSubmatchIndex(content, -1)
	for _, m := range matches {
		if len(m) >= 6 {
			method := strings.ToUpper(content[m[2]:m[3]])
			path := content[m[4]:m[5]]

			httpMethod := method
			if method == "USE" {
				httpMethod = "" // middleware, no specific HTTP method
			}

			results = append(results, model.Result{
				URL:             path,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(content, m[0]),
				DetectionMethod: "express_routes",
				HTTPMethod:      httpMethod,
				Category:        model.CatAPIEndpoint,
				Confidence:      model.ConfHigh,
				TechniqueID:     38,
			})
		}
	}

	return results
}

// extractGraphQLOps detects GraphQL queries, mutations, subscriptions,
// operation names, endpoint URLs, and gql tagged template literals.
func extractGraphQLOps(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	// Match query/mutation/subscription operation definitions
	opRe := model.GetRegex(`(?:query|mutation|subscription)\s+([A-Za-z_][A-Za-z0-9_]*)\s*(?:\([^)]*\)\s*)?\{`)
	opMatches := opRe.FindAllStringSubmatchIndex(content, -1)
	for _, m := range opMatches {
		if len(m) >= 4 {
			opName := content[m[2]:m[3]]
			results = append(results, model.Result{
				URL:             "graphql:" + opName,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(content, m[0]),
				DetectionMethod: "graphql_ops",
				Category:        model.CatAPIEndpoint,
				Confidence:      model.ConfMedium,
				TechniqueID:     39,
			})
		}
	}

	// Match anonymous query/mutation/subscription blocks
	anonRe := model.GetRegex(`(?:^|[^A-Za-z_])(query|mutation|subscription)\s*\{`)
	anonMatches := anonRe.FindAllStringSubmatchIndex(content, -1)
	for _, m := range anonMatches {
		if len(m) >= 4 {
			opType := content[m[2]:m[3]]
			results = append(results, model.Result{
				URL:             "graphql:" + opType,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(content, m[0]),
				DetectionMethod: "graphql_ops",
				Category:        model.CatAPIEndpoint,
				Confidence:      model.ConfMedium,
				TechniqueID:     39,
			})
		}
	}

	// Match GraphQL endpoint URLs
	endpointRe := model.GetRegex(`["']([^"']*(?:/graphql|/gql)[^"']*)["']`)
	epMatches := endpointRe.FindAllStringSubmatchIndex(content, -1)
	for _, m := range epMatches {
		if len(m) >= 4 {
			url := content[m[2]:m[3]]
			results = append(results, model.Result{
				URL:             url,
				SourceFile:      ctx.FileName,
				SourceLine:      model.LineNumber(content, m[0]),
				DetectionMethod: "graphql_ops",
				Category:        model.CatAPIEndpoint,
				Confidence:      model.ConfMedium,
				TechniqueID:     39,
			})
		}
	}

	// Match gql tagged template literals and extract operation names inside them
	gqlTagRe := model.GetRegex("gql\\s*`([^`]*)`")
	gqlMatches := gqlTagRe.FindAllStringSubmatchIndex(content, -1)
	for _, m := range gqlMatches {
		if len(m) >= 4 {
			body := content[m[2]:m[3]]
			innerRe := model.GetRegex(`(?:query|mutation|subscription)\s+([A-Za-z_][A-Za-z0-9_]*)`)
			innerMatches := innerRe.FindAllStringSubmatch(body, -1)
			for _, inner := range innerMatches {
				if len(inner) >= 2 {
					results = append(results, model.Result{
						URL:             "graphql:" + inner[1],
						SourceFile:      ctx.FileName,
						SourceLine:      model.LineNumber(content, m[0]),
						DetectionMethod: "graphql_ops",
						Category:        model.CatAPIEndpoint,
						Confidence:      model.ConfMedium,
						TechniqueID:     39,
					})
				}
			}
		}
	}

	return results
}

// extractRESTInference generates inferred CRUD endpoints from discovered API paths
// that look like REST resources (plural nouns under /api/).
func extractRESTInference(ctx *model.ExtractionContext) []model.Result {
	var results []model.Result
	content := ctx.Content

	// Find paths that look like /api/<resource> where resource is a plural noun
	re := model.GetRegex(`["'](/api/[a-z][a-z0-9_-]*s)(?:/[^"']*)?["']`)
	matches := re.FindAllStringSubmatchIndex(content, -1)

	seen := make(map[string]bool)
	for _, m := range matches {
		if len(m) >= 4 {
			basePath := content[m[2]:m[3]]

			// Normalize: strip trailing slash
			basePath = strings.TrimRight(basePath, "/")

			if seen[basePath] {
				continue
			}
			seen[basePath] = true

			srcLine := model.LineNumber(content, m[0])

			// Generate the inferred CRUD endpoints
			inferred := []string{
				basePath,
				basePath + "/{id}",
				basePath + "/create",
				basePath + "/update",
				basePath + "/delete",
			}

			for _, ep := range inferred {
				results = append(results, model.Result{
					URL:             ep,
					SourceFile:      ctx.FileName,
					SourceLine:      srcLine,
					DetectionMethod: "rest_inference",
					Category:        model.CatInferred,
					Confidence:      model.ConfLow,
					TechniqueID:     40,
				})
			}
		}
	}

	return results
}
