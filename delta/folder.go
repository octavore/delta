package delta

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
)

func DiffFolder(pathA, pathB string) error {
	f, err := os.Open(pathA)
	if err != nil {
		return err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return err
	}

	g, err := os.Open(pathB)
	if err != nil {
		return err
	}
	defer g.Close()
	gi, err := g.Stat()
	if err != nil {
		return err
	}
	if !gi.IsDir() {
		return err
	}

	a, err := GetFileList(pathA)
	if err != nil {
		return err
	}

	b, err := GetFileList(pathB)
	if err != nil {
		return err
	}

	fmt.Println(a)
	fmt.Println(b)
	return nil
}

type FileHeader struct {
	Name     string // e.g. foo.txt
	Path     string // e.g. /example
	FullPath string // e.g. /usr/local/repos/example/foo.txt
	Hash     string
	Size     int64 // bytes
}

var readDir = ioutil.ReadDir

func GetFileList(dir string) ([]*FileHeader, error) {
	return getFileListWithRoot(dir, "/")
}

// Read a file list from dir, with output path (in cloud) set to:
//   <baseOutPath>/<filename>
// Recurses into sub dirs
func getFileListWithRoot(dir, baseOutPath string) ([]*FileHeader, error) {
	lst, err := readDir(dir)
	if err != nil {
		return []*FileHeader{}, err
	}

	out := make([]*FileHeader, 0)
	for _, f := range lst {
		filePath := path.Join(dir, f.Name())
		if !f.IsDir() {
			// hashs of f.ModTime(), f.Name(), f.Size()
			hash, err := checksum(filePath)
			if err != nil {
				return nil, err
			}
			out = append(out, &FileHeader{
				Name:     f.Name(),
				Path:     baseOutPath,
				FullPath: filePath,
				Hash:     hash,
				Size:     f.Size(),
			})

		} else {
			b := path.Join(baseOutPath, f.Name())
			sublist, err := getFileListWithRoot(filePath, b)
			if err != nil {
				return nil, err
			}
			out = append(out, sublist...)
		}
	}

	return out, nil
}

func (f *FileHeader) PathName() string {
	return path.Join(f.Path, f.Name)
}

func checksum(fp string) (string, error) {
	f, err := os.Open(fp)
	if err != nil {
		return "", err
	}
	defer f.Close()
	hash := md5.New()
	for err == nil {
		_, err = io.CopyN(hash, f, 8192)
	}
	if err != nil && err != io.EOF {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (f *FileHeader) String() string {
	return fmt.Sprintf("%+v", *f)
}
