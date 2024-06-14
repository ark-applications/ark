package usecase

import (
	"context"

	"github.com/dkimot/ark/arkcluster/internal/models"
	"gorm.io/gorm"
)

func CreateDeployment(ctx context.Context, db *gorm.DB, stack *models.Stack) error {
  // get current stack definition
  var stackDef models.StackDef
  result := db.First(&stackDef, "stack_id = ?", stack.ID)
  if result.Error != nil {
    return result.Error
  }

  // save deployment to database
  result = db.Create(&models.Deployment{
    StackID: stack.ID,
    StackDefRaw: stackDef.RawDefinition,
  })
  if result.Error != nil {
    return result.Error
  }

  return nil
}
