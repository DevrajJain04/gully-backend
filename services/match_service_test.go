package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"gully-backend/models"
)

// ── Helpers ──

func newPlayerID() primitive.ObjectID { return primitive.NewObjectID() }

func makeLiveMatch(t1, t2 []primitive.ObjectID) *models.Match {
	return &models.Match{
		ID:              primitive.NewObjectID(),
		GroupID:         primitive.NewObjectID(),
		Team1IDs:        t1,
		Team2IDs:        t2,
		Team1Names:      []string{"Alice", "Bob"}[:len(t1)],
		Team2Names:      []string{"Charlie", "Dave"}[:len(t2)],
		Score1:          0,
		Score2:          0,
		ScoreHistory:    []models.ScoreEvent{},
		ServingTeam:     1,
		ServingPlayerID: t1[0].Hex(),
		Team1Positions:  toHexSlice(t1),
		Team2Positions:  toHexSlice(t2),
		Status:          models.MatchStatusLive,
	}
}

// ── CreateMatch tests ──

func TestCreateMatch_1v1_Success(t *testing.T) {
	matchRepo := new(MockMatchRepo)
	playerRepo := new(MockPlayerRepo)
	svc := NewMatchService(matchRepo, playerRepo)
	ctx := context.Background()

	p1, p2 := newPlayerID(), newPlayerID()
	groupID := primitive.NewObjectID()

	playerRepo.On("FindByID", ctx, p1).Return(&models.Player{ID: p1, Name: "Alice"}, nil)
	playerRepo.On("FindByID", ctx, p2).Return(&models.Player{ID: p2, Name: "Bob"}, nil)
	matchRepo.On("Create", ctx, mock.AnythingOfType("*models.Match")).Return(nil)

	match, err := svc.CreateMatch(ctx, groupID, []primitive.ObjectID{p1}, []primitive.ObjectID{p2})

	assert.NoError(t, err)
	assert.NotNil(t, match)
	assert.Equal(t, models.MatchStatusLive, match.Status)
	assert.Len(t, match.Team1IDs, 1)
	assert.Len(t, match.Team2IDs, 1)
	assert.Equal(t, "Alice", match.Team1Names[0])
	assert.Equal(t, "Bob", match.Team2Names[0])
	assert.Equal(t, 1, match.ServingTeam)
	assert.Equal(t, p1.Hex(), match.ServingPlayerID)
	matchRepo.AssertExpectations(t)
	playerRepo.AssertExpectations(t)
}

func TestCreateMatch_2v2_Success(t *testing.T) {
	matchRepo := new(MockMatchRepo)
	playerRepo := new(MockPlayerRepo)
	svc := NewMatchService(matchRepo, playerRepo)
	ctx := context.Background()

	p1, p2, p3, p4 := newPlayerID(), newPlayerID(), newPlayerID(), newPlayerID()
	groupID := primitive.NewObjectID()

	playerRepo.On("FindByID", ctx, p1).Return(&models.Player{ID: p1, Name: "Alice"}, nil)
	playerRepo.On("FindByID", ctx, p2).Return(&models.Player{ID: p2, Name: "Bob"}, nil)
	playerRepo.On("FindByID", ctx, p3).Return(&models.Player{ID: p3, Name: "Charlie"}, nil)
	playerRepo.On("FindByID", ctx, p4).Return(&models.Player{ID: p4, Name: "Dave"}, nil)
	matchRepo.On("Create", ctx, mock.AnythingOfType("*models.Match")).Return(nil)

	match, err := svc.CreateMatch(ctx, groupID,
		[]primitive.ObjectID{p1, p2},
		[]primitive.ObjectID{p3, p4})

	assert.NoError(t, err)
	assert.Len(t, match.Team1IDs, 2)
	assert.Len(t, match.Team2IDs, 2)
	assert.Equal(t, []string{"Alice", "Bob"}, match.Team1Names)
	assert.Equal(t, []string{"Charlie", "Dave"}, match.Team2Names)
	assert.Len(t, match.Team1Positions, 2)
	assert.Len(t, match.Team2Positions, 2)
	matchRepo.AssertExpectations(t)
}

func TestCreateMatch_EmptyTeam_Fails(t *testing.T) {
	matchRepo := new(MockMatchRepo)
	playerRepo := new(MockPlayerRepo)
	svc := NewMatchService(matchRepo, playerRepo)
	ctx := context.Background()

	_, err := svc.CreateMatch(ctx, primitive.NewObjectID(),
		[]primitive.ObjectID{},
		[]primitive.ObjectID{newPlayerID()})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least 1 player")
}

