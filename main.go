package main

import "log"

// Oido Google Drive MCP server. The OAuth access token is injected by oido-core
// as GOOGLE_ACCESS_TOKEN (refreshed at launch); this process just consumes it.
func main() {
	log.Println("Starting Oido Google Drive MCP Server v1.0.0...")
	RunMCPServer()
}
