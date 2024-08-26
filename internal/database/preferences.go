package database

import "gorm.io/gorm"

type Preferences struct {
	gorm.Model
	SavePath string
}

type PreferencesRepository struct {
	db *gorm.DB
}

func NewPreferencesRepository(db *gorm.DB, savePath string) *PreferencesRepository {
	db.AutoMigrate(&Preferences{})

	// check if a preferences exists, if not create one
	p := &Preferences{}
	err := db.First(p).Error
	if err != nil {
		db.Create(&Preferences{
			SavePath: savePath,
		})
	}

	if savePath != p.SavePath {
		db.Model(p).Update("save_path", savePath)
	}

	return &PreferencesRepository{
		db: db,
	}
}

func (r *PreferencesRepository) Create(preferences *Preferences) error {
	// if a preferences already exists, update it
	p := &Preferences{}
	exist := r.db.First(p).Error
	if exist == nil {
		preferences.ID = p.ID
		return r.db.Save(preferences).Error
	}
	return r.db.Create(preferences).Error
}

func (r *PreferencesRepository) GetSavePath() string {
	p := &Preferences{}
	err := r.db.First(p).Error
	if err != nil {
		return ""
	}
	return p.SavePath
}