func TestCreateMatch_TooManyPlayers_Fails(t *testing.T) {
	matchRepo := new(MockMatchRepo)
	playerRepo := new(MockPlayerRepo)
	svc := NewMatchService(matchRepo, playerRepo)
	ctx := context.Background()

	_, err := svc.CreateMatch(ctx, primitive.NewObjectID(),
		[]primitive.ObjectID{newPlayerID(), newPlayerID(), newPlayerID()},
		[]primitive.ObjectID{newPlayerID()})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at most 2 players")
}

func TestCreateMatch_PlayerNotFound_Fails(t *testing.T) {
	matchRepo := new(MockMatchRepo)
	playerRepo := new(MockPlayerRepo)
	svc := NewMatchService(matchRepo, playerRepo)
	ctx := context.Background()

	p1 := newPlayerID()
	playerRepo.On("FindByID", ctx, p1).Return(nil, errors.New("not found"))

	_, err := svc.CreateMatch(ctx, primitive.NewObjectID(),
		[]primitive.ObjectID{p1},
		[]primitive.ObjectID{newPlayerID()})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// ── UpdateScore tests ──

func TestUpdateScore_Team1Scores(t *testing.T) {
	matchRepo := new(MockMatchRepo)
	playerRepo := new(MockPlayerRepo)
	svc := NewMatchService(matchRepo, playerRepo)
	ctx := context.Background()

	p1, p2 := newPlayerID(), newPlayerID()
	match := makeLiveMatch([]primitive.ObjectID{p1}, []primitive.ObjectID{p2})

	matchRepo.On("FindByID", ctx, match.ID).Return(match, nil)
	matchRepo.On("Update", ctx, match).Return(nil)

	result, err := svc.UpdateScore(ctx, match.ID, 1, p1.Hex())

	assert.NoError(t, err)
	assert.Equal(t, 1, result.Score1)
	assert.Equal(t, 0, result.Score2)
	assert.Len(t, result.ScoreHistory, 1)
	assert.Equal(t, 1, result.ScoreHistory[0].Team)
	assert.Equal(t, p1.Hex(), result.ScoreHistory[0].PlayerID)
	assert.Equal(t, 1, result.ServingTeam)
	assert.Equal(t, p1.Hex(), result.ServingPlayerID)
	matchRepo.AssertExpectations(t)
}

func TestUpdateScore_Team2Scores(t *testing.T) {
	matchRepo := new(MockMatchRepo)
	playerRepo := new(MockPlayerRepo)
	svc := NewMatchService(matchRepo, playerRepo)
	ctx := context.Background()

	p1, p2 := newPlayerID(), newPlayerID()
	match := makeLiveMatch([]primitive.ObjectID{p1}, []primitive.ObjectID{p2})

	matchRepo.On("FindByID", ctx, match.ID).Return(match, nil)
	matchRepo.On("Update", ctx, match).Return(nil)

	result, err := svc.UpdateScore(ctx, match.ID, 2, p2.Hex())

	assert.NoError(t, err)
	assert.Equal(t, 0, result.Score1)
	assert.Equal(t, 1, result.Score2)
	assert.Equal(t, 2, result.ServingTeam)
	assert.Equal(t, p2.Hex(), result.ServingPlayerID)
}

func TestUpdateScore_InvalidTeam(t *testing.T) {
	matchRepo := new(MockMatchRepo)
	playerRepo := new(MockPlayerRepo)
	svc := NewMatchService(matchRepo, playerRepo)
	ctx := context.Background()

	p1, p2 := newPlayerID(), newPlayerID()
	match := makeLiveMatch([]primitive.ObjectID{p1}, []primitive.ObjectID{p2})

	matchRepo.On("FindByID", ctx, match.ID).Return(match, nil)

	_, err := svc.UpdateScore(ctx, match.ID, 3, p1.Hex())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid team number")
}

func TestUpdateScore_FinishedMatch_Fails(t *testing.T) {
	matchRepo := new(MockMatchRepo)
	playerRepo := new(MockPlayerRepo)
	svc := NewMatchService(matchRepo, playerRepo)
	ctx := context.Background()

	p1, p2 := newPlayerID(), newPlayerID()
	match := makeLiveMatch([]primitive.ObjectID{p1}, []primitive.ObjectID{p2})
	match.Status = models.MatchStatusFinished

	matchRepo.On("FindByID", ctx, match.ID).Return(match, nil)

	_, err := svc.UpdateScore(ctx, match.ID, 1, p1.Hex())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not live")
}

// ── Swap logic tests (doubles) ──

func TestUpdateScore_Doubles_ConsecutiveScores_SwapPositions(t *testing.T) {
	matchRepo := new(MockMatchRepo)
	playerRepo := new(MockPlayerRepo)
	svc := NewMatchService(matchRepo, playerRepo)
	ctx := context.Background()

	p1, p2, p3, p4 := newPlayerID(), newPlayerID(), newPlayerID(), newPlayerID()
	match := makeLiveMatch([]primitive.ObjectID{p1, p2}, []primitive.ObjectID{p3, p4})

	matchRepo.On("FindByID", ctx, match.ID).Return(match, nil)
	matchRepo.On("Update", ctx, match).Return(nil)

	// First score by p1 — no swap, positions stay [p1, p2]
	result, err := svc.UpdateScore(ctx, match.ID, 1, p1.Hex())
	assert.NoError(t, err)
	assert.Equal(t, []string{p1.Hex(), p2.Hex()}, result.Team1Positions)

	// Second score by p1 — same player consecutive → swap to [p2, p1]
	result, err = svc.UpdateScore(ctx, match.ID, 1, p1.Hex())
	assert.NoError(t, err)
	assert.Equal(t, []string{p2.Hex(), p1.Hex()}, result.Team1Positions)
}

func TestUpdateScore_Doubles_DifferentScorers_NoSwap(t *testing.T) {
	matchRepo := new(MockMatchRepo)
	playerRepo := new(MockPlayerRepo)
	svc := NewMatchService(matchRepo, playerRepo)
	ctx := context.Background()

	p1, p2, p3, p4 := newPlayerID(), newPlayerID(), newPlayerID(), newPlayerID()
	match := makeLiveMatch([]primitive.ObjectID{p1, p2}, []primitive.ObjectID{p3, p4})

	matchRepo.On("FindByID", ctx, match.ID).Return(match, nil)
	matchRepo.On("Update", ctx, match).Return(nil)

	// Score by p1
	result, err := svc.UpdateScore(ctx, match.ID, 1, p1.Hex())
	assert.NoError(t, err)
	assert.Equal(t, []string{p1.Hex(), p2.Hex()}, result.Team1Positions)

	// Score by p2 — different player → no swap
	result, err = svc.UpdateScore(ctx, match.ID, 1, p2.Hex())
	assert.NoError(t, err)
	assert.Equal(t, []string{p1.Hex(), p2.Hex()}, result.Team1Positions)
}

func TestUpdateScore_Singles_ConsecutiveScores_NoSwap(t *testing.T) {
	matchRepo := new(MockMatchRepo)
	playerRepo := new(MockPlayerRepo)
	svc := NewMatchService(matchRepo, playerRepo)
	ctx := context.Background()

	p1, p2 := newPlayerID(), newPlayerID()
	match := makeLiveMatch([]primitive.ObjectID{p1}, []primitive.ObjectID{p2})

	matchRepo.On("FindByID", ctx, match.ID).Return(match, nil)
	matchRepo.On("Update", ctx, match).Return(nil)

	// Two consecutive scores by p1 in singles — no swap (only 1 position)
	svc.UpdateScore(ctx, match.ID, 1, p1.Hex())
	result, err := svc.UpdateScore(ctx, match.ID, 1, p1.Hex())

	assert.NoError(t, err)
	assert.Len(t, result.Team1Positions, 1)
	assert.Equal(t, p1.Hex(), result.Team1Positions[0])
}

// ── UndoScore tests ──

func TestUndoScore_Success(t *testing.T) {
	matchRepo := new(MockMatchRepo)
	playerRepo := new(MockPlayerRepo)
	svc := NewMatchService(matchRepo, playerRepo)
	ctx := context.Background()

	p1, p2 := newPlayerID(), newPlayerID()
	match := makeLiveMatch([]primitive.ObjectID{p1}, []primitive.ObjectID{p2})
	match.Score1 = 2
	match.ScoreHistory = []models.ScoreEvent{
		{Team: 1, PlayerID: p1.Hex()},
		{Team: 1, PlayerID: p1.Hex()},
	}

	matchRepo.On("FindByID", ctx, match.ID).Return(match, nil)
	matchRepo.On("Update", ctx, match).Return(nil)

	result, err := svc.UndoScore(ctx, match.ID)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.Score1)
	assert.Len(t, result.ScoreHistory, 1)
	// Serve returns to previous scorer
	assert.Equal(t, 1, result.ServingTeam)
	assert.Equal(t, p1.Hex(), result.ServingPlayerID)
}

