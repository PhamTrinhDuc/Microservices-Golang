package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	token      = "mock_token"
	owner      = "PhamTrinhDuc"
	repo       = "Learn-Go"
	baseBranch = "main"
	newBranch  = "feat/mcp-product-brand-category-tools"
	repoPath   = "e:/workspace/AI/Learn-Go"
)

type RefResponse struct {
	Object struct {
		SHA string `json:"sha"`
	} `json:"object"`
}

type BlobRequest struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

type BlobResponse struct {
	SHA string `json:"sha"`
}

type TreeEntry struct {
	Path string  `json:"path"`
	Mode string  `json:"mode"`
	Type string  `json:"type"`
	SHA  *string `json:"sha"`
}

type TreeRequest struct {
	BaseTree string      `json:"base_tree"`
	Tree     []TreeEntry `json:"tree"`
}

type TreeResponse struct {
	SHA string `json:"sha"`
}

type CommitRequest struct {
	Message string   `json:"message"`
	Tree    string   `json:"tree"`
	Parents []string `json:"parents"`
}

type CommitResponse struct {
	SHA string `json:"sha"`
}

type PRRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Head  string `json:"head"`
	Base  string `json:"base"`
}

type PRResponse struct {
	HTMLURL string `json:"html_url"`
	Number  int    `json:"number"`
}

func main() {
	filesToAdd := []string{
		"backend/domain/catalog.go",
		"mcp-server/internal/server/mcp_handler.go",
		"mcp-server/internal/tools/api_branch.go",
		"mcp-server/internal/tools/api_category.go",
		"mcp-server/internal/tools/api_products.go",
	}

	filesToDelete := []string{}

	fmt.Println("Starting PR creation via GitHub API...")

	baseSHA, err := getRefSHA(baseBranch)
	if err != nil {
		fmt.Printf("Error getting base branch SHA: %v\n", err)
		return
	}
	fmt.Printf("Base branch (%s) SHA: %s\n", baseBranch, baseSHA)

	parentSHA := baseSHA
	_, err = getRefSHA(newBranch)
	if err != nil {
		err = createRef(newBranch, baseSHA)
		if err != nil {
			fmt.Printf("Error creating ref %s: %v\n", newBranch, err)
			return
		}
		fmt.Printf("Created branch ref: %s\n", newBranch)
	} else {
		fmt.Printf("Existing branch %s found, will force-update it after committing on top of base SHA\n", newBranch)
	}

	var treeEntries []TreeEntry

	for _, fileRelPath := range filesToAdd {
		fullPath := filepath.Join(repoPath, fileRelPath)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", fileRelPath, err)
			return
		}

		encoded := base64.StdEncoding.EncodeToString(content)
		blobSHA, err := createBlob(encoded)
		if err != nil {
			fmt.Printf("Error creating blob for %s: %v\n", fileRelPath, err)
			return
		}
		fmt.Printf("Created blob for %s -> SHA: %s\n", fileRelPath, blobSHA)

		shaCopy := blobSHA
		githubPath := filepath.ToSlash(fileRelPath)
		treeEntries = append(treeEntries, TreeEntry{
			Path: githubPath,
			Mode: "100644",
			Type: "blob",
			SHA:  &shaCopy,
		})
	}

	for _, fileRelPath := range filesToDelete {
		githubPath := filepath.ToSlash(fileRelPath)
		treeEntries = append(treeEntries, TreeEntry{
			Path: githubPath,
			Mode: "100644",
			Type: "blob",
			SHA:  nil,
		})
		fmt.Printf("Queued deletion for %s\n", fileRelPath)
	}

	treeSHA, err := createTree(parentSHA, treeEntries)
	if err != nil {
		fmt.Printf("Error creating tree: %v\n", err)
		return
	}
	fmt.Printf("Created new tree: %s\n", treeSHA)

	commitSHA, err := createCommit("feat(mcp): align product, brand, and category tools with backend catalog APIs\n\n- Replace outdated branch and stylist tools with brand and category tools\n- Add advanced filters to list_products (price, stock, spec, rating)\n- Register CategoryTool, BrandTool, and ProductTool in SSE handler\n- Add form tags to ProductSearchQuery in backend domain to support query binding", treeSHA, parentSHA)
	if err != nil {
		fmt.Printf("Error creating commit: %v\n", err)
		return
	}
	fmt.Printf("Created commit: %s\n", commitSHA)

	err = updateRef(newBranch, commitSHA)
	if err != nil {
		fmt.Printf("Error updating branch ref: %v\n", err)
		return
	}
	fmt.Printf("Updated branch ref %s to commit %s\n", newBranch, commitSHA)

	prURL, prNum, err := createPullRequest("feat(mcp): align product, brand, and category tools with backend catalog APIs", "This PR aligns the MCP server tools with the recently updated backend product catalog APIs: it replaces the outdated branch/stylist tools with brand/category tools, adds advanced search parameters to list_products, and registers all three catalog tools to the MCP server handler.", newBranch, baseBranch)
	if err != nil {
		fmt.Printf("Pull request creation skipped/already exists: %v\n", err)
	} else {
		fmt.Printf("\nSUCCESS! Pull Request #%d created/updated successfully!\nLink: %s\n", prNum, prURL)
	}
}

func doRequest(method, url string, body interface{}, responseObj interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %s returned status %d: %s", url, resp.StatusCode, string(respBody))
	}

	if responseObj != nil {
		return json.NewDecoder(resp.Body).Decode(responseObj)
	}
	return nil
}

func getRefSHA(branch string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/ref/heads/%s", owner, repo, branch)
	var ref RefResponse
	err := doRequest("GET", url, nil, &ref)
	if err != nil {
		return "", err
	}
	return ref.Object.SHA, nil
}

func createRef(branch, sha string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs", owner, repo)
	body := map[string]string{
		"ref": "refs/heads/" + branch,
		"sha": sha,
	}
	return doRequest("POST", url, body, nil)
}

func createBlob(contentBase64 string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/blobs", owner, repo)
	body := BlobRequest{
		Content:  contentBase64,
		Encoding: "base64",
	}
	var resp BlobResponse
	err := doRequest("POST", url, body, &resp)
	if err != nil {
		return "", err
	}
	return resp.SHA, nil
}

func createTree(baseTreeSHA string, entries []TreeEntry) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees", owner, repo)
	body := TreeRequest{
		BaseTree: baseTreeSHA,
		Tree:     entries,
	}
	var resp TreeResponse
	err := doRequest("POST", url, body, &resp)
	if err != nil {
		return "", err
	}
	return resp.SHA, nil
}

func createCommit(message, treeSHA, parentSHA string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/commits", owner, repo)
	body := CommitRequest{
		Message: message,
		Tree:    treeSHA,
		Parents: []string{parentSHA},
	}
	var resp CommitResponse
	err := doRequest("POST", url, body, &resp)
	if err != nil {
		return "", err
	}
	return resp.SHA, nil
}

func updateRef(branch, sha string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs/heads/%s", owner, repo, branch)
	body := map[string]interface{}{
		"sha":   sha,
		"force": true,
	}
	return doRequest("PATCH", url, body, nil)
}

func createPullRequest(title, body, head, base string) (string, int, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", owner, repo)
	reqBody := PRRequest{
		Title: title,
		Body:  body,
		Head:  head,
		Base:  base,
	}
	var resp PRResponse
	err := doRequest("POST", url, reqBody, &resp)
	if err != nil {
		return "", 0, err
	}
	return resp.HTMLURL, resp.Number, nil
}
