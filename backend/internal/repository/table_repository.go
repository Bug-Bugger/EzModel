package repository

import (
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TableRepository struct {
	db *gorm.DB
}

func NewTableRepository(db *gorm.DB) TableRepositoryInterface {
	return &TableRepository{db: db}
}

func (r *TableRepository) Create(table *models.Table) (uuid.UUID, error) {
	if err := r.db.Create(table).Error; err != nil {
		return uuid.Nil, err
	}
	return table.ID, nil
}

func (r *TableRepository) GetByID(id uuid.UUID) (*models.Table, error) {
	var table models.Table
	err := r.db.Preload("Fields").First(&table, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &table, nil
}

func (r *TableRepository) GetByProjectID(projectID uuid.UUID) ([]*models.Table, error) {
	var tables []*models.Table
	err := r.db.Preload("Fields").Where("project_id = ?", projectID).Find(&tables).Error
	if err != nil {
		return nil, err
	}
	return tables, nil
}

func (r *TableRepository) Update(table *models.Table) error {
	return r.db.Save(table).Error
}

func (r *TableRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Table{}, "id = ?", id).Error
}

func (r *TableRepository) UpdatePosition(id uuid.UUID, posX, posY float64) error {
	return r.db.Model(&models.Table{}).Where("id = ?", id).Updates(map[string]interface{}{
		"pos_x": posX,
		"pos_y": posY,
	}).Error
}