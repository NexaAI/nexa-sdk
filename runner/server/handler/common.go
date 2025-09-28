package handler

import (
	"encoding/base64"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func SaveURIToTempFile(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	var data []byte
	var fileExt string

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
		fileExt = getFileExtensionFromURL(u.Path)
	case "data":
		// format: data:[<mediatype>][;base64],<data>
		parts := strings.SplitN(u.Opaque, ",", 2)
		if len(parts) != 2 {
			return "", errors.New("invalid data URI format")
		}

		fileExt = getFileExtensionFromMediaType(parts[0])

		if strings.Contains(parts[0], ";base64") {
			data, err = base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				return "", err
			}
		} else {
			decoded, err := url.QueryUnescape(parts[1])
			if err != nil {
				return "", err
			}
			data = []byte(decoded)
		}
	default:
		data, err = os.ReadFile(u.Path) // try local file
		if err != nil {
			return "", err
		}
		fileExt = getFileExtensionFromURL(u.Path)
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

func getFileExtensionFromURL(path string) string {
	ext := filepath.Ext(path)
	if ext == "" {
		return ""
	}
	return ext
}

func getFileExtensionFromMediaType(mediaType string) string {
	mediaType = strings.Split(mediaType, ";")[0]
	mediaType = strings.TrimSpace(mediaType)
	if exts, err := mime.ExtensionsByType(mediaType); err == nil && len(exts) > 0 {
		return exts[0]
	}
	return "." + strings.SplitN(mediaType, "/", 2)[1]
}
