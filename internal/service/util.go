package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type MapSort struct {
	Key   string
	Value string
}

func sortParams(paramMap map[string]string) []MapSort {
	keys := make([]string, 0, len(paramMap))
	for k := range paramMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	mapSorted := []MapSort{}

	//var paramMapSorted map[string]string
	for _, k := range keys {
		value := paramMap[k]
		for n := range paramMap {
			if k == n {
				mapSorted = append(mapSorted, MapSort{Key: k, Value: value})
			}
		}
	}
	return mapSorted
}

func sendHttpGetRequest(requestUrl string) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest(http.MethodGet, requestUrl, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	link := resp.Header.Get("Location")
	fmt.Println("Link redirect to invoice page: ", link)
}

func sendHttpGetRequestWithHeader(requestUrl string, headerRequest map[string]string) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest(http.MethodGet, requestUrl, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	for key, value := range headerRequest {
		req.Header.Add(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Response status:", resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}
	fmt.Println("Response body:", string(body))
}

func generateSecureHash(stringToHash string, merchantHashCode string) string {
	keyHashHex, err := hex.DecodeString(merchantHashCode)
	if err != nil {
		panic(err)
	}
	secureHash := hmac.New(sha256.New, keyHashHex)
	secureHash.Write([]byte(stringToHash))
	secureHashToString := hex.EncodeToString(secureHash.Sum(nil))
	signUpper := strings.ToUpper(secureHashToString)
	return signUpper
}

func generateStringToHash(paramMapSorted []MapSort) string {
	stringToHash := ""
	for _, items := range paramMapSorted {
		key := items.Key
		value := items.Value
		pref4 := key[0:4]
		pref5 := key[0:5]
		if pref4 == "vpc_" || pref5 == "user_" {
			if key != "vpc_SecureHashType" && key != "vpc_SecureHash" {
				if len(value) > 0 {
					if len(stringToHash) > 0 {
						stringToHash += "&"
					}
					stringToHash += key + "=" + value
				}
			}
		}
	}
	return stringToHash
}

func createRequestSignatureITA(method string, uri string, headerSign map[string]string, signedHeaderNames []string, merchantId string, merchantHashCode string) string {
	created := time.Now().Unix()

	lowercaseHeaders := make(map[string]string)
	for key, value := range headerSign {
		lowercaseHeaders[strings.ToLower(key)] = value
	}
	lowercaseHeaders["(request-target)"] = strings.ToLower(method) + " " + uri
	lowercaseHeaders["(created)"] = fmt.Sprintf("%d", created)

	signingString := ""

	var headerNames string
	for _, element := range signedHeaderNames {
		headerName := element
		if _, ok := lowercaseHeaders[headerName]; !ok {
			panic("MissingRequiredHeaderException: " + headerName)
		}
		if signingString != "" {
			signingString += "\n"
		}
		signingString += headerName + ": " + lowercaseHeaders[headerName]

		if headerNames != "" {
			headerNames += " "
		}
		headerNames += headerName
	}

	fmt.Println("signingString=", signingString)

	hmacKey, _ := hexDecode(merchantHashCode)
	h := hmac.New(sha512.New, hmacKey)
	h.Write([]byte(signingString))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	signingAlgorithm := "hs2019"
	return fmt.Sprintf(`algorithm="%s", keyId="%s", headers="%s", created=%d, signature="%s"`,
		signingAlgorithm, merchantId, headerNames, created, signature)
}

func hexDecode(s string) ([]byte, error) {
	n := len(s)
	if n%2 != 0 {
		return nil, fmt.Errorf("input length must be even")
	}
	b := make([]byte, n/2)
	for i := 0; i < n; i += 2 {
		c, err := parseHexByte(s[i : i+2])
		if err != nil {
			return nil, err
		}
		b[i/2] = c
	}
	return b, nil
}

func parseHexByte(s string) (byte, error) {
	var c byte
	for _, r := range s {
		c <<= 4
		switch {
		case '0' <= r && r <= '9':
			c |= byte(r - '0')
		case 'a' <= r && r <= 'f':
			c |= byte(r-'a') + 10
		case 'A' <= r && r <= 'F':
			c |= byte(r-'A') + 10
		default:
			return 0, fmt.Errorf("invalid character '%c' in hex string", r)
		}
	}
	return c, nil
}

func sendHttpPostRequest(urlRequest string, merchantParam map[string]string) {
	// Tạo dữ liệu để gửi
	data := url.Values{}

	for key, value := range merchantParam {
		data.Set(key, value)
	}

	// Tạo HTTP client
	client := &http.Client{}

	// Tạo HTTP request
	req, err := http.NewRequest("POST", urlRequest, bytes.NewBufferString(data.Encode()))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Set header cho request
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Gửi request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Đọc response
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	// Hiển thị response
	fmt.Println("Response:", string(respBody))
}
