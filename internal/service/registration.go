package service

import (
	"ashno-onepay/internal/config"
	"ashno-onepay/internal/controller/dto"
	errs "ashno-onepay/internal/errors"
	"ashno-onepay/internal/model"
	"ashno-onepay/internal/repository"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RegistrationService interface {
	Register(registration dto.RegistrationRequest, clientIP string) (string, string, error)
	GetRegistration(ID string) (*model.Registration, error)
	UpdatePaymentStatus(ID, status string) error
	OnePayVerifySecureHash(u *url.URL) error
}

type registrationService struct {
	registrationRepo        repository.RegistrationRepository
	registrationOptionsRepo repository.RegistrationOptionRepository
	config                  *config.Config
}

func (r registrationService) OnePayVerifySecureHash(u *url.URL) error {
	queryParams := u.Query()
	queryParamsMap := make(map[string]string)
	for k, v := range queryParams {
		queryParamsMap[k] = strings.Join(v, "")
	}
	// Transaction result handling
	regID := queryParamsMap["vpc_MerchTxnRef"]
	if regID == "" {
		return errs.ErrBadRequest
	}
	reg, err := r.registrationRepo.GetRegistration(regID)
	if err != nil {
		return err
	}
	op := r.config.OnePayUSD
	if reg.Nationality == model.NationalityVietNam {
		op = r.config.OnePayVND
	}
	queryMapSorted := sortParams(queryParamsMap)
	stringToHash := generateStringToHash(queryMapSorted)
	onePaySecureHash := generateSecureHash(stringToHash, op.HashCode)
	merchantSecureHash := queryParamsMap["vpc_SecureHash"]
	fmt.Println("OnePay's Hash: ", onePaySecureHash)
	fmt.Println("Merchant's Hash: ", merchantSecureHash)
	if onePaySecureHash != merchantSecureHash {
		return errs.ErrForbidden.Reform("Invalid signature")
	}
	txnCode := queryParamsMap["vpc_TxnResponseCode"]
	message := queryParamsMap["vpc_Message"]
	var status string
	if txnCode == "0" {
		log.Println("Payment Success for ", regID)
		status = string(model.PaymentStatusDone)
		go func() {
			err := SendPaymentSuccessEmailWithQR(
				reg.Email, reg.FirstName, reg.Id,
				r.config.Server.Host+":"+r.config.Server.Port,
				r.config.SendGrip.ApiKey,
			)
			if err != nil {
				log.Printf("Send QR Failed for %s", reg.Id)
			}
		}()
	} else {
		log.Printf("Payment Failed for %s: %s", regID, message)
		status = string(model.PaymentStatusFail)
	}

	return r.registrationRepo.UpdatePaymentStatus(regID, status)
}

func (r registrationService) UpdatePaymentStatus(ID, status string) error {
	reg, err := r.GetRegistration(ID)
	if err != nil {
		return err
	}
	if status == string(model.PaymentStatusDone) {
		go func() {
			err := SendPaymentSuccessEmailWithQR(
				reg.Email, reg.FirstName, reg.Id,
				r.config.Server.Host+":"+r.config.Server.Port,
				r.config.SendGrip.ApiKey,
			)
			if err != nil {
				log.Printf("Send QR Failed for %s", reg.Id)
			}
		}()
	}
	return r.registrationRepo.UpdatePaymentStatus(ID, status)
}

func (r registrationService) GetRegistration(ID string) (*model.Registration, error) {
	return r.registrationRepo.GetRegistration(ID)
}

func (r registrationService) Register(request dto.RegistrationRequest, clientIP string) (string, string, error) {
	// check email registered
	oldReg, err := r.registrationRepo.GetByEmail(request.Email)
	if err != nil {
		return "", "", err
	}
	if oldReg != nil {
		if oldReg.PaymentStatus == string(model.PaymentStatusDone) {
			return "", "", errs.ErrInternal.Reform("email registered")
		}
		err = r.registrationRepo.Remove(oldReg.Id)
	}

	// setup registration
	reg, err := r.setupRegistration(request)
	if err != nil {
		return "", "", err
	}
	// generate paymentURL
	paymentURL, err := r.generatePaymentURL(&reg, clientIP)
	if err != nil {
		return "", "", err
	}
	// remove old request + insert registration
	_, err = r.registrationRepo.Create(reg)
	if err != nil {
		return "", "", err
	}
	return paymentURL, reg.Id, nil
}

func (r registrationService) setupRegistration(request dto.RegistrationRequest) (model.Registration, error) {
	reg := model.Registration{
		RegistrationCategory: request.RegistrationCategory,
		Nationality:          request.Nationality,
		DoctorateDegree:      request.DoctorateDegree,
		FirstName:            request.FirstName,
		MiddleName:           request.MiddleName,
		LastName:             request.LastName,
		DateOfBirth:          request.DateOfBirth,
		Institution:          request.Institution,
		Email:                request.Email,
		PhoneNumber:          request.PhoneNumber,
		Sponsor:              request.Sponsor,
	}
	reg.Id = uuid.New().String()
	OptionFilter := model.RegistrationOptionFilter{}
	switch request.RegistrationCategory {
	case string(model.DoctorCategory):
		OptionFilter.Category = string(model.DoctorCategory)
		if request.AttendGalaDinner {
			OptionFilter.Category = string(model.DoctorAndDinnerCategory)
		}
		OptionFilter.Subtype = DetermineRegistrationPeriod(time.Now())
	case string(model.StudentCategory):
		OptionFilter.Category = string(model.StudentCategory)
		if request.AttendGalaDinner {
			OptionFilter.Category = string(model.StudentAndDinnerCategory)
		}
	default:
		return model.Registration{}, errs.ErrNotFound.Reform("option not found")
	}

	option, err := r.registrationOptionsRepo.Find(OptionFilter)
	if err != nil {
		return model.Registration{}, errs.ErrNotFound.Reform("option not found")
	}
	reg.RegistrationOptionID = option.Id
	reg.RegistrationOption = *option

	return reg, nil
}

func DetermineRegistrationPeriod(now time.Time) string {
	earlyBirdEnd := time.Date(2025, 8, 31, 23, 59, 59, 0, time.UTC)
	regularEnd := time.Date(2025, 10, 31, 23, 59, 59, 0, time.UTC)
	onSiteStart := time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC)

	switch {
	case now.Before(earlyBirdEnd):
		return string(model.EarlyBird)
	case now.Before(regularEnd):
		return string(model.Regular)
	case now.After(onSiteStart):
		return string(model.OnSite)
	default:
		return string(model.Regular)
	}
}

