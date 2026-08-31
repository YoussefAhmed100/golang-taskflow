
package task_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"task-api/task"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&task.Task{})
	require.NoError(t, err)

	return db
}

func TestService_GetAll(t *testing.T) {
	db := setupTestDB(t)

	repository := task.NewRepository(db)
	service := task.NewService(repository)

	expectedTasks := []task.Task{
		{
			Title: "Task 1",
		},
		{
			Title: "Task 2",
		},
	}

	for _, expectedTask := range expectedTasks {
		_, err := repository.Create(expectedTask)
		require.NoError(t, err)
	}

	tasks, err := service.GetAll()

	require.NoError(t, err)
	assert.Len(t, tasks, 2)
	assert.Equal(t, "Task 1", tasks[0].Title)
	assert.Equal(t, "Task 2", tasks[1].Title)
}

func TestService_GetAll_Empty(t *testing.T) {
	db := setupTestDB(t)

	repository := task.NewRepository(db)
	service := task.NewService(repository)

	tasks, err := service.GetAll()

	require.NoError(t, err)
	assert.Empty(t, tasks)
}

func TestService_Create(t *testing.T) {
	db := setupTestDB(t)

	repository := task.NewRepository(db)
	service := task.NewService(repository)

	input := task.Task{
		Title: "Learn CI/CD",
	}

	createdTask, err := service.Create(input)

	require.NoError(t, err)

	assert.NotZero(t, createdTask.ID)
	assert.Equal(t, "Learn CI/CD", createdTask.Title)
}

func TestService_Create_MultipleTasks(t *testing.T) {
	db := setupTestDB(t)

	repository := task.NewRepository(db)
	service := task.NewService(repository)

	firstTask := task.Task{
		Title: "Learn Go",
	}

	secondTask := task.Task{
		Title: "Learn GitHub Actions",
	}

	createdFirst, err := service.Create(firstTask)
	require.NoError(t, err)

	createdSecond, err := service.Create(secondTask)
	require.NoError(t, err)

	assert.NotZero(t, createdFirst.ID)
	assert.NotZero(t, createdSecond.ID)
	assert.NotEqual(t, createdFirst.ID, createdSecond.ID)
}
func TestTemporaryFailure(t *testing.T) {
	t.Fatal("intentional CI failure")
}
