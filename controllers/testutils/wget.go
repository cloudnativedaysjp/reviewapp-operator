package testutils

import (
	"io/ioutil"
	"net/http"
)

func Wget(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(byteArray), nil
}
