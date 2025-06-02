package controller

import (
	"ashno-onepay/internal/controller/dto"
	"ashno-onepay/internal/errors"
	"ashno-onepay/internal/model"
	"ashno-onepay/internal/service"
	"github.com/gin-gonic/gin"
	"log"
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
	queryParams := ctx.Request.URL.Query()

	// Extract SecureHash
	receivedHash := queryParams.Get("vpc_SecureHash")
	if receivedHash == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing vpc_SecureHash"})
		return
	}

	// Remove vpc_SecureHash from map for HMAC calculation
	deleteQuery := ctx.Request.URL.Query()
	deleteQuery.Del("vpc_SecureHash")

	// Convert to map[string]string
	paramMap := make(map[string]string)
	for key := range deleteQuery {
		paramMap[key] = deleteQuery.Get(key)
	}

	// Validate HMAC
	if !u.registrationSvc.ValidateHMAC(paramMap, receivedHash) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Invalid signature"})
		return
	}

	// Transaction result handling
	txnRef := paramMap["vpc_MerchTxnRef"] // registrationID
	txnCode := paramMap["vpc_TxnResponseCode"]
	message := paramMap["vpc_Message"]

	if txnCode == "0" {
		// Payment Success
		log.Println("Payment Success for ", txnRef)
		err := u.registrationSvc.UpdatePaymentStatus(txnRef, string(model.PaymentStatusDone))
		if err != nil {
			handleError(ctx, errors.ErrInternal.Wrap(err))
			return
		}
		ctx.String(http.StatusOK, "responsecode=1&desc=confirm-success")
	} else {
		// Payment Failed
		log.Printf("Payment Failed for %s: %s", txnRef, message)
		err := u.registrationSvc.UpdatePaymentStatus(txnRef, string(model.PaymentStatusFail))
		if err != nil {
			handleError(ctx, errors.ErrInternal.Wrap(err))
			return
		}
		ctx.String(http.StatusOK, "payment_failed")
	}
}

func (u *RegistrationController) Test(ctx *gin.Context) {
	txnRef := ctx.Query("vpc_MerchTxnRef") // registrationID

	// Payment Success
	log.Println("Payment Success for ", txnRef)
	err := u.registrationSvc.UpdatePaymentStatus(txnRef, string(model.PaymentStatusDone))
	if err != nil {
		handleError(ctx, errors.ErrInternal.Wrap(err))
		return
	}
	ctx.String(http.StatusOK, "responsecode=1&desc=confirm-success")
}

func NewRegistrationController(registrationSvc service.RegistrationService) *RegistrationController {
	return &RegistrationController{
		registrationSvc: registrationSvc,
	}
}
