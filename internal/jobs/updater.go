package jobs

import (
	"os"

	gorealdebrid "github.com/TOomaAh/go-realdebrid"
	"github.com/TOomaAh/qbrdt/internal/database"
	"github.com/TOomaAh/qbrdt/pkg/downloader"
	"github.com/TOomaAh/qbrdt/pkg/logger"
)

type TorrentUpdater struct {
	client      *gorealdebrid.RealDebridClient
	torrents    *database.TorrentRepository
	download    *database.DownloadRepository
	preferences *database.PreferencesRepository
	logger      logger.Interface
	downloader  *downloader.Downloader
}

func NewTorrentUpdater(client *gorealdebrid.RealDebridClient,
	downloader *downloader.Downloader,
	torrents *database.TorrentRepository,
	download *database.DownloadRepository,
	preferences *database.PreferencesRepository,
	logger logger.Interface) *TorrentUpdater {

	download.CleanAllDownloads()
	torrents.UpdateTorrentsStatusToWaitingForDownload()

	allTorrent, err := torrents.FindAll()

	if err != nil {
		logger.Error("Error getting all torrents: %s", err)
	}

	for _, torrent := range allTorrent {
		downloads, err := download.FindAllByRdId(torrent.ID)
		if err != nil {
			logger.Error("Error getting downloads for torrent %d: %s", torrent.ID, err)
		}

		if len(downloads) == 0 {
			torrent.InternalStatus = database.TorrentInternalWaitingForDownload
			torrents.Update(&torrent)
		}

	}

	return &TorrentUpdater{
		client:      client,
		torrents:    torrents,
		download:    download,
		preferences: preferences,
		logger:      logger,
		downloader:  downloader,
	}
}

func (tu *TorrentUpdater) acceptTorrent(id string) error {
	tu.logger.Info("Accepting torrent %s", id)
	return tu.client.AcceptTorrent(id)
}

func (tu *TorrentUpdater) DeleteTorrent(id string) error {
	tu.logger.Info("Deleting torrent %s", id)
	return tu.client.DeleteTorrent(id)
}

