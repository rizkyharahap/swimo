// File: swagger-merge.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	oldPath := flag.String("old", "./docs/swagger/swagger.json", "Path to existing swagger.json")
	newPath := flag.String("new", "./docs/swagger/tmp/swagger.json", "Path to newly generated swagger.json")
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

	oldPaths, _ := oldMap["paths"].(map[string]any)
	newPaths, _ := newMap["paths"].(map[string]any)
	if oldPaths == nil {
		oldPaths = map[string]any{}
	}

	for k, v := range newPaths {
		if ov, exists := oldPaths[k]; exists {
			fmt.Printf("~ Updating existing path: %s\n", k)
			oldPaths[k] = smartMergePath(ov, v)
		} else {
			fmt.Printf("➕ Added new path: %s\n", k)
			oldPaths[k] = v
		}
	}
	oldMap["paths"] = oldPaths

	// Merge definitions (always additive)
	oldDefs, _ := oldMap["definitions"].(map[string]any)
	newDefs, _ := newMap["definitions"].(map[string]any)
	if oldDefs == nil {
		oldDefs = map[string]any{}
	}
	for k, v := range newDefs {
		oldDefs[k] = v
	}
	oldMap["definitions"] = oldDefs

	out, _ := json.MarshalIndent(oldMap, "", "    ")
	os.WriteFile(*oldPath, out, 0644)
	fmt.Println("✅ Swagger merge complete — examples preserved & outdated responses removed.")
}

func mergeDeep(oldVal, newVal any) any {
	switch old := oldVal.(type) {
	case map[string]any:
		newMap, ok := newVal.(map[string]any)
		if !ok {
			return oldVal
		}
		for k, v := range newMap {
			if ov, exists := old[k]; exists {
				old[k] = mergeDeep(ov, v)
			} else {
				old[k] = v
			}
		}
		return old
	default:
		if newVal != nil {
			return newVal
		}
		return oldVal
	}
}

func smartMergePath(oldVal, newVal any) any {
	oldMap, ok1 := oldVal.(map[string]any)
	newMap, ok2 := newVal.(map[string]any)
	if !ok1 || !ok2 {
		return newVal
	}

	for method, newMethodVal := range newMap {
		oldMethodVal, hasOld := oldMap[method]
		if !hasOld {
			oldMap[method] = newMethodVal
			continue
		}

		oldMethodMap, okOld := oldMethodVal.(map[string]any)
		newMethodMap, okNew := newMethodVal.(map[string]any)
		if !okOld || !okNew {
			oldMap[method] = newMethodVal
			continue
		}

		oldResp, _ := oldMethodMap["responses"].(map[string]any)
		newResp, _ := newMethodMap["responses"].(map[string]any)
		if newResp != nil {
			if oldResp == nil {
				oldResp = map[string]any{}
			}

			// 🔹 1. Remove response codes no longer present in new JSON
			for code := range oldResp {
				if _, stillExists := newResp[code]; !stillExists {
					delete(oldResp, code)
				}
			}

			// 🔹 2. Merge or add updated responses
			for code, newVal := range newResp {
				if oldVal, exists := oldResp[code]; exists {
					oldResp[code] = mergeDeep(oldVal, newVal) // preserve examples
				} else {
					oldResp[code] = newVal
				}
			}

			oldMethodMap["responses"] = oldResp
		}

		// Merge other fields (summary, desc, tags, produces, etc.)
		for k, v := range newMethodMap {
			if k != "responses" {
				oldMethodMap[k] = v
			}
		}
		oldMap[method] = oldMethodMap
	}
	return oldMap
}
