package controller

import (
	"ashno-onepay/internal/controller/dto"
	"ashno-onepay/internal/errors"
	"ashno-onepay/internal/model"
	"ashno-onepay/internal/service"
	"fmt"
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
// @Param body body dto.RegistrationRequest true "body"
// @Success 200 {object} dto.RegistrationResponse
// @Failure 400 {object} errors.AppError
// @Failure 500 {object} errors.AppError
// @Router /register [post]
func (u *RegistrationController) HandleRegister(ctx *gin.Context) {
	fmt.Println("--------------")
	var req dto.RegistrationRequest
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

// @Summary Get Registration Info
// @Id getRegistrationInfo
// @Tags register
// @version 1.0
// @Param registerID path string true "registerID"
// @Success 200 {object} model.Registration
// @Failure 400 {object} errors.AppError
// @Failure 500 {object} errors.AppError
// @Router /register/{registerID}/registration-info [get]
func (u *RegistrationController) HandlerGetRegistrationInfo(ctx *gin.Context) {
	registerID := ctx.Param("registerID")

	reg, err := u.registrationSvc.GetRegistration(registerID)
	if err != nil {
		handleError(ctx, errors.ErrInternal.Wrap(err))
		return
	}

	ctx.JSON(http.StatusOK, reg)
}

// @Summary OnePayIPN
// @Id onePayIPN
// @Tags register
// @version 1.0
// @Success 200 {string} string "responsecode=1&desc=confirm-success"
// @Failure 400 {object} errors.AppError
// @Failure 500 {object} errors.AppError
// @Router /onepay/ipn [get]
func (u *RegistrationController) HandlerOnePayIPN(ctx *gin.Context) {
	err := u.registrationSvc.OnePayVerifySecureHash(ctx.Request.URL)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.String(http.StatusOK, "responsecode=1&desc=confirm-success")
}

// @Summary OnePayIPN
// @Id onePayIPN
// @Tags register
// @version 1.0
// @Success 200 {object} model.RegistrationOption
// @Failure 400 {object} errors.AppError
// @Failure 500 {object} errors.AppError
// @Router /onepay/ipn [get]
func (u *RegistrationController) HandlerGetOption(ctx *gin.Context) {
	registrationOption := ctx.Query("registration_option")
	attendGalaDinner := ctx.Query("attend_gala_dinner") == "true"
	option, err := u.registrationSvc.GetRegistrationOption(model.RegistrationOptionFilter{
		Category:         registrationOption,
		AttendGalaDinner: attendGalaDinner,
	})
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, option)
}

func NewRegistrationController(registrationSvc service.RegistrationService) *RegistrationController {
	return &RegistrationController{
		registrationSvc: registrationSvc,
	}
}
