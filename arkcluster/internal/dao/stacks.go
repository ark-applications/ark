package dao

import (
	"context"
	"fmt"

	"github.com/dkimot/ark/arkcluster/internal/models"
	"gorm.io/gorm"
)

func GetStackByName(ctx context.Context, db *gorm.DB, stackName string) (*models.Stack, error) {
  var stack models.Stack
  result := db.First(&stack, "name = ?", stackName)
  if result.Error != nil {
    return nil, fmt.Errorf("could not get stack by name %s: %w", stackName, result.Error)
  }

  return &stack, nil
}
