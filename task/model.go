package task

type Task struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Title       string `json:"title" gorm:"not null"`
	Description string `json:"description"`
	Completed   bool   `json:"completed" gorm:"not null;default:false"`
}

//migrate create -ext sql -dir migrations -seq create_tasks_table