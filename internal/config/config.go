package config

import (
	"os"
	"strconv"

	"gopkg.in/yaml.v2"
)

type QBRDTConfig struct {
	RealDebrid struct {
		Token string `yaml:"token"`
	} `yaml:"realdebrid"`
	QBittorrent struct {
		Port     string `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"qbittorrent"`
	Qbrdt struct {
		TorrentRefreshInterval string `yaml:"torrent_refresh_interval"`
	} `yaml:"qbrdt"`
	Downloader struct {
		SavePath     string `yaml:"save_path"`
		Chunk        int    `yaml:"chunk"`
		SpeedLimit   int    `yaml:"speed_limit"`
		MaxDownloads int    `yaml:"max_downloads"`
	} `yaml:"downloader"`
	Logger struct {
		Level string `yaml:"level"`
	} `yaml:"logger"`
}

func NewQBRDTConfig() *QBRDTConfig {
	if os.Getenv("CONFIG_FILE") == "" {
		os.Setenv("CONFIG_FILE", "config.yml")
	}

	file, err := os.Open(os.Getenv("CONFIG_FILE"))

	if err != nil {
		panic(err)
	}

	defer file.Close()

	config := &QBRDTConfig{}

	err = yaml.NewDecoder(file).Decode(config)

	if err != nil {
		panic(err)
	}

	if os.Getenv("REALDEBRID_TOKEN") != "" {
		config.RealDebrid.Token = os.Getenv("REALDEBRID_TOKEN")
	}

	if os.Getenv("QB_PORT") != "" {
		config.QBittorrent.Port = os.Getenv("QB_PORT")
	}

	if os.Getenv("QB_USERNAME") != "" {
		config.QBittorrent.Username = os.Getenv("QB_USERNAME")
	}

	if os.Getenv("QB_PASSWORD") != "" {
		config.QBittorrent.Password = os.Getenv("QB_PASSWORD")
	}

	if os.Getenv("DOWNLOADER_SAVE_PATH") != "" {
		config.Downloader.SavePath = os.Getenv("DOWNLOADER_SAVE_PATH")
	}

	if os.Getenv("DOWNLOADER_CHUNK") != "" {
		config.Downloader.Chunk, err = strconv.Atoi(os.Getenv("DOWNLOADER_CHUNK"))

		if err != nil {
			panic(err)
		}

	}

	if os.Getenv("DOWNLOADER_SPEED_LIMIT") != "" {
		config.Downloader.SpeedLimit, err = strconv.Atoi(os.Getenv("DOWNLOADER_SPEED_LIMIT"))

		if err != nil {
			panic(err)
		}

	}

	return config
}
