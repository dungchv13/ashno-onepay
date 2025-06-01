package controller

import (
	"ashno-onepay/internal/controller/dto"
	"ashno-onepay/internal/errors"
	"ashno-onepay/internal/model"
	"ashno-onepay/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type RegistrationController struct {
	registrationSvc service.RegistrationService
}

// @Summary Register
// @Id register
// @Tags register
// @version 1.0
// @Param body body model.Registration true "body"
// @Success 200 {object} dto.RegistrationResponse
// @Failure 400 {object} errors.AppError
// @Failure 500 {object} errors.AppError
// @Router /register [post]
func (u *RegistrationController) HandleRegister(ctx *gin.Context) {
	var req model.Registration
	if err := ctx.BindJSON(&req); err != nil {
		handleError(ctx, errors.ErrBadRequest.Wrap(err).Reform("json marshal failed"))
		return
	}
	clientIP := ctx.ClientIP()

	url, userID, err := u.registrationSvc.Register(req, clientIP)
	if err != nil {
		handleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.RegistrationResponse{
		PaymentURL: url,
		UserID:     userID,
	})
}

// @Summary Register
// @Id register
// @Tags register
// @version 1.0
// @Param userID path string true "userID"
// @Success 200 {object} model.Registration
// @Failure 400 {object} errors.AppError
// @Failure 500 {object} errors.AppError
// @Router /user/{userID}/registration-info [get]
func (u *RegistrationController) GetRegistrationInfo(ctx *gin.Context) {
	userID := ctx.Param("userID")

	reg, err := u.registrationSvc.GetRegistration(userID)
	if err != nil {
		handleError(ctx, errors.ErrInternal.Wrap(err))
		return
	}

	ctx.JSON(http.StatusOK, reg)
}

func (u *RegistrationController) OnePayIPN(ctx *gin.Context) {
	userID := ctx.Param("userID")

	reg, err := u.registrationSvc.GetRegistration(userID)
	if err != nil {
		handleError(ctx, errors.ErrInternal.Wrap(err))
		return
	}

	ctx.JSON(http.StatusOK, reg)
}

func NewRegistrationController(registrationSvc service.RegistrationService) *RegistrationController {
	return &RegistrationController{
		registrationSvc: registrationSvc,
	}
}
