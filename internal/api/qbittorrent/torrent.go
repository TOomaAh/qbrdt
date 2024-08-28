package qbittorrent

import (
	"errors"
	"io"
	"os"
	"strings"
	"time"

	gorealdebrid "github.com/TOomaAh/go-realdebrid"
	"github.com/TOomaAh/qbrdt/internal/database"
	"github.com/TOomaAh/qbrdt/pkg/logger"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
)

type QBittorrentTorrentApi struct {
	cache      *cache.Cache
	preference *database.PreferencesRepository
	category   *database.CategoryRepository
	torrents   *database.TorrentRepository
	client     *gorealdebrid.RealDebridClient
	logger     logger.Interface
}

type QbittorentAddRequest struct {
	Urls     []string `json:"urls"`
	Category string   `json:"category"`
	Priority int      `json:"priority"`
}

type FileInfoResponse struct {
	Name string `json:"name"`
}

type TorrentPropertiesResponse struct {
	AdditionDate          int64   `json:"additionDate"`
	Comment               string  `json:"comment"`
	CompletionDate        int64   `json:"completionDate"`
	CreatedBy             string  `json:"createdBy"`
	CreationDate          int64   `json:"creationDate"`
	DlLimit               int64   `json:"dlLimit"`
	DlSpeed               int64   `json:"dlSpeed"`
	DlSpeedAvg            int64   `json:"dlSpeedAvg"`
	Eta                   int64   `json:"eta"`
	LastSeen              int64   `json:"lastSeen"`
	NbConnections         int     `json:"nbConnections"`
	NbConnectionsLimit    int     `json:"nbConnectionsLimit"`
	Peers                 int     `json:"peers"`
	PeersTotal            int     `json:"peersTotal"`
	PiecesHave            int     `json:"piecesHave"`
	PiecesNum             int     `json:"piecesNum"`
	PieceSize             int     `json:"pieceSize"`
	Reannounce            int64   `json:"reannounce"`
	SavePath              string  `json:"save_path"`
	SeedingTime           int64   `json:"seedingTime"`
	Seeds                 int     `json:"seeds"`
	SeedsTotal            int     `json:"seedsTotal"`
	ShareRatio            float64 `json:"shareRatio"`
	TimeElapsed           int64   `json:"timeElapsed"`
	TotalDownloaded       int64   `json:"totalDownloaded"`
	TotalDowloadedSession int64   `json:"totalDowloadedSession"`
	TotalSize             int64   `json:"totalSize"`
	TotalUploaded         int64   `json:"totalUploaded"`
	TotalUploadedSession  int64   `json:"totalUploadedSession"`
	TotalWasted           int64   `json:"totalWasted"`
	UpLimit               int64   `json:"upLimit"`
	UpSpeed               int64   `json:"upSpeed"`
	UpSpeedAvg            int64   `json:"upSpeedAvg"`
}

