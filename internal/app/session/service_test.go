package session

import (
	"testing"
	"time"
)

func TestGenerateUserId(t *testing.T) {
	// Create fake salt
	salt := Salt{
		Salt:    []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16},
		Created: time.Now(),
	}

	repoId := "my_fake_repo"
	userAgent := "Mozilla/5.0 (compatible; FakeUser/1.0; +http://www.example.com/bot.html)"
	clientIp := "127.0.0.1"
	hostDomain := "example.com"

	// Generate the userId
	userId := GenerateUserId(&salt, clientIp, userAgent, repoId, hostDomain)

	var expected uint64 = 10981375520814568898
	if userId != expected {
		// fatal if the user id is not the expected value
		t.Fatalf(`User id is not %d`, expected)
	}
}

func TestGenerateSessionId(t *testing.T) {

	// Create fake user id
	userId := uint64(10981375520814568898)

	// Create fake time for reproducibility
	time := time.Date(2019, time.January, 1, 15, 15, 0, 0, time.UTC)

	// Generate session id
	sessionId := GenerateSessionId(userId, time)

	var expected uint64 = 2259115543464263857
	if sessionId != expected {
		t.Fatalf(`Session id is not %d`, expected)
	}
}
