package utils

import (
	"crypto/md5"
	"fmt"
	"os"
)

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	} else {
		return true, nil
	}
}

func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	} else {
		return s.IsDir()
	}
}

func MD5(raw string) string {
	bs := md5.Sum([]byte(raw))
	return fmt.Sprintf("%x", bs)
}

// ESIdsLenLimit: VM label length limit is 16384 bytes.
// So, we limit 30 * 512 bytes(elasticsearch _doc id lenth limit) = 15360.
func ESIdsLenLimit(ids []string) []string {
	return ids[:30]
}
