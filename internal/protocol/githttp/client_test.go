package githttp

import (
	"context"
	"log"
	"os"
	"testing"
)

func TestDiscoverRef(t *testing.T) {
	gitUrl := os.Getenv("GIT_URL")
	baseContext := context.Background()
	client := NewGitHttpClient()
	reply, err := client.discoverRef(baseContext, gitUrl)
	if err != nil {
		t.Fatalf("failed to discover ref: %v", err)
	}
	log.Printf("%+v", reply)
}
