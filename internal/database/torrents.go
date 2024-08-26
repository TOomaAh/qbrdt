package database

import (
	"sync"

	"gorm.io/gorm"
)

type TorrentStatus string

const (
	TorrentStatusQueued      TorrentStatus = "queued"
	TorrentStatusDownloading TorrentStatus = "downloading"
	TorrentStatusDownloaded  TorrentStatus = "downloaded"
	TorrentStatusError       TorrentStatus = "error"
)

type AddedBy string

const (
	WebInterface AddedBy = "web"
	Qbittorent   AddedBy = "qbittorent"
)

type TorrentType string

const (
	TorrentTypeMagnet TorrentType = "magnet"
	TorrentTypeFile   TorrentType = "file"
)

type TorrentInternalStatus string

const (
	TorrentInternalWaiting            TorrentInternalStatus = "waiting"
	TorrentInternalWaitingForDownload TorrentInternalStatus = "waiting_for_download"
	TorrentInternalDownloading        TorrentInternalStatus = "downloading"
	TorrentInternalDownloaded         TorrentInternalStatus = "downloaded"
	TorrentInternalError              TorrentInternalStatus = "error"
)

type Torrent struct {
	gorm.Model
	Status         TorrentStatus         `json:"status"`
	Type           TorrentType           `json:"type"`
	Downloads      []Download            `json:"downloads"`
	Category       string                `json:"category"`
	AddedBy        AddedBy               `json:"added_by"`
	RDId           string                `json:"rd_id" gorm:"unique"`
	RDProgress     float64               `json:"rd_progress"`
	RDName         string                `json:"rd_name"`
	RDSize         int                   `json:"rd_size"`
	RDSplit        int                   `json:"rd_split"`
	RDHost         string                `json:"rd_host"`
	RDSpeed        int                   `json:"rd_speed"`
	RDSeeders      int                   `json:"rd_seeders"`
	InternalStatus TorrentInternalStatus `json:"internal_status"`
}

type TorrentRepository struct {
	Mutex *sync.Mutex
	db    *gorm.DB
}

func NewTorrentRepository(db *gorm.DB) *TorrentRepository {
	db.AutoMigrate(&Torrent{})
	return &TorrentRepository{
		Mutex: &sync.Mutex{},
		db:    db,
	}
}

func (r *TorrentRepository) Create(torrent *Torrent) error {
	return r.db.Create(torrent).Error
}

func (r *TorrentRepository) FindAll() ([]Torrent, error) {
	var torrents []Torrent
	err := r.db.Find(&torrents).Error
	return torrents, err
}

func (r *TorrentRepository) FindAllNotDownloaded() ([]Torrent, error) {
	var torrents []Torrent
	err := r.db.Where("internal_status != ?", TorrentInternalDownloaded).Find(&torrents).Error
	return torrents, err
}

func (r *TorrentRepository) FindByRDId(rdId string) (*Torrent, error) {
	var torrent Torrent
	err := r.db.Where("rd_id = ?", rdId).First(&torrent).Error
	return &torrent, err
}

func (r *TorrentRepository) Update(torrent *Torrent) error {
	return r.db.Save(torrent).Error
}

func (r *TorrentRepository) Delete(id uint) error {
	return r.db.Delete(&Torrent{}, id).Error
}

func (r *TorrentRepository) DeleteByRDId(id string) error {
	r.Mutex.Lock()
	err := r.db.Where("rd_id = ?", id).Delete(&Torrent{}).Error
	r.Mutex.Unlock()
	return err
}

func (r *TorrentRepository) FindByStatus(status TorrentStatus) ([]Torrent, error) {
	var torrents []Torrent
	err := r.db.Where("status = ?", status).Find(&torrents).Error
	return torrents, err
}

func (r *TorrentRepository) FindByCategory(category string) ([]Torrent, error) {
	var torrents []Torrent
	err := r.db.Where("category = ?", category).Find(&torrents).Error
	return torrents, err
}

func (r *TorrentRepository) FindByAddedBy(addedBy AddedBy) ([]Torrent, error) {
	var torrents []Torrent
	err := r.db.Where("added_by = ?", addedBy).Find(&torrents).Error
	return torrents, err
}

func (r *TorrentRepository) UpdateTorrentsStatusToWaitingForDownload() error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	return r.db.Model(&Torrent{}).Where("internal_status = ?", TorrentInternalDownloading).Update("internal_status", TorrentInternalWaitingForDownload).Error
}

func (r *TorrentRepository) AllDownloadsAreDownloaded(torrentId uint) bool {
	var count int64
	r.db.Model(&Download{}).Where("torrent_id = ? AND downloaded = ?", torrentId, false).Count(&count)
	return count == 0
}

func (r *TorrentRepository) HasDownload(torrentId uint) bool {
	var count int64
	r.db.Model(&Download{}).Where("torrent_id = ?", torrentId).Count(&count)
	return count > 0
}

func (r *TorrentRepository) UpdateTorrentStatusToDownloaded(torrentId uint) error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	return r.db.Model(&Torrent{}).Where("id = ?", torrentId).Updates(map[string]interface{}{"status": TorrentStatusDownloaded, "internal_status": TorrentInternalDownloaded}).Error

}

func (r *TorrentRepository) FindAllDownloadByRdId(torrentId uint) ([]Download, error) {
	var downloads []Download
	err := r.db.Where("torrent_id = ?", torrentId).Find(&downloads).Error
	return downloads, err
}
