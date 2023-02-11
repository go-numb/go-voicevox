package main

import (
	"fmt"
	"github.com/go-numb/go-voicevox"
	"log"
)

func main() {
	client := voicevox.New()

	speakers := client.GetSpeakers()
	if speakers == nil {
		log.Fatal("responce is nil")
	}

	for i := 0; i < len(speakers); i++ {
		fmt.Println(speakers[i])
	}

	spk := speakers[client.Speaker]
	if client.Style >= len(spk.Styles) {
		log.Fatal("style not found")
	}
	spkID := spk.Styles[client.Style].ID
	fmt.Println(spk.Name, spk.Styles[client.Style].Name, spkID)

	params, err := client.GetQuery(spkID, "テストしています")
	if err != nil {
		log.Fatal(err)
	}

	client.Set(params)
	fmt.Printf("%#v\n", params)

	b, err := client.Synth(spkID, params)
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Speaking(params, b[44:]); err != nil {
		log.Fatal(err)
	}

}
