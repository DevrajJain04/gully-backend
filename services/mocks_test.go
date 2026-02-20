package services

import (
	"context"

	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"gully-backend/models"
)

// ── Mock UserRepository ──

type MockUserRepo struct{ mock.Mock }

func (m *MockUserRepo) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// ── Mock GroupRepository ──

type MockGroupRepo struct{ mock.Mock }

func (m *MockGroupRepo) Create(ctx context.Context, group *models.Group) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *MockGroupRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Group, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Group), args.Error(1)
}

func (m *MockGroupRepo) FindByJoinCode(ctx context.Context, code string) (*models.Group, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Group), args.Error(1)
}

func (m *MockGroupRepo) FindByMember(ctx context.Context, userID primitive.ObjectID) ([]models.Group, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Group), args.Error(1)
}

func (m *MockGroupRepo) AddMember(ctx context.Context, groupID, userID primitive.ObjectID) error {
	args := m.Called(ctx, groupID, userID)
	return args.Error(0)
}

// ── Mock PlayerRepository ──

type MockPlayerRepo struct{ mock.Mock }

func (m *MockPlayerRepo) Create(ctx context.Context, player *models.Player) error {
	args := m.Called(ctx, player)
	return args.Error(0)
}

func (m *MockPlayerRepo) FindByGroupID(ctx context.Context, groupID primitive.ObjectID) ([]models.Player, error) {
	args := m.Called(ctx, groupID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Player), args.Error(1)
}

func (m *MockPlayerRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Player, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Player), args.Error(1)
}

func (m *MockPlayerRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPlayerRepo) FindByNameAndGroupID(ctx context.Context, name string, groupID primitive.ObjectID) (*models.Player, error) {
	args := m.Called(ctx, name, groupID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Player), args.Error(1)
}

func (m *MockPlayerRepo) Update(ctx context.Context, player *models.Player) error {
	args := m.Called(ctx, player)
	return args.Error(0)
}

// ── Mock MatchRepository ──

type MockMatchRepo struct{ mock.Mock }

func (m *MockMatchRepo) Create(ctx context.Context, match *models.Match) error {
	args := m.Called(ctx, match)
	return args.Error(0)
}

func (m *MockMatchRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Match, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Match), args.Error(1)
}

func (m *MockMatchRepo) FindByGroupID(ctx context.Context, groupID primitive.ObjectID) ([]models.Match, error) {
	args := m.Called(ctx, groupID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Match), args.Error(1)
}

func (m *MockMatchRepo) Update(ctx context.Context, match *models.Match) error {
	args := m.Called(ctx, match)
	return args.Error(0)
}

func (m *MockMatchRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMatchRepo) ReplacePlayerInMatches(ctx context.Context, groupID, sourceID, targetID primitive.ObjectID, sourceName, targetName string) error {
	args := m.Called(ctx, groupID, sourceID, targetID, sourceName, targetName)
	return args.Error(0)
}
