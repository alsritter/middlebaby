package grpcurl

import (
	"errors"
	"io"
	"os"
	"strings"
)

// TODO: support custom
func Accessor(filename string) (io.ReadCloser, error) {
	// 解析包内部直接引用仓库公共文件路径
	if strings.HasPrefix(filename, "alsritter.icu") {
		f, err := os.Open(filename)
		// 文件不存在尝试一级一级查找
		if errors.Is(err, os.ErrNotExist) {
			ss := strings.Split(filename, "/")
			// 最大尝试4级目录
			for i := 0; i < 4; i++ {
				b := i + 1
				if b < len(ss) {
					jf := strings.Join(ss[b:], "/")
					ff, fErr := os.Open(jf)
					if fErr == nil {
						return ff, fErr
					}
				}
			}
			return nil, os.ErrNotExist
		}
		return f, err
	}
	return os.Open(filename)
}
