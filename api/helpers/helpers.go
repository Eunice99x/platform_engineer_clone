package helpers

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
)

func WrapStrInErrMap(str string) map[string][]string {
	w := make(map[string][]string)
	w["errors"] = []string{str}
	return w
}

func WrapErrInErrMap(err error) map[string][]string {
	w := make(map[string][]string)
	w["errors"] = []string{err.Error()}
	return w
}

func ResponseBodyToString(io io.ReadCloser) (string, error) {
	var str string
	readBody, err := ioutil.ReadAll(io)
	if err != nil {
		return str, err
	}
	return string(readBody), nil
}

func AuthorizationHeaderBasicAuth(user, pass string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v:%v", user, pass)))
}
