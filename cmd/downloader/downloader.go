package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

// Structure pour la progression du téléchargement
type Progress struct {
	Downloaded int64
	Total      int64
	Percent    float64
	Speed      float64
	Remaining  time.Duration
}

// Fonction pour télécharger un chunk spécifique
func downloadChunk(url, filePath string, start, end int64, maxSpeedKBps int, progressChan chan<- Progress, wg *sync.WaitGroup) error {

	defer wg.Done()
	// Créer une requête HTTP GET avec un en-tête Range pour télécharger une portion du fichier
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Créer le fichier chunk
	chunkFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer chunkFile.Close()

	// Variables pour le suivi du téléchargement
	var downloadedSize int64
	startTime := time.Now()
	maxSpeed := int64(maxSpeedKBps * 1024)
	buffer := make([]byte, 32*1024) // 32 KB

	for {
		// Lire un morceau de données
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			// Écrire les données dans le fichier chunk
			chunkFile.Write(buffer[:n])

			// Mettre à jour la taille téléchargée
			downloadedSize += int64(n)

			// Calculer la vitesse de téléchargement et le temps restant
			elapsedTime := time.Since(startTime).Seconds()
			speed := float64(downloadedSize) / elapsedTime
			remaining := time.Duration(float64(end-start-downloadedSize)/speed) * time.Second

			// Envoyer la progression sur le canal
			progressChan <- Progress{
				Downloaded: downloadedSize,
				Total:      end - start + 1,
				Percent:    float64(downloadedSize) / float64(end-start+1) * 100,
				Speed:      speed,
				Remaining:  remaining,
			}

			// Limiter la vitesse si nécessaire, sauf si maxSpeedKBps est à 0
			if maxSpeedKBps > 0 && speed > float64(maxSpeed) {
				sleepDuration := time.Duration(float64(n)/float64(maxSpeed)*1000) * time.Millisecond
				time.Sleep(sleepDuration)
			}
		}

		// Si la lecture est terminée, quitter la boucle
		if err == io.EOF {
			break
		}

		// Gérer les autres erreurs possibles
		if err != nil {
			return err
		}
	}

	return nil
}

// Fonction pour télécharger le fichier en plusieurs chunks
func downloadFile(url, filePath string, numChunks int, maxSpeedKBps int, progressChan chan<- Progress) error {
	// Faire une requête HEAD pour obtenir la taille du fichier
	resp, err := http.Head(url)
	if err != nil {
		return err
	}

	// Vérifier si la requête a réussi
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("échec de la requête : %s", resp.Status)
	}

	// Obtenir la taille totale du fichier
	totalSize := resp.ContentLength
	chunkSize := totalSize / int64(numChunks)

	wg := sync.WaitGroup{}

	for i := 0; i < numChunks; i++ {
		// Calculer la plage de bytes pour ce chunk
		start := int64(i) * chunkSize
		end := start + chunkSize - 1
		if i == numChunks-1 {
			end = totalSize - 1 // Le dernier chunk peut être plus grand
		}

		// Télécharger ce chunk
		chunkFileName := fmt.Sprintf("%s.part%d", filePath, i)
		wg.Add(1)
		go func() {
			err := downloadChunk(url, chunkFileName, start, end, maxSpeedKBps, progressChan, &wg)

			if err != nil {
				fmt.Printf("Erreur lors du téléchargement du chunk %d : %v\n", i, err)
			}
		}()

	}
	wg.Wait()
	// Fusionner les fichiers chunks en un seul fichier final
	if err := mergeChunks(filePath, numChunks); err != nil {
		return err
	}

	return nil
}

// Fonction pour fusionner les chunks en un seul fichier
func mergeChunks(filePath string, numChunks int) error {
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	for i := 0; i < numChunks; i++ {
		chunkFileName := fmt.Sprintf("%s.part%d", filePath, i)
		chunkFile, err := os.Open(chunkFileName)
		if err != nil {
			return err
		}

		_, err = io.Copy(out, chunkFile)
		if err != nil {
			chunkFile.Close()
			return err
		}
		chunkFile.Close()

		// Supprimer le fichier chunk après la fusion
		os.Remove(chunkFileName)
	}

	return nil
}

func main() {
	// URL du fichier à télécharger
	url := ""

	// Chemin où enregistrer le fichier téléchargé
	filePath := os.Getenv("URL")

	// Nombre de chunks
	numChunks := 10

	// Limite de la bande passante en KB/s (par exemple, 500 KB/s, 0 pour illimité)
	maxSpeedKBps := 0

	// Créer un canal pour la progression
	progressChan := make(chan Progress)

	// Lancer le téléchargement dans une goroutine
	go func() {
		err := downloadFile(url, filePath, numChunks, maxSpeedKBps, progressChan)
		if err != nil {
			fmt.Printf("Erreur lors du téléchargement : %v\n", err)
			close(progressChan)
		}
	}()

	// Afficher la progression au fur et à mesure sur la même ligne
	for progress := range progressChan {
		fmt.Printf("\rTéléchargé %d/%d octets (%.2f%%) à %.2f KB/s, temps restant estimé : %v",
			progress.Downloaded, progress.Total, progress.Percent, progress.Speed/1024, progress.Remaining)
	}

	fmt.Println("\nTéléchargement terminé avec succès !")
}
