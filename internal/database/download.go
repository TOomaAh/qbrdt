package database

import "gorm.io/gorm"

type Download struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	UserId     int64  `json:"user_id"`
	TorrentId  uint   `json:"torrent_id"`
	FileName   string `json:"file_name"`
	FileSize   int64  `json:"file_size"`
	FilePath   string `json:"file_path"`
	SavePath   string `json:"save_path"`
	Url        string `json:"url"`
	Downloaded bool   `json:"downloaded"`
}

func NewDownload(userId int64, torrentId uint, fileName string, fileSize int64, filePath string, url string, downloaded bool) *Download {
	return &Download{
		UserId:     userId,
		TorrentId:  torrentId,
		FileName:   fileName,
		FileSize:   fileSize,
		FilePath:   filePath,
		Url:        url,
		Downloaded: downloaded,
	}
}

func (d *Download) SplitIntoChunks(chunks int, chunkSize int64) [][2]int64 {
	arr := make([][2]int64, chunks)
	for i := 0; i < chunks; i++ {
		if i == 0 {
			arr[i][0] = 0
			arr[i][1] = chunkSize
		} else if i == chunks-1 {
			arr[i][0] = arr[i-1][1] + 1
			arr[i][1] = d.FileSize - 1
		} else {
			arr[i][0] = arr[i-1][1] + 1
			arr[i][1] = arr[i][0] + chunkSize
		}
	}

	return arr
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

func (r *DownloadRepository) CleanAllDownloads() error {
	return r.db.Where("downloaded=?", 0).Delete(&Download{}).Error
}

func (r *DownloadRepository) Update(download *Download) error {
	return r.db.Save(download).Error
}
