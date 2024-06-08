package arkd

import "testing"

func Test_NewImageRef(t *testing.T) {
	tests := []struct {
		name  string
		image string
		want  ImageRef
	}{
		{
			"full_image_url",
			"registry-1.docker.io/library/ubuntu:latest",
			ImageRef{
				Registry:   "registry-1.docker.io",
				Repository: "library/ubuntu",
				Tag:        "latest",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := NewImageRef(tt.image); got != tt.want {
				t.Errorf("NewImageRef() = %v, want %v", got, tt.want)
			}
		})
	}
}