type QbittorentTorrent struct {
	// Time (Unix Epoch) when the torrent was added to the client
	AddedOn int64 `json:"added_on"`
	// Amount of data left to download (bytes)
	AmountLeft int64 `json:"amount_left"`
	// Whether this torrent is managed by Automatic Torrent Management
	AutoTMM bool `json:"auto_tmm"`
	// Percentage of file pieces currently available
	Availability float64 `json:"availability"`
	// Category of the torrent
	Category string `json:"category"`
	// Amount of transfer data completed (bytes)
	Completed int64 `json:"completed"`
	// Time (Unix Epoch) when the torrent completed
	CompletionOn int64 `json:"completion_on"`
	// Absolute path of torrent content (root path for multifile torrents, absolute file path for single-file torrents)
	ContentPath string `json:"content_path"`
	// Torrent download speed limit (bytes/s). -1 if unlimited.
	DLLimit int64 `json:"dl_limit"`
	// Torrent download speed (bytes/s)
	DLSpeed int64 `json:"dlspeed"`
	// Amount of data downloaded
	Downloaded int64 `json:"downloaded"`
	// Amount of data downloaded this session
	DownloadedSession int64 `json:"downloaded_session"`
	// Torrent ETA (seconds)
	ETA int64 `json:"eta"`
	// True if first and last piece are prioritized
	FLPiecePrio bool `json:"f_l_piece_prio"`
	// True if force start is enabled for this torrent
	ForceStart bool `json:"force_start"`
	// Torrent hash
	Hash string `json:"hash"`
	// True if torrent is from a private tracker
	IsPrivate bool `json:"isPrivate"`
	// Last time (Unix Epoch) when a chunk was downloaded/uploaded
	LastActivity int64 `json:"last_activity"`
	// Magnet URI corresponding to this torrent
	MagnetURI string `json:"magnet_uri"`
	// Maximum share ratio until torrent is stopped from seeding/uploading
	MaxRatio float64 `json:"max_ratio"`
	// Maximum seeding time (seconds) until torrent is stopped from seeding
	MaxSeedingTime int64 `json:"max_seeding_time"`
	// Torrent name
	Name string `json:"name"`
	// Number of seeds in the swarm
	NumComplete int `json:"num_complete"`
	// Number of leechers in the swarm
	NumIncomplete int `json:"num_incomplete"`
	// Number of leechers connected to
	NumLeechs int `json:"num_leechs"`
	// Number of seeds connected to
	NumSeeds int `json:"num_seeds"`
	// Torrent priority. Returns -1 if queuing is disabled or torrent is in seed mode
	Priority int `json:"priority"`
	// Torrent progress (percentage/100)
	Progress float64 `json:"progress"`
	// Torrent share ratio. Max ratio value: 9999.
	Ratio float64 `json:"ratio"`
	// Similar to max_ratio, limits share ratio
	RatioLimit float64 `json:"ratio_limit"`
	// Path where this torrent's data is stored
	SavePath string `json:"save_path"`
	// Torrent elapsed time while complete (seconds)
	SeedingTime int64 `json:"seeding_time"`
	// Seeding time limit, related to max_seeding_time with additional rules
	SeedingTimeLimit int64 `json:"seeding_time_limit"`
	// Time (Unix Epoch) when this torrent was last seen complete
	SeenComplete int64 `json:"seen_complete"`
	// True if sequential download is enabled
	SeqDL bool `json:"seq_dl"`
	// Total size (bytes) of files selected for download
	Size int64 `json:"size"`
	// Torrent state (e.g., downloading, paused, completed)
	State string `json:"state"`
	// True if super seeding is enabled
	SuperSeeding bool `json:"super_seeding"`
	// Comma-concatenated tag list of the torrent
	Tags string `json:"tags"`
	// Total active time (seconds)
	TimeActive int64 `json:"time_active"`
	// Total size (bytes) of all files in this torrent (including unselected ones)
	TotalSize int64 `json:"total_size"`
	// The first tracker with working status. Returns empty string if no tracker is working.
	Tracker string `json:"tracker"`
	// Torrent upload speed limit (bytes/s). -1 if unlimited.
	UpLimit int64 `json:"up_limit"`
	// Amount of data uploaded
	Uploaded int64 `json:"uploaded"`
	// Amount of data uploaded this session
	UploadedSession int64 `json:"uploaded_session"`
	// Torrent upload speed (bytes/s)
	UpSpeed int64 `json:"upspeed"`
}

func NewQbittorrentTorrentApi(l logger.Interface,
	e *echo.Group,
	preference *database.PreferencesRepository,
	category *database.CategoryRepository,
	torrents *database.TorrentRepository,
	client *gorealdebrid.RealDebridClient,
) *QBittorrentTorrentApi {

	torrentApi := &QBittorrentTorrentApi{
		cache:      cache.New(cache.NoExpiration, cache.NoExpiration),
		preference: preference,
		category:   category,
		torrents:   torrents,
		client:     client,
		logger:     l,
	}

	g := e.Group("/torrents")
	g.GET("/categories", torrentApi.categories)
	g.POST("/categories", torrentApi.categories)
	g.GET("/createCategory", torrentApi.saveCatergories)
	g.POST("/createCategory", torrentApi.saveCatergories)
	g.GET("/info", torrentApi.torrentsInfo)
	g.POST("/info", torrentApi.torrentsInfo)
	g.GET("/files", torrentApi.torrentsFiles)
	g.POST("/files", torrentApi.torrentsFiles)
	g.GET("/properties", torrentApi.torrentsProperties)
	g.POST("/properties", torrentApi.torrentsProperties)
	g.GET("/add", torrentApi.addTorrentFromUrls)
	g.POST("/add", torrentApi.addTorrentFromFile)
	g.GET("/delete", torrentApi.deleteTorrent)
	g.POST("/delete", torrentApi.deleteTorrent)

	return torrentApi

}

