package service

import (
	"ashno-onepay/internal/config"
	"ashno-onepay/internal/controller/dto"
	errs "ashno-onepay/internal/errors"
	"ashno-onepay/internal/model"
	"ashno-onepay/internal/repository"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const RateUSDVND = 26000

type RegistrationService interface {
	Register(registration dto.RegistrationRequest, clientIP string) (string, string, error)
	GetRegistration(ID string) (*model.Registration, error)
	OnePayVerifySecureHash(u *url.URL) error
	GetRegistrationOption(filter model.RegistrationOptionFilter) (*model.RegistrationOption, error)
	RegisterForAccompanyPersons(email string, accompanyPersons model.AccompanyPersonList, clientIP string) (string, error)
}

type registrationService struct {
	registrationRepo        repository.RegistrationRepository
	registrationOptionsRepo repository.RegistrationOptionRepository
	config                  *config.Config
}

func (r registrationService) GetRegistrationOption(filter model.RegistrationOptionFilter) (*model.RegistrationOption, error) {
	var registrationOption *model.RegistrationOption
	var err error
	if filter.Category != "" {
		switch filter.Category {
		case string(model.DoctorCategory):
			filter.Category = string(model.DoctorCategory)
			if filter.AttendGalaDinner {
				filter.Category = string(model.DoctorAndDinnerCategory)
			}
			filter.Subtype = DetermineRegistrationPeriod(time.Now())
		case string(model.StudentCategory):
			filter.Category = string(model.StudentCategory)
			if filter.AttendGalaDinner {
				filter.Category = string(model.StudentAndDinnerCategory)
			}
		default:
			return nil, errs.ErrNotFound.Reform("option not found")
		}
		registrationOption, err = r.registrationOptionsRepo.Find(filter)
		if err != nil {
			return nil, err
		}
		if filter.NumberAccompanyPersons > 0 {
			registrationOption.FeeUSD = registrationOption.FeeUSD + float64(filter.NumberAccompanyPersons)*model.GalaDinnerOnlyOption.FeeUSD
			registrationOption.FeeVND = registrationOption.FeeVND + int64(filter.NumberAccompanyPersons)*model.GalaDinnerOnlyOption.FeeVND
		}
		return registrationOption, nil
	}
	return &model.RegistrationOption{
		FeeUSD:   float64(filter.NumberAccompanyPersons) * model.GalaDinnerOnlyOption.FeeUSD,
		FeeVND:   int64(filter.NumberAccompanyPersons) * model.GalaDinnerOnlyOption.FeeVND,
		Category: string(model.GalaDinnerOnlyOption.Category),
	}, nil
}

func (r registrationService) OnePayVerifySecureHash(u *url.URL) error {
	queryParams := u.Query()
	queryParamsMap := make(map[string]string)
	for k, v := range queryParams {
		queryParamsMap[k] = strings.Join(v, "")
	}

	// Extract required fields
	txnRef := queryParamsMap["vpc_MerchTxnRef"]
	orderInfo := queryParamsMap["vpc_OrderInfo"]
	txnCode := queryParamsMap["vpc_TxnResponseCode"]
	merchantSecureHash := queryParamsMap["vpc_SecureHash"]
	message := queryParamsMap["vpc_Message"]

	if txnRef == "" {
		return errs.ErrBadRequest
	}

	// Verify secure hash
	op := r.config.OnePay
	queryMapSorted := sortParams(queryParamsMap)
	stringToHash := generateStringToHash(queryMapSorted)
	onePaySecureHash := generateSecureHash(stringToHash, op.HashCode)
	log.Println("OnePay's Hash: ", onePaySecureHash)
	log.Println("Merchant's Hash: ", merchantSecureHash)
	if onePaySecureHash != merchantSecureHash {
		return errs.ErrForbidden.Reform("Invalid signature")
	}

	// Handle payment result by order type
	switch {
	case strings.HasPrefix(orderInfo, "ORDER"):
		// Main registration payment
		regID := txnRef
		reg, err := r.registrationRepo.GetRegistration(regID)
		if err != nil {
			return err
		}
		if reg == nil {
			return errs.ErrNotFound.Reform("registration not found")
		}
		var status string
		if txnCode == "0" {
			log.Println("Payment Success for ", regID)
			status = string(model.PaymentStatusDone)
			// Mark all accompany persons as paid if they were pending
			for i := range reg.AccompanyPersons {
				if reg.AccompanyPersons[i].PaymentStatus == model.AccompanyPersonsPaymentStatusPending {
					reg.AccompanyPersons[i].PaymentStatus = model.AccompanyPersonsPaymentStatusDone
				}
			}
			err = r.registrationRepo.UpdateAccompanyPersonsByID(reg.Id, reg.AccompanyPersons)
			if err != nil {
				return err
			}
			// Send registration email in background
			go func() {
				var registrationFee, locale string
				if reg.Nationality == model.NationalityVietNam {
					registrationFee = strconv.FormatInt(reg.RegistrationOption.FeeVND, 10) + " VND"
					locale = "vi"
				} else {
					registrationFee = strconv.FormatFloat(float64(reg.RegistrationOption.FeeUSD), 'f', -1, 64) + " USD"
					locale = "en"
				}
				fullName := fmt.Sprintf("%s %s %s", reg.FirstName, reg.MiddleName, reg.LastName)
				err := SendRegistrationEmailWithTemplate(
					reg.Email, reg.FirstName, reg.Id, locale, fullName, reg.PhoneNumber, registrationFee, r.config,
				)
				if err != nil {
					log.Printf("Send QR Failed for %s", err.Error())
					log.Printf("Send QR Failed for %s", reg.Id)
				}
			}()
		} else {
			log.Printf("Payment Failed for %s: %s", regID, message)
			status = string(model.PaymentStatusFail)
		}
		return r.registrationRepo.UpdatePaymentStatus(regID, status)

	case strings.HasPrefix(orderInfo, "ACCOM"):
		// Accompany person payment
		transactionID := strings.TrimPrefix(orderInfo, "ACCOM")
		accompanyPersons, err := r.registrationRepo.GetAccompanyPersonsByTransactionAndRegistration(transactionID)
		if err != nil {
			return err
		}
		if len(accompanyPersons) == 0 {
			return err
		}
		regID := accompanyPersons[0].RegistrationID
		reg, err := r.registrationRepo.GetRegistration(regID)
		if err != nil {
			return err
		}
		if reg == nil {
			return errs.ErrNotFound.Reform("registration not found")
		}
		for _, person := range accompanyPersons {
			reg.AccompanyPersons = append(reg.AccompanyPersons, model.AccompanyPerson{
				FirstName:     person.FirstName,
				MiddleName:    person.MiddleName,
				LastName:      person.LastName,
				DateOfBirth:   person.DateOfBirth,
				PaymentStatus: model.AccompanyPersonsPaymentStatusDone,
			})
		}
		if txnCode == "0" {
			log.Println("Payment Success for accompany person ", regID)
			err = r.registrationRepo.UpdateAccompanyPersonsByID(reg.Id, reg.AccompanyPersons)
			if err != nil {
				return err
			}
		} else {
			log.Printf("Accompany person payment failed for %s: %s", regID, message)
		}
	}

	return nil
}

func (r registrationService) GetRegistration(ID string) (*model.Registration, error) {
	reg, err := r.registrationRepo.GetRegistration(ID)
	if err != nil {
		return nil, err
	}
	if reg == nil {
		return nil, errs.ErrNotFound.Reform("registration not found")
	}
	if len(reg.AccompanyPersons) > 0 {
		reg.RegistrationOption.FeeUSD = reg.RegistrationOption.FeeUSD + float64(len(reg.AccompanyPersons))*model.GalaDinnerOnlyOption.FeeUSD
		reg.RegistrationOption.FeeVND = reg.RegistrationOption.FeeVND + int64(len(reg.AccompanyPersons))*model.GalaDinnerOnlyOption.FeeVND
	}
	return reg, nil
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
		if err != nil {
			return "", "", err
		}
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
		AccompanyPersons:     []model.AccompanyPerson{},
	}
	for _, p := range request.AccompanyPersons {
		p.PaymentStatus = model.AccompanyPersonsPaymentStatusPending
		reg.AccompanyPersons = append(reg.AccompanyPersons, p)
	}
	reg.Id = uuid.New().String()
	OptionFilter := model.RegistrationOptionFilter{}
	switch request.RegistrationOption {
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

func (r registrationService) RegisterForAccompanyPersons(email string, accompanyPersons model.AccompanyPersonList, clientIP string) (string, error) {
	reg, err := r.registrationRepo.GetByEmail(email)
	if err != nil {
		return "", err
	}
	if reg == nil || reg.PaymentStatus != string(model.PaymentStatusDone) {
		return "", errs.ErrNotFound.Reform("registration with email %s not found", email)
	}
	transactionID := RandomString(16)
	var accompanyPersonsDB []model.AccompanyPersonDB
	for i := range accompanyPersons {
		accompanyPersonsDB = append(accompanyPersonsDB, model.AccompanyPersonDB{
			TransactionID:  transactionID,
			RegistrationID: reg.Id,
			FirstName:      accompanyPersons[i].FirstName,
			MiddleName:     accompanyPersons[i].MiddleName,
			LastName:       accompanyPersons[i].LastName,
			DateOfBirth:    accompanyPersons[i].DateOfBirth,
		})
	}
	err = r.registrationRepo.SaveAccompanyPersons(accompanyPersonsDB)
	if err != nil {
		return "", err
	}
	// Update the in-memory reg object for payment calculation
	reg.AccompanyPersons = accompanyPersons
	paymentURL, err := r.generatePaymentURLForAccompanyPersons(reg, accompanyPersons, clientIP, transactionID)
	if err != nil {
		return "", err
	}
	return paymentURL, nil
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

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandomString(length int) string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func (r registrationService) generatePaymentURL(reg *model.Registration, clientIP string) (string, error) {
	op := r.config.OnePay
	locale := "en"
	currency := "VND"
	// adding AccompanyPersons fee
	optionUSDFee := reg.RegistrationOption.FeeUSD + float64(len(reg.AccompanyPersons))*model.GalaDinnerOnlyOption.FeeUSD
	optionVNDFee := reg.RegistrationOption.FeeVND + int64(len(reg.AccompanyPersons))*model.GalaDinnerOnlyOption.FeeVND

	usd := int64(optionUSDFee)
	amount := strconv.FormatInt(usd*RateUSDVND*100, 10)
	if reg.Nationality == model.NationalityVietNam {
		locale = "vn"
		amount = strconv.FormatInt(optionVNDFee*100, 10)
	}

	merchantQueryMap := map[string]string{
		"vpc_Version":     "2",
		"vpc_Currency":    currency,
		"vpc_Command":     "pay",
		"vpc_AccessCode":  op.AccessCode,
		"vpc_Merchant":    op.MerchantID,
		"vpc_Locale":      locale,
		"vpc_ReturnURL":   op.ReturnURL + "/" + reg.Id,
		"vpc_MerchTxnRef": reg.Id,
		"vpc_OrderInfo":   fmt.Sprintf("ORDER%s", RandomString(16)),
		"vpc_Amount":      amount,
		"vpc_TicketNo":    clientIP,
		"vpc_CallbackURL": r.config.Server.Host + "/onepay/ipn",
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

func (r registrationService) generatePaymentURLForAccompanyPersons(reg *model.Registration, accompanyPersons model.AccompanyPersonList, clientIP, transactionID string) (string, error) {
	op := r.config.OnePay
	locale := "en"
	currency := "VND"

	numAccompany := len(accompanyPersons)
	optionUSDFee := float64(numAccompany) * model.GalaDinnerOnlyOption.FeeUSD
	optionVNDFee := int64(numAccompany) * model.GalaDinnerOnlyOption.FeeVND

	usd := int64(optionUSDFee)
	amount := strconv.FormatInt(usd*RateUSDVND*100, 10)
	if reg.Nationality == model.NationalityVietNam {
		locale = "vn"
		amount = strconv.FormatInt(optionVNDFee*100, 10)
	}

	merchantQueryMap := map[string]string{
		"vpc_Version":     "2",
		"vpc_Currency":    currency,
		"vpc_Command":     "pay",
		"vpc_AccessCode":  op.AccessCode,
		"vpc_Merchant":    op.MerchantID,
		"vpc_Locale":      locale,
		"vpc_ReturnURL":   op.ReturnURL + "/" + reg.Id,
		"vpc_MerchTxnRef": transactionID,
		"vpc_OrderInfo":   fmt.Sprintf("ACCOM%s", transactionID),
		"vpc_Amount":      amount,
		"vpc_TicketNo":    clientIP,
		"vpc_CallbackURL": r.config.Server.Host + "/onepay/ipn",
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
