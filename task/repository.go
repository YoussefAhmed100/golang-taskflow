package task

import (
	"fmt"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (repository *Repository) GetAll() ([]Task, error) {
	var tasks []Task

	if err := repository.db.Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	return tasks, nil
}

func (repository *Repository) Create(task Task) (Task, error) {
	if err := repository.db.Create(&task).Error; err != nil {
		return Task{}, fmt.Errorf("failed to create task: %w", err)
	}

	return task, nil
}