package controller

import (
	"ashno-onepay/internal/controller/dto"
	"ashno-onepay/internal/errors"
	"ashno-onepay/internal/model"
	"ashno-onepay/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RegistrationController struct {
	registrationSvc service.RegistrationService
}

// @Summary Register a New User for the Event
// @Id register
// @Tags register
// @version 1.0
// @Param body body dto.RegistrationRequest true "body"
// @Success 200 {object} dto.RegistrationResponse
// @Failure 400 {object} errors.AppError
// @Failure 500 {object} errors.AppError
// @Router /register [post]
func (u *RegistrationController) HandleRegister(ctx *gin.Context) {
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

// @Summary Get Registration Information by ID
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

// @Summary OnePay Payment Notification (IPN) Handler
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

// @Summary Get Registration Option Details
// @Id getRegistrationOption
// @Tags register
// @version 1.0
// @Success 200 {object} model.RegistrationOption
// @Failure 400 {object} errors.AppError
// @Failure 500 {object} errors.AppError
// @Router /register/option [get]
func (u *RegistrationController) HandlerGetOption(ctx *gin.Context) {
	registrationOption := ctx.Query("registration_option")
	attendGalaDinner := ctx.Query("attend_gala_dinner") == "true"
	numberAccompanyPersons, _ := strconv.Atoi(ctx.Query("numbers_accompany_persons"))
	option, err := u.registrationSvc.GetRegistrationOption(model.RegistrationOptionFilter{
		Category:               registrationOption,
		AttendGalaDinner:       attendGalaDinner,
		NumberAccompanyPersons: numberAccompanyPersons,
	})
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, option)
}

// @Summary Register Accompanying Persons for an Existing Registration
// @Id registerAccompanyPersons
// @Tags register
// @version 1.0
// @Param body body dto.AccompanyPersonRegistrationRequest true "body"
// @Success 200 {object} dto.AccompanyPersonRegistrationResponse
// @Failure 400 {object} errors.AppError
// @Failure 500 {object} errors.AppError
// @Router /register/accompany-persons [post]
func (u *RegistrationController) HandleRegisterAccompanyPersons(ctx *gin.Context) {
	var req dto.AccompanyPersonRegistrationRequest
	if err := ctx.BindJSON(&req); err != nil {
		handleError(ctx, errors.ErrBadRequest.Wrap(err).Reform("json marshal failed"))
		return
	}
	clientIP := ctx.ClientIP()
	paymentURL, err := u.registrationSvc.RegisterForAccompanyPersons(req.Email, req.AccompanyPersons, clientIP)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, dto.RegistrationResponse{
		PaymentURL: paymentURL,
	})
}

func NewRegistrationController(registrationSvc service.RegistrationService) *RegistrationController {
	return &RegistrationController{
		registrationSvc: registrationSvc,
	}
}
