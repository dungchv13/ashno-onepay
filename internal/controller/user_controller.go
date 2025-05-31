package controller

import (
	"ashno-onepay/internal/controller/dto"
	"ashno-onepay/internal/errors"
	"ashno-onepay/internal/jwt"
	"ashno-onepay/internal/model"
	"ashno-onepay/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UserController struct {
	userSvc     service.UserService
	tokenIssuer jwt.Issuer
}

// @Summary Login
// @Id login
// @Tags User
// @version 1.0
// @Param body body dto.LoginRequest true "body"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} errors.AppError
// @Failure 500 {object} errors.AppError
// @Router /user/login [post]
func (u *UserController) HandleLogin(ctx *gin.Context) {
	var req dto.LoginRequest
	if err := ctx.BindJSON(&req); err != nil {
		handleError(ctx, errors.ErrBadRequest.Wrap(err).Reform("json marshal failed"))
		return
	}

	user, err := u.userSvc.UserLogin(req.Email, req.Password)
	if err != nil {
		handleError(ctx, errors.ErrInvalidIdentifier.Wrap(err))
		return
	}

	token, err := u.tokenIssuer.Issue(ctx, jwt.NewUserClaims(user, jwt.UserRole))
	if err != nil {
		handleError(ctx, errors.ErrOtherService.Wrap(err))
		return
	}

	ctx.JSON(http.StatusOK, dto.LoginResponse{
		Id:    int64(user.Id),
		Role:  jwt.UserRole,
		Token: token,
	})
}

// @Summary Register
// @Id register
// @Tags User
// @version 1.0
// @Param body body dto.RegisterRequest true "body"
// @Success 201
// @Failure 400 {object} errors.AppError
// @Failure 500 {object} errors.AppError
// @Router /user/register [post]
func (u *UserController) HandleRegister(ctx *gin.Context) {
	var req dto.RegisterRequest
	if err := ctx.BindJSON(&req); err != nil {
		handleError(ctx, errors.ErrBadRequest.Wrap(err).Reform("json marshal failed"))
		return
	}

	err := u.userSvc.UserRegister(model.User{
		Email:       req.Email,
		Password:    req.Password,
		Phone:       req.Phone,
		Nationality: req.Nationality,
		FirstName:   req.FirstName,
		MiddleName:  req.MiddleName,
		LastName:    req.LastName,
		DateOfBirth: req.DateOfBirth,
		Institution: req.Institution,
		SponsorBy:   req.SponsorBy,
	})
	if err != nil {
		handleError(ctx, errors.ErrInternal.Wrap(err))
		return
	}

	ctx.Status(http.StatusCreated)
}

func NewUserController(tokenIssuer jwt.Issuer, userSvc service.UserService) *UserController {
	return &UserController{
		userSvc:     userSvc,
		tokenIssuer: tokenIssuer,
	}
}
