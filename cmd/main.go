package main

import (
	"log"

	"github.com/jeffrydegrande/kongo"
)

func main() {
	kong := kongo.NewKong("http://localhost:8001")
	endpoints, err := kong.GetEndpoints()
	if err != nil {
		panic(err)
	}

	for _, endpoint := range endpoints {
		log.Println(endpoint.Name, "(", endpoint.Path, ")", "=>", endpoint.TargetUrl)
		plugins, err := kong.GetPlugins(endpoint.Name)
		if err != nil {
			panic(err)
		}
		for _, plugin := range plugins {
			log.Printf("\t%s %t %#v\n", plugin.Name, plugin.Value)
		}
	}
}
