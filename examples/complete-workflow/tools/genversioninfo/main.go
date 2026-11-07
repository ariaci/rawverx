package main

import (
	"log"

	"github.com/ariaci/rawverx/genverx"
)

func main() {
	info, err := genverx.NewInfoWithVersionFromFile(
		"../../version/VERSION",
		genverx.AdditionalInfo{
			ProductName: "example",
			Author:      "Example Author",
			Copyright:   "© Example Author. Licensed under MIT.",
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	for _, arch := range []genverx.Arch{
		genverx.ArchI386,
		genverx.ArchAMD64,
		genverx.ArchARM64,
	} {
		if err := genverx.Generate(info, arch, "../../version/winres"); err != nil {
			log.Fatal(err)
		}
	}
}
