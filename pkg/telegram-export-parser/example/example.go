package main

import (
	"fmt"
	"log"

	"github.com/Nikalively/telegram-export-parser/parser"
)

func main() {
	events, err := parser.ParseSingleFile("path/to/result.json")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Parsed %d events from single file\n", len(events))

	// Example: Merge from folders
	folders := []string{"path/to/export1", "path/to/export2"}
	merged, err := parser.MergeFromFolders(folders)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Merged %d unique events\n", len(merged.Events))

	// Output first event as example
	if len(merged.Events) > 0 {
		fmt.Printf("First event: ID=%d, FromID=%s, Text=%s\n", merged.Events[0].ID, merged.Events[0].FromID, merged.Events[0].Text)
	}
}
