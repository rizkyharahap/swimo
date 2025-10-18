package training

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/rizkyharahap/swimo/pkg/middleware"
	"github.com/rizkyharahap/swimo/pkg/response"
	"github.com/rizkyharahap/swimo/pkg/validator"
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
// @Failure 404 {object} response.Message "Training not found"
// @Security ApiKeyAuth
// @Router /trainings/{id} [get]
func (h *TrainingHandler) GetById(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	training, err := h.trainingUseCase.GetById(r.Context(), id)
	if err != nil {
		if err == ErrTrainingNotFound {
			response.JSON(w, http.StatusNotFound, response.Message{Message: "Training not found"})
			return
		}

		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, response.Success{Data: training})
}

// GetTrainings handles getting paginated list of trainings
// @Summary Get trainings with pagination
// @Description Retrieve a paginated list of trainings with optional search and sorting
// @Tags Training
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Number of items per page" default(10) minimum(1) maximum(100)
// @Param sort query string false "Sort field and direction" Enums(name.asc,name.desc,level.asc,level.desc,created_at.asc,created_at.desc) default(created_at.desc)
// @Param search query string false "Search term for training name and description"
// @Success 200 {object} response.SuccessPagination{data=[]TrainingItemResponse} "Trainings retrieved successfully"
// @Failure 404 {object} response.SuccessPagination{data=[]TrainingItemResponse} "Training not found"
// @Security ApiKeyAuth
// @Router /trainings [get]
func (h *TrainingHandler) GetTrainings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters with default values
	query := TrainingsQuery{
		Page:  1,
		Limit: 10,
		Sort:  "created_at.desc",
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			query.Page = page
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			query.Limit = limit
		}
	}

	if sort := r.URL.Query().Get("sort"); sort != "" {
		query.Sort = sort
	}

	query.Search = r.URL.Query().Get("search")

	if err := query.Validate(); err != nil {
		response.ValidationError(w, err.Errors)
		return
	}

	// Get paginated trainings from usecase
	trainingItems, totalPages, err := h.trainingUseCase.GetTrainings(ctx, &query)
	if err != nil {
		if err == ErrTrainingNotFound {
			response.JSON(w, http.StatusNotFound, response.SuccessPagination{
				Data: trainingItems,
				Pagination: response.Pagination{
					Page:       query.Page,
					Limit:      query.Limit,
					TotalPages: totalPages,
				},
			})
			return
		}

		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, response.SuccessPagination{
		Data: trainingItems,
		Pagination: response.Pagination{
			Page:       query.Page,
			Limit:      query.Limit,
			TotalPages: totalPages,
		},
	})
}

// CreateTraining handles creating a new training
// @Summary Create a new training
// @Description Create a new training with the provided details
// @Tags Training
// @Accept json
// @Produce json
// @Param request body TrainingRequest true "Training creation request"
// @Success 201 {object} response.Success{data=TrainingResponse} "Training created successfully"
// @Failure 409 {object} response.Message "Training already exists"
// @Failure 422 {object} response.Error "Validation errors"
// @Security ApiKeyAuth
// @Router /trainings [post]
func (h *TrainingHandler) CreateTraining(w http.ResponseWriter, r *http.Request) {
	var req TrainingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w)
		return
	}

	if err := req.Validate(); err != nil {
		response.ValidationError(w, err.(*validator.ValidationError).Errors)
		return
	}

	training, err := h.trainingUseCase.CreateTraining(r.Context(), &req)
	if err != nil {
		if err == ErrorTrainingExists {
			response.JSON(w, http.StatusConflict, response.Message{Message: "Training already exists"})
			return
		}
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusCreated, response.Success{Data: training})
}

// GetLastTraining handles getting user's last training session
// @Summary Get user's last training session
// @Description Retrieve the most recent training session
// @Tags Training
// @Accept json
// @Produce json
// @Success 200 {object} response.Success{data=TrainingSessionResponse} "Last training session retrieved successfully"
// @Failure 404 {object} response.Message "No training sessions found"
// @Security ApiKeyAuth
// @Router /trainings/sessions/last [get]
func (h *TrainingHandler) GetLastSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claim := middleware.AuthFromContext(ctx)

	trainingSession, err := h.trainingUseCase.GetLastSession(ctx, *claim.Uid)
	if err != nil {
		if err == ErrTrainingSessionNotFound {
			response.JSON(w, http.StatusNotFound, response.Message{Message: "No training sessions found"})
			return
		}

		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusOK, response.Success{Data: trainingSession})
}

func (h *TrainingHandler) FinishSession(w http.ResponseWriter, r *http.Request) {
	var req TrainingFinishSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w)
		return
	}

	if err := req.Validate(); err != nil {
		response.ValidationError(w, err.(*validator.ValidationError).Errors)
		return
	}

	ctx := r.Context()
	claim := middleware.AuthFromContext(ctx)
	id := r.PathValue("id")

	training, err := h.trainingUseCase.FinishSession(r.Context(), *claim.Uid, id, &req)
	if err != nil {
		if err == ErrorTrainingExists {
			response.JSON(w, http.StatusConflict, response.Message{Message: "Training already exists"})
			return
		}
		response.InternalError(w)
		return
	}

	response.JSON(w, http.StatusCreated, response.Success{Data: training})
}
