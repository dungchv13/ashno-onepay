package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
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

func generateSecureHash(stringToHash string, merchantHashCode string) string {
	keyHashHex, err := hex.DecodeString(merchantHashCode)
	if err != nil {
		fmt.Println(err)
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
	fmt.Println(paramMapSorted)
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
