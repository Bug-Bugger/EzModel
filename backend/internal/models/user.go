package models

import (
	"time"
)

type User struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"type:varchar(255);not null"`
	CreatedTime time.Time `json:"created_time" gorm:"autoCreateTime"`
}
