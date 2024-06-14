package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dkimot/ark"
	"github.com/dkimot/ark/arkcluster/internal/dao"
	"github.com/dkimot/ark/arkcluster/internal/models"
	"gorm.io/gorm"
)

func UpsertStackDefinition(ctx context.Context, db *gorm.DB, stackName string, stackDef ark.StackDefinition) error {
  stack, err := dao.GetStackByName(ctx, db, stackName)
  if err != nil {
    return err
  }

  defBytes, err := json.Marshal(&stackDef)
  if err != nil {
    return err
  }

  result := db.Create(&models.StackDef{
    StackID: stack.ID,
    RootApp: stackDef.RootApp,
    RawDefinition: defBytes,
  })
  if result.Error != nil {
    return fmt.Errorf("could not create StackDef: %w", err)
  }

  return nil
}
