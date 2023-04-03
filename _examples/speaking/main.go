package main

import (
	"fmt"
	"log"
	"time"

	"github.com/go-numb/go-voicevox"
)

var (
	texts = []string{
		"萌え袖って……いいよねぇ︎",
		"最近 CM ではめでたく? ノーリミットになってました",
	}
)

func main() {
	client := voicevox.New("http://localhost:50021")

	speakers, err := client.GetSpeakers()
	if err != nil {
		log.Fatal("responce is nil")
	}

	for i := 0; i < len(speakers); i++ {
		fmt.Println(speakers[i])
	}

	spk := speakers[voicevox.NURSEROBOT]
	if client.Style >= len(spk.Styles) {
		log.Fatal("style not found")
	}
	spkID := spk.Styles[client.Style].ID
	fmt.Println(spk.Name, spk.Styles[client.Style].Name, spkID)

	for i := 0; i < len(texts); i++ {
		now := time.Now()
		params, err := client.GetQuery(spkID, texts[i])
		if err != nil {
			log.Fatal(err)
		}

		client.Set(params)
		fmt.Printf("%#v\n", params)

		b, err := client.Synth(spkID, params)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("処理時間: %.3fs\n", time.Since(now).Seconds())

		if err := client.Speaking(params, b[44:]); err != nil {
			log.Fatal(err)
		}
	}

}
