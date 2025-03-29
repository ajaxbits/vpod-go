package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

func main() {
	cmd := exec.Command("yt-dlp", "-J", "https://www.youtube.com/@Monoanalysis")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	var c YouTubeChannel
	err = json.Unmarshal(out.Bytes(), &c)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(c.Title)
}
