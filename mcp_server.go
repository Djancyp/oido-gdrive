package main

import (
	"context"
	"fmt"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type handler struct{}

type ListArgs struct {
	Query string `json:"query" jsonschema:"Optional Drive query, e.g. name contains 'report' or mimeType='application/pdf'. Empty returns recent files."`
	Count int    `json:"count" jsonschema:"Max results (default 20)"`
}

type ReadArgs struct {
	FileID string `json:"file_id" jsonschema:"ID of the file to read"`
}

type CreateArgs struct {
	Name     string `json:"name" jsonschema:"Name of the new file"`
	Content  string `json:"content" jsonschema:"Text content of the file"`
	FolderID string `json:"folder_id" jsonschema:"Optional parent folder ID"`
}

type UpdateArgs struct {
	FileID  string `json:"file_id" jsonschema:"ID of the file to overwrite"`
	Content string `json:"content" jsonschema:"New text content"`
}

type DeleteArgs struct {
	FileID string `json:"file_id" jsonschema:"ID of the file to delete permanently"`
}

type CreateFolderArgs struct {
	Name     string `json:"name" jsonschema:"Name of the new folder"`
	ParentID string `json:"parent_id" jsonschema:"Optional parent folder ID"`
}

type DeleteFolderArgs struct {
	FolderID string `json:"folder_id" jsonschema:"ID of the folder to delete permanently (with its contents)"`
}

func textResult(s string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: s}}}
}

func errResult(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "Error: " + err.Error()}}, IsError: true}
}

func (h *handler) List(ctx context.Context, _ *mcp.CallToolRequest, a ListArgs) (*mcp.CallToolResult, any, error) {
	files, err := ListFiles(ctx, a.Query, a.Count)
	if err != nil {
		return errResult(err), nil, nil
	}
	if len(files) == 0 {
		return textResult("No files found."), nil, nil
	}
	out := fmt.Sprintf("Files (%d):\n\nID | Name | Type | Modified\n", len(files))
	for _, f := range files {
		out += fmt.Sprintf("%s | %s | %s | %s\n", f.Id, f.Name, f.MimeType, f.ModifiedTime)
	}
	return textResult(out), nil, nil
}

func (h *handler) Read(ctx context.Context, _ *mcp.CallToolRequest, a ReadArgs) (*mcp.CallToolResult, any, error) {
	content, err := ReadFile(ctx, a.FileID)
	if err != nil {
		return errResult(err), nil, nil
	}
	return textResult(content), nil, nil
}

func (h *handler) Create(ctx context.Context, _ *mcp.CallToolRequest, a CreateArgs) (*mcp.CallToolResult, any, error) {
	f, err := CreateFile(ctx, a.Name, a.Content, a.FolderID)
	if err != nil {
		return errResult(err), nil, nil
	}
	return textResult(fmt.Sprintf("Created file %q (id: %s)", f.Name, f.Id)), nil, nil
}

func (h *handler) Update(ctx context.Context, _ *mcp.CallToolRequest, a UpdateArgs) (*mcp.CallToolResult, any, error) {
	f, err := UpdateFile(ctx, a.FileID, a.Content)
	if err != nil {
		return errResult(err), nil, nil
	}
	return textResult(fmt.Sprintf("Updated file %q (id: %s)", f.Name, f.Id)), nil, nil
}

func (h *handler) Delete(ctx context.Context, _ *mcp.CallToolRequest, a DeleteArgs) (*mcp.CallToolResult, any, error) {
	if err := DeleteFile(ctx, a.FileID); err != nil {
		return errResult(err), nil, nil
	}
	return textResult("Deleted file " + a.FileID), nil, nil
}

func (h *handler) CreateFolder(ctx context.Context, _ *mcp.CallToolRequest, a CreateFolderArgs) (*mcp.CallToolResult, any, error) {
	f, err := CreateFolder(ctx, a.Name, a.ParentID)
	if err != nil {
		return errResult(err), nil, nil
	}
	return textResult(fmt.Sprintf("Created folder %q (id: %s)", f.Name, f.Id)), nil, nil
}

func (h *handler) DeleteFolder(ctx context.Context, _ *mcp.CallToolRequest, a DeleteFolderArgs) (*mcp.CallToolResult, any, error) {
	if err := DeleteFile(ctx, a.FolderID); err != nil {
		return errResult(err), nil, nil
	}
	return textResult("Deleted folder " + a.FolderID), nil, nil
}

// RunMCPServer registers tools and serves over stdio.
func RunMCPServer() {
	h := &handler{}
	server := mcp.NewServer(&mcp.Implementation{Name: "oido-gdrive", Version: "1.0.0"}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "drive_list",
		Description: "List or search Google Drive files. Optional query filters; empty returns recent files.",
	}, h.List)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "drive_read_file",
		Description: "Read a Drive file's text content by ID. Google Docs/Sheets/Slides are exported to plain text.",
	}, h.Read)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "drive_create_file",
		Description: "Create a new text file in Drive, optionally inside a folder.",
	}, h.Create)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "drive_update_file",
		Description: "Overwrite a Drive file's content by ID.",
	}, h.Update)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "drive_delete_file",
		Description: "Permanently delete a Drive file by ID.",
	}, h.Delete)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "drive_create_folder",
		Description: "Create a folder in Drive, optionally inside a parent folder.",
	}, h.CreateFolder)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "drive_delete_folder",
		Description: "Permanently delete a Drive folder (and its contents) by ID.",
	}, h.DeleteFolder)

	log.Println("Oido Google Drive MCP Server starting on stdio...")
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
