package training

import (
	"net/http"

	"github.com/rizkyharahap/swimo/pkg/middleware"
	"github.com/rizkyharahap/swimo/pkg/response"
)

type TrainingHandler struct {
	trainingUseCase TrainingUsecase
}

func NewTrainingHandler(trainingUseCase TrainingUsecase) *TrainingHandler {
	return &TrainingHandler{trainingUseCase}
}

// GetById handles getting training by ID
// @Summary Get training by ID
// @Description Retrieve detailed training information by training ID
// @Tags Training
// @Accept json
// @Produce json
// @Param id path string true "Training ID" example("8c4a2d27-56e2-4ef3-8a6e-43b812345abc")
// @Success 200 {object} response.Success{data=TrainingResponse} "Training retrieved successfully"
// @Failure 401 {object} response.Error "Unauthorized"
// @Failure 404 {object} response.Error "Training not found"
// @Failure 500 {object} response.Error "Internal server error"
// @Security ApiKeyAuth
// @Router /training/{id} [get]
func (h *TrainingHandler) GetById(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	training, err := h.trainingUseCase.GetById(r.Context(), id)
	if err != nil {
		if err == ErrTrainingNotFound {
			response.JSON(w, http.StatusNotFound, response.Error{Message: "Training not found"})
			return
		}

		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, response.Success{Data: training})
}

// GetLastTraining handles getting user's last training session
// @Summary Get user's last training session
// @Description Retrieve the most recent training session for the authenticated user
// @Tags Training
// @Accept json
// @Produce json
// @Success 200 {object} response.Success{data=TrainingSessionResponse} "Last training session retrieved successfully"
// @Failure 404 {object} response.Error "No training sessions found"
// @Failure 401 {object} response.Error "Unauthorized"
// @Failure 500 {object} response.Error "Internal server error"
// @Security ApiKeyAuth
// @Router /training/last [get]
func (h *TrainingHandler) GetLastTraining(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claim := middleware.AuthFromContext(ctx)

	trainingSession, err := h.trainingUseCase.GetLastTraining(ctx, *claim.Uid)
	if err != nil {
		if err == ErrTrainingSessionNotFound {
			response.JSON(w, http.StatusNotFound, response.Error{Message: "No training sessions found"})
			return
		}

		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, response.Success{Data: trainingSession})
}
