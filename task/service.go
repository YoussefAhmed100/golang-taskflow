package task

import (
	"fmt"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) GetAll() ([]Task, error) {
	tasks, err := s.repository.GetAll()
	if err != nil {
		return nil, fmt.Errorf("service: failed to get tasks: %w", err)
	}

	return tasks, nil
}

func (s *Service) Create(task Task) (Task, error) {
	createdTask, err := s.repository.Create(task)
	if err != nil {
		return Task{}, fmt.Errorf("service: failed to create task: %w", err)
	}

	return createdTask, nil
}