package cli

import "fmt"

// PrintHelp prints available CLI commands.
func PrintHelp() {
	fmt.Println(cs("Commands:", bold, cyan))
	fmt.Println(cs("status", bold, green) + " " + c("Show HUD", dim))
	fmt.Println(cs("explore", bold, green) + " " + c("Explore for events/loot/enemies", dim))
	fmt.Println(cs("hunt [extra_sp]", bold, green) + " " + c("Hunt enemies; stake extra SP", dim))
	fmt.Println(cs("rest [sp]", bold, green) + " " + c("Convert SP into HP", dim))
	fmt.Println(cs("use <item_id>", bold, green) + " " + c("Use an item", dim))
	fmt.Println(cs("save", bold, green) + " " + c("Save game", dim))
	fmt.Println(cs("exit / quit", bold, green) + " " + c("Save and exit", dim))
}
