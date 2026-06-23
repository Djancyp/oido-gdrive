package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// driveService builds a Drive client from the injected access token.
func driveService(ctx context.Context) (*drive.Service, error) {
	tok := os.Getenv("GOOGLE_ACCESS_TOKEN")
	if tok == "" {
		return nil, fmt.Errorf("not connected: GOOGLE_ACCESS_TOKEN is empty — open the extension settings and click Connect with Google")
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: tok})
	return drive.NewService(ctx, option.WithTokenSource(ts))
}

// ListFiles lists files, optionally filtered by a Drive query string. An empty
// query returns the most recently modified files. (ponytail: search is just a
// query, so list covers both — no separate search tool.)
func ListFiles(ctx context.Context, query string, count int) ([]*drive.File, error) {
	svc, err := driveService(ctx)
	if err != nil {
		return nil, err
	}
	if count <= 0 {
		count = 20
	}
	call := svc.Files.List().
		PageSize(int64(count)).
		OrderBy("modifiedTime desc").
		Fields("files(id,name,mimeType,modifiedTime,size)")
	if strings.TrimSpace(query) != "" {
		call = call.Q(query)
	}
	res, err := call.Do()
	if err != nil {
		return nil, err
	}
	return res.Files, nil
}

// ReadFile returns the text content of a file. Google-native docs are exported
// as text/plain; everything else is downloaded as-is.
func ReadFile(ctx context.Context, fileID string) (string, error) {
	svc, err := driveService(ctx)
	if err != nil {
		return "", err
	}
	meta, err := svc.Files.Get(fileID).Fields("mimeType,name").Do()
	if err != nil {
		return "", err
	}
	var resp io.ReadCloser
	if strings.HasPrefix(meta.MimeType, "application/vnd.google-apps") {
		r, err := svc.Files.Export(fileID, "text/plain").Download()
		if err != nil {
			return "", err
		}
		resp = r.Body
	} else {
		r, err := svc.Files.Get(fileID).Download()
		if err != nil {
			return "", err
		}
		resp = r.Body
	}
	defer resp.Close()
	b, err := io.ReadAll(resp)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// CreateFile creates a plain-text file, optionally inside a parent folder.
func CreateFile(ctx context.Context, name, content, folderID string) (*drive.File, error) {
	svc, err := driveService(ctx)
	if err != nil {
		return nil, err
	}
	f := &drive.File{Name: name}
	if folderID != "" {
		f.Parents = []string{folderID}
	}
	return svc.Files.Create(f).Media(strings.NewReader(content)).Fields("id,name,mimeType").Do()
}

// UpdateFile overwrites a file's content with new text.
func UpdateFile(ctx context.Context, fileID, content string) (*drive.File, error) {
	svc, err := driveService(ctx)
	if err != nil {
		return nil, err
	}
	return svc.Files.Update(fileID, &drive.File{}).Media(strings.NewReader(content)).Fields("id,name").Do()
}

// DeleteFile permanently deletes a file.
func DeleteFile(ctx context.Context, fileID string) error {
	svc, err := driveService(ctx)
	if err != nil {
		return err
	}
	return svc.Files.Delete(fileID).Do()
}