func (r registrationService) generatePaymentURL(reg *model.Registration, clientIP string) (string, error) {
	op := r.config.OnePayUSD
	locale := "en"
	currency := "USD"
	amount := strconv.FormatFloat(reg.RegistrationOption.FeeUSD*100, 'f', 2, 64)
	if reg.Nationality == model.NationalityVietNam {
		locale = "vn"
		currency = "VND"
		amount = strconv.FormatInt(reg.RegistrationOption.FeeVND*100, 10)
		op = r.config.OnePayVND
	}
	merchantQueryMap := map[string]string{
		"vpc_Version":     "2",
		"vpc_Currency":    currency, // "USD" or "VND"
		"vpc_Command":     "pay",
		"vpc_AccessCode":  op.AccessCode,
		"vpc_Merchant":    op.MerchantID,
		"vpc_Locale":      locale, // e.g., "en" or "vn"
		"vpc_ReturnURL":   op.ReturnURL,
		"vpc_MerchTxnRef": reg.Id,
		"vpc_OrderInfo":   "REG " + reg.FirstName + " " + reg.MiddleName + " " + reg.LastName, // display info
		"vpc_Amount":      amount,
		"vpc_TicketNo":    clientIP,
		"vpc_CallbackURL": r.config.Server.Host + ":" + r.config.Server.Port + "/onepay/ipn",
	}

	queryParamSorted := sortParams(merchantQueryMap)
	stringTohash := generateStringToHash(queryParamSorted)
	merchantGenSecureHash := generateSecureHash(stringTohash, op.HashCode)
	merchantQueryMap["vpc_SecureHash"] = merchantGenSecureHash

	params := url.Values{}
	for key, value := range merchantQueryMap {
		params.Add(key, value)
	}
	requestUrl := op.Endpoint + "?" + params.Encode()
	return requestUrl, nil
}

var registrationServiceInstance RegistrationService
var registrationServiceOnce sync.Once

func GetRegistrationServiceInstance(
	registrationRepo repository.RegistrationRepository,
	registrationOptionsRepo repository.RegistrationOptionRepository,
	config *config.Config,
) RegistrationService {
	registrationServiceOnce.Do(func() {
		registrationServiceInstance = NewRegistrationService(
			registrationRepo, registrationOptionsRepo, config,
		)
	})
	return registrationServiceInstance
}

func NewRegistrationService(
	registrationRepo repository.RegistrationRepository,
	registrationOptionsRepo repository.RegistrationOptionRepository,
	config *config.Config,
) RegistrationService {
	return &registrationService{
		registrationRepo:        registrationRepo,
		registrationOptionsRepo: registrationOptionsRepo,
		config:                  config,
	}
}
