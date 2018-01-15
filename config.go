package main

import (
    "os"
    "encoding/json"
)

type Config struct {
    Env string `json:"env"`
    Port string `json:"port"`
    Endpoints []string `json:"endpoints"`
}

func LoadConfiguration(file string) Config {
    var config Config
    configFile, err := os.Open(file)
    defer configFile.Close()
    if err != nil {
        panic(err.Error())
    }
    jsonParser := json.NewDecoder(configFile)
    jsonParser.Decode(&config)
    return config
}