func TestUndoScore_BackToInitial(t *testing.T) {
	matchRepo := new(MockMatchRepo)
	playerRepo := new(MockPlayerRepo)
	svc := NewMatchService(matchRepo, playerRepo)
	ctx := context.Background()

	p1, p2 := newPlayerID(), newPlayerID()
	match := makeLiveMatch([]primitive.ObjectID{p1}, []primitive.ObjectID{p2})
	match.Score1 = 1
	match.ScoreHistory = []models.ScoreEvent{
		{Team: 1, PlayerID: p1.Hex()},
	}

	matchRepo.On("FindByID", ctx, match.ID).Return(match, nil)
	matchRepo.On("Update", ctx, match).Return(nil)

	result, err := svc.UndoScore(ctx, match.ID)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.Score1)
	assert.Empty(t, result.ScoreHistory)
	// Back to initial: serve to team 1, first player
	assert.Equal(t, 1, result.ServingTeam)
	assert.Equal(t, p1.Hex(), result.ServingPlayerID)
}

func TestUndoScore_NoHistory_Fails(t *testing.T) {
	matchRepo := new(MockMatchRepo)
	playerRepo := new(MockPlayerRepo)
	svc := NewMatchService(matchRepo, playerRepo)
	ctx := context.Background()

	p1, p2 := newPlayerID(), newPlayerID()
	match := makeLiveMatch([]primitive.ObjectID{p1}, []primitive.ObjectID{p2})
	// Empty history

	matchRepo.On("FindByID", ctx, match.ID).Return(match, nil)

	_, err := svc.UndoScore(ctx, match.ID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no scores to undo")
}

func TestUndoScore_FinishedMatch_Fails(t *testing.T) {
	matchRepo := new(MockMatchRepo)
	playerRepo := new(MockPlayerRepo)
	svc := NewMatchService(matchRepo, playerRepo)
	ctx := context.Background()

	p1, p2 := newPlayerID(), newPlayerID()
	match := makeLiveMatch([]primitive.ObjectID{p1}, []primitive.ObjectID{p2})
	match.Status = models.MatchStatusFinished

	matchRepo.On("FindByID", ctx, match.ID).Return(match, nil)

	_, err := svc.UndoScore(ctx, match.ID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not live")
}

// ── FinishMatch tests ──

func TestFinishMatch_Success(t *testing.T) {
	matchRepo := new(MockMatchRepo)
	playerRepo := new(MockPlayerRepo)
	svc := NewMatchService(matchRepo, playerRepo)
	ctx := context.Background()

	p1, p2 := newPlayerID(), newPlayerID()
	match := makeLiveMatch([]primitive.ObjectID{p1}, []primitive.ObjectID{p2})

	matchRepo.On("FindByID", ctx, match.ID).Return(match, nil)
	matchRepo.On("Update", ctx, match).Return(nil)

	result, err := svc.FinishMatch(ctx, match.ID)

	assert.NoError(t, err)
	assert.Equal(t, models.MatchStatusFinished, result.Status)
	assert.NotNil(t, result.FinishedAt)
	matchRepo.AssertExpectations(t)
}

func TestFinishMatch_AlreadyFinished_Fails(t *testing.T) {
	matchRepo := new(MockMatchRepo)
	playerRepo := new(MockPlayerRepo)
	svc := NewMatchService(matchRepo, playerRepo)
	ctx := context.Background()

	p1, p2 := newPlayerID(), newPlayerID()
	match := makeLiveMatch([]primitive.ObjectID{p1}, []primitive.ObjectID{p2})
	match.Status = models.MatchStatusFinished

	matchRepo.On("FindByID", ctx, match.ID).Return(match, nil)

	_, err := svc.FinishMatch(ctx, match.ID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already finished")
}

// ── EditScore tests ──

func TestEditScore_Success(t *testing.T) {
	matchRepo := new(MockMatchRepo)
	playerRepo := new(MockPlayerRepo)
	svc := NewMatchService(matchRepo, playerRepo)
	ctx := context.Background()

	p1, p2 := newPlayerID(), newPlayerID()
	match := makeLiveMatch([]primitive.ObjectID{p1}, []primitive.ObjectID{p2})

	matchRepo.On("FindByID", ctx, match.ID).Return(match, nil)
	matchRepo.On("Update", ctx, match).Return(nil)

	result, err := svc.EditScore(ctx, match.ID, 21, 15)

	assert.NoError(t, err)
	assert.Equal(t, 21, result.Score1)
	assert.Equal(t, 15, result.Score2)
	matchRepo.AssertExpectations(t)
}

// ── AddResult tests ──

func TestAddResult_Success(t *testing.T) {
	matchRepo := new(MockMatchRepo)
	playerRepo := new(MockPlayerRepo)
	svc := NewMatchService(matchRepo, playerRepo)
	ctx := context.Background()

	p1, p2 := newPlayerID(), newPlayerID()
	groupID := primitive.NewObjectID()

	playerRepo.On("FindByID", ctx, p1).Return(&models.Player{ID: p1, Name: "Alice"}, nil)
	playerRepo.On("FindByID", ctx, p2).Return(&models.Player{ID: p2, Name: "Bob"}, nil)
	matchRepo.On("Create", ctx, mock.AnythingOfType("*models.Match")).Return(nil)

	match, err := svc.AddResult(ctx, groupID,
		[]primitive.ObjectID{p1},
		[]primitive.ObjectID{p2},
		21, 18)

	assert.NoError(t, err)
	assert.Equal(t, models.MatchStatusFinished, match.Status)
	assert.Equal(t, 21, match.Score1)
	assert.Equal(t, 18, match.Score2)
	assert.NotNil(t, match.FinishedAt)
	matchRepo.AssertExpectations(t)
}

// ── DeleteMatch tests ──

func TestDeleteMatch_Success(t *testing.T) {
	matchRepo := new(MockMatchRepo)
	playerRepo := new(MockPlayerRepo)
	svc := NewMatchService(matchRepo, playerRepo)
	ctx := context.Background()

	matchID := primitive.NewObjectID()
	matchRepo.On("Delete", ctx, matchID).Return(nil)

	err := svc.DeleteMatch(ctx, matchID)

	assert.NoError(t, err)
	matchRepo.AssertExpectations(t)
}

// ── Helper function tests ──

func TestToHexSlice(t *testing.T) {
	id1 := primitive.NewObjectID()
	id2 := primitive.NewObjectID()

	result := toHexSlice([]primitive.ObjectID{id1, id2})

	assert.Len(t, result, 2)
	assert.Equal(t, id1.Hex(), result[0])
	assert.Equal(t, id2.Hex(), result[1])
}

func TestToHexSlice_Empty(t *testing.T) {
	result := toHexSlice([]primitive.ObjectID{})
	assert.Empty(t, result)
}

// ── Undo with position rebuild in doubles ──

func TestUndoScore_Doubles_RebuildPositions(t *testing.T) {
	matchRepo := new(MockMatchRepo)
	playerRepo := new(MockPlayerRepo)
	svc := NewMatchService(matchRepo, playerRepo)
	ctx := context.Background()

	p1, p2, p3, p4 := newPlayerID(), newPlayerID(), newPlayerID(), newPlayerID()
	match := makeLiveMatch([]primitive.ObjectID{p1, p2}, []primitive.ObjectID{p3, p4})

	// Simulate: p1 scored twice (triggers swap), then we undo the second
	match.Score1 = 2
	match.ScoreHistory = []models.ScoreEvent{
		{Team: 1, PlayerID: p1.Hex()},
		{Team: 1, PlayerID: p1.Hex()},
	}
	// After two consecutive scores by p1, positions would be swapped
	match.Team1Positions = []string{p2.Hex(), p1.Hex()}

	matchRepo.On("FindByID", ctx, match.ID).Return(match, nil)
	matchRepo.On("Update", ctx, match).Return(nil)

	result, err := svc.UndoScore(ctx, match.ID)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.Score1)
	// After undo, only 1 score by p1 — no consecutive same-player, so positions should be initial
	assert.Equal(t, []string{p1.Hex(), p2.Hex()}, result.Team1Positions)
}
