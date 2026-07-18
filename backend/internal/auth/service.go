package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Signup(request SignupRequest) error {

	_, err := s.repository.FindByEmail(request.Email)

	if err == nil {
		return errors.New("email already exists")
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(request.Password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return err
	}

	user := User{
		Name:         request.Name,
		Email:        request.Email,
		PasswordHash: string(hash),
	}

	return s.repository.CreateUser(&user)
}
