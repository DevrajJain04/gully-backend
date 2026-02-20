package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"gully-backend/models"
)

func TestRegister_Success(t *testing.T) {
	userRepo := new(MockUserRepo)
	svc := NewAuthService(userRepo, "test-secret")
	ctx := context.Background()

	userRepo.On("FindByUsername", ctx, "alice").Return(nil, errors.New("not found"))
	userRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).Return(nil)

	user, err := svc.Register(ctx, "alice", "password123")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "alice", user.Username)
	// Password should be hashed
	assert.NotEqual(t, "password123", user.Password)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("password123")))
	userRepo.AssertExpectations(t)
}

func TestRegister_UsernameTaken(t *testing.T) {
	userRepo := new(MockUserRepo)
	svc := NewAuthService(userRepo, "test-secret")
	ctx := context.Background()

	existing := &models.User{
		ID:       primitive.NewObjectID(),
		Username: "alice",
	}
	userRepo.On("FindByUsername", ctx, "alice").Return(existing, nil)

	user, err := svc.Register(ctx, "alice", "password123")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "username already taken")
	userRepo.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	userRepo := new(MockUserRepo)
	svc := NewAuthService(userRepo, "test-secret")
	ctx := context.Background()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	existingUser := &models.User{
		ID:       primitive.NewObjectID(),
		Username: "alice",
		Password: string(hashed),
	}
	userRepo.On("FindByUsername", ctx, "alice").Return(existingUser, nil)

	token, user, err := svc.Login(ctx, "alice", "password123")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, "alice", user.Username)
	userRepo.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	userRepo := new(MockUserRepo)
	svc := NewAuthService(userRepo, "test-secret")
	ctx := context.Background()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.DefaultCost)
	existingUser := &models.User{
		ID:       primitive.NewObjectID(),
		Username: "alice",
		Password: string(hashed),
	}
	userRepo.On("FindByUsername", ctx, "alice").Return(existingUser, nil)

	token, user, err := svc.Login(ctx, "alice", "wrongpass")

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "invalid credentials")
	userRepo.AssertExpectations(t)
}

func TestLogin_UserNotFound(t *testing.T) {
	userRepo := new(MockUserRepo)
	svc := NewAuthService(userRepo, "test-secret")
	ctx := context.Background()

	userRepo.On("FindByUsername", ctx, "nobody").Return(nil, errors.New("not found"))

	token, user, err := svc.Login(ctx, "nobody", "password123")

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Nil(t, user)
	userRepo.AssertExpectations(t)
}

func TestGenerateJWT_ContainsUserID(t *testing.T) {
	userRepo := new(MockUserRepo)
	svc := NewAuthService(userRepo, "test-secret")
	ctx := context.Background()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
	user := &models.User{
		ID:       primitive.NewObjectID(),
		Username: "alice",
		Password: string(hashed),
	}
	userRepo.On("FindByUsername", ctx, "alice").Return(user, nil)

	token, _, err := svc.Login(ctx, "alice", "pass")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	// Token should be a valid JWT (3 parts separated by dots)
	parts := 0
	for _, ch := range token {
		if ch == '.' {
			parts++
		}
	}
	assert.Equal(t, 2, parts, "JWT should have 3 parts separated by 2 dots")
}
