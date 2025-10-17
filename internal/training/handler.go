package training

import (
	"net/http"
	"strconv"

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

// GetLastTraining handles getting user's last training session
// @Summary Get user's last training session
// @Description Retrieve the most recent training session
// @Tags Training
// @Accept json
// @Produce json
// @Success 200 {object} response.Success{data=TrainingSessionResponse} "Last training session retrieved successfully"
// @Failure 404 {object} response.Message "No training sessions found"
// @Security ApiKeyAuth
// @Router /trainings/last [get]
func (h *TrainingHandler) GetLastTraining(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claim := middleware.AuthFromContext(ctx)

	trainingSession, err := h.trainingUseCase.GetLastTraining(ctx, *claim.Uid)
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
