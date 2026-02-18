package services

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"gully-backend/models"
	"gully-backend/repositories"
)

type MatchService struct {
	matchRepo  *repositories.MatchRepo
	playerRepo *repositories.PlayerRepo
}

func NewMatchService(matchRepo *repositories.MatchRepo, playerRepo *repositories.PlayerRepo) *MatchService {
	return &MatchService{matchRepo: matchRepo, playerRepo: playerRepo}
}

func (s *MatchService) CreateMatch(ctx context.Context, groupID, player1ID, player2ID primitive.ObjectID) (*models.Match, error) {
	p1, err := s.playerRepo.FindByID(ctx, player1ID)
	if err != nil {
		return nil, errors.New("player 1 not found")
	}
	p2, err := s.playerRepo.FindByID(ctx, player2ID)
	if err != nil {
		return nil, errors.New("player 2 not found")
	}

	match := &models.Match{
		GroupID:      groupID,
		Player1ID:    player1ID,
		Player2ID:    player2ID,
		Player1Name:  p1.Name,
		Player2Name:  p2.Name,
		Score1:       0,
		Score2:       0,
		ScoreHistory: []string{},
		Status:       models.MatchStatusLive,
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

// UpdateScore increments the score for the given player (1 or 2).
func (s *MatchService) UpdateScore(ctx context.Context, matchID primitive.ObjectID, player int) (*models.Match, error) {
	match, err := s.matchRepo.FindByID(ctx, matchID)
	if err != nil {
		return nil, err
	}
	if match.Status != models.MatchStatusLive {
		return nil, errors.New("match is not live")
	}

	switch player {
	case 1:
		match.Score1++
		match.ScoreHistory = append(match.ScoreHistory, "p1")
	case 2:
		match.Score2++
		match.ScoreHistory = append(match.ScoreHistory, "p2")
	default:
		return nil, fmt.Errorf("invalid player number: %d", player)
	}

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

	switch last {
	case "p1":
		match.Score1--
	case "p2":
		match.Score2--
	}

	if err := s.matchRepo.Update(ctx, match); err != nil {
		return nil, err
	}
	return match, nil
}

// FinishMatch marks the match as finished.
func (s *MatchService) FinishMatch(ctx context.Context, matchID primitive.ObjectID) (*models.Match, error) {
	match, err := s.matchRepo.FindByID(ctx, matchID)
	if err != nil {
		return nil, err
	}
	if match.Status != models.MatchStatusLive {
		return nil, errors.New("match is already finished")
	}

	match.Status = models.MatchStatusFinished
	if err := s.matchRepo.Update(ctx, match); err != nil {
		return nil, err
	}
	return match, nil
}
