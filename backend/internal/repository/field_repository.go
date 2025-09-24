package repository

import (
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FieldRepository struct {
	db *gorm.DB
}

func NewFieldRepository(db *gorm.DB) FieldRepositoryInterface {
	return &FieldRepository{db: db}
}

func (r *FieldRepository) Create(field *models.Field) (uuid.UUID, error) {
	if err := r.db.Create(field).Error; err != nil {
		return uuid.Nil, err
	}
	return field.ID, nil
}

func (r *FieldRepository) GetByID(id uuid.UUID) (*models.Field, error) {
	var field models.Field
	err := r.db.First(&field, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &field, nil
}

func (r *FieldRepository) GetByTableID(tableID uuid.UUID) ([]*models.Field, error) {
	var fields []*models.Field
	err := r.db.Where("table_id = ?", tableID).Order("position ASC").Find(&fields).Error
	if err != nil {
		return nil, err
	}
	return fields, nil
}

func (r *FieldRepository) Update(field *models.Field) error {
	return r.db.Save(field).Error
}

func (r *FieldRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Field{}, "id = ?", id).Error
}

func (r *FieldRepository) ReorderFields(tableID uuid.UUID, fieldPositions map[uuid.UUID]int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for fieldID, position := range fieldPositions {
			if err := tx.Model(&models.Field{}).Where("id = ? AND table_id = ?", fieldID, tableID).Update("position", position).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
