package service

import (
	"ashno-onepay/internal/config"
	errs "ashno-onepay/internal/errors"
	"ashno-onepay/internal/model"
	"ashno-onepay/internal/repository"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

type RegistrationService interface {
	Register(registration model.Registration, clientIP string) (string, string, error)
	GetRegistration(ID string) (*model.Registration, error)
	ValidateHMAC(params map[string]string, receivedHash string) bool
	UpdatePaymentStatus(ID, status string) error
}

type registrationService struct {
	registrationRepo        repository.RegistrationRepository
	registrationOptionsRepo repository.RegistrationOptionRepository
	config                  *config.Config
}

func (r registrationService) UpdatePaymentStatus(ID, status string) error {
	//reg, err := r.GetRegistration(ID)
	//if err != nil {
	//	return err
	//}
	//if status == string(model.PaymentStatusDone) {
	//	go SendPaymentSuccessEmailWithQR(reg.Email, reg.FirstName, reg.Id, )
	//}
	return r.registrationRepo.UpdatePaymentStatus(ID, status)
}

func (r registrationService) ValidateHMAC(params map[string]string, receivedHash string) bool {
	// Sort keys
	keys := make([]string, 0, len(params))
	for k := range params {
		if strings.HasPrefix(k, "vpc_") || strings.HasPrefix(k, "user_") {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var rawData []string
	for _, k := range keys {
		v := params[k]
		if v != "" {
			rawData = append(rawData, k+"="+v)
		}
	}
	dataToHash := strings.Join(rawData, "&")

	hmacHash := hmac.New(sha256.New, []byte(r.config.OnePay.HashCode))
	hmacHash.Write([]byte(dataToHash))
	calculatedHash := strings.ToUpper(hex.EncodeToString(hmacHash.Sum(nil)))

	return calculatedHash == receivedHash
}

func (r registrationService) GetRegistration(ID string) (*model.Registration, error) {
	return r.registrationRepo.GetRegistration(ID)
}

func (r registrationService) Register(registration model.Registration, clientIP string) (string, string, error) {
	// check email registered
	oldReg, err := r.registrationRepo.GetByEmail(registration.Email)
	if err != nil {
		return "", "", err
	}
	if oldReg != nil && oldReg.PaymentStatus == string(model.PaymentStatusDone) {
		return "", "", errs.ErrInternal.Reform("email registered")
	}
	// setup registration
	err = r.setupRegistration(&registration)
	if err != nil {
		return "", "", err
	}
	// generate paymentURL
	paymentURL, err := r.generatePaymentURL(&registration, clientIP)
	if err != nil {
		return "", "", err
	}
	// insert registration
	_, err = r.registrationRepo.Create(registration)
	if err != nil {
		return "", "", err
	}
	return paymentURL, registration.Id, nil
}

func (r registrationService) setupRegistration(reg *model.Registration) error {
	reg.Id = uuid.New().String()
	OptionFilter := model.RegistrationOptionFilter{}
	switch reg.RegistrationCategory {
	case string(model.DoctorCategory):
		OptionFilter.Category = string(model.DoctorCategory)
		OptionFilter.Subtype = DetermineRegistrationPeriod(time.Now())
	case string(model.StudentCategory):
		OptionFilter.Category = string(model.StudentCategory)
	case string(model.DinnerCategory):
		OptionFilter.Category = string(model.DinnerCategory)
	default:
		return errs.ErrNotFound.Reform("option not found")
	}
	fmt.Println(OptionFilter)
	option, err := r.registrationOptionsRepo.Find(OptionFilter)
	if err != nil {
		return errs.ErrNotFound.Reform("option not found")
	}
	reg.RegistrationOptionID = option.Id
	reg.RegistrationOption = *option

	return nil
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
	op := r.config.OnePay
	locale := "en"
	currency := "USD"
	amount := reg.RegistrationOption.FeeUSD
	if reg.Nationality == model.NationalityVietNam {
		locale = "vn"
		currency = "VND"
		amount = reg.RegistrationOption.FeeVND
	}
	merchantQueryMap := map[string]string{
		"vpc_Version":     "2",
		"vpc_Currency":    currency, // "USD" or "VND"
		"vpc_Command":     "pay",
		"vpc_AccessCode":  op.AccessCode,
		"vpc_Merchant":    op.MerchantID,
		"vpc_Locale":      locale,                                                             // e.g., "en" or "vn"
		"vpc_ReturnURL":   op.ReturnURL,                                                       //TODO: add reg.Id
		"vpc_MerchTxnRef": reg.Id,                                                             // unique transaction ID
		"vpc_OrderInfo":   "REG " + reg.FirstName + " " + reg.MiddleName + " " + reg.LastName, // display info
		"vpc_Amount":      fmt.Sprintf("%d", amount*100),                                      // in smallest unit
		"vpc_TicketNo":    clientIP,                                                           // client IP
		//"vpc_CallbackURL": "TODO"
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
	//sendHttpGetRequest(requestUrl)
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
