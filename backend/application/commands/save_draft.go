package commands

import (
	"context"
	"fmt"

	"github.com/agopalakrishnan/teams360/backend/domain/healthcheck"
)

// SaveDraftCommand represents the command to upsert an in-progress survey draft.
// Unlike SubmitHealthCheckCommand, responses may be partial (a dimension may have
// no trend, or no score yet), since this is autosaved before the survey is complete.
type SaveDraftCommand struct {
	TeamID           string
	UserID           string
	SurveyType       string
	AssessmentPeriod string
	CurrentDimension int
	Responses        []HealthCheckResponseCommand
	ClientUpdatedAt  int64
}

// SaveDraftHandler handles the save draft command
type SaveDraftHandler struct {
	repository healthcheck.Repository
}

// NewSaveDraftHandler creates a new command handler
func NewSaveDraftHandler(repository healthcheck.Repository) *SaveDraftHandler {
	return &SaveDraftHandler{repository: repository}
}

// Handle executes the command
func (h *SaveDraftHandler) Handle(ctx context.Context, cmd SaveDraftCommand) (*healthcheck.HealthCheckDraft, error) {
	if err := h.validate(cmd); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	surveyType := cmd.SurveyType
	if surveyType == "" {
		surveyType = healthcheck.SurveyTypeIndividual
	}

	draft := &healthcheck.HealthCheckDraft{
		TeamID:           cmd.TeamID,
		UserID:           cmd.UserID,
		SurveyType:       surveyType,
		AssessmentPeriod: cmd.AssessmentPeriod,
		CurrentDimension: cmd.CurrentDimension,
		Responses:        make([]healthcheck.HealthCheckResponse, len(cmd.Responses)),
		ClientUpdatedAt:  cmd.ClientUpdatedAt,
	}

	for i, resp := range cmd.Responses {
		draft.Responses[i] = healthcheck.HealthCheckResponse{
			DimensionID: resp.DimensionID,
			Score:       resp.Score,
			Trend:       resp.Trend,
			Comment:     resp.Comment,
		}
	}

	if err := h.repository.SaveDraft(ctx, draft); err != nil {
		return nil, fmt.Errorf("failed to save draft: %w", err)
	}

	return draft, nil
}

// validate ensures the command is valid. Score/trend are intentionally not range-checked here
// (unlike SubmitHealthCheckCommand) because a draft may hold partial, in-progress answers.
func (h *SaveDraftHandler) validate(cmd SaveDraftCommand) error {
	if cmd.TeamID == "" {
		return fmt.Errorf("teamId is required")
	}

	if cmd.UserID == "" {
		return fmt.Errorf("userId is required")
	}

	if cmd.AssessmentPeriod == "" {
		return fmt.Errorf("assessmentPeriod is required")
	}

	if cmd.SurveyType != "" && cmd.SurveyType != healthcheck.SurveyTypeIndividual && cmd.SurveyType != healthcheck.SurveyTypePostWorkshop {
		return fmt.Errorf("surveyType must be 'individual' or 'post_workshop'")
	}

	if cmd.ClientUpdatedAt <= 0 {
		return fmt.Errorf("clientUpdatedAt is required")
	}

	for i, resp := range cmd.Responses {
		if resp.DimensionID == "" {
			return fmt.Errorf("response %d: dimensionId is required", i)
		}
		if resp.Score < 0 || resp.Score > 3 {
			return fmt.Errorf("response %d: score must be between 0 and 3", i)
		}
	}

	return nil
}
