package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/m0thm4n/Dusty/bot"
	"github.com/m0thm4n/Dusty/config"
	"github.com/m0thm4n/Dusty/youtube"
)

var cfg config.Config

func readConfig(cfg *config.Config, configFileName string) {
	configFileName, _ = filepath.Abs(configFileName)
	log.Printf("Loading config: %v", configFileName)

	configFile, err := os.Open(configFileName)
	if err != nil {
		log.Fatal("File error: ", err.Error())
	}
	defer configFile.Close()
	jsonParser := json.NewDecoder(configFile)
	if err := jsonParser.Decode(&cfg); err != nil {
		log.Fatal("Config error: ", err.Error())
	}
}

func banner() {
	b, err := ioutil.ReadFile("asciiart.txt")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func main() {
	banner()
	readConfig(&cfg, "config.json")
	log.Println("Starting Dusty.")

	//make api connections
	youtubeAPI := youtube.NewYoutubeAPI(cfg.Youtube.ApiKey)

	err := bot.InitBot(cfg.Discord.Token, youtubeAPI, &cfg)
	if err != nil {
		log.Println(err)
	}
}
