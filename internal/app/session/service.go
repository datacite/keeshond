package session

import (
	"crypto/rand"
	"fmt"
	"hash/fnv"
	"io"
	"time"

	"github.com/datacite/keeshond/internal/app"
	"github.com/dchest/siphash"
	"gorm.io/gorm"
)

type SessionService struct {
	repository SessionRepositoryReader
	config     *app.Config
}

// NewSessionService creates a new session service
func NewSessionService(repository SessionRepositoryReader, config *app.Config) *SessionService {
	return &SessionService{
		repository: repository,
		config:     config,
	}
}

func (service *SessionService) CreateSalt() error {
	// Generate random salt
	salt, err := generateSalt()
	if err != nil {
		return err
	}

	// Store the salt
	err = service.repository.Create(&salt)

	return err
}

func generateSalt() (Salt, error) {
	// Generate a new random salt byte array
	saltBytes := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, saltBytes)
	if err != nil {
		return Salt{}, err
	}

	// Create the salt
	salt := Salt{
		Salt: saltBytes,
		Created: time.Now(),
	}

	return salt, nil
}

func (service *SessionService) GetSalt() (Salt, error) {

	// Get the current salt
	currentSalt, err := service.repository.Get()

	// TODO: The following setup probably should not happen on a get call, instead
	// the salt should be created ahead of time and a rotation job seperatly.

	// If not found, create a new salt
	if err == gorm.ErrRecordNotFound {
		err = service.CreateSalt()

		// If there was an error, return it
		if err != nil {
			return currentSalt, err
		}

		// Get the current salt again
		currentSalt, err = service.repository.Get()
	}

	// Generate a new salt if salt created is older than a day
	if currentSalt.Created.Add(time.Hour * 24).Before(time.Now()) {
		err = service.CreateSalt()

		// If there was an error, return it
		if err != nil {
			return currentSalt, err
		}

		// Get the current salt again
		currentSalt, err = service.repository.Get()
	}

	return currentSalt, nil
}

func GenerateSessionId(user_id uint64, time time.Time) uint64 {
	// Construct a session id based timestamp date + hour time slice + user id
	// sessionId := now.Format("2006-01-02") + "|" + now.Format("15") + "|" + user_id
	sessionId := fmt.Sprintf("%s|%s|%x", time.Format("2006-01-02"), time.Format("15"), user_id)

	// Hash the session id into a 64 bit integer
	h := fnv.New64a()
	h.Write([]byte(sessionId))
	return h.Sum64()
}


func GenerateUserId(salt *Salt, client_ip string, user_agent string, repo_id string, host_domain string) uint64 {
	// Build a salted integer user id
	// User_id is based upon a daily salt, the ip from the client,
	// the original user_agent from the request
	// a repo_id to ensure that the user id is unique per repo
	// and finally the host domain, this is to ensure it's unique
	// if it's the same repo but different websites.

	userId := fmt.Sprintf("%s|%s|%s|%s", client_ip, user_agent, repo_id, host_domain)

	// Siphash is a fast cryptographic hash function that can be used to hash
	// a string or byte array into a 64 bit integer.
	h := siphash.New(salt.Salt)
	h.Write([]byte(userId))

	return h.Sum64()
}
