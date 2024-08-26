package qbittorrent

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
)

var (
	Ok = func(c echo.Context) error {
		return c.String(200, "Ok.")
	}

	OkBody = func(body string, c echo.Context) error {
		return c.String(200, body)
	}

	Fails = func(c echo.Context) error {
		return c.String(500, "Fails.")
	}
)

func CalculateRemainingTime(totalSize int64, startTime time.Time, percentageDownloaded float64) (time.Duration, error) {
	if percentageDownloaded <= 0 || percentageDownloaded >= 100 {
		return 0, fmt.Errorf("le pourcentage téléchargé doit être compris entre 0 et 100")
	}

	elapsedTime := time.Since(startTime)
	downloadedSize := (percentageDownloaded / 100) * float64(totalSize)

	// Calcul de la vitesse de téléchargement en octets par seconde
	downloadSpeed := downloadedSize / elapsedTime.Seconds()
	if downloadSpeed == 0 {
		return 0, fmt.Errorf("la vitesse de téléchargement est nulle, impossible de calculer le temps restant")
	}

	remainingSize := float64(totalSize) - downloadedSize
	remainingSeconds := remainingSize / downloadSpeed

	return time.Duration(remainingSeconds) * time.Second, nil
}