func (q *QBittorrentTorrentApi) categories(c echo.Context) error {

	categories := q.category.GetTorrentCategoriesDistinct()
	path := q.preference.GetSavePath()

	var cats = make(map[string]map[string]string)
	for _, v := range categories {
		cats[v] = map[string]string{
			"name":      v,
			"save_path": path + string(os.PathSeparator) + v,
		}
	}

	return c.JSON(200, cats)

}

func (a *QBittorrentTorrentApi) saveCatergories(c echo.Context) error {
	// print body
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return Fails(c)
	}
	values := strings.Split(string(body), "=")

	if len(values) != 2 {
		return Fails(c)
	}

	category := values[1]

	if category == "" {
		return Fails(c)
	}

	cat := database.NewCategory(category)

	if err := a.category.Create(cat); err != nil {
		return Fails(c)
	}

	err = os.MkdirAll(a.preference.GetSavePath()+"/"+category, os.ModePerm)
	if err != nil {
		return Fails(c)
	}

	return c.JSON(200, cat)
}

func (q *QBittorrentTorrentApi) torrentsInfo(c echo.Context) error {
	var queryCategory string
	if c.Request().Method == "POST" {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return Fails(c)
		}

		values := strings.Split(string(body), "=")

		if len(values) != 2 {
			return Fails(c)
		}

		queryCategory = values[1]

		if queryCategory == "" {
			return Fails(c)
		}
	} else {
		queryCategory = c.QueryParam("category")
	}

	torrents, err := q.torrents.FindByCategory(queryCategory)

	if err != nil {
		return Fails(c)
	}

	var torrentsInfo = make([]QbittorentTorrent, len(torrents))

	for i, v := range torrents {
		remainingTime, err := CalculateRemainingTime(int64(v.RDSize), v.CreatedAt, v.RDProgress)

		if err != nil {
			remainingTime = 0
		}

		var status string
		if v.Status == database.TorrentStatusDownloading || (v.Status == database.TorrentStatusDownloaded && v.InternalStatus == database.TorrentInternalDownloading) {
			status = "downloading"
		} else if v.Status == database.TorrentStatusDownloaded && v.InternalStatus == database.TorrentInternalDownloaded {
			status = "pausedUP"
		} else if v.Status == database.TorrentStatusError {
			status = "error"
		} else if v.Status == database.TorrentStatusDownloading && v.RDSeeders == 0 {
			status = "stalledDL"
		} else {
			status = "paused"
		}

		torrentsInfo[i] = QbittorentTorrent{
			AddedOn:           v.CreatedAt.Unix(),
			AmountLeft:        int64(v.RDSize) * int64(v.RDProgress) / 100,
			AutoTMM:           false,
			Availability:      0,
			Category:          v.Category,
			Completed:         0,
			CompletionOn:      time.Unix(0, int64(remainingTime.Nanoseconds())).Unix(),
			ContentPath:       "",
			DLLimit:           0,
			DLSpeed:           0,
			Downloaded:        0,
			DownloadedSession: 0,
			ETA:               0,
			FLPiecePrio:       false,
			ForceStart:        false,
			Hash:              v.RDHash,
			IsPrivate:         false,
			LastActivity:      v.UpdatedAt.Unix(),
			MagnetURI:         "",
			MaxRatio:          0,
			MaxSeedingTime:    0,
			Name:              v.RDName,
			NumComplete:       0,
			NumIncomplete:     0,
			NumLeechs:         0,
			NumSeeds:          0,
			Priority:          0,
			Progress:          v.RDProgress / 100,
			Ratio:             0,
			RatioLimit:        0,
			SavePath:          q.preference.GetSavePath() + string(os.PathSeparator) + v.Category,
			SeedingTime:       0,
			SeedingTimeLimit:  0,
			SeenComplete:      0,
			SeqDL:             false,
			Size:              int64(v.RDSize),
			State:             status,
			SuperSeeding:      false,
			Tags:              "",
			TimeActive:        0,
			TotalSize:         0,
			Tracker:           "",
			UpLimit:           0,
			Uploaded:          0,
			UploadedSession:   0,
			UpSpeed:           0,
		}
	}

	return c.JSON(200, torrentsInfo)
}

