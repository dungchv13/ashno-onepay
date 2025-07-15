package controller

import (
	"ashno-onepay/internal/config"
	"ashno-onepay/internal/controller/dto"
	"ashno-onepay/internal/errors"
	"ashno-onepay/internal/model"
	"ashno-onepay/internal/service"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/gin-gonic/gin"
)

type RegistrationController struct {
	registrationSvc service.RegistrationService
	config          *config.Config
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
	email := ctx.Query("email")
	option, err := u.registrationSvc.GetRegistrationOption(model.RegistrationOptionFilter{
		Category:               registrationOption,
		AttendGalaDinner:       attendGalaDinner,
		NumberAccompanyPersons: numberAccompanyPersons,
		Email:                  email,
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
// @Success 200 {object} dto.RegistrationResponse
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

// @Summary Export Registrations as XLSX
// @Id exportRegistrationsXLSX
// @Tags register
// @version 1.0
// @Param start_time query string false "Start time (YYYY-MM-DD)"
// @Param end_time query string false "End time (YYYY-MM-DD)"
// @Success 200 {file} xlsx "XLSX file"
// @Failure 500 {object} errors.AppError
// @Router /register/file [get]
func (u *RegistrationController) HandleGetFile(ctx *gin.Context) {
	startTimeStr := ctx.Query("start_time")
	endTimeStr := ctx.Query("end_time")

	var startTime, endTime time.Time
	var err error
	if startTimeStr != "" {
		startTime, err = time.Parse(time.DateOnly, startTimeStr)
		if err != nil {
			handleError(ctx, errors.ErrBadRequest.Reform("invalid start_time format, must be YYYY-MM-DD"))
			return
		}
	}
	if endTimeStr != "" {
		endTime, err = time.Parse(time.DateOnly, endTimeStr)
		if err != nil {
			handleError(ctx, errors.ErrBadRequest.Reform("invalid end_time format, must be YYYY-MM-DD"))
			return
		}
	}

	regs, err := u.registrationSvc.GetRegistrations(startTime, endTime)
	if err != nil {
		handleError(ctx, err)
		return
	}

	f := excelize.NewFile()
	sheet := "Registrations"
	f.SetSheetName(f.GetSheetName(0), sheet)
	headers := []string{"VerifyLink", "Category", "Nationality", "DoctorateDegree", "FirstName", "MiddleName", "LastName", "FullName", "DateOfBirth", "Institution", "Email", "PhoneNumber", "Sponsor", "PaymentStatus", "AccompanyPersons"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	for rowIdx, reg := range regs {
		var accompanyList []string
		for _, p := range reg.AccompanyPersons {
			accompanyList = append(accompanyList,
				p.FirstName+" "+p.MiddleName+" "+p.LastName+
					" (DOB: "+p.DateOfBirth+")",
			)
		}
		accompanyStr := strings.Join(accompanyList, "\n")

		verifyLinkCell, _ := excelize.CoordinatesToCellName(1, rowIdx+2)
		verifyURL := fmt.Sprintf("%s/%s", u.config.OnePay.ReturnURL, reg.Id)
		f.SetCellValue(sheet, verifyLinkCell, "VerifyLink")
		f.SetCellHyperLink(sheet, verifyLinkCell, verifyURL, "External")

		row := []interface{}{
			// skip the first column, already set
			reg.RegistrationCategory,
			reg.Nationality,
			reg.DoctorateDegree,
			reg.FirstName,
			reg.MiddleName,
			reg.LastName,
			reg.FirstName + " " + reg.MiddleName + " " + reg.LastName,
			reg.DateOfBirth,
			reg.Institution,
			reg.Email,
			reg.PhoneNumber,
			reg.Sponsor,
			reg.PaymentStatus,
			accompanyStr,
		}
		for colIdx, val := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+2, rowIdx+2)
			f.SetCellValue(sheet, cell, val)
		}
	}
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", "attachment; filename=registrations.xlsx")
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Expires", "0")
	if err := f.Write(ctx.Writer); err != nil {
		handleError(ctx, errors.ErrInternal.Wrap(err).Reform("failed to write xlsx file"))
		return
	}
}

func NewRegistrationController(registrationSvc service.RegistrationService, config *config.Config) *RegistrationController {
	return &RegistrationController{
		registrationSvc: registrationSvc,
		config:          config,
	}
}
