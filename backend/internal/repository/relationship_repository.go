package repository

import (
	"github.com/Bug-Bugger/ezmodel/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RelationshipRepository struct {
	db *gorm.DB
}

func NewRelationshipRepository(db *gorm.DB) RelationshipRepositoryInterface {
	return &RelationshipRepository{db: db}
}

func (r *RelationshipRepository) Create(relationship *models.Relationship) (uuid.UUID, error) {
	if err := r.db.Create(relationship).Error; err != nil {
		return uuid.Nil, err
	}
	return relationship.ID, nil
}

func (r *RelationshipRepository) GetByID(id uuid.UUID) (*models.Relationship, error) {
	var relationship models.Relationship
	err := r.db.First(&relationship, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &relationship, nil
}

func (r *RelationshipRepository) GetByProjectID(projectID uuid.UUID) ([]*models.Relationship, error) {
	var relationships []*models.Relationship
	err := r.db.Where("project_id = ?", projectID).Find(&relationships).Error
	if err != nil {
		return nil, err
	}
	return relationships, nil
}

func (r *RelationshipRepository) GetByTableID(tableID uuid.UUID) ([]*models.Relationship, error) {
	var relationships []*models.Relationship
	err := r.db.Where("source_table_id = ? OR target_table_id = ?", tableID, tableID).Find(&relationships).Error
	if err != nil {
		return nil, err
	}
	return relationships, nil
}

func (r *RelationshipRepository) Update(relationship *models.Relationship) error {
	return r.db.Save(relationship).Error
}

func (r *RelationshipRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Relationship{}, "id = ?", id).Error
}