func (q *QBittorrentTorrentApi) torrentsFiles(c echo.Context) error {
	var queryHash string
	if c.Request().Method == "POST" {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return Fails(c)
		}

		values := strings.Split(string(body), "=")

		if len(values) != 2 {
			return Fails(c)
		}

		queryHash = values[1]

		if queryHash == "" {
			return Fails(c)
		}
	} else {
		queryHash = c.QueryParam("hash")

		if queryHash == "" {
			return Fails(c)
		}
	}

	queryHash = strings.ToUpper(queryHash)
	torrent, err := q.torrents.FindByRDId(queryHash)

	if err != nil {
		return Fails(c)
	}

	downlaods, err := q.torrents.FindAllDownloadByRdId(torrent.ID)

	if err != nil {
		return Fails(c)
	}

	var files = make([]FileInfoResponse, len(downlaods))

	for i, v := range downlaods {
		files[i] = FileInfoResponse{
			Name: string(os.PathSeparator) + torrent.RDName + string(os.PathSeparator) + v.FileName,
		}
	}

	return c.JSON(200, files)
}

func (q *QBittorrentTorrentApi) torrentsProperties(c echo.Context) error {
	var queryHash string
	if c.Request().Method == "POST" {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return Fails(c)
		}

		values := strings.Split(string(body), "=")

		if len(values) != 2 {
			return Fails(c)
		}

		queryHash = values[1]

		if queryHash == "" {
			return Fails(c)
		}
	} else {
		queryHash = c.QueryParam("hash")

		if queryHash == "" {
			return Fails(c)
		}
	}

	queryHash = strings.ToUpper(queryHash)
	torrent, err := q.torrents.FindByRDId(queryHash)

	if err != nil {
		return Fails(c)
	}

	var properties = TorrentPropertiesResponse{
		AdditionDate:          torrent.CreatedAt.Unix(),
		Comment:               "QBRDT",
		CompletionDate:        torrent.UpdatedAt.Unix(),
		CreatedBy:             "QBRDT",
		CreationDate:          torrent.CreatedAt.Unix(),
		DlLimit:               -1,
		DlSpeed:               int64(torrent.RDSpeed),
		DlSpeedAvg:            int64(torrent.RDSpeed),
		Eta:                   0,
		LastSeen:              torrent.UpdatedAt.Unix(),
		NbConnections:         0,
		NbConnectionsLimit:    100,
		Peers:                 torrent.RDSeeders,
		PeersTotal:            torrent.RDSeeders,
		PiecesHave:            len(torrent.Downloads),
		PiecesNum:             len(torrent.Downloads),
		PieceSize:             0,
		Reannounce:            0,
		SavePath:              q.preference.GetSavePath() + string(os.PathSeparator) + string(os.PathSeparator) + torrent.Category,
		SeedingTime:           1,
		Seeds:                 torrent.RDSeeders,
		SeedsTotal:            torrent.RDSeeders,
		ShareRatio:            0,
		TimeElapsed:           int64(time.Since(torrent.CreatedAt).Seconds()),
		TotalDownloaded:       int64(torrent.RDSize),
		TotalDowloadedSession: int64(torrent.RDSize),
		TotalSize:             int64(torrent.RDSize),
		TotalUploaded:         0,
		TotalUploadedSession:  0,
		TotalWasted:           0,
		UpLimit:               -1,
		UpSpeed:               0,
		UpSpeedAvg:            0,
	}

	return c.JSON(200, properties)

}

