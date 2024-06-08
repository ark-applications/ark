package arkd

import (
	dockerparser "github.com/novln/docker-parser"
)

func NewImageRef(image string) (ImageRef, error) {
	ref, err := dockerparser.Parse(image)
	if err != nil {
		return ImageRef{}, err
	}

	return ImageRef{
		FullName:   image,
		Registry:   ref.Registry(),
		Repository: ref.ShortName(),
		Tag:        ref.Tag(),
	}, nil
}

type ImageRef struct {
	FullName   string `json:"full_name"`
	Registry   string `json:"registry"`
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	Digest     string `json:"digest"`
}
