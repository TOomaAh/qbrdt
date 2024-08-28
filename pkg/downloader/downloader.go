package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/TOomaAh/qbrdt/pkg/logger"
)

type Downloader struct {
	chunk        int
	speedLimit   int
	downloadLock chan struct{}
	logger       logger.Interface
	OnStart      func(download *Download)
	OnUpdate     func(download *Download)
	OnFinish     func(download *Download)
}

type Progress struct {
	Downloaded int64
	Total      int64
	Percent    float64
	Speed      float64
	Remaining  time.Duration
}

type Download struct {
	Index      int
	Url        string
	FileName   string
	FileSize   int64
	SavePath   string
	Progress   int
	Downloaded int64
	Remaining  time.Duration
	Object     interface{}
	lock       sync.Mutex
}

func NewDownloader(chunk, speedLimit, maxDownlaods int, logger logger.Interface) *Downloader {
	logger.Info("Initialisation of downloader with %d chunks, speed limit %d KB/s and %d simultaneous downloads", chunk, speedLimit, maxDownlaods)
	return &Downloader{
		chunk:        chunk,
		speedLimit:   speedLimit,
		downloadLock: make(chan struct{}, maxDownlaods),
		logger:       logger,
		OnStart:      func(download *Download) {},
		OnUpdate:     func(download *Download) {},
		OnFinish:     func(download *Download) {},
	}
}

func (d *Downloader) AddDownload(download *Download) {
	progressChan := make(chan Progress)
	download.lock = sync.Mutex{}

	// Lancer le téléchargement dans une goroutine
	go func() {
		// Verrouiller le téléchargement
		d.downloadLock <- struct{}{}
		defer func() {
			// Déverrouiller le téléchargement
			<-d.downloadLock
		}()
		err := d.downloadFile(download, progressChan)
		if err != nil {
			fmt.Printf("Erreur lors du téléchargement : %v\n", err)
		}
		close(progressChan)

		// Appeler la fonction de rappel onFinish
		d.OnFinish(download)

	}()
	for progress := range progressChan {

		download.lock.Lock()

		// update download object
		download.Progress = int(progress.Percent) / d.chunk
		download.Downloaded = download.Downloaded + progress.Downloaded
		download.Remaining = progress.Remaining * time.Duration(d.chunk)

		download.lock.Unlock()

		d.OnUpdate(download)
	}
}

// Fonction pour fusionner les chunks en un seul fichier
func (d *Downloader) mergeChunks(filePath string) error {
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	for i := 0; i < d.chunk; i++ {
		chunkFileName := fmt.Sprintf("%s.part%d", filePath, i)
		chunkFile, err := os.Open(chunkFileName)
		if err != nil {
			return err
		}

		defer chunkFile.Close()

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

// Fonction pour télécharger le fichier en plusieurs chunks
func (d *Downloader) downloadFile(download *Download, progressChan chan<- Progress) error {
	// Faire une requête HEAD pour obtenir la taille du fichier

	// Obtenir la taille totale du fichier
	totalSize := download.FileSize
	chunkSize := totalSize / int64(d.chunk)

	wg := sync.WaitGroup{}

	filename := download.SavePath + string(os.PathSeparator) + download.FileName

	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			// create folder
			err := os.MkdirAll(download.SavePath, os.ModePerm)
			if err != nil {
				return err
			}
		}
	}

	for i := 0; i < d.chunk; i++ {
		// Calculer la plage de bytes pour ce chunk
		start := int64(i) * chunkSize
		end := start + chunkSize - 1
		if i == d.chunk-1 {
			end = totalSize - 1 // Le dernier chunk peut être plus grand
		}

		// Télécharger ce chunk
		chunkFileName := fmt.Sprintf("%s.part%d", filename, i)
		wg.Add(1)
		go func() {
			err := d.downloadChunk(download.Url, chunkFileName, start, end, progressChan, &wg)

			if err != nil {
				fmt.Printf("Erreur lors du téléchargement du chunk %d : %v\n", i, err)
			}
		}()

	}
	wg.Wait()
	// Fusionner les fichiers chunks en un seul fichier final
	if err := d.mergeChunks(filename); err != nil {
		return err
	}

	return nil
}

func (d *Downloader) downloadChunk(url, filePath string, start, end int64, progressChan chan<- Progress, wg *sync.WaitGroup) error {

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
	maxSpeed := int64(d.speedLimit * 1024)
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
			if d.speedLimit > 0 && speed > float64(maxSpeed) {
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