func (tu *TorrentUpdater) Run() {

	tu.logger.Info("Running torrent updater")
	tu.torrents.Mutex.Lock()

	defer tu.torrents.Mutex.Unlock()

	torrents, err := tu.torrents.FindAllNotDownloaded()
	if err != nil {
		return
	}

	for _, torrent := range torrents {
		if torrent.RDId == "" {
			tu.torrents.Delete(torrent.ID)
		}

		// if torrent has pending downloads, set it to waiting for download
		if tu.torrents.HavePendingDownloads(torrent.ID) {
			if torrent.InternalStatus != database.TorrentInternalDownloading {
				torrent.InternalStatus = database.TorrentInternalWaitingForDownload
				tu.torrents.Update(&torrent)
			}
			continue
		}

		if torrent.InternalStatus == database.TorrentInternalDownloaded && torrent.Status == database.TorrentStatusDownloaded && !tu.torrents.HasDownload(torrent.ID) {
			torrent.InternalStatus = database.TorrentInternalWaitingForDownload
			tu.torrents.Update(&torrent)
		}

		info, err := tu.client.GetTorrent(torrent.RDId)

		// if torrent is not found, delete it
		if err != nil {
			tu.logger.Error("Error getting torrent info: %s", err)
			tu.DeleteTorrent(torrent.RDId)
			tu.torrents.Delete(torrent.ID)
			continue
		}

		var needUpdate bool
		if torrent.RDProgress != info.Progress {
			torrent.RDProgress = info.Progress
			needUpdate = true
		}

		if info.Seeders != nil && torrent.RDSeeders != *info.Seeders {
			torrent.RDSeeders = *info.Seeders
			needUpdate = true
		}

		if info.Speed != nil && torrent.RDSpeed != *info.Speed {
			torrent.RDSpeed = *info.Speed
			needUpdate = true
		}

		if needUpdate {
			tu.torrents.Update(&torrent)
		}

		// if torrent is waiting for files selection, accept it
		if info.Status == "waiting_files_selection" {
			tu.acceptTorrent(torrent.RDId)
			torrent.InternalStatus = database.TorrentInternalWaitingForDownload
			tu.torrents.Update(&torrent)
		}

		if info.Status == "downloaded" {
			if torrent.Status != database.TorrentStatusDownloaded {
				torrent.Status = database.TorrentStatusDownloaded
				tu.torrents.Update(&torrent)
			}
		}

		// if torrent is dead, delete it
		if info.Status == "dead" {
			tu.logger.Error("Torrent " + torrent.RDId + " is dead, deleting it")
			tu.DeleteTorrent(torrent.RDId)
			tu.torrents.Delete(torrent.ID)
			continue
		}

		// if torrent is downloading, wait for download to finish
		if info.Status == "downloading" {
			if torrent.Status != database.TorrentStatusDownloading {
				torrent.InternalStatus = database.TorrentInternalDownloading
				tu.torrents.Update(&torrent)
			}
			continue
		}

		// if torrent is downloaded, but have pending downloads, set it to waiting for download
		if info.Status == "downloaded" && tu.torrents.HavePendingDownloads(torrent.ID) {
			var needUpdate bool
			if torrent.Status != database.TorrentStatusDownloaded {
				torrent.Status = database.TorrentStatusDownloaded
				needUpdate = true
			}

			if torrent.InternalStatus != database.TorrentInternalDownloading {
				torrent.InternalStatus = database.TorrentInternalWaitingForDownload
				needUpdate = true
			}

			if needUpdate {
				tu.torrents.Update(&torrent)
			}
			continue
		}

		if torrent.Status == database.TorrentStatusDownloaded && torrent.InternalStatus == database.TorrentInternalWaitingForDownload {
			torrent.InternalStatus = database.TorrentInternalDownloading
			tu.torrents.Update(&torrent)
			tu.saveDownload(&torrent, info)
			continue
		}

		// if torrent is queued, set it to queued
		if info.Status == "queue" {
			if torrent.Status != database.TorrentStatusQueued {
				torrent.Status = database.TorrentStatusQueued
				tu.torrents.Update(&torrent)
			}
			continue
		}

		// if torrent is uploading, set it to checkingUP
		if info.Status == "uploading" {
			if torrent.Status != "checkingUP" {
				torrent.Status = "checkingUP"
				tu.torrents.Update(&torrent)
			}
			continue
		}

		// if torrent is in unknown status, set it to unknown
		if info.Status == "" {
			torrent.Status = "unknown"
			tu.torrents.Update(&torrent)
			continue
		}

	}

}

func (tu *TorrentUpdater) saveDownload(torrent *database.Torrent, info *gorealdebrid.Torrent) {
	for _, link := range info.Links {
		debrid, err := tu.client.DebridTorrent(link)
		if err != nil {
			tu.logger.Error("Error debriding torrent: %s", err)
			continue
		}

		download := &database.Download{
			UserId:       0,
			TorrentId:    torrent.ID,
			FileName:     debrid.Filename,
			FileSize:     debrid.FileSize,
			IsDownloaded: false,
			Url:          debrid.Download,
			SavePath:     tu.preferences.GetSavePath() + string(os.PathSeparator) + torrent.Category + string(os.PathSeparator) + torrent.RDName,
		}

		d := tu.download.Create(download)

		if d != nil {
			tu.logger.Error("Error saving download: %s", d)
		}

		tu.logger.Info("Start downloading %s", download.FileName)

		go tu.downloader.AddDownload(&downloader.Download{
			Url:      download.Url,
			FileName: download.FileName,
			FileSize: download.FileSize,
			SavePath: download.SavePath,
			Object:   download,
		})

	}
}
