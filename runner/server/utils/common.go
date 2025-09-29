package utils

import (
	"encoding/base64"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func SaveURIToTempFile(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	var data []byte
	// windows drive letter
	if len(u.Scheme) == 1 && (u.Scheme[0] >= 'a' && u.Scheme[0] <= 'z' || u.Scheme[0] >= 'A' && u.Scheme[0] <= 'Z') {
		u.Scheme = "file"
		u.Path = uri
	}
	switch u.Scheme {
	case "http", "https":
		resp, err := http.Get(uri)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", errors.New("http download failed: " + resp.Status)
		}
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
	case "data":
		parts := strings.SplitN(u.Opaque, ",", 2)
		if len(parts) != 2 {
			return "", errors.New("invalid data URI format")
		}
		// format: data:[<mediatype>][;base64],<data>
		if strings.Contains(parts[0], ";base64") {
			data, err = base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				return "", err
			}

		} else {
			// format: data:[<mediatype>],<data>
			decoded, err := url.QueryUnescape(parts[1])
			if err != nil {
				return "", err
			}
			data = []byte(decoded)
		}
	case "file", "":
		data, err = os.ReadFile(u.Path)
		if err != nil {
			return "", err
		}
	default:

		return "", errors.New("unsupported scheme: " + u.Scheme)
	}

	fileExt := ""
	if exts, err := mime.ExtensionsByType(http.DetectContentType(data)); err == nil && len(exts) > 0 {
		fileExt = exts[0]
	}
	tmpFile, err := os.CreateTemp("", "uri-*"+fileExt)
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	_, err = tmpFile.Write(data)
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}
	return tmpFile.Name(), nil
}
