package database

import (
	"gorm.io/gorm"
)

type Category struct {
	gorm.Model
	Name string `gorm:"unique"`
}

func NewCategory(name string) *Category {
	return &Category{
		Name: name,
	}
}

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	db.AutoMigrate(&Category{})
	return &CategoryRepository{
		db: db,
	}
}

func (r *CategoryRepository) Create(category *Category) error {
	return r.db.Create(category).Error
}

func (r *CategoryRepository) FindAll() ([]Category, error) {
	var categories []Category
	err := r.db.Find(&categories).Error
	return categories, err
}

func (r *CategoryRepository) Exist(name string) bool {
	var category Category
	err := r.db.Where("name = ?", name).First(&category).Error
	if err != nil {
		return false
	}
	return true
}

func (r *CategoryRepository) GetTorrentCategoriesDistinct() []string {
	var categories []string
	err := r.db.Model(&Category{}).Select("name").Find(&categories).Error
	if err != nil {
		return []string{}
	}
	return categories
}
