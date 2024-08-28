package database

import "gorm.io/gorm"

type Download struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	UserId       int64  `json:"user_id"`
	TorrentId    uint   `json:"torrent_id"`
	FileName     string `json:"file_name"`
	FileSize     int64  `json:"file_size"`
	FilePath     string `json:"file_path"`
	SavePath     string `json:"save_path"`
	Url          string `json:"url"`
	IsDownloaded bool   `json:"is_downloaded"`
	Progress     int    `json:"progress"`
	Downloaded   int64  `json:"downloaded"`
}

func NewDownload(userId int64, torrentId uint, fileName string, fileSize int64, filePath string, url string, downloaded bool) *Download {
	return &Download{
		UserId:       userId,
		TorrentId:    torrentId,
		FileName:     fileName,
		FileSize:     fileSize,
		FilePath:     filePath,
		Url:          url,
		IsDownloaded: downloaded,
		Progress:     0,
		Downloaded:   0,
	}
}

type DownloadRepository struct {
	db *gorm.DB
}

func NewDownloadRepository(db *gorm.DB) *DownloadRepository {
	db.AutoMigrate(&Download{})
	return &DownloadRepository{
		db: db,
	}
}

func (r *DownloadRepository) Create(download *Download) error {
	return r.db.Create(download).Error
}

func (r *DownloadRepository) FindAllByRdId(rdId uint) ([]Download, error) {
	var downloads []Download
	err := r.db.Where("torrent_id=?", rdId).Find(&downloads).Error
	return downloads, err
}

func (r *DownloadRepository) CleanAllDownloads() error {
	return r.db.Where("is_downloaded=?", 0).Delete(&Download{}).Error
}

func (r *DownloadRepository) Update(download *Download) error {
	return r.db.Save(download).Error
}
