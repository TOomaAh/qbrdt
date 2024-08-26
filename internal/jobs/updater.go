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

func (tu *TorrentUpdater) updateInternalStatus(torrent *database.Torrent, rdtTorrent *gorealdebrid.Torrent) {
	if (torrent.InternalStatus == database.TorrentInternalWaitingForDownload ||
		torrent.InternalStatus == database.TorrentInternalError ||
		torrent.InternalStatus == database.TorrentInternalDownloading) && rdtTorrent.Status == "downloaded" {
		return
	}
	switch rdtTorrent.Status {
	case "downloaded":
		torrent.InternalStatus = database.TorrentInternalWaitingForDownload
	case "dead":
		torrent.InternalStatus = database.TorrentInternalError
	default:
		torrent.InternalStatus = database.TorrentInternalWaiting
	}
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

		// if all of downloads is terminated return
		if tu.torrents.HasDownload(torrent.ID) && tu.torrents.AllDownloadsAreDownloaded(torrent.ID) {
			torrent.Status = "downloaded"
			torrent.InternalStatus = database.TorrentInternalDownloaded
			tu.torrents.Update(&torrent)
			continue
		} else {
			info, err := tu.client.GetTorrent(torrent.RDId)
			if err != nil {
				tu.logger.Error("Error getting torrent info: %s", err)
				tu.DeleteTorrent(torrent.RDId)
				tu.torrents.Delete(torrent.ID)
				continue
			}
			tu.updateInternalStatus(&torrent, info)

			switch info.Status {
			case "downloaded":
				if torrent.InternalStatus == database.TorrentInternalWaitingForDownload {
					tu.saveDownload(&torrent, info)
					torrent.InternalStatus = database.TorrentInternalDownloading
				}

				if torrent.InternalStatus == database.TorrentInternalDownloading {
					torrent.Status = "downloading"
				} else {
					torrent.Status = "pausedUP"
				}
			case "dead":
				torrent.Status = "error"
			case "queue":
				torrent.Status = "queuedUP"
			case "downloading":
				torrent.Status = "downloading"
			case "uploading":
				torrent.Status = "checkingUP"
			case "waiting_files_selection":
				tu.acceptTorrent(torrent.RDId)
			default:
				torrent.Status = "unknown"
			}

			torrent.RDProgress = info.Progress

			if info.Seeders != nil {
				torrent.RDSeeders = *info.Seeders
			}
			if info.Speed != nil {
				torrent.RDSpeed = *info.Speed
			}

			tu.torrents.Update(&torrent)
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
			UserId:     0,
			TorrentId:  torrent.ID,
			FileName:   debrid.Filename,
			FileSize:   debrid.FileSize,
			Downloaded: false,
			Url:        debrid.Download,
			SavePath:   tu.preferences.GetSavePath() + string(os.PathSeparator) + torrent.Category + string(os.PathSeparator) + torrent.RDName,
		}

		d := tu.download.Create(download)

		if d != nil {
			tu.logger.Error("Error saving download: %s", d)
		}

		tu.logger.Info("Start downloading %s", download.FileName)

		tu.downloader.AddDownload(&downloader.Download{
			Url:      download.Url,
			FileName: download.FileName,
			FileSize: download.FileSize,
			SavePath: download.SavePath,
			Object:   download,
		})

	}
}
