package main

import (
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/goccy/go-json"
)

type Number interface {
	~float32 | ~float64 |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// SumBy summarizes the values in a collection using the given return value from the iteration function. If collection is empty 0 is returned.
func SumBy[T any, R Number](s []T, valueFn func(item T) R) R {
	var sum R = 0
	for i := range s {
		sum += valueFn(s[i])
	}
	return sum
}

const ErrHash = "ERR"

type HashMethod func() hash.Hash

// 计算文件的hash值, 如果失败返回"ERR", 否则返回hash值, 如果fullRead为true, 则读取全部内容, 否则读取5个32k
func FileHash(path string, hashMethod HashMethod, fullRead bool) string {
	f, err := os.Open(path)
	if err != nil {
		return ErrHash
	}
	defer f.Close()

	const base int64 = 32 << 10 //32k

	hash := hashMethod()
	defer hash.Reset()

	stat, _ := f.Stat()

	if fullRead || stat.Size() <= 5*base {
		_, err = io.Copy(hash, f) //160k以下的文件，取全部
	} else {
		size := stat.Size()
		for i := int64(0); i < 10; i++ { //320k以上的文件，取10个32k
			if _, err = f.Seek(size*i/10, io.SeekStart); err == nil {
				_, err = io.CopyN(hash, f, base)
			}
		}
	}

	if err != nil {
		return ErrHash
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func May[T any](v T, err error) T {
	if err != nil {
		return v
	}
	return v
}

// HumanSize 将大小转换为人类可读的格式
func HumanSize[T Number](size T) string {
	const humanUnits = "EPTGMK"
	for i, c := range humanUnits {
		shift := len(humanUnits) - i
		base := 1 << (shift * 10)
		if float64(size) >= float64(base) {
			return fmt.Sprintf("%.2f %cB", float64(size)/float64(base), c)
		}
	}
	return fmt.Sprintf("%d B", int(size))
}

func ToBytes(value any) (b []byte) {
	if value == nil {
		return
	}

	var s string

	switch v := value.(type) {
	case time.Duration:
		s = v.String()
	case time.Time:
		s = v.Format(time.RFC3339)
	case string:
		s = v
	case []byte:
		b = v
	case bool:
		s = strconv.FormatBool(v)
	case int:
		s = strconv.Itoa(v)
	case int8:
		s = strconv.Itoa(int(v))
	case int16:
		s = strconv.Itoa(int(v))
	case int32:
		s = strconv.Itoa(int(v))
	case int64:
		s = strconv.FormatInt(v, 10)
	case uint:
		s = strconv.FormatUint(uint64(v), 10)
	case uint8:
		s = strconv.FormatUint(uint64(v), 10)
	case uint16:
		s = strconv.FormatUint(uint64(v), 10)
	case uint32:
		s = strconv.FormatUint(uint64(v), 10)
	case uint64:
		s = strconv.FormatUint(v, 10)
	case float32:
		s = strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		s = strconv.FormatFloat(v, 'f', -1, 64)
	default:
		b = May(json.Marshal(v))
	}

	if s != "" {
		b = []byte(s)
	}
	return b
}

func FromBytes(b []byte, value any) {
	if len(b) == 0 {
		return
	}

	switch v := value.(type) {
	case *time.Duration:
		*v = May(time.ParseDuration(string(b)))
	case *time.Time:
		*v = May(time.Parse(time.RFC3339, string(b)))
	case *string:
		*v = string(b)
	case *[]byte:
		*v = b
	case *bool:
		*v = May(strconv.ParseBool(string(b)))
	case *int:
		*v = May(strconv.Atoi(string(b)))
	case *int8:
		*v = int8(May(strconv.ParseInt(string(b), 10, 8)))
	case *int16:
		*v = int16(May(strconv.ParseInt(string(b), 10, 16)))
	case *int32:
		*v = int32(May(strconv.ParseInt(string(b), 10, 32)))
	case *int64:
		*v = int64(May(strconv.ParseInt(string(b), 10, 64)))
	case *uint:
		*v = uint(May(strconv.ParseUint(string(b), 10, 0)))
	case *uint8:
		*v = uint8(May(strconv.ParseUint(string(b), 10, 8)))
	case *uint16:
		*v = uint16(May(strconv.ParseUint(string(b), 10, 16)))
	case *uint32:
		*v = uint32(May(strconv.ParseUint(string(b), 10, 32)))
	case *uint64:
		*v = uint64(May(strconv.ParseUint(string(b), 10, 64)))
	case *float32:
		*v = float32(May(strconv.ParseFloat(string(b), 32)))
	case *float64:
		*v = float64(May(strconv.ParseFloat(string(b), 64)))
	default:
		json.Unmarshal(b, value) //nolint: errcheck
	}
}
