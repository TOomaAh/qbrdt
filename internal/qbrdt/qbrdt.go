package qbrdt

import (
	gorealdebrid "github.com/TOomaAh/go-realdebrid"
	"github.com/TOomaAh/qbrdt/internal/api/qbittorrent"
	"github.com/TOomaAh/qbrdt/internal/config"
	"github.com/TOomaAh/qbrdt/internal/database"
	"github.com/TOomaAh/qbrdt/internal/jobs"
	"github.com/TOomaAh/qbrdt/pkg/downloader"
	"github.com/TOomaAh/qbrdt/pkg/logger"
	"github.com/bamzi/jobrunner"
	"github.com/labstack/echo/v4"
)

type QBRDT struct {
	logger      logger.Interface
	conf        *config.QBRDTConfig
	downloader  *downloader.Downloader
	preferences *database.PreferencesRepository
	categories  *database.CategoryRepository
	torrents    *database.TorrentRepository
	downloads   *database.DownloadRepository
	client      *gorealdebrid.RealDebridClient
}

func New(logger logger.Interface, conf *config.QBRDTConfig) *QBRDT {
	db := database.NewDatabase(logger)
	logger.Info("All downloads will be saved in %s", conf.Downloader.SavePath)
	preferences := database.NewPreferencesRepository(db, conf.Downloader.SavePath)
	categories := database.NewCategoryRepository(db)
	torrents := database.NewTorrentRepository(db)
	downloads := database.NewDownloadRepository(db)
	client := gorealdebrid.NewRealDebridClient(conf.RealDebrid.Token)
	return &QBRDT{
		logger:      logger,
		conf:        conf,
		preferences: preferences,
		categories:  categories,
		torrents:    torrents,
		downloads:   downloads,
		client:      client,
		downloader: downloader.NewDownloader(
			conf.Downloader.Chunk,
			conf.Downloader.SpeedLimit,
			conf.Downloader.MaxDownloads,
			logger,
			func(download *downloader.Download) {
			},
			func(download *downloader.Download) {
				download.Object.(*database.Download).Downloaded = true
				downloads.Update(download.Object.(*database.Download))
				// if all downloads are downloaded, update torrent status to downloaded
				if torrents.AllDownloadsAreDownloaded(download.Object.(*database.Download).TorrentId) {
					torrents.UpdateTorrentStatusToDownloaded(download.Object.(*database.Download).TorrentId)
				}
			}),
	}
}

func (qbrdt *QBRDT) Run() {

	e := echo.New()

	e.HideBanner = true
	//e.Use(middleware.Logger())

	// on startup make clean all downloads and put torrent with status downloading to waiting_for_download
	qbrdt.downloads.CleanAllDownloads()
	qbrdt.torrents.UpdateTorrentsStatusToWaitingForDownload()

	jobrunner.Start()
	jobrunner.Schedule("@every "+qbrdt.conf.Qbrdt.TorrentRefreshInterval+"s", jobs.NewTorrentUpdater(
		qbrdt.client,
		qbrdt.downloader,
		qbrdt.torrents,
		qbrdt.downloads,
		qbrdt.preferences,
		qbrdt.logger,
	))
	qbrdt.logger.Info("Torrent updater job scheduled every " + qbrdt.conf.Qbrdt.TorrentRefreshInterval + "s")

	defer jobrunner.Stop()

	api := e.Group("/api/v2")
	qbittorrent.NewQbittorrentAuthenticationApi(api, qbrdt.conf.QBittorrent.Username, qbrdt.conf.QBittorrent.Password)
	qbittorrent.NewQbittorrentAppApi(api, qbrdt.preferences)
	qbittorrent.NewQbittorrentTorrentApi(qbrdt.logger, api, qbrdt.preferences, qbrdt.categories, qbrdt.torrents, qbrdt.client)

	e.Logger.Fatal(e.Start(":" + qbrdt.conf.QBittorrent.Port))

}
