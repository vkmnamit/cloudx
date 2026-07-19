package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repository *Repository
	JWTService *JWTService
}

func NewService(repository *Repository, jwtService *JWTService) *Service {
	return &Service{
		repository: repository,
		JWTService: jwtService,
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

func (s *Service) Login(request LoginRequest) (*LoginResponse, error) {
	user, err := s.repository.FindByEmail(request.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(request.Password),
	)

	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	token, err := s.JWTService.GenerateToken(user.ID.String(), user.Email)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken: token,
	}, nil
}
