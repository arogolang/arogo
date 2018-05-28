package util

import (
	"bytes"
	"crypto/md5"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/arogolang/arogo/errlog"
	"github.com/ender-wan/ewlog"
	"github.com/golang/crypto/argon2"
)

type File interface {
	io.Reader
	io.ReaderAt
	io.Seeker
	io.Closer
}

func NoDirListing(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func ReadFileToString(tplPath string) (data string, err error) {
	file, err := os.Open(tplPath)
	if err != nil {
		errlog.Errorf("Open error %s", err)
		return
	}

	dataBytes, err := ioutil.ReadAll(file)
	if err != nil {
		errlog.Errorf("read error %s", err)
		file.Close()
		return
	}

	data = string(dataBytes)

	file.Close()
	return
}

func MD5FileAndSize(file File) (string, int64, error) {
	md5hash := md5.New()
	fileSize, err := io.Copy(md5hash, file)
	if err != nil {
		ewlog.Error(err)
		return "", -1, err
	}
	if _, err := file.Seek(0, 0); err != nil {
		ewlog.Error(err)
		return "", -1, err
	}
	return hex.EncodeToString(md5hash.Sum(nil)), fileSize, nil
}

func MD5String(data string) string {
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

func GetAroHash(data string) (r []byte) {
	h := sha512.New()
	h.Write([]byte(data))
	r = h.Sum(nil)

	for i := 0; i < 5; i++ {
		h.Reset()
		h.Write(r)
		r = h.Sum(nil)
	}

	return
	//return hex.EncodeToString(r)
}

func FileExists(f string) bool {
	_, err := os.Stat(f)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func PostDataToNode(nodeurl, data string) (ok bool, retData string, err error) {
	hc := http.Client{}
	req, err := http.NewRequest("POST", nodeurl+"/mine.php?q=submitNonce", strings.NewReader(data))
	if err != nil {
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	resp, err := hc.Do(req)
	if err != nil {
		return
	}

	if resp.Body == nil {
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var o = make(map[string]*json.RawMessage)
	if err = json.Unmarshal(body, &o); err != nil {
		return
	}

	retDataB, okData := o["data"]
	if okData && retDataB != nil {
		retData = string(*retDataB)
	}

	retStatus, okData := o["data"]
	if okData && retStatus != nil {
		if string(*retStatus) == "ok" {
			ok = true
			err = nil
		} else {
			err = fmt.Errorf("post %v data:%v ret:%v", nodeurl, data, retData)
		}
	}

	return
}

func SubmitNonceToNode(nodeurl string, argon string, nonce string, pubkey string, privkey string) (ok bool, err error) {
	form := url.Values{}
	form.Add("argon", argon)
	form.Add("nonce", nonce)
	form.Add("private_key", privkey)
	form.Add("public_key", pubkey)

	hc := http.Client{}
	req, err := http.NewRequest("POST", nodeurl+"/mine.php?q=submitNonce", strings.NewReader(form.Encode()))
	if err != nil {
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	resp, err := hc.Do(req)
	if err != nil {
		return
	}

	if resp.Body == nil {
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var o = make(map[string]*json.RawMessage)
	if err = json.Unmarshal(body, &o); err != nil {
		return
	}

	data, okData := o["data"]
	if okData && data != nil {
		if string(*data) == "ok" {
			ok = true
			err = nil
		}
	}

	return
}

func Argon2Verify(base string, hash string) (ok bool) {

	hashArr := strings.Split(hash, "$")

	padding := ""
	for left := len(hashArr[1]) % 4; left > 0; left-- {
		padding = padding + "="
	}
	hashArr[1] = hashArr[1] + padding

	padding = ""
	for left := len(hashArr[2]) % 4; left > 0; left-- {
		padding = padding + "="
	}
	hashArr[2] = hashArr[2] + padding

	salt, _ := base64.StdEncoding.DecodeString(hashArr[1])
	if len(salt) != 16 {
		return
	}

	hashDst, _ := base64.StdEncoding.DecodeString(hashArr[2])
	if len(hashDst) != 32 {
		return
	}

	hashCalc := argon2.Key([]byte(base), salt, 1, 524288, 1, 32)

	if bytes.Equal(hashCalc, hashDst) {
		ok = true
	}

	return
}
