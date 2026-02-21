package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"gully-backend/models"
	"gully-backend/repositories"
)

type MatchService struct {
	matchRepo  repositories.MatchRepository
	playerRepo repositories.PlayerRepository
}

func NewMatchService(matchRepo repositories.MatchRepository, playerRepo repositories.PlayerRepository) *MatchService {
	return &MatchService{matchRepo: matchRepo, playerRepo: playerRepo}
}

// CreateMatch creates a live match supporting 1v1, 1v2, or 2v2.
func (s *MatchService) CreateMatch(ctx context.Context, groupID primitive.ObjectID, team1IDs, team2IDs []primitive.ObjectID) (*models.Match, error) {
	if len(team1IDs) == 0 || len(team2IDs) == 0 {
		return nil, errors.New("each team must have at least 1 player")
	}
	if len(team1IDs) > 2 || len(team2IDs) > 2 {
		return nil, errors.New("each team can have at most 2 players")
	}

	// Look up player names
	team1Names, err := s.resolvePlayerNames(ctx, team1IDs)
	if err != nil {
		return nil, fmt.Errorf("team 1: %w", err)
	}
	team2Names, err := s.resolvePlayerNames(ctx, team2IDs)
	if err != nil {
		return nil, fmt.Errorf("team 2: %w", err)
	}

	now := time.Now()
	match := &models.Match{
		GroupID:         groupID,
		Team1IDs:        team1IDs,
		Team2IDs:        team2IDs,
		Team1Names:      team1Names,
		Team2Names:      team2Names,
		Score1:          0,
		Score2:          0,
		ScoreHistory:    []models.ScoreEvent{},
		ServingTeam:     1,
		ServingPlayerID: team1IDs[0].Hex(),
		Team1Positions:  toHexSlice(team1IDs),
		Team2Positions:  toHexSlice(team2IDs),
		Status:          models.MatchStatusLive,
		StartedAt:       now,
	}
	if err := s.matchRepo.Create(ctx, match); err != nil {
		return nil, err
	}
	return match, nil
}

// AddResult creates a finished match with final scores (for past matches).
func (s *MatchService) AddResult(ctx context.Context, groupID primitive.ObjectID, team1IDs, team2IDs []primitive.ObjectID, score1, score2 int) (*models.Match, error) {
	team1Names, err := s.resolvePlayerNames(ctx, team1IDs)
	if err != nil {
		return nil, fmt.Errorf("team 1: %w", err)
	}
	team2Names, err := s.resolvePlayerNames(ctx, team2IDs)
	if err != nil {
		return nil, fmt.Errorf("team 2: %w", err)
	}

	now := time.Now()
	match := &models.Match{
		GroupID:         groupID,
		Team1IDs:        team1IDs,
		Team2IDs:        team2IDs,
		Team1Names:      team1Names,
		Team2Names:      team2Names,
		Score1:          score1,
		Score2:          score2,
		ScoreHistory:    []models.ScoreEvent{},
		ServingTeam:     0,
		ServingPlayerID: "",
		Team1Positions:  toHexSlice(team1IDs),
		Team2Positions:  toHexSlice(team2IDs),
		Status:          models.MatchStatusFinished,
		StartedAt:       now,
		FinishedAt:      &now,
		DurationSecs:    0,
	}
	if err := s.matchRepo.Create(ctx, match); err != nil {
		return nil, err
	}
	return match, nil
}

func (s *MatchService) GetMatches(ctx context.Context, groupID primitive.ObjectID) ([]models.Match, error) {
	return s.matchRepo.FindByGroupID(ctx, groupID)
}

func (s *MatchService) GetMatch(ctx context.Context, id primitive.ObjectID) (*models.Match, error) {
	return s.matchRepo.FindByID(ctx, id)
}