func (q *QBittorrentTorrentApi) addTorrent(c echo.Context, content io.Reader, category string) error {
	add, err := q.client.AddTorrent(content)

	if err != nil {
		q.logger.Error("Failed to add torrent %s", err.Error())
		return Fails(c)
	}

	rdTorrent, err := q.client.GetTorrent(add.Id)

	if err != nil {
		q.logger.Error("Failed to get torrent %s", err.Error())
		return Fails(c)
	}

	var speed = 0
	if rdTorrent.Speed != nil {
		speed = *rdTorrent.Speed
	}

	var seeders = 0
	if rdTorrent.Seeders != nil {
		seeders = *rdTorrent.Seeders
	}

	var torrent = &database.Torrent{
		Status:     database.TorrentStatusDownloading,
		Type:       database.TorrentTypeFile,
		Category:   category,
		AddedBy:    database.Qbittorent,
		RDId:       add.Id,
		RDProgress: rdTorrent.Progress,
		RDName:     rdTorrent.Filename,
		RDSize:     rdTorrent.Bytes,
		RDSplit:    rdTorrent.Split,
		RDHost:     rdTorrent.Host,
		RDSpeed:    speed,
		RDSeeders:  seeders,
		RDHash:     rdTorrent.Hash,
	}

	if err := q.torrents.Create(torrent); err != nil {
		q.logger.Error("Failed to save torrent %s", err.Error())
		return Fails(c)
	}

	return Ok(c)
}

func (q *QBittorrentTorrentApi) addTorrentFromUrls(c echo.Context) error {
	return nil
}

func (q *QBittorrentTorrentApi) addTorrentFromFile(c echo.Context) error {

	var addTorrentRequest QbittorentAddRequest

	if err := c.Bind(&addTorrentRequest); err != nil {
		return Fails(c)
	}

	form, err := c.MultipartForm()

	if err != nil {
		q.logger.Error("Failed to parse form %s", err.Error())
		return Fails(c)
	}

	files := form.File["torrents"]

	if len(files) == 0 {
		q.logger.Error("No files found")
		return Fails(c)
	}

	categoryForm := form.Value["category"]
	var category string
	if len(categoryForm) == 0 {
		category = ""
	} else {
		category = categoryForm[0]
	}

	if !q.category.Exist(category) && category != "" {
		q.category.Create(database.NewCategory(category))
	}

	for _, file := range files {
		src, err := file.Open()

		if err != nil {
			q.logger.Error("Failed to open file %s", err.Error())
			return Fails(c)
		}

		defer src.Close()

		err = q.addTorrent(c, src, category)

		if errors.Is(err, Fails(c)) && err != nil {
			q.logger.Error("Failed to add torrent %s", err.Error())
			return Fails(c)
		}

	}

	return Ok(c)
}

func (q *QBittorrentTorrentApi) deleteTorrent(c echo.Context) error {
	// print body
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return Fails(c)
	}

	values := strings.Split(string(body), "&")

	savePath := q.preference.GetSavePath()

	// exemple value = "hashes=5cw3eirtiyij4&deleteFiles=true"
	// get hash but search for "hashes=" and remove it
	for _, v := range values {
		if strings.Contains(v, "hashes=") {
			hash := strings.Replace(v, "hashes=", "", 1)
			// split hash with '|'
			hashes := strings.Split(hash, "|")
			for _, h := range hashes {
				h = strings.ToUpper(h)
				torrent, err := q.torrents.FindByRDId(h)

				if err != nil {
					return Fails(c)
				}

				if err := os.RemoveAll(savePath + string(os.PathSeparator) + torrent.Category + string(os.PathSeparator) + torrent.RDName); err != nil {
					return Fails(c)
				}

				if err := q.torrents.DeleteByRDId(h); err != nil {
					return Fails(c)
				}
			}
		}
	}

	return Ok(c)

}
