package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	oldPath := flag.String("old", "./docs/swagger/swagger.json", "Path to existing swagger.json (old)")
	newPath := flag.String("new", "./docs/swagger/tmp/swagger.json", "Path to newly generated swagger.json (new)")
	outPath := flag.String("out", "./docs/swagger/swagger.json", "Output path for merged swagger.json")
	flag.Parse()

	oldData, err := os.ReadFile(*oldPath)
	if err != nil {
		fmt.Printf("⚠️  Could not read old swagger file: %v\n", err)
		return
	}

	newData, err := os.ReadFile(*newPath)
	if err != nil {
		fmt.Printf("⚠️  Could not read new swagger file: %v\n", err)
		return
	}

	var oldMap, newMap map[string]any
	json.Unmarshal(oldData, &oldMap)
	json.Unmarshal(newData, &newMap)

	examplesMap := extractExamples(oldMap)
	fmt.Printf("🔍 Found %d examples from old swagger.\n", len(examplesMap))

	applied := injectExamples(newMap, examplesMap)

	out, _ := json.MarshalIndent(newMap, "", "    ")
	os.WriteFile(*outPath, out, 0644)

	fmt.Printf("✅ Restored %d examples (skipped %d unmatched)\n", applied, len(examplesMap)-applied)
}

// extractExamples scans all paths/methods/responses and saves examples
func extractExamples(sw map[string]any) map[string]any {
	result := map[string]any{}
	paths, ok := sw["paths"].(map[string]any)
	if !ok {
		return result
	}

	for path, pathVal := range paths {
		methods, ok := pathVal.(map[string]any)
		if !ok {
			continue
		}
		for method, methodVal := range methods {
			methodMap, ok := methodVal.(map[string]any)
			if !ok {
				continue
			}
			responses, ok := methodMap["responses"].(map[string]any)
			if !ok {
				continue
			}
			for code, respVal := range responses {
				respMap, ok := respVal.(map[string]any)
				if !ok {
					continue
				}
				if examples, exists := respMap["examples"]; exists {
					key := fmt.Sprintf("%s|%s|%s", strings.ToLower(path), strings.ToLower(method), code)
					result[key] = examples
				}
			}
		}
	}
	return result
}

// injectExamples merges examples from old swagger into new swagger
func injectExamples(sw map[string]any, examples map[string]any) int {
	paths, ok := sw["paths"].(map[string]any)
	if !ok {
		return 0
	}

	applied := 0
	for key, examplesVal := range examples {
		parts := strings.Split(key, "|")
		if len(parts) != 3 {
			continue
		}
		path, method, code := parts[0], parts[1], parts[2]

		// cari path yang cocok
		for newPathKey, pathVal := range paths {
			if strings.ToLower(newPathKey) != path {
				continue
			}

			methods, ok := pathVal.(map[string]any)
			if !ok {
				continue
			}

			for newMethodKey, methodVal := range methods {
				if strings.ToLower(newMethodKey) != method {
					continue
				}

				methodMap, ok := methodVal.(map[string]any)
				if !ok {
					continue
				}
				responses, ok := methodMap["responses"].(map[string]any)
				if !ok {
					continue
				}
				respMap, ok := responses[code].(map[string]any)
				if !ok {
					continue
				}

				respMap["examples"] = examplesVal
				responses[code] = respMap
				methodMap["responses"] = responses
				methods[newMethodKey] = methodMap
				paths[newPathKey] = methods
				applied++
			}
		}
	}

	sw["paths"] = paths
	return applied
}
