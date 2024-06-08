package arkd

import (
	"io"
	"os"

	"github.com/oklog/ulid/v2"
)

func FindOrCreateWorkerIdFromFile() (string, error) {
  f, err := os.Open("./arkd_worker_id")
  if os.IsNotExist(err) {
    if f, err = os.Create("./arkd_worker_id"); err != nil {
      return "", err
    }
  } else if err != nil {
    return "", err
  }

  buf, err := io.ReadAll(f)
  if err != nil {
    return "", err
  }

  wid := string(buf)

  if wid == "" {
    wid = ulid.Make().String()
    if err := os.WriteFile("./arkd_worker_id", []byte(wid), 0666); err != nil {
      return "", err
    }
  }

  return string(wid), nil
}