// UpdateScore increments the score for the given team and records the scorer.
// All position/serve logic is handled by the frontend from scoreHistory.
func (s *MatchService) UpdateScore(ctx context.Context, matchID primitive.ObjectID, team int, scorerID string) (*models.Match, error) {
	match, err := s.matchRepo.FindByID(ctx, matchID)
	if err != nil {
		return nil, err
	}
	if match.Status != models.MatchStatusLive {
		return nil, errors.New("match is not live")
	}

	event := models.ScoreEvent{
		Team:     team,
		PlayerID: scorerID,
	}

	switch team {
	case 1:
		match.Score1++
	case 2:
		match.Score2++
	default:
		return nil, fmt.Errorf("invalid team number: %d", team)
	}

	match.ScoreHistory = append(match.ScoreHistory, event)

	// Simple serve tracking (frontend computes the real server)
	match.ServingTeam = team
	match.ServingPlayerID = scorerID

	if err := s.matchRepo.Update(ctx, match); err != nil {
		return nil, err
	}
	return match, nil
}

// UndoScore reverts the last score entry.
func (s *MatchService) UndoScore(ctx context.Context, matchID primitive.ObjectID) (*models.Match, error) {
	match, err := s.matchRepo.FindByID(ctx, matchID)
	if err != nil {
		return nil, err
	}
	if match.Status != models.MatchStatusLive {
		return nil, errors.New("match is not live")
	}
	if len(match.ScoreHistory) == 0 {
		return nil, errors.New("no scores to undo")
	}

	last := match.ScoreHistory[len(match.ScoreHistory)-1]
	match.ScoreHistory = match.ScoreHistory[:len(match.ScoreHistory)-1]

	switch last.Team {
	case 1:
		match.Score1--
	case 2:
		match.Score2--
	}

	// Restore simple serve state from previous event
	if len(match.ScoreHistory) > 0 {
		prev := match.ScoreHistory[len(match.ScoreHistory)-1]
		match.ServingTeam = prev.Team
		match.ServingPlayerID = prev.PlayerID
	} else {
		match.ServingTeam = 1
		if len(match.Team1IDs) > 0 {
			match.ServingPlayerID = match.Team1IDs[0].Hex()
		}
	}

	if err := s.matchRepo.Update(ctx, match); err != nil {
		return nil, err
	}
	return match, nil
}

// FinishMatch marks the match as finished and records the duration.
func (s *MatchService) FinishMatch(ctx context.Context, matchID primitive.ObjectID) (*models.Match, error) {
	match, err := s.matchRepo.FindByID(ctx, matchID)
	if err != nil {
		return nil, err
	}
	if match.Status != models.MatchStatusLive {
		return nil, errors.New("match is already finished")
	}

	now := time.Now()
	match.Status = models.MatchStatusFinished
	match.FinishedAt = &now
	match.DurationSecs = int(now.Sub(match.StartedAt).Seconds())

	if err := s.matchRepo.Update(ctx, match); err != nil {
		return nil, err
	}
	return match, nil
}

// DeleteMatch removes a match.
func (s *MatchService) DeleteMatch(ctx context.Context, matchID primitive.ObjectID) error {
	return s.matchRepo.Delete(ctx, matchID)
}

// EditScore directly sets scores (admin only).
func (s *MatchService) EditScore(ctx context.Context, matchID primitive.ObjectID, score1, score2 int) (*models.Match, error) {
	match, err := s.matchRepo.FindByID(ctx, matchID)
	if err != nil {
		return nil, err
	}
	match.Score1 = score1
	match.Score2 = score2
	if err := s.matchRepo.Update(ctx, match); err != nil {
		return nil, err
	}
	return match, nil
}

// ── Helpers ──

func (s *MatchService) resolvePlayerNames(ctx context.Context, ids []primitive.ObjectID) ([]string, error) {
	names := make([]string, len(ids))
	for i, id := range ids {
		p, err := s.playerRepo.FindByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("player %s not found", id.Hex())
		}
		names[i] = p.Name
	}
	return names, nil
}

func toHexSlice(ids []primitive.ObjectID) []string {
	s := make([]string, len(ids))
	for i, id := range ids {
		s[i] = id.Hex()
	}
	return s
}
