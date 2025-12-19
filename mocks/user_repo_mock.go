package mocks

import (
	"phase3-api-architecture/models"

	"github.com/stretchr/testify/mock"
)

type UserRepoMock struct {
	mock.Mock
}

func (m *UserRepoMock) Register(u models.User) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *UserRepoMock) GetByEmail(email string) (models.User, error) {
	args := m.Called(email)
	return args.Get(0).(models.User), args.Error(1)
}
