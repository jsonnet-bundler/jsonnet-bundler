package jsonnetfile

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/jsonnet-bundler/jsonnet-bundler/spec"
	"github.com/pkg/errors"
)

const File = "jsonnetfile.json"
const LockFile = "jsonnetfile.lock.json"

var ErrNoFile = errors.New("no jsonnetfile")

func Choose(dir string) (string, bool, error) {
	jsonnetfileLock := path.Join(dir, LockFile)
	jsonnetfile := path.Join(dir, File)

	lockExists, err := fileExists(jsonnetfileLock)
	if err != nil {
		return "", false, err
	}
	if lockExists {
		return jsonnetfileLock, true, nil
	}

	fileExists, err := fileExists(jsonnetfile)
	if err != nil {
		return "", false, err
	}
	if fileExists {
		return jsonnetfile, false, nil
	}

	return "", false, ErrNoFile
}

func Load(filepath string) (spec.JsonnetFile, error) {
	m := spec.JsonnetFile{}

	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return m, errors.Wrap(err, "failed to read file")
	}

	if err := json.Unmarshal(bytes, &m); err != nil {
		return m, errors.Wrap(err, "failed to unmarshal file")
	}

	return m, nil
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}
