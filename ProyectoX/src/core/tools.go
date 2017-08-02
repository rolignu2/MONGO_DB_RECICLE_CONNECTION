package core

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"os"
	"time"
)

func StrConcat(str ...string) string {

	var buffer bytes.Buffer

	for i := 0; i < len(str); i++ {
		buffer.WriteString(str[i])
	}

	return buffer.String()
}

func CreateFileLog(path string, ltype string) (string, bool) {

	Cdate := time.Now().Local().Format("2006-01-02")
	Cat := StrConcat(path, ltype, Cdate, ".log")

	f, err := ioutil.ReadFile(Cat)

	if err != nil {
		k, ierr := os.Create(Cat)

		defer k.Close()

		if ierr != nil {
			return string(f), false
		}
	}

	return Cat, true
}

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}
